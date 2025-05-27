// Copyright (c) 2025 Proton AG
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
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/logging"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
)

func (s *Service) HandleLabelEvents(ctx context.Context, events []proton.LabelEvent) error {
	s.log.Debug("handling label event")

	for _, event := range events {
		switch event.Action {
		case proton.EventCreate:
			if !WantLabel(event.Label) {
				continue
			}

			updates, err := onLabelCreated(ctx, s, event)
			if err != nil {
				return fmt.Errorf("failed to handle create label event: %w", err)
			}

			if err := waitOnIMAPUpdates(ctx, updates); err != nil {
				return err
			}

		case proton.EventUpdateFlags, proton.EventUpdate:
			if !WantLabel(event.Label) {
				continue
			}

			updates, err := onLabelUpdated(ctx, s, event)
			if err != nil {
				return fmt.Errorf("failed to handle update label event: %w", err)
			}

			if err := waitOnIMAPUpdates(ctx, updates); err != nil {
				return err
			}

		case proton.EventDelete:
			updates := onLabelDeleted(ctx, s, event)

			if err := waitOnIMAPUpdates(ctx, updates); err != nil {
				return fmt.Errorf("failed to handle delete label event in gluon: %w", err)
			}
		}
	}

	return nil
}

func onLabelCreated(ctx context.Context, s *Service, event proton.LabelEvent) ([]imap.Update, error) {
	updates := []imap.Update{}

	s.log.WithFields(logrus.Fields{
		"labelID": event.ID,
		"name":    logging.Sensitive(event.Label.Name),
	}).Info("Handling label created event")

	wr := s.labels.Write()
	defer wr.Close()

	wr.SetLabel(event.Label.ID, event.Label, "onLabelCreated")

	labelConflictResolver := s.labelConflictManager.NewConflictResolver(maps.Values(s.connectors))
	conflictUpdatesGenerator, err := labelConflictResolver.ResolveConflict(ctx, event.Label, make(map[string]bool))
	if err != nil {
		return updates, err
	}

	for _, updateCh := range maps.Values(s.connectors) {
		conflictUpdates := conflictUpdatesGenerator()
		updateCh.publishUpdate(ctx, conflictUpdates...)
		updates = append(updates, conflictUpdates...)

		update := newMailboxCreatedUpdate(imap.MailboxID(event.ID), GetMailboxName(event.Label))
		updateCh.publishUpdate(ctx, update)
		updates = append(updates, update)
	}

	s.eventPublisher.PublishEvent(ctx, events.UserLabelCreated{
		UserID:  s.identityState.UserID(),
		LabelID: event.Label.ID,
		Name:    event.Label.Name,
	})

	return updates, nil
}

func onLabelUpdated(ctx context.Context, s *Service, event proton.LabelEvent) ([]imap.Update, error) {
	var updates []imap.Update

	s.log.WithFields(logrus.Fields{
		"labelID": event.ID,
		"name":    logging.Sensitive(event.Label.Name),
	}).Info("Handling label updated event")

	stack := []proton.Label{event.Label}

	wr := s.labels.Write()
	defer wr.Close()

	for len(stack) > 0 {
		label := stack[0]
		stack = stack[1:]

		// Only update the label if it exists; we don't want to create it as a client may have just deleted it.
		if _, ok := wr.GetLabel(label.ID); ok {
			wr.SetLabel(label.ID, event.Label, "onLabelUpdatedLabelEventID")
		}

		// API doesn't notify us that the path has changed. We need to fetch it again.
		apiLabel, err := s.client.GetLabel(ctx, label.ID, label.Type)
		if apiErr := new(proton.APIError); errors.As(err, &apiErr) && apiErr.Status == http.StatusUnprocessableEntity {
			s.log.WithError(apiErr).Warn("Failed to get label: label does not exist")
			continue
		} else if err != nil {
			return nil, fmt.Errorf("failed to get label %q: %w", label.ID, err)
		}

		// Update the label in the map.
		wr.SetLabel(apiLabel.ID, apiLabel, "onLabelUpdatedApiID")

		// Resolve potential conflicts
		labelConflictResolver := s.labelConflictManager.NewConflictResolver(maps.Values(s.connectors))
		conflictUpdatesGenerator, err := labelConflictResolver.ResolveConflict(ctx, event.Label, make(map[string]bool))
		if err != nil {
			return updates, err
		}

		// Notify the IMAP clients.
		for _, updateCh := range maps.Values(s.connectors) {
			conflictUpdates := conflictUpdatesGenerator()
			updateCh.publishUpdate(ctx, conflictUpdates...)
			updates = append(updates, conflictUpdates...)

			update := imap.NewMailboxUpdated(
				imap.MailboxID(apiLabel.ID),
				GetMailboxName(apiLabel),
			)
			updateCh.publishUpdate(ctx, update)
			updates = append(updates, update)
		}

		s.eventPublisher.PublishEvent(ctx, events.UserLabelUpdated{
			UserID:  s.identityState.UserID(),
			LabelID: apiLabel.ID,
			Name:    apiLabel.Name,
		})

		children := xslices.Filter(wr.GetLabels(), func(other proton.Label) bool {
			return other.ParentID == label.ID
		})

		stack = append(stack, children...)
	}

	return updates, nil
}

func onLabelDeleted(ctx context.Context, s *Service, event proton.LabelEvent) []imap.Update {
	updates := make([]imap.Update, 0, len(s.connectors))

	s.log.WithField("labelID", event.ID).Info("Handling label deleted event")

	for _, updateCh := range maps.Values(s.connectors) {
		update := imap.NewMailboxDeleted(imap.MailboxID(event.ID))
		updateCh.publishUpdate(ctx, update)
		updates = append(updates, update)
	}

	wr := s.labels.Write()
	wr.Close()

	wr.Delete(event.ID, "onLabelDeleted")

	s.eventPublisher.PublishEvent(ctx, events.UserLabelDeleted{
		UserID:  s.identityState.UserID(),
		LabelID: event.ID,
	})

	return updates
}
