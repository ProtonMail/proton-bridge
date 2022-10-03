package user

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/queue"
	"github.com/ProtonMail/gluon/wait"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/pool"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
	"github.com/emersion/go-smtp"
	"github.com/sirupsen/logrus"
	"gitlab.protontech.ch/go/liteapi"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

var (
	DefaultEventPeriod = 20 * time.Second
	DefaultEventJitter = 20 * time.Second
)

type User struct {
	vault   *vault.User
	client  *liteapi.Client
	builder *pool.Pool[request, *imap.MessageCreated]
	eventCh *queue.QueuedChannel[events.Event]

	apiUser  liteapi.User
	apiAddrs *addrList
	userKR   *crypto.KeyRing
	addrKRs  map[string]*crypto.KeyRing
	settings liteapi.MailSettings

	updateCh map[string]*queue.QueuedChannel[imap.Update]
	syncWG   wait.Group
}

func New(
	ctx context.Context,
	encVault *vault.User,
	client *liteapi.Client,
	apiUser liteapi.User,
	apiAddrs []liteapi.Address,
	userKR *crypto.KeyRing,
	addrKRs map[string]*crypto.KeyRing,
) (*User, error) {
	if encVault.EventID() == "" {
		eventID, err := client.GetLatestEventID(ctx)
		if err != nil {
			return nil, err
		}

		if err := encVault.SetEventID(eventID); err != nil {
			return nil, err
		}
	}

	settings, err := client.GetMailSettings(ctx)
	if err != nil {
		return nil, err
	}

	user := &User{
		vault:   encVault,
		client:  client,
		builder: newBuilder(client, runtime.NumCPU()*runtime.NumCPU(), runtime.NumCPU()*runtime.NumCPU()),
		eventCh: queue.NewQueuedChannel[events.Event](0, 0),

		apiUser:  apiUser,
		apiAddrs: newAddrList(apiAddrs),

		userKR:   userKR,
		addrKRs:  addrKRs,
		settings: settings,

		updateCh: make(map[string]*queue.QueuedChannel[imap.Update]),
	}

	// Initialize update channels for each of the user's addresses.
	for _, addrID := range user.apiAddrs.addrIDs() {
		user.updateCh[addrID] = queue.NewQueuedChannel[imap.Update](0, 0)

		// If in combined mode, we only need one update channel.
		if encVault.AddressMode() == vault.CombinedMode {
			break
		}
	}

	// When we receive an auth object, we update it in the store.
	// This will be used to authorize the user on the next run.
	client.AddAuthHandler(func(auth liteapi.Auth) {
		if err := user.vault.SetAuth(auth.UID, auth.RefreshToken); err != nil {
			logrus.WithError(err).Error("Failed to update auth")
		}
	})

	// When we are deauthorized, we send a deauth event to the notify channel.
	// Bridge will catch this and log the user out.
	client.AddDeauthHandler(func() {
		user.eventCh.Enqueue(events.UserDeauth{
			UserID: user.ID(),
		})
	})

	// When we receive an API event, we attempt to handle it.
	// If successful, we update the event ID in the vault.
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		for event := range user.client.NewEventStreamer(DefaultEventPeriod, DefaultEventJitter, encVault.EventID()).Subscribe() {
			if err := user.handleAPIEvent(ctx, event); err != nil {
				logrus.WithError(err).Error("Failed to handle event")
			} else if err := user.vault.SetEventID(event.EventID); err != nil {
				logrus.WithError(err).Error("Failed to update event ID")
			}
		}
	}()

	return user, nil
}

// ID returns the user's ID.
func (user *User) ID() string {
	return user.apiUser.ID
}

// Name returns the user's username.
func (user *User) Name() string {
	return user.apiUser.Name
}

// Match matches the given query against the user's username and email addresses.
func (user *User) Match(query string) bool {
	if query == user.apiUser.Name {
		return true
	}

	return slices.Contains(user.apiAddrs.emails(), query)
}

// Emails returns all the user's email addresses.
func (user *User) Emails() []string {
	return user.apiAddrs.emails()
}

// GetAddressMode returns the user's current address mode.
func (user *User) GetAddressMode() vault.AddressMode {
	return user.vault.AddressMode()
}

// SetAddressMode sets the user's address mode.
func (user *User) SetAddressMode(ctx context.Context, mode vault.AddressMode) error {
	for _, updateCh := range user.updateCh {
		updateCh.Close()
	}

	user.updateCh = make(map[string]*queue.QueuedChannel[imap.Update])

	for _, addrID := range user.apiAddrs.addrIDs() {
		user.updateCh[addrID] = queue.NewQueuedChannel[imap.Update](0, 0)

		if mode == vault.CombinedMode {
			break
		}
	}

	if err := user.vault.SetAddressMode(mode); err != nil {
		return fmt.Errorf("failed to set address mode: %w", err)
	}

	return nil
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
	return user.vault.SetGluonID(addrID, gluonID)
}

// GluonKey returns the user's gluon key from the vault.
func (user *User) GluonKey() []byte {
	return user.vault.GluonKey()
}

// BridgePass returns the user's bridge password, used for authentication over SMTP and IMAP.
func (user *User) BridgePass() []byte {
	return []byte(user.vault.BridgePass())
}

// UsedSpace returns the total space used by the user on the API.
func (user *User) UsedSpace() int {
	return user.apiUser.UsedSpace
}

// MaxSpace returns the amount of space the user can use on the API.
func (user *User) MaxSpace() int {
	return user.apiUser.MaxSpace
}

// HasSync returns whether the user has finished syncing.
func (user *User) HasSync() bool {
	return user.vault.HasSync()
}

// AbortSync aborts any ongoing sync.
// TODO: This should abort the sync rather than just waiting.
// Should probably be done automatically when one of the user's IMAP connectors is closed.
func (user *User) AbortSync(ctx context.Context) error {
	user.syncWG.Wait()

	return nil
}

// DoSync performs a sync for the user.
func (user *User) DoSync(ctx context.Context) <-chan error {
	errCh := queue.NewQueuedChannel[error](0, 0)

	user.syncWG.Go(func() {
		defer errCh.Close()

		user.eventCh.Enqueue(events.SyncStarted{
			UserID: user.ID(),
		})

		errCh.Enqueue(func() error {
			if err := user.syncLabels(ctx, maps.Keys(user.updateCh)...); err != nil {
				return fmt.Errorf("failed to sync labels: %w", err)
			}

			if err := user.syncMessages(ctx); err != nil {
				return fmt.Errorf("failed to sync messages: %w", err)
			}

			user.syncWait()

			if err := user.vault.SetSync(true); err != nil {
				return fmt.Errorf("failed to set sync status: %w", err)
			}

			return nil
		}())

		user.eventCh.Enqueue(events.SyncFinished{
			UserID: user.ID(),
		})
	})

	return errCh.GetChannel()
}

// GetEventCh returns a channel which notifies of events happening to the user (such as deauth, address change)
func (user *User) GetEventCh() <-chan events.Event {
	return user.eventCh.GetChannel()
}

// NewIMAPConnector returns an IMAP connector for the given address.
// If not in split mode, this function returns an error.
func (user *User) NewIMAPConnector(addrID string) (connector.Connector, error) {
	var emails []string

	switch user.vault.AddressMode() {
	case vault.CombinedMode:
		if addrID != user.apiAddrs.primary() {
			return nil, fmt.Errorf("cannot create IMAP connector for non-primary address in combined mode")
		}

		emails = user.apiAddrs.emails()

	case vault.SplitMode:
		emails = []string{user.apiAddrs.email(addrID)}
	}

	return newIMAPConnector(
		user.client,
		user.updateCh[addrID].GetChannel(),
		user.vault.BridgePass(),
		emails...,
	), nil
}

// NewIMAPConnectors returns IMAP connectors for each of the user's addresses.
// In combined mode, this is just the user's primary address.
// In split mode, this is all the user's addresses.
func (user *User) NewIMAPConnectors() (map[string]connector.Connector, error) {
	imapConn := make(map[string]connector.Connector)

	for addrID := range user.updateCh {
		conn, err := user.NewIMAPConnector(addrID)
		if err != nil {
			return nil, fmt.Errorf("failed to create IMAP connector: %w", err)
		}

		imapConn[addrID] = conn
	}

	return imapConn, nil
}

// NewSMTPSession returns an SMTP session for the user.
func (user *User) NewSMTPSession(username string) smtp.Session {
	return newSMTPSession(user.client, username, user.apiAddrs.addrMap(), user.settings, user.userKR, user.addrKRs)
}

// Logout logs the user out from the API.
func (user *User) Logout(ctx context.Context) error {
	return user.client.AuthDelete(ctx)
}

// Close closes ongoing connections and cleans up resources.
func (user *User) Close(ctx context.Context) error {
	// Wait for ongoing syncs to finish.
	user.syncWG.Wait()

	// Close the user's message builder.
	user.builder.Done()

	// Close the user's API client.
	user.client.Close()

	// Close the user's update channels.
	for _, updateCh := range user.updateCh {
		updateCh.Close()
	}

	// Close the user's notify channel.
	user.eventCh.Close()

	return nil
}
