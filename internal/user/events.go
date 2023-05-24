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

	"github.com/ProtonMail/gluon"
	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/internal"
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

	// Abort the event stream
	defer user.pollAbort.Abort()

	// Re-sync messages after the user, address and label refresh.
	defer user.goSync()

	return user.syncUserAddressesLabelsAndClearSync(ctx, false)
}

func (user *User) syncUserAddressesLabelsAndClearSync(ctx context.Context, cancelEventPool bool) error {
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
			UserID:          user.apiUser.ID,
			CancelEventPool: cancelEventPool,
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
				user.reportError("Failed to apply address create event", err)
				return fmt.Errorf("failed to handle create address event: %w", err)
			}

		case proton.EventUpdate, proton.EventUpdateFlags:
			if err := user.handleUpdateAddressEvent(ctx, event); err != nil {
				if errors.Is(err, ErrAddressDoesNotExist) {
					logrus.Debugf("Address %v does not exist, will try create instead", event.Address.ID)
					if createErr := user.handleCreateAddressEvent(ctx, event); createErr != nil {
						user.reportError("Failed to apply address update event (with create)", createErr)
						return fmt.Errorf("failed to handle update address event (with create): %w", createErr)
					}

					return nil
				}

				user.reportError("Failed to apply address update event", err)
				return fmt.Errorf("failed to handle update address event: %w", err)
			}

		case proton.EventDelete:
			if err := user.handleDeleteAddressEvent(ctx, event); err != nil {
				user.reportError("Failed to apply address delete event", err)
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

		// If the address is disabled.
		if event.Address.Status != proton.AddressStatusEnabled {
			return nil
		}

		// If the address is enabled, we need to hook it up to the update channels.
		switch user.vault.AddressMode() {
		case vault.CombinedMode:
			primAddr, err := getPrimaryAddr(user.apiAddrs)
			if err != nil {
				return fmt.Errorf("failed to get primary address: %w", err)
			}

			user.updateCh[event.Address.ID] = user.updateCh[primAddr.ID]

		case vault.SplitMode:
			user.updateCh[event.Address.ID] = async.NewQueuedChannel[imap.Update](0, 0, user.panicHandler)
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
		if event.Address.Status != proton.AddressStatusEnabled {
			return nil
		}

		if user.vault.AddressMode() == vault.SplitMode {
			if err := syncLabels(ctx, user.apiLabels, user.updateCh[event.Address.ID]); err != nil {
				return fmt.Errorf("failed to sync labels to new address: %w", err)
			}
		}

		return nil
	}, user.apiAddrsLock, user.apiLabelsLock, user.updateChLock)
}

var ErrAddressDoesNotExist = errors.New("address does not exist")

func (user *User) handleUpdateAddressEvent(_ context.Context, event proton.AddressEvent) error { //nolint:unparam
	return safe.LockRet(func() error {
		user.log.WithFields(logrus.Fields{
			"addressID": event.ID,
			"email":     logging.Sensitive(event.Address.Email),
		}).Info("Handling address updated event")

		oldAddr, ok := user.apiAddrs[event.Address.ID]
		if !ok {
			return ErrAddressDoesNotExist
		}

		user.apiAddrs[event.Address.ID] = event.Address

		switch {
		// If the address was newly enabled:
		case oldAddr.Status != proton.AddressStatusEnabled && event.Address.Status == proton.AddressStatusEnabled:
			switch user.vault.AddressMode() {
			case vault.CombinedMode:
				primAddr, err := getPrimaryAddr(user.apiAddrs)
				if err != nil {
					return fmt.Errorf("failed to get primary address: %w", err)
				}

				user.updateCh[event.Address.ID] = user.updateCh[primAddr.ID]

			case vault.SplitMode:
				user.updateCh[event.Address.ID] = async.NewQueuedChannel[imap.Update](0, 0, user.panicHandler)
			}

			user.eventCh.Enqueue(events.UserAddressEnabled{
				UserID:    user.apiUser.ID,
				AddressID: event.Address.ID,
				Email:     event.Address.Email,
			})

		// If the address was newly disabled:
		case oldAddr.Status == proton.AddressStatusEnabled && event.Address.Status != proton.AddressStatusEnabled:
			if user.vault.AddressMode() == vault.SplitMode {
				user.updateCh[event.ID].CloseAndDiscardQueued()
			}

			delete(user.updateCh, event.ID)

			user.eventCh.Enqueue(events.UserAddressDisabled{
				UserID:    user.apiUser.ID,
				AddressID: event.Address.ID,
				Email:     event.Address.Email,
			})

		// Otherwise it's just an update:
		default:
			user.eventCh.Enqueue(events.UserAddressUpdated{
				UserID:    user.apiUser.ID,
				AddressID: event.Address.ID,
				Email:     event.Address.Email,
			})
		}

		return nil
	}, user.apiAddrsLock, user.updateChLock)
}

func (user *User) handleDeleteAddressEvent(_ context.Context, event proton.AddressEvent) error {
	return safe.LockRet(func() error {
		user.log.WithField("addressID", event.ID).Info("Handling address deleted event")

		addr, ok := user.apiAddrs[event.ID]
		if !ok {
			user.log.Debugf("Address %q does not exist", event.ID)
			return nil
		}

		delete(user.apiAddrs, event.ID)

		// If the address was disabled to begin with, we don't need to do anything.
		if addr.Status != proton.AddressStatusEnabled {
			return nil
		}

		// Otherwise, in split mode, drop the update queue.
		if user.vault.AddressMode() == vault.SplitMode {
			user.updateCh[event.ID].CloseAndDiscardQueued()
		}

		// And in either mode, remove the address from the update channel map.
		delete(user.updateCh, event.ID)

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

func (user *User) handleCreateLabelEvent(_ context.Context, event proton.LabelEvent) ([]imap.Update, error) { //nolint:unparam
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

		stack := []proton.Label{event.Label}

		for len(stack) > 0 {
			label := stack[0]
			stack = stack[1:]

			// Only update the label if it exists; we don't want to create it as a client may have just deleted it.
			if _, ok := user.apiLabels[label.ID]; ok {
				user.apiLabels[label.ID] = event.Label
			}

			// API doesn't notify us that the path has changed. We need to fetch it again.
			apiLabel, err := user.client.GetLabel(ctx, label.ID, label.Type)
			if apiErr := new(proton.APIError); errors.As(err, &apiErr) && apiErr.Status == http.StatusUnprocessableEntity {
				user.log.WithError(apiErr).Warn("Failed to get label: label does not exist")
				continue
			} else if err != nil {
				return nil, fmt.Errorf("failed to get label %q: %w", label.ID, err)
			}

			// Update the label in the map.
			user.apiLabels[apiLabel.ID] = apiLabel

			// Notify the IMAP clients.
			for _, updateCh := range xslices.Unique(maps.Values(user.updateCh)) {
				update := imap.NewMailboxUpdated(
					imap.MailboxID(apiLabel.ID),
					getMailboxName(apiLabel),
				)
				updateCh.Enqueue(update)
				updates = append(updates, update)
			}

			user.eventCh.Enqueue(events.UserLabelUpdated{
				UserID:  user.apiUser.ID,
				LabelID: apiLabel.ID,
				Name:    apiLabel.Name,
			})

			children := xslices.Filter(maps.Values(user.apiLabels), func(other proton.Label) bool {
				return other.ParentID == label.ID
			})

			stack = append(stack, children...)
		}

		return updates, nil
	}, user.apiLabelsLock, user.updateChLock)
}

func (user *User) handleDeleteLabelEvent(_ context.Context, event proton.LabelEvent) ([]imap.Update, error) { //nolint:unparam
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
			updates, err := user.handleCreateMessageEvent(logging.WithLogrusField(ctx, "action", "create message"), event.Message)
			if err != nil {
				user.reportError("Failed to apply create message event", err)
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
					user.reportError("Failed to apply update draft message event", err)
					return fmt.Errorf("failed to handle update draft event: %w", err)
				}

				if err := waitOnIMAPUpdates(ctx, updates); err != nil {
					return err
				}

				continue
			}

			// GODT-2028 - Use better events here. It should be possible to have 3 separate events that refrain to
			// whether the flags, labels or read only data (header+body) has been changed. This requires fixing proton
			// first so that it correctly reports those cases.
			// Issue regular update to handle mailboxes and flag changes.
			updates, err := user.handleUpdateMessageEvent(
				logging.WithLogrusField(ctx, "action", "update message"),
				event.Message,
			)
			if err != nil {
				user.reportError("Failed to apply update message event", err)
				return fmt.Errorf("failed to handle update message event: %w", err)
			}

			// If the update fails on the gluon side because it doesn't exist, we try to create the message instead.
			if err := waitOnIMAPUpdates(ctx, updates); gluon.IsNoSuchMessage(err) {
				user.log.WithError(err).Error("Failed to handle update message event in gluon, will try creating it")

				updates, err := user.handleCreateMessageEvent(ctx, event.Message)
				if err != nil {
					return fmt.Errorf("failed to handle update message event as create: %w", err)
				}

				if err := waitOnIMAPUpdates(ctx, updates); err != nil {
					return err
				}
			} else if err != nil {
				return err
			}

		case proton.EventDelete:
			updates, err := user.handleDeleteMessageEvent(
				logging.WithLogrusField(ctx, "action", "delete message"),
				event,
			)
			if err != nil {
				user.reportError("Failed to apply delete message event", err)
				return fmt.Errorf("failed to handle delete message event: %w", err)
			}

			if err := waitOnIMAPUpdates(ctx, updates); err != nil {
				return fmt.Errorf("failed to handle delete message event in gluon: %w", err)
			}
		}
	}

	return nil
}

func (user *User) handleCreateMessageEvent(ctx context.Context, message proton.MessageMetadata) ([]imap.Update, error) {
	user.log.WithFields(logrus.Fields{
		"messageID": message.ID,
		"subject":   logging.Sensitive(message.Subject),
	}).Info("Handling message created event")

	full, err := user.client.GetFullMessage(ctx, message.ID, newProtonAPIScheduler(user.panicHandler), proton.NewDefaultAttachmentAllocator())
	if err != nil {
		// If the message is not found, it means that it has been deleted before we could fetch it.
		if apiErr := new(proton.APIError); errors.As(err, &apiErr) && apiErr.Status == http.StatusUnprocessableEntity {
			user.log.WithField("messageID", message.ID).Warn("Cannot create new message: full message is missing on API")
			return nil, nil
		}

		return nil, fmt.Errorf("failed to get full message: %w", err)
	}

	return safe.RLockRetErr(func() ([]imap.Update, error) {
		var update imap.Update

		if err := withAddrKR(user.apiUser, user.apiAddrs[message.AddressID], user.vault.KeyPass(), func(_, addrKR *crypto.KeyRing) error {
			res := buildRFC822(user.apiLabels, full, addrKR, new(bytes.Buffer))

			if res.err != nil {
				user.log.WithError(err).Error("Failed to build RFC822 message")

				if err := user.vault.AddFailedMessageID(message.ID); err != nil {
					user.log.WithError(err).Error("Failed to add failed message ID to vault")
				}

				user.reportErrorAndMessageID("Failed to build message (event create)", res.err, res.messageID)

				return nil
			}

			if err := user.vault.RemFailedMessageID(message.ID); err != nil {
				user.log.WithError(err).Error("Failed to remove failed message ID from vault")
			}

			update = imap.NewMessagesCreated(false, res.update)
			didPublish, err := safePublishMessageUpdate(user, full.AddressID, update)
			if err != nil {
				return err
			}

			if !didPublish {
				update = nil
			}

			return nil
		}); err != nil {
			return nil, err
		}

		if update == nil {
			return nil, nil
		}

		return []imap.Update{update}, nil
	}, user.apiUserLock, user.apiAddrsLock, user.apiLabelsLock, user.updateChLock)
}

func (user *User) handleUpdateMessageEvent(_ context.Context, message proton.MessageMetadata) ([]imap.Update, error) { //nolint:unparam
	return safe.RLockRetErr(func() ([]imap.Update, error) {
		user.log.WithFields(logrus.Fields{
			"messageID": message.ID,
			"subject":   logging.Sensitive(message.Subject),
		}).Info("Handling message updated event")

		flags := imap.NewFlagSet()

		if message.Seen() {
			flags.AddToSelf(imap.FlagSeen)
		}

		if message.Starred() {
			flags.AddToSelf(imap.FlagFlagged)
		}

		if message.IsDraft() {
			flags.AddToSelf(imap.FlagDraft)
		}

		if message.IsRepliedAll == true || message.IsReplied == true { //nolint: gosimple
			flags.AddToSelf(imap.FlagAnswered)
		}

		update := imap.NewMessageMailboxesUpdated(
			imap.MessageID(message.ID),
			mapTo[string, imap.MailboxID](wantLabels(user.apiLabels, message.LabelIDs)),
			flags,
		)

		didPublish, err := safePublishMessageUpdate(user, message.AddressID, update)
		if err != nil {
			return nil, err
		}

		if !didPublish {
			return nil, nil
		}

		return []imap.Update{update}, nil
	}, user.apiLabelsLock, user.updateChLock)
}

func (user *User) handleDeleteMessageEvent(_ context.Context, event proton.MessageEvent) ([]imap.Update, error) {
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

func (user *User) handleUpdateDraftEvent(ctx context.Context, event proton.MessageEvent) ([]imap.Update, error) {
	return safe.RLockRetErr(func() ([]imap.Update, error) {
		user.log.WithFields(logrus.Fields{
			"messageID": event.ID,
			"subject":   logging.Sensitive(event.Message.Subject),
		}).Info("Handling draft updated event")

		full, err := user.client.GetFullMessage(ctx, event.Message.ID, newProtonAPIScheduler(user.panicHandler), proton.NewDefaultAttachmentAllocator())
		if err != nil {
			// If the message is not found, it means that it has been deleted before we could fetch it.
			if apiErr := new(proton.APIError); errors.As(err, &apiErr) && apiErr.Status == http.StatusUnprocessableEntity {
				user.log.WithField("messageID", event.Message.ID).Warn("Cannot update draft: full message is missing on API")
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

				user.reportErrorAndMessageID("Failed to build draft message (event update)", res.err, res.messageID)

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
				true, // Is the message doesn't exist, silently create it.
			)

			didPublish, err := safePublishMessageUpdate(user, full.AddressID, update)
			if err != nil {
				return err
			}

			if !didPublish {
				update = nil
			}

			return nil
		}); err != nil {
			return nil, err
		}

		if update == nil {
			return nil, nil
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

func (user *User) reportError(title string, err error) {
	user.reportErrorNoContextCancel(title, err, reporter.Context{})
}

func (user *User) reportErrorAndMessageID(title string, err error, messgeID string) {
	user.reportErrorNoContextCancel(title, err, reporter.Context{"messageID": messgeID})
}

func (user *User) reportErrorNoContextCancel(title string, err error, reportContext reporter.Context) {
	if !errors.Is(err, context.Canceled) {
		reportContext["error"] = err
		reportContext["error_type"] = internal.ErrCauseType(err)
		if rerr := user.reporter.ReportMessageWithContext(title, reportContext); rerr != nil {
			user.log.WithError(err).WithField("title", title).Error("Failed to report message")
		}
	}
}

// safePublishMessageUpdate handles the rare case where the address' update channel may have been deleted in the same
// event. This rare case can take place if in the same event fetch request there is an update for delete address and
// create/update message.
// If the user is in combined mode, we simply push the update to the primary address. If the user is in split mode
// we do not publish the update as the address no longer exists.
func safePublishMessageUpdate(user *User, addressID string, update imap.Update) (bool, error) {
	v, ok := user.updateCh[addressID]
	if !ok {
		if user.GetAddressMode() == vault.CombinedMode {
			primAddr, err := getPrimaryAddr(user.apiAddrs)
			if err != nil {
				return false, fmt.Errorf("failed to get primary address: %w", err)
			}
			primaryCh, ok := user.updateCh[primAddr.ID]
			if !ok {
				return false, fmt.Errorf("primary address channel is not available")
			}

			primaryCh.Enqueue(update)

			return true, nil
		}

		logrus.Warnf("Update channel not found for address %v, it may have been already deleted", addressID)
		_ = user.reporter.ReportMessage("Message Update channel does not exist")

		return false, nil
	}

	v.Enqueue(update)

	return true, nil
}
