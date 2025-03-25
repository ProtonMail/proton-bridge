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

package syncservice

import (
	"context"
	"fmt"
	"time"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/network"
	"github.com/sirupsen/logrus"
)

const DefaultRetryCoolDown = 20 * time.Second
const NumSyncStages = 4

type LabelMap = map[string]proton.Label

// Handler is the interface from which we control the syncing of the IMAP data. One instance should be created for each
// user and used for every subsequent sync request.
type Handler struct {
	regulator      Regulator
	client         APIClient
	userID         string
	syncState      StateProvider
	log            *logrus.Entry
	group          *async.Group
	syncFinishedCh chan error
	panicHandler   async.PanicHandler
	downloadCache  *DownloadCache
}

func NewHandler(
	regulator Regulator,
	client APIClient,
	userID string,
	state StateProvider,
	log *logrus.Entry,
	panicHandler async.PanicHandler,
) *Handler {
	return &Handler{
		client:         client,
		userID:         userID,
		syncState:      state,
		log:            log,
		syncFinishedCh: make(chan error),
		group:          async.NewGroup(context.Background(), panicHandler),
		regulator:      regulator,
		panicHandler:   panicHandler,
		downloadCache:  newDownloadCache(),
	}
}

func (t *Handler) Close() {
	t.group.CancelAndWait()
	close(t.syncFinishedCh)
}

func (t *Handler) CancelAndWait() {
	t.group.CancelAndWait()
}

func (t *Handler) Cancel() {
	t.group.Cancel()
}

func (t *Handler) OnSyncFinishedCH() <-chan error {
	return t.syncFinishedCh
}

func (t *Handler) Execute(
	syncReporter Reporter,
	labels LabelMap,
	updateApplier UpdateApplier,
	messageBuilder MessageBuilder,
	coolDown time.Duration,
) {
	t.log.Info("Sync triggered")
	t.group.Once(func(ctx context.Context) {
		start := time.Now()
		t.log.WithField("start", start).Info("Beginning user sync")

		syncReporter.OnStart(ctx)
		var err error
		for {
			if err = ctx.Err(); err != nil {
				t.log.WithError(err).Error("Sync aborted")
				break
			} else if err = t.run(ctx, syncReporter, labels, updateApplier, messageBuilder); err != nil {
				t.log.WithError(err).Error("Failed to sync, will retry later")
				sleepCtx(ctx, coolDown)
			} else {
				break
			}
		}

		if err != nil {
			syncReporter.OnError(ctx, err)
		} else {
			syncReporter.OnFinished(ctx)
		}

		t.log.WithField("duration", time.Since(start)).Info("Finished user sync")
		select {
		case <-ctx.Done():
			return
		case t.syncFinishedCh <- err:
		}
	})
}

func (t *Handler) run(ctx context.Context,
	syncReporter Reporter,
	labels LabelMap,
	updateApplier UpdateApplier,
	messageBuilder MessageBuilder,
) error {
	syncStatus, err := t.syncState.GetSyncStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to get sync status: %w", err)
	}

	if syncStatus.IsComplete() {
		t.log.Info("Sync already complete, updating labels")

		if err := updateApplier.SyncLabels(ctx, labels); err != nil {
			t.log.WithError(err).Error("Failed to sync labels")
			return err
		}

		return nil
	}

	if !syncStatus.HasLabels {
		t.log.Info("Syncing labels")
		if err := updateApplier.SyncLabels(ctx, labels); err != nil {
			return fmt.Errorf("failed to sync labels: %w", err)
		}

		if err := t.syncState.SetHasLabels(ctx, true); err != nil {
			return fmt.Errorf("failed to set has labels: %w", err)
		}

		t.log.Info("Synced labels")
	}

	if !syncStatus.HasMessageCount {
		wrapper := network.NewClientRetryWrapper(t.client, &network.ExpCoolDown{})

		messageCounts, err := network.RetryWithClient(ctx, wrapper, func(ctx context.Context, c APIClient) ([]proton.MessageGroupCount, error) {
			return c.GetGroupedMessageCount(ctx)
		})
		if err != nil {
			return fmt.Errorf("failed to retrieve message ids: %w", err)
		}

		var totalMessageCount int64

		for _, gc := range messageCounts {
			if gc.LabelID == proton.AllMailLabel {
				totalMessageCount = int64(gc.Total)
				break
			}
		}

		if err := t.syncState.SetMessageCount(ctx, totalMessageCount); err != nil {
			return fmt.Errorf("failed to store message count: %w", err)
		}

		syncStatus.TotalMessageCount = totalMessageCount
	}

	syncReporter.InitializeProgressCounter(ctx, syncStatus.NumSyncedMessages*NumSyncStages, syncStatus.TotalMessageCount*NumSyncStages)

	if !syncStatus.HasMessages {
		t.log.Info("Syncing messages")

		stageContext := NewJob(
			ctx,
			t.client,
			t.userID,
			labels,
			messageBuilder,
			updateApplier,
			syncReporter,
			t.syncState,
			t.panicHandler,
			t.downloadCache,
			t.log,
		)

		stageContext.metadataFetched = syncStatus.NumSyncedMessages
		stageContext.totalMessageCount = syncStatus.TotalMessageCount

		if err := t.regulator.Sync(ctx, stageContext); err != nil {
			stageContext.onError(err)
			_ = stageContext.waitAndClose(ctx)
			return fmt.Errorf("failed to start sync job: %w", err)
		}

		// Wait on reply
		if err := stageContext.waitAndClose(ctx); err != nil {
			return fmt.Errorf("failed sync messages: %w", err)
		}

		if err := t.syncState.SetHasMessages(ctx, true); err != nil {
			return fmt.Errorf("failed to set sync as completed: %w", err)
		}

		t.log.Info("Synced messages")
	} else {
		t.log.Info("Messages are already synced, skipping")
	}

	return nil
}
