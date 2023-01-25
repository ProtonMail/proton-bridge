// Copyright (c) 2023 Proton AG
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
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/queue"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/async"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/logging"
	"github.com/ProtonMail/proton-bridge/v3/internal/safe"
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
	log *logrus.Entry

	vault    *vault.User
	client   *proton.Client
	reporter reporter.Reporter
	sendHash *sendRecorder

	eventCh   *queue.QueuedChannel[events.Event]
	eventLock safe.RWMutex

	apiUser     proton.User
	apiUserLock safe.RWMutex

	apiAddrs     map[string]proton.Address
	apiAddrsLock safe.RWMutex

	apiLabels     map[string]proton.Label
	apiLabelsLock safe.RWMutex

	updateCh     map[string]*queue.QueuedChannel[imap.Update]
	updateChLock safe.RWMutex

	tasks     *async.Group
	abortable async.Abortable
	goSync    func()

	pollAPIEventsCh chan chan struct{}
	goPollAPIEvents func(wait bool)

	syncWorkers int
	showAllMail uint32
}

// New returns a new user.
//
// nolint:funlen
func New(
	ctx context.Context,
	encVault *vault.User,
	client *proton.Client,
	reporter reporter.Reporter,
	apiUser proton.User,
	crashHandler async.PanicHandler,
	syncWorkers int,
	showAllMail bool,
) (*User, error) { //nolint:funlen
	logrus.WithField("userID", apiUser.ID).Info("Creating new user")

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

	// Create the user object.
	user := &User{
		log: logrus.WithField("userID", apiUser.ID),

		vault:    encVault,
		client:   client,
		reporter: reporter,
		sendHash: newSendRecorder(sendEntryExpiry),

		eventCh:   queue.NewQueuedChannel[events.Event](0, 0),
		eventLock: safe.NewRWMutex(),

		apiUser:     apiUser,
		apiUserLock: safe.NewRWMutex(),

		apiAddrs:     groupBy(apiAddrs, func(addr proton.Address) string { return addr.ID }),
		apiAddrsLock: safe.NewRWMutex(),

		apiLabels:     groupBy(apiLabels, func(label proton.Label) string { return label.ID }),
		apiLabelsLock: safe.NewRWMutex(),

		updateCh:     make(map[string]*queue.QueuedChannel[imap.Update]),
		updateChLock: safe.NewRWMutex(),

		tasks:           async.NewGroup(context.Background(), crashHandler),
		pollAPIEventsCh: make(chan chan struct{}),

		syncWorkers: syncWorkers,
		showAllMail: b32(showAllMail),
	}

	// Initialize the user's update channels for its current address mode.
	user.initUpdateCh(encVault.AddressMode())

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
		user.log.Infof("%v: %v %v", r.Status(), r.Request.Method, r.Request.URL)
		return nil
	})

	// Stream events from the API, logging any errors that occur.
	// This does nothing until the sync has been marked as complete.
	// When we receive an API event, we attempt to handle it.
	// If successful, we update the event ID in the vault.
	user.tasks.Once(func(ctx context.Context) {
		ticker := proton.NewTicker(EventPeriod, EventJitter)
		defer ticker.Stop()

		for {
			var doneCh chan struct{}

			select {
			case <-ctx.Done():
				return

			case doneCh = <-user.pollAPIEventsCh:
				// ...

			case <-ticker.C:
				// ...
			}

			user.log.Debug("Event poll triggered")

			if !user.vault.SyncStatus().IsComplete() {
				user.log.Debug("Sync is incomplete, skipping event poll")
			} else if err := user.doEventPoll(ctx); err != nil {
				user.log.WithError(err).Error("Failed to poll events")
			}

			if doneCh != nil {
				close(doneCh)
			}
		}
	})

	// When triggered, poll the API for events, optionally blocking until the poll is complete.
	user.goPollAPIEvents = func(wait bool) {
		doneCh := make(chan struct{})

		go func() { user.pollAPIEventsCh <- doneCh }()

		if wait {
			<-doneCh
		}
	}

	// When triggered, attempt to sync the user.
	user.goSync = user.tasks.Trigger(func(ctx context.Context) {
		user.log.Debug("Sync triggered")

		user.abortable.Do(ctx, func(ctx context.Context) {
			if user.vault.SyncStatus().IsComplete() {
				user.log.Debug("Sync is already complete, skipping")
			} else if err := user.doSync(ctx); err != nil {
				user.log.WithError(err).Error("Failed to sync, will retry later")
				time.AfterFunc(SyncRetryCooldown, user.goSync)
			}
		})
	})

	// Trigger an initial sync (if necessary).
	user.goSync()

	return user, nil
}

// ID returns the user's ID.
func (user *User) ID() string {
	return safe.RLockRet(func() string {
		return user.apiUser.ID
	}, user.apiUserLock)
}

// Name returns the user's username.
func (user *User) Name() string {
	return safe.RLockRet(func() string {
		return user.apiUser.Name
	}, user.apiUserLock)
}

// Match matches the given query against the user's username and email addresses.
func (user *User) Match(query string) bool {
	return safe.RLockRet(func() bool {
		if query == user.apiUser.Name {
			return true
		}

		for _, addr := range user.apiAddrs {
			if query == addr.Email {
				return true
			}
		}

		return false
	}, user.apiUserLock, user.apiAddrsLock)
}

// Emails returns all the user's email addresses.
// It returns them in sorted order; the user's primary address is first.
func (user *User) Emails() []string {
	return safe.RLockRet(func() []string {
		addresses := maps.Values(user.apiAddrs)

		slices.SortFunc(addresses, func(a, b proton.Address) bool {
			return a.Order < b.Order
		})

		return xslices.Map(addresses, func(addr proton.Address) string {
			return addr.Email
		})
	}, user.apiAddrsLock)
}

// GetAddressMode returns the user's current address mode.
func (user *User) GetAddressMode() vault.AddressMode {
	return user.vault.AddressMode()
}

// SetAddressMode sets the user's address mode.
func (user *User) SetAddressMode(_ context.Context, mode vault.AddressMode) error {
	user.log.WithField("mode", mode).Info("Setting address mode")

	user.abortable.Abort()
	defer user.goSync()

	return safe.LockRet(func() error {
		user.initUpdateCh(mode)

		if err := user.vault.SetAddressMode(mode); err != nil {
			return fmt.Errorf("failed to set address mode: %w", err)
		}

		if err := user.vault.ClearSyncStatus(); err != nil {
			return fmt.Errorf("failed to clear sync status: %w", err)
		}

		return nil
	}, user.apiAddrsLock, user.updateChLock)
}

// SetShowAllMail sets whether to show the All Mail mailbox.
func (user *User) SetShowAllMail(show bool) {
	user.log.WithField("show", show).Info("Setting show all mail")

	atomic.StoreUint32(&user.showAllMail, b32(show))
}

// GetGluonIDs returns the users gluon IDs.
func (user *User) GetGluonIDs() map[string]string {
	return user.vault.GetGluonIDs()
}

// GetGluonID returns the gluon ID for the given address, if present.
func (user *User) GetGluonID(addrID string) (string, bool) {
	gluonID, ok := user.vault.GetGluonIDs()[addrID]
	if !ok {
		return "", false
	}

	return gluonID, true
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
func (user *User) UsedSpace() int {
	return safe.RLockRet(func() int {
		return user.apiUser.UsedSpace
	}, user.apiUserLock)
}

// MaxSpace returns the amount of space the user can use on the API.
func (user *User) MaxSpace() int {
	return safe.RLockRet(func() int {
		return user.apiUser.MaxSpace
	}, user.apiUserLock)
}

// GetEventCh returns a channel which notifies of events happening to the user (such as deauth, address change).
func (user *User) GetEventCh() <-chan events.Event {
	return user.eventCh.GetChannel()
}

// NewIMAPConnector returns an IMAP connector for the given address.
// If not in split mode, this must be the primary address.
func (user *User) NewIMAPConnector(addrID string) connector.Connector {
	return newIMAPConnector(user, addrID)
}

// NewIMAPConnectors returns IMAP connectors for each of the user's addresses.
// In combined mode, this is just the user's primary address.
// In split mode, this is all the user's addresses.
func (user *User) NewIMAPConnectors() (map[string]connector.Connector, error) {
	return safe.RLockRetErr(func() (map[string]connector.Connector, error) {
		imapConn := make(map[string]connector.Connector)

		switch user.vault.AddressMode() {
		case vault.CombinedMode:
			primAddr, err := getAddrIdx(user.apiAddrs, 0)
			if err != nil {
				return nil, fmt.Errorf("failed to get primary address: %w", err)
			}

			imapConn[primAddr.ID] = newIMAPConnector(user, primAddr.ID)

		case vault.SplitMode:
			for addrID := range user.apiAddrs {
				imapConn[addrID] = newIMAPConnector(user, addrID)
			}
		}

		return imapConn, nil
	}, user.apiAddrsLock)
}

// SendMail sends an email from the given address to the given recipients.
//
// nolint:funlen
func (user *User) SendMail(authID string, from string, to []string, r io.Reader) error {
	defer user.goPollAPIEvents(true)

	if len(to) == 0 {
		return ErrInvalidRecipient
	}

	return user.sendMail(authID, from, to, r)
}

// CheckAuth returns whether the given email and password can be used to authenticate over IMAP or SMTP with this user.
// It returns the address ID of the authenticated address.
func (user *User) CheckAuth(email string, password []byte) (string, error) {
	user.log.WithField("email", logging.Sensitive(email)).Debug("Checking authentication")

	if email == "crash@bandicoot" {
		panic("your wish is my command.. I crash")
	}

	dec, err := algo.B64RawDecode(password)
	if err != nil {
		return "", fmt.Errorf("failed to decode password: %w", err)
	}

	if subtle.ConstantTimeCompare(user.vault.BridgePass(), dec) != 1 {
		return "", fmt.Errorf("invalid password")
	}

	return safe.RLockRetErr(func() (string, error) {
		for _, addr := range user.apiAddrs {
			if addr.Status != proton.AddressStatusEnabled {
				continue
			}

			if strings.EqualFold(addr.Email, email) {
				return addr.ID, nil
			}
		}

		return "", fmt.Errorf("invalid email")
	}, user.apiAddrsLock)
}

// OnStatusUp is called when the connection goes up.
func (user *User) OnStatusUp(context.Context) {
	user.log.Info("Connection is up")

	user.goSync()
}

// OnStatusDown is called when the connection goes down.
func (user *User) OnStatusDown(context.Context) {
	user.log.Info("Connection is down")

	user.abortable.Abort()
}

// ClearSyncStatus clears the sync status of the user. This triggers a resync.
func (user *User) ClearSyncStatus() error {
	user.abortable.Abort()
	defer user.goSync()

	return user.vault.ClearSyncStatus()
}

// Logout logs the user out from the API.
func (user *User) Logout(ctx context.Context, withAPI bool) error {
	user.log.WithField("withAPI", withAPI).Info("Logging out user")

	user.log.Debug("Canceling ongoing tasks")

	user.tasks.CancelAndWait()

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

	// Close the user's API client.
	user.client.Close()

	// Close the user's update channels.
	safe.RLock(func() {
		for _, updateCh := range xslices.Unique(maps.Values(user.updateCh)) {
			updateCh.CloseAndDiscardQueued()
		}
	}, user.updateChLock)

	// Close the user's notify channel.
	user.eventCh.CloseAndDiscardQueued()

	// Close the user's vault.
	if err := user.vault.Close(); err != nil {
		user.log.WithError(err).Error("Failed to close vault")
	}
}

// initUpdateCh initializes the user's update channels in the given address mode.
// It is assumed that user.apiAddrs and user.updateCh are already locked.
func (user *User) initUpdateCh(mode vault.AddressMode) {
	for _, updateCh := range xslices.Unique(maps.Values(user.updateCh)) {
		updateCh.CloseAndDiscardQueued()
	}

	user.updateCh = make(map[string]*queue.QueuedChannel[imap.Update])

	switch mode {
	case vault.CombinedMode:
		primaryUpdateCh := queue.NewQueuedChannel[imap.Update](0, 0)

		for addrID := range user.apiAddrs {
			user.updateCh[addrID] = primaryUpdateCh
		}

	case vault.SplitMode:
		for addrID := range user.apiAddrs {
			user.updateCh[addrID] = queue.NewQueuedChannel[imap.Update](0, 0)
		}
	}
}

// doEventPoll is called whenever API events should be polled.
func (user *User) doEventPoll(ctx context.Context) error {
	user.eventLock.Lock()
	defer user.eventLock.Unlock()

	event, err := user.client.GetEvent(ctx, user.vault.EventID())
	if err != nil {
		return fmt.Errorf("failed to get event: %w", err)
	}

	// If the event ID hasn't changed, there are no new events.
	if event.EventID == user.vault.EventID() {
		user.log.Debug("No new API events")
		return nil
	}

	user.log.WithFields(logrus.Fields{
		"old": user.vault.EventID(),
		"new": event,
	}).Info("Received new API event")

	// Handle the event.
	if err := user.handleAPIEvent(ctx, event); err != nil {
		// If the error is a network error, return error to retry later.
		if netErr := new(proton.NetError); errors.As(err, &netErr) {
			return fmt.Errorf("failed to handle event due to network issue: %w", err)
		}

		// If the error is a server-side issue, return error to retry later.
		if apiErr := new(proton.APIError); errors.As(err, &apiErr) && apiErr.Status >= 500 {
			return fmt.Errorf("failed to handle event due to server error: %w", err)
		}

		// Otherwise, the error is a client-side issue; notify bridge to handle it.
		user.log.WithField("event", event).Warn("Failed to handle API event")

		user.eventCh.Enqueue(events.UserBadEvent{
			UserID: user.ID(),
			Error:  err,
		})

		return fmt.Errorf("failed to handle event due to client error: %w", err)
	}

	user.log.WithField("event", event).Debug("Handled API event")

	// Update the event ID in the vault. If this fails, notify bridge to handle it.
	if err := user.vault.SetEventID(event.EventID); err != nil {
		user.eventCh.Enqueue(events.UserBadEvent{
			UserID: user.ID(),
			Error:  err,
		})

		return fmt.Errorf("failed to update event ID: %w", err)
	}

	user.log.WithField("eventID", event.EventID).Debug("Updated event ID in vault")

	return nil
}

// b32 returns a uint32 0 or 1 representing b.
func b32(b bool) uint32 {
	if b {
		return 1
	}

	return 0
}
