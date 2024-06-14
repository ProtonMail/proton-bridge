// Copyright (c) 2024 Proton AG
//
// This file is part of Proton Mail Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package user

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/configstatus"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/safe"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/imapservice"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/orderedtasks"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/sendrecorder"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/smtp"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/syncservice"
	telemetryservice "github.com/ProtonMail/proton-bridge/v3/internal/services/telemetry"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/userevents"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/useridentity"
	"github.com/ProtonMail/proton-bridge/v3/internal/telemetry"
	"github.com/ProtonMail/proton-bridge/v3/internal/usertypes"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/ProtonMail/proton-bridge/v3/pkg/algo"
	"github.com/bradenaw/juniper/xslices"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

var (
	EventPeriod = 20 * time.Second // nolint:gochecknoglobals,revive
	EventJitter = 20 * time.Second // nolint:gochecknoglobals,revive
)

const (
	SyncRetryCooldown = 20 * time.Second
)

type User struct {
	id  string
	log *logrus.Entry

	vault    *vault.User
	client   *proton.Client
	reporter reporter.Reporter
	sendHash *sendrecorder.SendRecorder

	eventCh   *async.QueuedChannel[events.Event]
	eventLock safe.RWMutex

	tasks *async.Group

	maxSyncMemory uint64

	panicHandler     async.PanicHandler
	configStatus     *configstatus.ConfigurationStatus
	telemetryManager telemetry.Availability
	// goStatusProgress triggers a check/sending if progress is needed.
	goStatusProgress func()

	eventService     *userevents.Service
	identityService  *useridentity.Service
	smtpService      *smtp.Service
	imapService      *imapservice.Service
	telemetryService *telemetryservice.Service

	serviceGroup *orderedtasks.OrderedCancelGroup
}

func New(
	ctx context.Context,
	encVault *vault.User,
	client *proton.Client,
	reporter reporter.Reporter,
	apiUser proton.User,
	crashHandler async.PanicHandler,
	showAllMail bool,
	maxSyncMemory uint64,
	statsDir string,
	telemetryManager telemetry.Availability,
	imapServerManager imapservice.IMAPServerManager,
	smtpServerManager smtp.ServerManager,
	eventSubscription events.Subscription,
	syncService syncservice.Regulator,
	syncConfigDir string,
	isNew bool,
) (*User, error) {
	user, err := newImpl(
		ctx,
		encVault,
		client,
		reporter,
		apiUser,
		crashHandler,
		showAllMail,
		maxSyncMemory,
		statsDir,
		telemetryManager,
		imapServerManager,
		smtpServerManager,
		eventSubscription,
		syncService,
		syncConfigDir,
		isNew,
	)
	if err != nil {
		// Cleanup any pending resources on error
		if user != nil {
			user.Close()
		}

		return nil, err
	}

	return user, nil
}

// New returns a new user.
func newImpl(
	ctx context.Context,
	encVault *vault.User,
	client *proton.Client,
	reporter reporter.Reporter,
	apiUser proton.User,
	crashHandler async.PanicHandler,
	showAllMail bool,
	maxSyncMemory uint64,
	statsDir string,
	telemetryManager telemetry.Availability,
	imapServerManager imapservice.IMAPServerManager,
	smtpServerManager smtp.ServerManager,
	eventSubscription events.Subscription,
	syncService syncservice.Regulator,
	syncConfigDir string,
	isNew bool,
) (*User, error) {
	logrus.WithField("userID", apiUser.ID).Info("Creating new user")

	// Migrate Sync Status from Vault.
	if err := migrateSyncStatusFromVault(encVault, syncConfigDir, apiUser.ID); err != nil {
		return nil, err
	}

	// Get the user's API addresses.
	apiAddrs, err := client.GetAddresses(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get addresses: %w", err)
	}

	// Get the user's API labels.
	apiLabels, err := client.GetLabels(ctx, proton.LabelTypeSystem, proton.LabelTypeFolder, proton.LabelTypeLabel)
	if err != nil {
		return nil, fmt.Errorf("failed to get labels: %w", err)
	}

	identityState := useridentity.NewState(apiUser, apiAddrs, client)

	logrus.WithFields(logrus.Fields{
		"userID":    apiUser.ID,
		"numAddr":   len(apiAddrs),
		"numLabels": len(apiLabels),
	}).Info("Creating user object")

	configStatusFile := filepath.Join(statsDir, apiUser.ID+".json")
	configStatus, err := configstatus.LoadConfigurationStatus(configStatusFile)
	if err != nil {
		return nil, fmt.Errorf("failed to init configuration status file: %w", err)
	}

	sendRecorder := sendrecorder.NewSendRecorder(sendrecorder.SendEntryExpiry)

	// Create the user object.
	user := &User{
		log: logrus.WithField("userID", apiUser.ID),
		id:  apiUser.ID,

		vault:    encVault,
		client:   client,
		reporter: reporter,
		sendHash: sendRecorder,

		eventCh:   async.NewQueuedChannel[events.Event](0, 0, crashHandler, fmt.Sprintf("bridge-user-%v", apiUser.ID)),
		eventLock: safe.NewRWMutex(),

		tasks: async.NewGroup(context.Background(), crashHandler),

		maxSyncMemory: maxSyncMemory,

		panicHandler: crashHandler,

		configStatus:     configStatus,
		telemetryManager: telemetryManager,

		serviceGroup: orderedtasks.NewOrderedCancelGroup(crashHandler),
		smtpService:  nil,
	}

	user.eventService = userevents.NewService(
		apiUser.ID,
		client,
		userevents.NewVaultEventIDStore(encVault),
		user,
		EventPeriod,
		EventJitter,
		5*time.Minute,
		crashHandler,
		eventSubscription,
	)

	addressMode := usertypes.VaultToAddressMode(encVault.AddressMode())

	user.identityService = useridentity.NewService(user.eventService, user, identityState, encVault, user)

	user.telemetryService = telemetryservice.NewService(apiUser.ID, client, user.eventService)

	user.smtpService = smtp.NewService(
		apiUser.ID,
		client,
		sendRecorder,
		crashHandler,
		reporter,
		encVault,
		encVault,
		user,
		user.eventService,
		addressMode,
		identityState.Clone(),
		smtpServerManager,
	)

	user.imapService = imapservice.NewService(
		client,
		identityState.Clone(),
		user,
		user.eventService,
		imapServerManager,
		user,
		encVault,
		encVault,
		crashHandler,
		sendRecorder,
		user,
		reporter,
		addressMode,
		eventSubscription,
		syncConfigDir,
		user.maxSyncMemory,
		showAllMail,
	)

	// Check for status_progress when triggered.
	user.goStatusProgress = user.tasks.PeriodicOrTrigger(configstatus.ProgressCheckInterval, 0, func(ctx context.Context) {
		user.SendConfigStatusProgress(ctx)
	})
	defer user.goStatusProgress()

	// When we receive an auth object, we update it in the vault.
	// This will be used to authorize the user on the next run.
	user.client.AddAuthHandler(func(auth proton.Auth) {
		if err := user.vault.SetAuth(auth.UID, auth.RefreshToken); err != nil {
			user.log.WithError(err).Error("Failed to update auth in vault")
		}
	})

	// When we are deauthorized, we send a deauth event to the event channel.
	// Bridge will react to this event by logging out the user.
	user.client.AddDeauthHandler(func() {
		user.eventCh.Enqueue(events.UserDeauth{
			UserID: user.ID(),
		})
	})

	// Log all requests made by the user.
	user.client.AddPostRequestHook(func(_ *resty.Client, r *resty.Response) error {
		user.log.WithField("pkg", "gpa/client").Infof("%v: %v %v", r.Status(), r.Request.Method, r.Request.URL)
		return nil
	})

	// If it's not a fresh user check the eventID and evaluate whether it is valid. If it's a new user, we don't
	// need to perform this check.
	if !isNew {
		if err := checkIrrecoverableEventID(ctx, encVault.EventID(), apiUser.ID, syncConfigDir, user); err != nil {
			return nil, err
		}
	}

	// Start Event Service
	lastEventID, err := user.eventService.Start(ctx, user.serviceGroup)
	if err != nil {
		return user, fmt.Errorf("failed to start event service: %w", err)
	}

	// Start Telemetry Service
	user.telemetryService.Start(ctx, user.serviceGroup)

	// Start Identity Service
	user.identityService.Start(ctx, user.serviceGroup)

	// Start SMTP Service
	if err := user.smtpService.Start(ctx, user.serviceGroup); err != nil {
		return user, fmt.Errorf("failed to start smtp service: %w", err)
	}

	// Start IMAP Service
	if err := user.imapService.Start(ctx, user.serviceGroup, syncService, lastEventID); err != nil {
		return user, fmt.Errorf("failed to start imap service: %w", err)
	}

	user.eventService.Resume()

	return user, nil
}

// ID returns the user's ID.
func (user *User) ID() string {
	return user.id
}

// Name returns the user's username.
func (user *User) Name() string {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute))
	defer cancel()

	apiUser, err := user.identityService.GetAPIUser(ctx)
	if err != nil {
		return ""
	}

	return apiUser.Name
}

// Match matches the given query against the user's username and email addresses.
func (user *User) Match(query string) bool {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute))
	defer cancel()

	apiUser, err := user.identityService.GetAPIUser(ctx)
	if err != nil {
		return false
	}

	if query == apiUser.Name {
		return true
	}

	apiAddrs, err := user.identityService.GetAddresses(ctx)
	if err != nil {
		return false
	}

	for _, addr := range apiAddrs {
		if query == addr.Email {
			return true
		}
	}

	return false
}

// DisplayNames returns a map of the email addresses and their associated display names.
func (user *User) DisplayNames() map[string]string {
	addresses := user.protonAddresses()
	if addresses == nil {
		return nil
	}

	result := make(map[string]string)
	for _, address := range addresses {
		result[address.Email] = address.DisplayName
	}

	return result
}

// Emails returns all the user's active email addresses.
// It returns them in sorted order; the user's primary address is first.
func (user *User) Emails() []string {
	addresses := user.protonAddresses()
	if addresses == nil {
		return nil
	}

	return xslices.Map(addresses, func(addr proton.Address) string {
		return addr.Email
	})
}

// GetAddressMode returns the user's current address mode.
func (user *User) GetAddressMode() vault.AddressMode {
	return user.vault.AddressMode()
}

// SetAddressMode sets the user's address mode.
func (user *User) SetAddressMode(ctx context.Context, mode vault.AddressMode) error {
	user.log.WithField("mode", mode).Info("Setting address mode")

	if err := user.vault.SetAddressMode(mode); err != nil {
		return fmt.Errorf("failed to set address mode: %w", err)
	}

	if err := user.smtpService.SetAddressMode(ctx, usertypes.VaultToAddressMode(mode)); err != nil {
		return fmt.Errorf("failed to set smtp address mode: %w", err)
	}

	if err := user.imapService.SetAddressMode(ctx, usertypes.VaultToAddressMode(mode)); err != nil {
		return fmt.Errorf("failed to imap address mode: %w", err)
	}

	return nil
}

// BadEventFeedbackResync sends user feedback whether should do message re-sync.
func (user *User) BadEventFeedbackResync(ctx context.Context) error {
	if err := user.imapService.OnBadEventResync(ctx); err != nil {
		return fmt.Errorf("failed to execute bad event request on imap service: %w", err)
	}

	if err := user.identityService.Resync(ctx); err != nil {
		return fmt.Errorf("failed to resync identity service: %w", err)
	}

	if err := user.smtpService.Resync(ctx); err != nil {
		return fmt.Errorf("failed to resync smtp service: %w", err)
	}

	if err := user.imapService.Resync(ctx); err != nil {
		return fmt.Errorf("failed to resync imap service: %w", err)
	}

	user.eventService.Resume()

	return nil
}

func (user *User) OnBadEvent(ctx context.Context) {
	if err := user.imapService.OnBadEvent(ctx); err != nil {
		user.log.WithError(err).Error("Failed to notify imap service of bad event")
	}
}

// SetShowAllMail sets whether to show the All Mail mailbox.
func (user *User) SetShowAllMail(show bool) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute))
	defer cancel()

	user.log.WithField("show", show).Info("Setting show all mail")

	if err := user.imapService.ShowAllMail(ctx, show); err != nil {
		user.log.WithError(err).Error("Failed to set show all mail")
	}
}

// GetGluonIDs returns the users gluon IDs.
func (user *User) GetGluonIDs() map[string]string {
	return user.vault.GetGluonIDs()
}

// GetGluonID returns the gluon ID for the given address, if present.
func (user *User) GetGluonID(addrID string) (string, bool) {
	if gluonID, ok := user.vault.GetGluonIDs()[addrID]; ok {
		return gluonID, true
	}

	if user.vault.AddressMode() != vault.CombinedMode {
		return "", false
	}

	// If there is only one address, return its gluon ID.
	// This can happen if we are in combined mode and the primary address ID has changed.
	if gluonIDs := maps.Values(user.vault.GetGluonIDs()); len(gluonIDs) == 1 {
		if err := user.vault.SetGluonID(addrID, gluonIDs[0]); err != nil {
			user.log.WithError(err).Error("Failed to set gluon ID for updated primary address")
		}

		return gluonIDs[0], true
	}

	return "", false
}

// SetGluonID sets the gluon ID for the given address.
func (user *User) SetGluonID(addrID, gluonID string) error {
	user.log.WithFields(logrus.Fields{
		"addrID":  addrID,
		"gluonID": gluonID,
	}).Info("Setting gluon ID")

	return user.vault.SetGluonID(addrID, gluonID)
}

// RemoveGluonID removes the gluon ID for the given address.
func (user *User) RemoveGluonID(addrID, gluonID string) error {
	user.log.WithFields(logrus.Fields{
		"addrID":  addrID,
		"gluonID": gluonID,
	}).Info("Removing gluon ID")

	return user.vault.RemoveGluonID(addrID, gluonID)
}

// GluonKey returns the user's gluon key from the vault.
func (user *User) GluonKey() []byte {
	return user.vault.GluonKey()
}

// BridgePass returns the user's bridge password, used for authentication over SMTP and IMAP.
func (user *User) BridgePass() []byte {
	return algo.B64RawEncode(user.vault.BridgePass())
}

// UsedSpace returns the total space used by the user on the API.
func (user *User) UsedSpace() uint64 {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute))
	defer cancel()

	apiUser, err := user.identityService.GetAPIUser(ctx)
	if err != nil {
		return 0
	}

	return apiUser.UsedSpace
}

// MaxSpace returns the amount of space the user can use on the API.
func (user *User) MaxSpace() uint64 {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute))
	defer cancel()

	apiUser, err := user.identityService.GetAPIUser(ctx)
	if err != nil {
		return 0
	}

	return apiUser.MaxSpace
}

// GetEventCh returns a channel which notifies of events happening to the user (such as deauth, address change).
func (user *User) GetEventCh() <-chan events.Event {
	return user.eventCh.GetChannel()
}

// CheckAuth returns whether the given email and password can be used to authenticate over IMAP or SMTP with this user.
// It returns the address ID of the authenticated address.
func (user *User) CheckAuth(email string, password []byte) (string, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute))
	defer cancel()

	return user.identityService.CheckAuth(ctx, email, password)
}

// Logout logs the user out from the API.
func (user *User) Logout(ctx context.Context, withAPI bool) error {
	user.log.WithField("withAPI", withAPI).Info("Logging out user")

	user.log.Debug("Canceling ongoing tasks")

	if err := user.smtpService.OnLogout(ctx); err != nil {
		return fmt.Errorf("failed to remove user from smtp server: %w", err)
	}

	if err := user.imapService.OnLogout(ctx); err != nil {
		return fmt.Errorf("failed to remove user from imap server: %w", err)
	}

	user.tasks.CancelAndWait()

	// Stop Services
	user.serviceGroup.CancelAndWait()

	// Cleanup Event Service.
	user.eventService.Close()

	// Close imap service.
	user.imapService.Close()

	if withAPI {
		user.log.Debug("Logging out from API")

		if err := user.client.AuthDelete(ctx); err != nil {
			user.log.WithError(err).Warn("Failed to delete auth")
		}
	}

	user.log.Debug("Clearing vault secrets")

	if err := user.vault.Clear(); err != nil {
		return fmt.Errorf("failed to clear vault: %w", err)
	}

	return nil
}

// Close closes ongoing connections and cleans up resources.
func (user *User) Close() {
	user.log.Info("Closing user")

	// Stop any ongoing background tasks.
	user.tasks.CancelAndWait()

	// Stop Services
	user.serviceGroup.CancelAndWait()

	// Cleanup Event Service.
	user.eventService.Close()

	// Close imap service.
	user.imapService.Close()

	// Close the user's API client.
	user.client.Close()

	// Close the user's notify channel.
	user.eventCh.CloseAndDiscardQueued()

	// Close the user's vault.
	if err := user.vault.Close(); err != nil {
		user.log.WithError(err).Error("Failed to close vault")
	}
}

// IsTelemetryEnabled check if the telemetry is enabled or disabled for this user.
func (user *User) IsTelemetryEnabled(ctx context.Context) bool {
	return user.telemetryService.IsTelemetryEnabled(ctx)
}

// SendTelemetry send telemetry request.
func (user *User) SendTelemetry(ctx context.Context, data []byte) error {
	var req proton.SendStatsReq
	if err := json.Unmarshal(data, &req); err != nil {
		user.log.WithError(err).Error("Failed to build telemetry request.")
		if err := user.reporter.ReportMessageWithContext("Failed to build telemetry request.", reporter.Context{
			"error": err,
		}); err != nil {
			logrus.WithError(err).Error("Failed to report telemetry request build error")
		}
		return err
	}
	err := user.client.SendDataEvent(ctx, req)
	if err != nil {
		user.log.WithError(err).Error("Failed to send telemetry.")
		return err
	}
	return nil
}

func (user *User) ReportSMTPAuthFailed(username string) {
	emails := user.Emails()
	for _, mail := range emails {
		if mail == username {
			user.ReportConfigStatusFailure("SMTP invalid username or password")
		}
	}
}

func (user *User) ReportSMTPAuthSuccess(ctx context.Context) {
	user.SendConfigStatusSuccess(ctx)
}

func (user *User) GetSMTPService() *smtp.Service {
	return user.smtpService
}

func (user *User) PublishEvent(_ context.Context, event events.Event) {
	user.eventCh.Enqueue(event)
}

func (user *User) PauseEventLoop() {
	user.eventService.Pause()
}

func (user *User) PauseEventLoopWithWaiter() *userevents.EventPollWaiter {
	return user.eventService.PauseWithWaiter()
}

func (user *User) ResumeEventLoop() {
	user.eventService.Resume()
}

func (user *User) protonAddresses() []proton.Address {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute))
	defer cancel()

	apiAddrs, err := user.identityService.GetAddresses(ctx)
	if err != nil {
		return nil
	}

	addresses := xslices.Filter(maps.Values(apiAddrs), func(addr proton.Address) bool {
		return addr.Status == proton.AddressStatusEnabled && addr.Type != proton.AddressTypeExternal
	})

	slices.SortFunc(addresses, func(a, b proton.Address) bool {
		return a.Order < b.Order
	})

	return addresses
}

func (user *User) VerifyResyncAndExecute() {
	user.log.Info("Checking whether logged in user should re-sync. UserID:", user.ID())
	if user.vault.GetShouldResync() {
		user.log.Info("User should re-sync, starting re-sync process. UserID:", user.ID())

		if err := user.vault.SetShouldSync(false); err != nil {
			user.log.WithError(err).Error("Failed to disable re-sync flag in user vault. UserID:", user.ID())
		}

		user.SendRepairDeferredTrigger(context.Background())
		if err := user.resyncIMAP(); err != nil {
			user.log.WithError(err).Error("Failed re-syncing IMAP for userID", user.ID())
		}
	}
}

func (user *User) TriggerRepair() error {
	user.SendRepairTrigger(context.Background())
	return user.resyncIMAP()
}

func (user *User) resyncIMAP() error {
	return user.imapService.Resync(context.Background())
}
