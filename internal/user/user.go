// Copyright (c) 2022 Proton AG
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
	"bytes"
	"context"
	"crypto/subtle"
	"fmt"
	"io"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/queue"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/internal/async"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/safe"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
	"github.com/ProtonMail/proton-bridge/v2/pkg/message"
	"github.com/ProtonMail/proton-bridge/v2/pkg/message/parser"
	"github.com/bradenaw/juniper/xslices"
	"github.com/bradenaw/juniper/xsync"
	"github.com/sirupsen/logrus"
	"gitlab.protontech.ch/go/liteapi"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

var (
	EventPeriod = 20 * time.Second // nolint:gochecknoglobals,revive
	EventJitter = 20 * time.Second // nolint:gochecknoglobals,revive
)

type User struct {
	log *logrus.Entry

	vault    *vault.User
	client   *liteapi.Client
	eventCh  *queue.QueuedChannel[events.Event]
	sendHash *sendRecorder

	apiUser     liteapi.User
	apiUserLock safe.RWMutex

	apiAddrs     map[string]liteapi.Address
	apiAddrsLock safe.RWMutex

	apiLabels     map[string]liteapi.Label
	apiLabelsLock safe.RWMutex

	updateCh     map[string]*queue.QueuedChannel[imap.Update]
	updateChLock safe.RWMutex

	tasks     *xsync.Group
	abortable async.Abortable
	goSync    func()
	goPoll    func()

	syncWorkers int
	syncBuffer  int
	showAllMail uint32
}

// New returns a new user.
//
// nolint:funlen
func New(
	ctx context.Context,
	encVault *vault.User,
	client *liteapi.Client,
	apiUser liteapi.User,
	syncWorkers, syncBuffer int,
	showAllMail bool,
) (*User, error) { //nolint:funlen
	logrus.WithField("userID", apiUser.ID).Debug("Creating new user")

	// Get the user's API addresses.
	apiAddrs, err := client.GetAddresses(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get addresses: %w", err)
	}

	// Check we can unlock the keyrings.
	if _, _, err := liteapi.Unlock(apiUser, apiAddrs, encVault.KeyPass()); err != nil {
		return nil, fmt.Errorf("failed to unlock user: %w", err)
	}

	// Get the user's API labels.
	apiLabels, err := client.GetLabels(ctx, liteapi.LabelTypeSystem, liteapi.LabelTypeFolder, liteapi.LabelTypeLabel)
	if err != nil {
		return nil, fmt.Errorf("failed to get labels: %w", err)
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

	// Create update channels for each of the user's addresses.
	// In combined mode, the addresses all share the same update channel.
	updateCh := make(map[string]*queue.QueuedChannel[imap.Update])

	switch encVault.AddressMode() {
	case vault.CombinedMode:
		primaryUpdateCh := queue.NewQueuedChannel[imap.Update](0, 0)

		for _, addr := range apiAddrs {
			updateCh[addr.ID] = primaryUpdateCh
		}

	case vault.SplitMode:
		for _, addr := range apiAddrs {
			updateCh[addr.ID] = queue.NewQueuedChannel[imap.Update](0, 0)
		}
	}

	user := &User{
		log: logrus.WithField("userID", apiUser.ID),

		vault:    encVault,
		client:   client,
		eventCh:  queue.NewQueuedChannel[events.Event](0, 0),
		sendHash: newSendRecorder(sendEntryExpiry),

		apiUser:     apiUser,
		apiUserLock: safe.NewRWMutex(),

		apiAddrs:     groupBy(apiAddrs, func(addr liteapi.Address) string { return addr.ID }),
		apiAddrsLock: safe.NewRWMutex(),

		apiLabels:     groupBy(apiLabels, func(label liteapi.Label) string { return label.ID }),
		apiLabelsLock: safe.NewRWMutex(),

		updateCh:     updateCh,
		updateChLock: safe.NewRWMutex(),

		tasks: xsync.NewGroup(context.Background()),

		syncWorkers: syncWorkers,
		syncBuffer:  syncBuffer,
		showAllMail: b32(showAllMail),
	}

	// When we receive an auth object, we update it in the vault.
	// This will be used to authorize the user on the next run.
	user.client.AddAuthHandler(func(auth liteapi.Auth) {
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

	// Stream events from the API, logging any errors that occur.
	// This does nothing until the sync has been marked as complete.
	// When we receive an API event, we attempt to handle it.
	// If successful, we update the event ID in the vault.
	user.goPoll = user.tasks.PeriodicOrTrigger(EventPeriod, EventJitter, func(ctx context.Context) {
		if !user.vault.SyncStatus().IsComplete() {
			user.log.Debug("Sync is incomplete, skipping event stream")
			return
		}

		event, err := user.client.GetEvent(ctx, user.vault.EventID())
		if err != nil {
			user.log.WithError(err).Error("Failed to get event")
			return
		}

		if event.EventID == user.vault.EventID() {
			user.log.Debug("No new events")
			return
		}

		if err := user.handleAPIEvent(ctx, event); err != nil {
			user.log.WithError(err).Error("Failed to handle API event")
			return
		}

		if err := user.vault.SetEventID(event.EventID); err != nil {
			user.log.WithError(err).Error("Failed to update event ID in vault")
			return
		}

		user.log.WithField("eventID", event.EventID).Debug("Updated event ID")
	})

	// When triggered, attempt to sync the user.
	user.goSync = user.tasks.Trigger(func(ctx context.Context) {
		user.abortable.Do(ctx, func(ctx context.Context) {
			if !user.vault.SyncStatus().IsComplete() {
				if err := user.doSync(ctx); err != nil {
					return
				}
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

		slices.SortFunc(addresses, func(a, b liteapi.Address) bool {
			return a.Order < b.Order
		})

		return xslices.Map(addresses, func(addr liteapi.Address) string {
			return addr.Email
		})
	}, user.apiAddrsLock)
}

// GetAddressMode returns the user's current address mode.
func (user *User) GetAddressMode() vault.AddressMode {
	return user.vault.AddressMode()
}

// SetAddressMode sets the user's address mode.
func (user *User) SetAddressMode(ctx context.Context, mode vault.AddressMode) error {
	user.abortable.Abort()
	defer user.goSync()

	return safe.LockRet(func() error {
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

		if err := user.vault.SetAddressMode(mode); err != nil {
			return fmt.Errorf("failed to set address mode: %w", err)
		}

		if err := user.vault.ClearSyncStatus(); err != nil {
			return fmt.Errorf("failed to clear sync status: %w", err)
		}

		return nil
	}, user.apiAddrsLock, user.updateChLock)
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
	return b64Encode(user.vault.BridgePass())
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
	defer user.goPoll()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if len(to) == 0 {
		return ErrInvalidRecipient
	}

	return safe.RLockRet(func() error {
		if _, err := getAddrID(user.apiAddrs, from); err != nil {
			return ErrInvalidReturnPath
		}

		emails := xslices.Map(maps.Values(user.apiAddrs), func(addr liteapi.Address) string {
			return addr.Email
		})

		// Read the message to send.
		b, err := io.ReadAll(r)
		if err != nil {
			return fmt.Errorf("failed to read message: %w", err)
		}

		// Compute the hash of the message (to match it against SMTP messages).
		hash, err := getMessageHash(b)
		if err != nil {
			return err
		}

		// Check if we already tried to send this message recently.
		if ok, err := user.sendHash.tryInsertWait(ctx, hash, to, time.Now().Add(90*time.Second)); err != nil {
			return fmt.Errorf("failed to check send hash: %w", err)
		} else if !ok {
			user.log.Warn("A duplicate message was already sent recently, skipping")
			return nil
		}

		// If we fail to send this message, we should remove the hash from the send recorder.
		defer user.sendHash.removeOnFail(hash)

		// Create a new message parser from the reader.
		parser, err := parser.New(bytes.NewReader(b))
		if err != nil {
			return fmt.Errorf("failed to create parser: %w", err)
		}

		// If the message contains a sender, use it instead of the one from the return path.
		if sender, ok := getMessageSender(parser); ok {
			from = sender
		}

		// Load the user's mail settings.
		settings, err := user.client.GetMailSettings(ctx)
		if err != nil {
			return fmt.Errorf("failed to get mail settings: %w", err)
		}

		addrID, err := getAddrID(user.apiAddrs, from)
		if err != nil {
			return err
		}

		return withAddrKR(user.apiUser, user.apiAddrs[addrID], user.vault.KeyPass(), func(userKR, addrKR *crypto.KeyRing) error {
			// Use the first key for encrypting the message.
			addrKR, err := addrKR.FirstKey()
			if err != nil {
				return fmt.Errorf("failed to get first key: %w", err)
			}

			// If we have to attach the public key, do it now.
			if settings.AttachPublicKey == liteapi.AttachPublicKeyEnabled {
				key, err := addrKR.GetKey(0)
				if err != nil {
					return fmt.Errorf("failed to get sending key: %w", err)
				}

				pubKey, err := key.GetArmoredPublicKey()
				if err != nil {
					return fmt.Errorf("failed to get public key: %w", err)
				}

				parser.AttachPublicKey(pubKey, fmt.Sprintf("publickey - %v - %v", addrKR.GetIdentities()[0].Name, key.GetFingerprint()[:8]))
			}

			// Parse the message we want to send (after we have attached the public key).
			message, err := message.ParseWithParser(parser)
			if err != nil {
				return fmt.Errorf("failed to parse message: %w", err)
			}

			// Send the message using the correct key.
			sent, err := sendWithKey(
				ctx,
				user.client,
				authID,
				user.vault.AddressMode(),
				settings,
				userKR, addrKR,
				emails, from, to,
				message,
			)
			if err != nil {
				return fmt.Errorf("failed to send message: %w", err)
			}

			// If the message was successfully sent, we can update the message ID in the record.
			user.sendHash.addMessageID(hash, sent.ID)

			return nil
		})
	}, user.apiUserLock, user.apiAddrsLock)
}

// CheckAuth returns whether the given email and password can be used to authenticate over IMAP or SMTP with this user.
// It returns the address ID of the authenticated address.
func (user *User) CheckAuth(email string, password []byte) (string, error) {
	dec, err := b64Decode(password)
	if err != nil {
		return "", fmt.Errorf("failed to decode password: %w", err)
	}

	if subtle.ConstantTimeCompare(user.vault.BridgePass(), dec) != 1 {
		return "", fmt.Errorf("invalid password")
	}

	return safe.RLockRetErr(func() (string, error) {
		for _, addr := range user.apiAddrs {
			if strings.EqualFold(addr.Email, email) {
				return addr.ID, nil
			}
		}

		return "", fmt.Errorf("invalid email")
	}, user.apiAddrsLock)
}

// OnStatusUp is called when the connection goes up.
func (user *User) OnStatusUp(context.Context) {
	user.goSync()
}

// OnStatusDown is called when the connection goes down.
func (user *User) OnStatusDown(context.Context) {
	user.abortable.Abort()
}

// Logout logs the user out from the API.
func (user *User) Logout(ctx context.Context, withAPI bool) error {
	user.tasks.Wait()

	if withAPI {
		if err := user.client.AuthDelete(ctx); err != nil {
			return fmt.Errorf("failed to delete auth: %w", err)
		}
	}

	if err := user.vault.Clear(); err != nil {
		return fmt.Errorf("failed to clear vault: %w", err)
	}

	return nil
}

// Close closes ongoing connections and cleans up resources.
func (user *User) Close() {
	// Stop any ongoing background tasks.
	user.tasks.Wait()

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

// SetShowAllMail sets whether to show the All Mail mailbox.
func (user *User) SetShowAllMail(show bool) {
	atomic.StoreUint32(&user.showAllMail, b32(show))
}

// b32 returns a uint32 0 or 1 representing b.
func b32(b bool) uint32 {
	if b {
		return 1
	}

	return 0
}
