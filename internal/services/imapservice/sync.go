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

package imapservice

import (
	"context"
	"fmt"
	"time"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/useridentity"
	"github.com/ProtonMail/proton-bridge/v3/internal/usertypes"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type updatePublisher interface {
	publishUpdate(ctx context.Context, update imap.Update)
}

type syncJob struct {
	client         APIClient
	userID         string
	labels         labelMap
	updaters       map[string]updatePublisher
	addressMode    usertypes.AddressMode
	syncState      SyncStateProvider
	eventPublisher events.EventPublisher
	log            *logrus.Entry
	identityState  *useridentity.State
	panicHandler   async.PanicHandler
	reporter       reporter.Reporter
	keyProvider    useridentity.KeyPassProvider
	maxSyncMemory  uint64
}

const SyncRetryCoolDown = 20 * time.Second

func (s *syncJob) run(ctx context.Context) error {
	s.log.Info("Sync triggered")
	s.eventPublisher.PublishEvent(ctx, events.SyncStarted{UserID: s.userID})

	if s.syncState.GetSyncStatus().IsComplete() {
		s.log.Info("Sync already complete, only system labels will be updated")

		if err := s.syncSystemLabels(ctx); err != nil {
			s.log.WithError(err).Error("Failed to sync system labels")
			s.eventPublisher.PublishEvent(ctx, events.SyncFailed{
				UserID: s.userID,
				Error:  err,
			})
			return err
		}
		s.eventPublisher.PublishEvent(ctx, events.SyncFinished{UserID: s.userID})
		return nil
	}

	for {
		if err := ctx.Err(); err != nil {
			s.log.WithError(err).Error("Sync aborted")
			return fmt.Errorf("sync aborted: %w", ctx.Err())
		} else if err := s.doSync(ctx); err != nil {
			s.log.WithError(err).Error("Failed to sync, will retry later")
			sleepCtx(ctx, SyncRetryCoolDown)
		} else {
			break
		}
	}

	return nil
}

func (s *syncJob) syncSystemLabels(ctx context.Context) error {
	var updates []imap.Update

	for _, label := range s.labels {
		if !WantLabel(label) {
			continue
		}

		for _, connector := range s.updaters {
			update := newSystemMailboxCreatedUpdate(imap.MailboxID(label.ID), label.Name)
			connector.publishUpdate(ctx, update)
			updates = append(updates, update)
		}
	}

	if err := waitOnIMAPUpdates(ctx, updates); err != nil {
		return fmt.Errorf("could not sync system labels: %w", err)
	}

	return nil
}

func (s *syncJob) doSync(ctx context.Context) error {
	start := time.Now()

	s.log.WithField("start", start).Info("Beginning user sync")

	s.eventPublisher.PublishEvent(ctx, events.SyncStarted{
		UserID: s.userID,
	})

	if err := s.sync(ctx); err != nil {
		s.log.WithError(err).Warn("Failed to sync user")

		s.eventPublisher.PublishEvent(ctx, events.SyncFailed{
			UserID: s.userID,
			Error:  err,
		})

		return fmt.Errorf("failed to sync: %w", err)
	}

	s.log.WithField("duration", time.Since(start)).Info("Finished user sync")

	s.eventPublisher.PublishEvent(ctx, events.SyncFinished{
		UserID: s.userID,
	})

	return nil
}

func (s *syncJob) sync(ctx context.Context) error {
	syncStatus := s.syncState.GetSyncStatus()

	if !syncStatus.HasLabels {
		s.log.Info("Syncing labels")

		if err := syncLabels(ctx, s.labels, maps.Values(s.updaters)...); err != nil {
			return fmt.Errorf("failed to sync labels: %w", err)
		}

		if err := s.syncState.SetHasLabels(true); err != nil {
			return fmt.Errorf("failed to set has labels: %w", err)
		}

		s.log.Info("Synced labels")
	}

	if !syncStatus.HasMessages {
		s.log.Info("Syncing messages")

		// Determine which messages to sync.
		messageIDs, err := s.client.GetAllMessageIDs(ctx, "")
		if err != nil {
			return fmt.Errorf("failed to get message IDs to sync: %w", err)
		}

		s.log.Debugf("User has the following failed synced message ids: %v", syncStatus.FailedMessageIDs)

		// Remove any messages that have already failed to sync.
		messageIDs = xslices.Filter(messageIDs, func(messageID string) bool {
			return !slices.Contains(syncStatus.FailedMessageIDs, messageID)
		})

		// Reverse the order of the message IDs so that the newest messages are synced first.
		xslices.Reverse(messageIDs)

		// If we have a message ID that we've already synced, then we can skip all messages before it.
		if idx := xslices.Index(messageIDs, syncStatus.LastMessageID); idx >= 0 {
			messageIDs = messageIDs[idx+1:]
		}

		// Sync the messages.
		if err := s.syncMessages(
			ctx,
			messageIDs,
		); err != nil {
			return fmt.Errorf("failed to sync messages: %w", err)
		}

		if err := s.syncState.SetHasMessages(true); err != nil {
			return fmt.Errorf("failed to set has messages: %w", err)
		}

		s.log.Info("Synced messages")
	} else {
		s.log.Info("Messages are already synced, skipping")
	}

	return nil
}
