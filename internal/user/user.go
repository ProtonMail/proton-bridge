package user

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/queue"
	"github.com/ProtonMail/gluon/wait"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/safe"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
	"github.com/bradenaw/juniper/xslices"
	"github.com/emersion/go-smtp"
	"github.com/sirupsen/logrus"
	"gitlab.protontech.ch/go/liteapi"
)

var (
	EventPeriod = 20 * time.Second // nolint:gochecknoglobals
	EventJitter = 20 * time.Second // nolint:gochecknoglobals
)

type User struct {
	vault   *vault.User
	client  *liteapi.Client
	eventCh *queue.QueuedChannel[events.Event]

	apiUser  *safe.Value[liteapi.User]
	apiAddrs *safe.Slice[liteapi.Address]
	settings *safe.Value[liteapi.MailSettings]

	userKR  *crypto.KeyRing
	addrKRs map[string]*crypto.KeyRing

	updateCh   map[string]*queue.QueuedChannel[imap.Update]
	syncStopCh chan struct{}
	syncWG     wait.Group
}

func New(ctx context.Context, encVault *vault.User, client *liteapi.Client, apiUser liteapi.User) (*User, error) {
	// Get the user's API addresses.
	apiAddrs, err := client.GetAddresses(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get addresses: %w", err)
	}

	// Unlock the user's keyrings.
	userKR, addrKRs, err := liteapi.Unlock(apiUser, apiAddrs, encVault.KeyPass())
	if err != nil {
		return nil, fmt.Errorf("failed to unlock user: %w", err)
	}

	// Get the latest event ID.
	if encVault.EventID() == "" {
		eventID, err := client.GetLatestEventID(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest event ID: %w", err)
		}

		if err := encVault.SetEventID(eventID); err != nil {
			return nil, fmt.Errorf("failed to set event ID: %w", err)
		}
	}

	// Get the user's mail settings.
	settings, err := client.GetMailSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get mail settings: %w", err)
	}

	// Create update channels for each of the user's addresses (if in combined mode, just the primary).
	updateCh := make(map[string]*queue.QueuedChannel[imap.Update])

	for _, addr := range apiAddrs {
		updateCh[addr.ID] = queue.NewQueuedChannel[imap.Update](0, 0)

		if encVault.AddressMode() == vault.CombinedMode {
			break
		}
	}

	user := &User{
		vault:   encVault,
		client:  client,
		eventCh: queue.NewQueuedChannel[events.Event](0, 0),

		apiUser:  safe.NewValue(apiUser),
		apiAddrs: safe.NewSlice(apiAddrs),
		settings: safe.NewValue(settings),

		userKR:  userKR,
		addrKRs: addrKRs,

		updateCh:   updateCh,
		syncStopCh: make(chan struct{}),
	}

	// When we receive an auth object, we update it in the vault.
	// This will be used to authorize the user on the next run.
	client.AddAuthHandler(func(auth liteapi.Auth) {
		if err := user.vault.SetAuth(auth.UID, auth.RefreshToken); err != nil {
			logrus.WithError(err).Error("Failed to update auth in vault")
		}
	})

	// When we are deauthorized, we send a deauth event to the event channel.
	// Bridge will react to this event by logging out the user.
	client.AddDeauthHandler(func() {
		user.eventCh.Enqueue(events.UserDeauth{
			UserID: user.ID(),
		})
	})

	// If we haven't synced yet, do it first.
	// If it fails, we don't start the event loop.
	// Otherwise, begin processing API events, logging any errors that occur.
	go func() {
		if status := user.vault.SyncStatus(); !status.HasMessages {
			if err := <-user.startSync(); err != nil {
				return
			}
		}

		for err := range user.streamEvents() {
			logrus.WithError(err).Error("Error while streaming events")
		}
	}()

	return user, nil
}

// ID returns the user's ID.
func (user *User) ID() string {
	return safe.GetType(user.apiUser, func(apiUser liteapi.User) string {
		return apiUser.ID
	})
}

// Name returns the user's username.
func (user *User) Name() string {
	return safe.GetType(user.apiUser, func(apiUser liteapi.User) string {
		return apiUser.Name
	})
}

// Match matches the given query against the user's username and email addresses.
func (user *User) Match(query string) bool {
	return safe.GetType(user.apiUser, func(apiUser liteapi.User) bool {
		return safe.GetSlice(user.apiAddrs, func(apiAddrs []liteapi.Address) bool {
			if query == apiUser.Name {
				return true
			}

			for _, addr := range apiAddrs {
				if addr.Email == query {
					return true
				}
			}

			return false
		})
	})
}

// Emails returns all the user's email addresses.
func (user *User) Emails() []string {
	return safe.GetSlice(user.apiAddrs, func(apiAddrs []liteapi.Address) []string {
		return xslices.Map(apiAddrs, func(addr liteapi.Address) string {
			return addr.Email
		})
	})
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

	user.apiAddrs.Get(func(apiAddrs []liteapi.Address) {
		for _, addr := range apiAddrs {
			user.updateCh[addr.ID] = queue.NewQueuedChannel[imap.Update](0, 0)

			if mode == vault.CombinedMode {
				break
			}
		}
	})

	if err := user.vault.SetAddressMode(mode); err != nil {
		return fmt.Errorf("failed to set address mode: %w", err)
	}

	user.stopSync()

	if err := user.vault.ClearSyncStatus(); err != nil {
		return fmt.Errorf("failed to clear sync status: %w", err)
	}

	go func() {
		if err := <-user.startSync(); err != nil {
			logrus.WithError(err).Error("Failed to sync after setting address mode")
		}
	}()

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
	buf := new(bytes.Buffer)

	if _, err := hex.NewEncoder(buf).Write(user.vault.BridgePass()); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

// UsedSpace returns the total space used by the user on the API.
func (user *User) UsedSpace() int {
	return safe.GetType(user.apiUser, func(apiUser liteapi.User) int {
		return apiUser.UsedSpace
	})
}

// MaxSpace returns the amount of space the user can use on the API.
func (user *User) MaxSpace() int {
	return safe.GetType(user.apiUser, func(apiUser liteapi.User) int {
		return apiUser.MaxSpace
	})
}

// GetEventCh returns a channel which notifies of events happening to the user (such as deauth, address change)
func (user *User) GetEventCh() <-chan events.Event {
	return user.eventCh.GetChannel()
}

// NewIMAPConnector returns an IMAP connector for the given address.
// If not in split mode, this function returns an error.
func (user *User) NewIMAPConnector(addrID string) (connector.Connector, error) {
	return safe.GetSliceErr(user.apiAddrs, func(apiAddrs []liteapi.Address) (connector.Connector, error) {
		var emails []string

		switch user.vault.AddressMode() {
		case vault.CombinedMode:
			if addrID != apiAddrs[0].ID {
				return nil, fmt.Errorf("cannot create IMAP connector for non-primary address in combined mode")
			}

			emails = xslices.Map(apiAddrs, func(addr liteapi.Address) string {
				return addr.Email
			})

		case vault.SplitMode:
			email, err := getAddrEmail(apiAddrs, addrID)
			if err != nil {
				return nil, err
			}

			emails = []string{email}
		}

		return newIMAPConnector(
			user.client,
			user.updateCh[addrID].GetChannel(),
			user.BridgePass(),
			emails...,
		), nil
	})
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
func (user *User) NewSMTPSession(email string) (smtp.Session, error) {
	return newSMTPSession(user, email)
}

// Logout logs the user out from the API.
// If withVault is true, the user's vault is also cleared.
func (user *User) Logout(ctx context.Context) error {
	if err := user.client.AuthDelete(ctx); err != nil {
		return fmt.Errorf("failed to delete auth: %w", err)
	}

	if err := user.vault.Clear(); err != nil {
		return fmt.Errorf("failed to clear vault: %w", err)
	}

	return nil
}

// Close closes ongoing connections and cleans up resources.
func (user *User) Close() error {
	// Cancel ongoing syncs.
	user.stopSync()

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

// streamEvents begins streaming API events for the user.
// When we receive an API event, we attempt to handle it.
// If successful, we update the event ID in the vault.
func (user *User) streamEvents() <-chan error {
	errCh := make(chan error)

	go func() {
		defer close(errCh)

		for event := range user.client.NewEventStreamer(EventPeriod, EventJitter, user.vault.EventID()).Subscribe() {
			if err := user.handleAPIEvent(context.Background(), event); err != nil {
				errCh <- fmt.Errorf("failed to handle API event: %w", err)
			} else if err := user.vault.SetEventID(event.EventID); err != nil {
				errCh <- fmt.Errorf("failed to update event ID: %w", err)
			}
		}
	}()

	return errCh
}

// startSync begins a startSync for the user.
func (user *User) startSync() <-chan error {
	errCh := make(chan error)

	user.syncWG.Go(func() {
		defer close(errCh)

		ctx, cancel := contextWithStopCh(context.Background(), user.syncStopCh)
		defer cancel()

		user.eventCh.Enqueue(events.SyncStarted{
			UserID: user.ID(),
		})

		if err := user.sync(ctx); err != nil {
			user.eventCh.Enqueue(events.SyncFailed{
				UserID: user.ID(),
				Err:    err,
			})

			errCh <- err
		} else {
			user.eventCh.Enqueue(events.SyncFinished{
				UserID: user.ID(),
			})
		}
	})

	return errCh
}

// AbortSync aborts any ongoing sync.
// TODO: Should probably be done automatically when one of the user's IMAP connectors is closed.
func (user *User) stopSync() {
	select {
	case user.syncStopCh <- struct{}{}:
		user.syncWG.Wait()

	default:
		// ...
	}
}

func getAddrID(apiAddrs []liteapi.Address, email string) (string, error) {
	for _, addr := range apiAddrs {
		if addr.Email == email {
			return addr.ID, nil
		}
	}

	return "", fmt.Errorf("address %s not found", email)
}

func getAddrEmail(apiAddrs []liteapi.Address, addrID string) (string, error) {
	for _, addr := range apiAddrs {
		if addr.ID == addrID {
			return addr.Email, nil
		}
	}

	return "", fmt.Errorf("address %s not found", addrID)
}

// contextWithStopCh returns a new context that is cancelled when the stop channel is closed or a value is sent to it.
func contextWithStopCh(ctx context.Context, stopCh <-chan struct{}) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		select {
		case <-stopCh:
			cancel()

		case <-ctx.Done():
			// ...
		}
	}()

	return ctx, cancel
}
