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

package imapservice

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ProtonMail/gluon"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/internal/logging"
	obsMetrics "github.com/ProtonMail/proton-bridge/v3/internal/services/imapservice/observabilitymetrics/evtloopmsgevents"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/observability"
	"github.com/ProtonMail/proton-bridge/v3/internal/usertypes"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
)

func (s *Service) HandleMessageEvents(ctx context.Context, events []proton.MessageEvent) error {
	s.log.Debug("handling message event")

	for _, event := range events {
		ctx = logging.WithLogrusField(ctx, "messageID", event.ID)

		switch event.Action {
		case proton.EventCreate:
			updates, err := onMessageCreated(logging.WithLogrusField(ctx, "action", "create message"), s, event.Message, false)
			if err != nil {
				s.observabilitySender.AddDistinctMetrics(observability.EventLoopError, obsMetrics.GenerateMessageEventFailureCreateMessageMetric())
				return fmt.Errorf("failed to handle create message event: %w", err)
			}

			if err := waitOnIMAPUpdates(ctx, updates); err != nil {
				return err
			}

		case proton.EventUpdate, proton.EventUpdateFlags:
			// Draft update means to completely remove old message and upload the new data again, but we should
			// only do this if the event is of type EventUpdate otherwise label switch operations will not work.
			if (event.Message.IsDraft() || (event.Message.Flags&proton.MessageFlagSent != 0)) && event.Action == proton.EventUpdate {
				updates, err := onMessageUpdateDraftOrSent(
					logging.WithLogrusField(ctx, "action", "update draft or sent message"),
					s,
					event,
				)
				if err != nil {
					s.observabilitySender.AddDistinctMetrics(observability.EventLoopError, obsMetrics.GenerateMessageEventFailureUpdateMetric())
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
			updates, err := onMessageUpdate(
				logging.WithLogrusField(ctx, "action", "update message"),
				s,
				event.Message,
			)
			if err != nil {
				s.observabilitySender.AddDistinctMetrics(observability.EventLoopError, obsMetrics.GenerateMessageEventFailureUpdateMetric())
				return fmt.Errorf("failed to handle update message event: %w", err)
			}

			// If the update fails on the gluon side because it doesn't exist, we try to create the message instead.
			if err := waitOnIMAPUpdates(ctx, updates); gluon.IsNoSuchMessage(err) {
				s.log.WithError(err).Error("Failed to handle update message event in gluon, will try creating it")

				updates, err := onMessageCreated(ctx, s, event.Message, false)
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
			updates := onMessageDeleted(
				logging.WithLogrusField(ctx, "action", "delete message"),
				s,
				event,
			)

			if err := waitOnIMAPUpdates(ctx, updates); err != nil {
				s.observabilitySender.AddDistinctMetrics(observability.EventLoopError, obsMetrics.GenerateMessageEventFailureDeleteMessageMetric())
				return fmt.Errorf("failed to handle delete message event in gluon: %w", err)
			}
		}
	}

	return nil
}

func onMessageCreated(
	ctx context.Context,
	s *Service,
	message proton.MessageMetadata,
	allowUnknownLabels bool,
) ([]imap.Update, error) {
	s.log.WithFields(logrus.Fields{
		"messageID": message.ID,
		"subject":   logging.Sensitive(message.Subject),
		"date":      message.Time,
	}).Info("Handling message created event")

	full, err := s.client.GetFullMessage(ctx, message.ID, usertypes.NewProtonAPIScheduler(s.panicHandler), proton.NewDefaultAttachmentAllocator())
	if err != nil {
		// If the message is not found, it means that it has been deleted before we could fetch it.
		if apiErr := new(proton.APIError); errors.As(err, &apiErr) && apiErr.Status == http.StatusUnprocessableEntity {
			s.log.WithField("messageID", message.ID).Warn("Cannot create new message: full message is missing on API")
			return nil, nil
		}

		return nil, fmt.Errorf("failed to get full message: %w", err)
	}

	var update imap.Update

	apiLabels := s.labels.GetLabelMap()

	if err := s.identityState.WithAddrKR(message.AddressID, func(_, addrKR *crypto.KeyRing) error {
		res := buildRFC822(apiLabels, full, addrKR, new(bytes.Buffer))

		if res.err != nil {
			s.log.WithError(err).Error("Failed to build RFC822 message")

			if err := s.syncStateProvider.AddFailedMessageID(ctx, message.ID); err != nil {
				s.log.WithError(err).Error("Failed to add failed message ID to vault")
			}

			s.observabilitySender.AddDistinctMetrics(observability.EventLoopError, obsMetrics.GenerateMessageEventFailedToBuildMessage())
			return nil
		}

		if err := s.syncStateProvider.RemFailedMessageID(ctx, message.ID); err != nil {
			s.log.WithError(err).Error("Failed to remove failed message ID from vault")
		}

		update = imap.NewMessagesCreated(allowUnknownLabels, res.update)
		didPublish, err := safePublishMessageUpdate(ctx, s, full.AddressID, update)
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
}

func onMessageUpdateDraftOrSent(ctx context.Context, s *Service, event proton.MessageEvent) ([]imap.Update, error) {
	s.log.WithFields(logrus.Fields{
		"messageID": event.ID,
		"subject":   logging.Sensitive(event.Message.Subject),
		"isDraft":   event.Message.IsDraft(),
	}).Info("Handling draft or sent updated event")

	full, err := s.client.GetFullMessage(ctx, event.Message.ID, usertypes.NewProtonAPIScheduler(s.panicHandler), proton.NewDefaultAttachmentAllocator())
	if err != nil {
		// If the message is not found, it means that it has been deleted before we could fetch it.
		if apiErr := new(proton.APIError); errors.As(err, &apiErr) && apiErr.Status == http.StatusUnprocessableEntity {
			s.log.WithField("messageID", event.Message.ID).Warn("Cannot update message: full message is missing on API")
			return nil, nil
		}

		return nil, fmt.Errorf("failed to get full draft: %w", err)
	}

	var update imap.Update

	apiLabels := s.labels.GetLabelMap()

	if err := s.identityState.WithAddrKR(event.Message.AddressID, func(_, addrKR *crypto.KeyRing) error {
		res := buildRFC822(apiLabels, full, addrKR, new(bytes.Buffer))

		if res.err != nil {
			logrus.WithError(err).Error("Failed to build RFC822 message")

			if err := s.syncStateProvider.AddFailedMessageID(ctx, event.ID); err != nil {
				s.log.WithError(err).Error("Failed to add failed message ID to vault")
			}

			s.observabilitySender.AddDistinctMetrics(observability.EventLoopError, obsMetrics.GenerateMessageEventFailedToBuildDraft())
			return nil
		}

		if err := s.syncStateProvider.RemFailedMessageID(ctx, event.ID); err != nil {
			s.log.WithError(err).Error("Failed to remove failed message ID from vault")
		}

		update = imap.NewMessageUpdated(
			res.update.Message,
			res.update.Literal,
			res.update.MailboxIDs,
			res.update.ParsedMessage,
			true, // Is the message doesn't exist, silently create it.
		)

		didPublish, err := safePublishMessageUpdate(ctx, s, full.AddressID, update)
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
}

func onMessageUpdate(ctx context.Context, s *Service, message proton.MessageMetadata) ([]imap.Update, error) {
	s.log.WithFields(logrus.Fields{
		"messageID": message.ID,
		"subject":   logging.Sensitive(message.Subject),
	}).Info("Handling message updated event")

	flags := BuildFlagSetFromMessageMetadata(message)

	update := imap.NewMessageMailboxesUpdated(
		imap.MessageID(message.ID),
		usertypes.MapTo[string, imap.MailboxID](wantLabels(s.labels.GetLabelMap(), message.LabelIDs)),
		flags,
	)

	didPublish, err := safePublishMessageUpdate(ctx, s, message.AddressID, update)
	if err != nil {
		return nil, err
	}

	if !didPublish {
		return nil, nil
	}

	return []imap.Update{update}, nil
}

func onMessageDeleted(ctx context.Context, s *Service, event proton.MessageEvent) []imap.Update {
	s.log.WithField("messageID", event.ID).Info("Handling message deleted event")

	updates := make([]imap.Update, 0, len(s.connectors))

	for _, updateCh := range maps.Values(s.connectors) {
		update := imap.NewMessagesDeleted(imap.MessageID(event.ID))
		updateCh.publishUpdate(ctx, update)
		updates = append(updates, update)
	}

	return updates
}

// safePublishMessageUpdate handles the rare case where the address' update channel may have been deleted in the same
// event. This rare case can take place if in the same event fetch request there is an update for delete address and
// create/update message.
// If the user is in combined mode, we simply push the update to the primary address. If the user is in split mode
// we do not publish the update as the address no longer exists.
func safePublishMessageUpdate(ctx context.Context, s *Service, addressID string, update imap.Update) (bool, error) {
	v, ok := s.connectors[addressID]
	if !ok {
		if s.addressMode == usertypes.AddressModeCombined {
			primAddr, err := s.identityState.GetPrimaryAddress()
			if err != nil {
				return false, fmt.Errorf("failed to get primary address: %w", err)
			}
			primaryCh, ok := s.connectors[primAddr.ID]
			if !ok {
				return false, fmt.Errorf("primary address channel is not available")
			}

			primaryCh.publishUpdate(ctx, update)

			return true, nil
		}

		logrus.Warnf("Update channel not found for address %v, it may have been already deleted", addressID)
		s.observabilitySender.AddDistinctMetrics(observability.EventLoopError, obsMetrics.GenerateMessageEventUpdateChannelDoesNotExist())

		return false, nil
	}

	v.publishUpdate(ctx, update)

	return true, nil
}
