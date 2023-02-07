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
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/queue"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/logging"
	"github.com/ProtonMail/proton-bridge/v3/internal/safe"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
)

// handleAPIEvent handles the given proton.Event.
func (user *User) handleAPIEvent(ctx context.Context, event proton.Event) error {
	if event.Refresh&proton.RefreshMail != 0 {
		return user.handleRefreshEvent(ctx, event.Refresh, event.EventID)
	}

	if event.User != nil {
		user.handleUserEvent(ctx, *event.User)
	}

	if len(event.Addresses) > 0 {
		if err := user.handleAddressEvents(ctx, event.Addresses); err != nil {
			return err
		}
	}

	if len(event.Labels) > 0 {
		if err := user.handleLabelEvents(ctx, event.Labels); err != nil {
			return err
		}
	}

	if len(event.Messages) > 0 {
		if err := user.handleMessageEvents(ctx, event.Messages); err != nil {
			return err
		}
	}

	if event.UsedSpace != nil {
		user.handleUsedSpaceChange(*event.UsedSpace)
	}

	return nil
}

func (user *User) handleRefreshEvent(ctx context.Context, refresh proton.RefreshFlag, eventID string) error {
	l := user.log.WithFields(logrus.Fields{
		"eventID": eventID,
		"refresh": refresh,
	})

	l.Info("Handling refresh event")

	if err := user.reporter.ReportMessageWithContext("Warning: refresh occurred", map[string]interface{}{
		"EventLoop": map[string]interface{}{
			"EventID": eventID,
			"Refresh": refresh,
		},
	}); err != nil {
		l.WithError(err).Error("Failed to report refresh to sentry")
	}

	// Cancel the event stream once this refresh is done.
	defer user.pollAbort.Abort()

	// Resync after the refresh.
	defer user.goSync()

	return safe.LockRet(func() error {
		// Fetch latest user info.
		apiUser, err := user.client.GetUser(ctx)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		// Fetch latest address info.
		apiAddrs, err := user.client.GetAddresses(ctx)
		if err != nil {
			return fmt.Errorf("failed to get addresses: %w", err)
		}

		// Fetch latest label info.
		apiLabels, err := user.client.GetLabels(ctx, proton.LabelTypeSystem, proton.LabelTypeFolder, proton.LabelTypeLabel)
		if err != nil {
			return fmt.Errorf("failed to get labels: %w", err)
		}

		// Update the API info in the user.
		user.apiUser = apiUser
		user.apiAddrs = groupBy(apiAddrs, func(addr proton.Address) string { return addr.ID })
		user.apiLabels = groupBy(apiLabels, func(label proton.Label) string { return label.ID })

		// Clear sync status; we want to sync everything again.
		if err := user.clearSyncStatus(); err != nil {
			return fmt.Errorf("failed to clear sync status: %w", err)
		}

		// The user was refreshed.
		user.eventCh.Enqueue(events.UserRefreshed{
			UserID: user.apiUser.ID,
		})

		return nil
	}, user.apiUserLock, user.apiAddrsLock, user.apiLabelsLock, user.updateChLock)
}

// handleUserEvent handles the given user event.
func (user *User) handleUserEvent(_ context.Context, userEvent proton.User) {
	safe.Lock(func() {
		user.log.WithFields(logrus.Fields{
			"userID":   userEvent.ID,
			"username": logging.Sensitive(userEvent.Name),
		}).Info("Handling user event")

		user.apiUser = userEvent

		user.eventCh.Enqueue(events.UserChanged{
			UserID: user.apiUser.ID,
		})
	}, user.apiUserLock)
}

// handleAddressEvents handles the given address events.
// GODT-1945: If split address mode, need to signal back to bridge to update the addresses.
func (user *User) handleAddressEvents(ctx context.Context, addressEvents []proton.AddressEvent) error {
	for _, event := range addressEvents {
		switch event.Action {
		case proton.EventCreate:
			if err := user.handleCreateAddressEvent(ctx, event); err != nil {
				if rerr := user.reporter.ReportMessageWithContext("Failed to apply address create event", reporter.Context{
					"error": err,
				}); rerr != nil {
					user.log.WithError(err).Error("Failed to report address create event error")
				}
				return fmt.Errorf("failed to handle create address event: %w", err)
			}

		case proton.EventUpdate, proton.EventUpdateFlags:
			if err := user.handleUpdateAddressEvent(ctx, event); err != nil {
				if rerr := user.reporter.ReportMessageWithContext("Failed to apply address update event", reporter.Context{
					"error": err,
				}); rerr != nil {
					user.log.WithError(err).Error("Failed to report address update event error")
				}
				return fmt.Errorf("failed to handle update address event: %w", err)
			}

		case proton.EventDelete:
			if err := user.handleDeleteAddressEvent(ctx, event); err != nil {
				if rerr := user.reporter.ReportMessageWithContext("Failed to apply address delete event", reporter.Context{
					"error": err,
				}); rerr != nil {
					user.log.WithError(err).Error("Failed to report address delete event error")
				}
				return fmt.Errorf("failed to delete address: %w", err)
			}
		}
	}

	return nil
}

func (user *User) handleCreateAddressEvent(ctx context.Context, event proton.AddressEvent) error {
	if err := safe.LockRet(func() error {
		user.log.WithFields(logrus.Fields{
			"addressID": event.ID,
			"email":     logging.Sensitive(event.Address.Email),
		}).Info("Handling address created event")

		if _, ok := user.apiAddrs[event.Address.ID]; ok {
			user.log.Debugf("Address %q already exists", event.ID)
			return nil
		}

		user.apiAddrs[event.Address.ID] = event.Address

		switch user.vault.AddressMode() {
		case vault.CombinedMode:
			primAddr, err := getAddrIdx(user.apiAddrs, 0)
			if err != nil {
				return fmt.Errorf("failed to get primary address: %w", err)
			}

			user.updateCh[event.Address.ID] = user.updateCh[primAddr.ID]

		case vault.SplitMode:
			user.updateCh[event.Address.ID] = queue.NewQueuedChannel[imap.Update](0, 0)
		}

		user.eventCh.Enqueue(events.UserAddressCreated{
			UserID:    user.apiUser.ID,
			AddressID: event.Address.ID,
			Email:     event.Address.Email,
		})

		return nil
	}, user.apiAddrsLock, user.updateChLock); err != nil {
		return fmt.Errorf("failed to handle create address event: %w", err)
	}

	// Perform the sync in an RLock.
	return safe.RLockRet(func() error {
		if user.vault.AddressMode() == vault.SplitMode {
			if err := syncLabels(ctx, user.apiLabels, user.updateCh[event.Address.ID]); err != nil {
				return fmt.Errorf("failed to sync labels to new address: %w", err)
			}
		}

		return nil
	}, user.apiAddrsLock, user.apiLabelsLock, user.updateChLock)
}

func (user *User) handleUpdateAddressEvent(_ context.Context, event proton.AddressEvent) error { //nolint:unparam
	return safe.LockRet(func() error {
		user.log.WithFields(logrus.Fields{
			"addressID": event.ID,
			"email":     logging.Sensitive(event.Address.Email),
		}).Info("Handling address updated event")

		if _, ok := user.apiAddrs[event.Address.ID]; !ok {
			user.log.Debugf("Address %q does not exist", event.Address.ID)
			return nil
		}

		user.apiAddrs[event.Address.ID] = event.Address

		user.eventCh.Enqueue(events.UserAddressUpdated{
			UserID:    user.apiUser.ID,
			AddressID: event.Address.ID,
			Email:     event.Address.Email,
		})

		return nil
	}, user.apiAddrsLock)
}

func (user *User) handleDeleteAddressEvent(_ context.Context, event proton.AddressEvent) error {
	return safe.LockRet(func() error {
		user.log.WithField("addressID", event.ID).Info("Handling address deleted event")

		addr, ok := user.apiAddrs[event.ID]
		if !ok {
			user.log.Debugf("Address %q does not exist", event.ID)
			return nil
		}

		if user.vault.AddressMode() == vault.SplitMode {
			user.updateCh[event.ID].CloseAndDiscardQueued()
			delete(user.updateCh, event.ID)
		}

		delete(user.apiAddrs, event.ID)

		user.eventCh.Enqueue(events.UserAddressDeleted{
			UserID:    user.apiUser.ID,
			AddressID: event.ID,
			Email:     addr.Email,
		})

		return nil
	}, user.apiAddrsLock, user.updateChLock)
}

// handleLabelEvents handles the given label events.
func (user *User) handleLabelEvents(ctx context.Context, labelEvents []proton.LabelEvent) error {
	for _, event := range labelEvents {
		switch event.Action {
		case proton.EventCreate:
			updates, err := user.handleCreateLabelEvent(ctx, event)
			if err != nil {
				return fmt.Errorf("failed to handle create label event: %w", err)
			}

			if err := waitOnIMAPUpdates(ctx, updates); err != nil {
				return err
			}

		case proton.EventUpdate, proton.EventUpdateFlags:
			updates, err := user.handleUpdateLabelEvent(ctx, event)
			if err != nil {
				return fmt.Errorf("failed to handle update label event: %w", err)
			}

			if err := waitOnIMAPUpdates(ctx, updates); err != nil {
				return err
			}

		case proton.EventDelete:
			updates, err := user.handleDeleteLabelEvent(ctx, event)
			if err != nil {
				return fmt.Errorf("failed to handle delete label event: %w", err)
			}

			if err := waitOnIMAPUpdates(ctx, updates); err != nil {
				return fmt.Errorf("failed to handle delete label event in gluon: %w", err)
			}
		}
	}

	return nil
}

func (user *User) handleCreateLabelEvent(ctx context.Context, event proton.LabelEvent) ([]imap.Update, error) { //nolint:unparam
	return safe.LockRetErr(func() ([]imap.Update, error) {
		var updates []imap.Update

		user.log.WithFields(logrus.Fields{
			"labelID": event.ID,
			"name":    logging.Sensitive(event.Label.Name),
		}).Info("Handling label created event")

		user.apiLabels[event.Label.ID] = event.Label

		for _, updateCh := range xslices.Unique(maps.Values(user.updateCh)) {
			update := newMailboxCreatedUpdate(imap.MailboxID(event.ID), getMailboxName(event.Label))
			updateCh.Enqueue(update)
			updates = append(updates, update)
		}

		user.eventCh.Enqueue(events.UserLabelCreated{
			UserID:  user.apiUser.ID,
			LabelID: event.Label.ID,
			Name:    event.Label.Name,
		})

		return updates, nil
	}, user.apiLabelsLock, user.updateChLock)
}

func (user *User) handleUpdateLabelEvent(ctx context.Context, event proton.LabelEvent) ([]imap.Update, error) { //nolint:unparam
	return safe.LockRetErr(func() ([]imap.Update, error) {
		var updates []imap.Update

		user.log.WithFields(logrus.Fields{
			"labelID": event.ID,
			"name":    logging.Sensitive(event.Label.Name),
		}).Info("Handling label updated event")

		// Only update the label if it exists; we don't want to create it as a client may have just deleted it.
		if _, ok := user.apiLabels[event.Label.ID]; ok {
			user.apiLabels[event.Label.ID] = event.Label
		}

		for _, updateCh := range xslices.Unique(maps.Values(user.updateCh)) {
			update := imap.NewMailboxUpdated(
				imap.MailboxID(event.ID),
				getMailboxName(event.Label),
			)
			updateCh.Enqueue(update)
			updates = append(updates, update)
		}

		user.eventCh.Enqueue(events.UserLabelUpdated{
			UserID:  user.apiUser.ID,
			LabelID: event.Label.ID,
			Name:    event.Label.Name,
		})

		return updates, nil
	}, user.apiLabelsLock, user.updateChLock)
}

func (user *User) handleDeleteLabelEvent(ctx context.Context, event proton.LabelEvent) ([]imap.Update, error) { //nolint:unparam
	return safe.LockRetErr(func() ([]imap.Update, error) {
		var updates []imap.Update

		user.log.WithField("labelID", event.ID).Info("Handling label deleted event")

		for _, updateCh := range xslices.Unique(maps.Values(user.updateCh)) {
			update := imap.NewMailboxDeleted(imap.MailboxID(event.ID))
			updateCh.Enqueue(update)
			updates = append(updates, update)
		}

		delete(user.apiLabels, event.ID)

		user.eventCh.Enqueue(events.UserLabelDeleted{
			UserID:  user.apiUser.ID,
			LabelID: event.ID,
		})

		return updates, nil
	}, user.apiLabelsLock, user.updateChLock)
}

// handleMessageEvents handles the given message events.
func (user *User) handleMessageEvents(ctx context.Context, messageEvents []proton.MessageEvent) error {
	for _, event := range messageEvents {
		ctx = logging.WithLogrusField(ctx, "messageID", event.ID)

		switch event.Action {
		case proton.EventCreate:
			updates, err := user.handleCreateMessageEvent(logging.WithLogrusField(ctx, "action", "create message"), event)
			if err != nil {
				if rerr := user.reporter.ReportMessageWithContext("Failed to apply create message event", reporter.Context{
					"error": err,
				}); rerr != nil {
					user.log.WithError(err).Error("Failed to report create message event error")
				}

				return fmt.Errorf("failed to handle create message event: %w", err)
			}

			if err := waitOnIMAPUpdates(ctx, updates); err != nil {
				return err
			}

		case proton.EventUpdate, proton.EventUpdateFlags:
			// Draft update means to completely remove old message and upload the new data again, but we should
			// only do this if the event is of type EventUpdate otherwise label switch operations will not work.
			if event.Message.IsDraft() && event.Action == proton.EventUpdate {
				updates, err := user.handleUpdateDraftEvent(
					logging.WithLogrusField(ctx, "action", "update draft"),
					event,
				)
				if err != nil {
					if rerr := user.reporter.ReportMessageWithContext("Failed to apply update draft message event", reporter.Context{
						"error": err,
					}); rerr != nil {
						user.log.WithError(err).Error("Failed to report update draft message event error")
					}
					return fmt.Errorf("failed to handle update draft event: %w", err)
				}

				if err := waitOnIMAPUpdates(ctx, updates); err != nil {
					return err
				}

				return nil
			}

			// GODT-2028 - Use better events here. It should be possible to have 3 separate events that refrain to
			// whether the flags, labels or read only data (header+body) has been changed. This requires fixing proton
			// first so that it correctly reports those cases.
			// Issue regular update to handle mailboxes and flag changes.
			updates, err := user.handleUpdateMessageEvent(
				logging.WithLogrusField(ctx, "action", "update message"),
				event,
			)
			if err != nil {
				if rerr := user.reporter.ReportMessageWithContext("Failed to apply update message event", reporter.Context{
					"error": err,
				}); rerr != nil {
					user.log.WithError(err).Error("Failed to report update message event error")
				}
				return fmt.Errorf("failed to handle update message event: %w", err)
			}

			if err := waitOnIMAPUpdates(ctx, updates); err != nil {
				return err
			}

		case proton.EventDelete:
			updates, err := user.handleDeleteMessageEvent(
				logging.WithLogrusField(ctx, "action", "delete message"),
				event,
			)
			if err != nil {
				if rerr := user.reporter.ReportMessageWithContext("Failed to apply delete message event", reporter.Context{
					"error": err,
				}); rerr != nil {
					user.log.WithError(err).Error("Failed to report delete message event error")
				}
				return fmt.Errorf("failed to handle delete message event: %w", err)
			}

			if err := waitOnIMAPUpdates(ctx, updates); err != nil {
				return fmt.Errorf("failed to handle delete message event in gluon: %w", err)
			}
		}
	}

	return nil
}

func (user *User) handleCreateMessageEvent(ctx context.Context, event proton.MessageEvent) ([]imap.Update, error) {
	full, err := user.client.GetFullMessage(ctx, event.Message.ID, newProtonAPIScheduler(), proton.NewDefaultAttachmentAllocator())
	if err != nil {
		// If the message is not found, it means that it has been deleted before we could fetch it.
		if apiErr := new(proton.APIError); errors.As(err, &apiErr) && apiErr.Status == http.StatusUnprocessableEntity {
			user.log.WithField("messageID", event.Message.ID).Warn("Cannot add new message: full message is missing on API")
			return nil, nil
		}

		return nil, fmt.Errorf("failed to get full message: %w", err)
	}

	return safe.RLockRetErr(func() ([]imap.Update, error) {
		user.log.WithFields(logrus.Fields{
			"messageID": event.ID,
			"subject":   logging.Sensitive(event.Message.Subject),
		}).Info("Handling message created event")

		var update imap.Update
		if err := withAddrKR(user.apiUser, user.apiAddrs[event.Message.AddressID], user.vault.KeyPass(), func(_, addrKR *crypto.KeyRing) error {
			res := buildRFC822(user.apiLabels, full, addrKR, new(bytes.Buffer))

			if res.err != nil {
				user.log.WithError(err).Error("Failed to build RFC822 message")

				if err := user.vault.AddFailedMessageID(event.ID); err != nil {
					user.log.WithError(err).Error("Failed to add failed message ID to vault")
				}

				if err := user.reporter.ReportMessageWithContext("Failed to build message (event create)", reporter.Context{
					"messageID": res.messageID,
					"error":     res.err,
				}); err != nil {
					user.log.WithError(err).Error("Failed to report message build error")
				}

				return nil
			}

			if err := user.vault.RemFailedMessageID(event.ID); err != nil {
				user.log.WithError(err).Error("Failed to remove failed message ID from vault")
			}

			update = imap.NewMessagesCreated(false, res.update)
			user.updateCh[full.AddressID].Enqueue(update)

			return nil
		}); err != nil {
			return nil, err
		}

		return []imap.Update{update}, nil
	}, user.apiUserLock, user.apiAddrsLock, user.apiLabelsLock, user.updateChLock)
}

func (user *User) handleUpdateMessageEvent(ctx context.Context, event proton.MessageEvent) ([]imap.Update, error) { //nolint:unparam
	return safe.RLockRetErr(func() ([]imap.Update, error) {
		user.log.WithFields(logrus.Fields{
			"messageID": event.ID,
			"subject":   logging.Sensitive(event.Message.Subject),
		}).Info("Handling message updated event")

		update := imap.NewMessageMailboxesUpdated(
			imap.MessageID(event.ID),
			mapTo[string, imap.MailboxID](wantLabels(user.apiLabels, event.Message.LabelIDs)),
			event.Message.Seen(),
			event.Message.Starred(),
		)

		user.updateCh[event.Message.AddressID].Enqueue(update)

		return []imap.Update{update}, nil
	}, user.apiLabelsLock, user.updateChLock)
}

func (user *User) handleDeleteMessageEvent(ctx context.Context, event proton.MessageEvent) ([]imap.Update, error) { //nolint:unparam
	return safe.RLockRetErr(func() ([]imap.Update, error) {
		user.log.WithField("messageID", event.ID).Info("Handling message deleted event")

		var updates []imap.Update

		for _, updateCh := range xslices.Unique(maps.Values(user.updateCh)) {
			update := imap.NewMessagesDeleted(imap.MessageID(event.ID))
			updateCh.Enqueue(update)
			updates = append(updates, update)
		}

		return updates, nil
	}, user.updateChLock)
}

func (user *User) handleUpdateDraftEvent(ctx context.Context, event proton.MessageEvent) ([]imap.Update, error) { //nolint:unparam
	return safe.RLockRetErr(func() ([]imap.Update, error) {
		user.log.WithFields(logrus.Fields{
			"messageID": event.ID,
			"subject":   logging.Sensitive(event.Message.Subject),
		}).Info("Handling draft updated event")

		full, err := user.client.GetFullMessage(ctx, event.Message.ID, newProtonAPIScheduler(), proton.NewDefaultAttachmentAllocator())
		if err != nil {
			// If the message is not found, it means that it has been deleted before we could fetch it.
			if apiErr := new(proton.APIError); errors.As(err, &apiErr) && apiErr.Status == http.StatusUnprocessableEntity {
				user.log.WithField("messageID", event.Message.ID).Warn("Cannot add new draft: full message is missing on API")
				return nil, nil
			}

			return nil, fmt.Errorf("failed to get full draft: %w", err)
		}

		var update imap.Update

		if err := withAddrKR(user.apiUser, user.apiAddrs[event.Message.AddressID], user.vault.KeyPass(), func(_, addrKR *crypto.KeyRing) error {
			res := buildRFC822(user.apiLabels, full, addrKR, new(bytes.Buffer))

			if res.err != nil {
				logrus.WithError(err).Error("Failed to build RFC822 message")

				if err := user.vault.AddFailedMessageID(event.ID); err != nil {
					user.log.WithError(err).Error("Failed to add failed message ID to vault")
				}

				if err := user.reporter.ReportMessageWithContext("Failed to build draft message (event update)", reporter.Context{
					"messageID": res.messageID,
					"error":     res.err,
				}); err != nil {
					logrus.WithError(err).Error("Failed to report message build error")
				}

				return nil
			}

			if err := user.vault.RemFailedMessageID(event.ID); err != nil {
				user.log.WithError(err).Error("Failed to remove failed message ID from vault")
			}

			update = imap.NewMessageUpdated(
				res.update.Message,
				res.update.Literal,
				res.update.MailboxIDs,
				res.update.ParsedMessage,
			)

			user.updateCh[full.AddressID].Enqueue(update)

			return nil
		}); err != nil {
			return nil, err
		}

		return []imap.Update{update}, nil
	}, user.apiUserLock, user.apiAddrsLock, user.apiLabelsLock, user.updateChLock)
}

func (user *User) handleUsedSpaceChange(usedSpace int) {
	safe.Lock(func() {
		if user.apiUser.UsedSpace == usedSpace {
			return
		}

		user.apiUser.UsedSpace = usedSpace
		user.eventCh.Enqueue(events.UsedSpaceChanged{
			UserID:    user.apiUser.ID,
			UsedSpace: usedSpace,
		})
	}, user.apiUserLock)
}

func getMailboxName(label proton.Label) []string {
	var name []string

	switch label.Type {
	case proton.LabelTypeFolder:
		name = append([]string{folderPrefix}, label.Path...)

	case proton.LabelTypeLabel:
		name = append([]string{labelPrefix}, label.Path...)

	case proton.LabelTypeContactGroup:
		fallthrough
	case proton.LabelTypeSystem:
		fallthrough
	default:
		name = label.Path
	}

	return name
}

func waitOnIMAPUpdates(ctx context.Context, updates []imap.Update) error {
	for _, update := range updates {
		if err, ok := update.WaitContext(ctx); ok && err != nil {
			return fmt.Errorf("failed to apply gluon update %v: %w", update.String(), err)
		}
	}

	return nil
}
