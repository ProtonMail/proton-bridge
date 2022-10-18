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
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/queue"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
	"github.com/bradenaw/juniper/stream"
	"github.com/bradenaw/juniper/xslices"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gitlab.protontech.ch/go/liteapi"
)

const (
	maxUpdateSize = 1 << 25
	maxBatchSize  = 1 << 8
)

func (user *User) sync(ctx context.Context) error {
	return user.withAddrKRs(func(addrKRs map[string]*crypto.KeyRing) error {
		logrus.Info("Beginning sync")

		if !user.vault.SyncStatus().HasLabels {
			logrus.Info("Syncing labels")

			if err := user.updateCh.ValuesErr(func(updateCh []*queue.QueuedChannel[imap.Update]) error {
				return syncLabels(ctx, user.client, xslices.Unique(updateCh)...)
			}); err != nil {
				return fmt.Errorf("failed to sync labels: %w", err)
			}

			if err := user.vault.SetHasLabels(true); err != nil {
				return fmt.Errorf("failed to set has labels: %w", err)
			}
		} else {
			logrus.Info("Labels are already synced, skipping")
		}

		if !user.vault.SyncStatus().HasMessages {
			logrus.Info("Syncing labels")

			if err := user.updateCh.MapErr(func(updateCh map[string]*queue.QueuedChannel[imap.Update]) error {
				return syncMessages(ctx, user.ID(), user.client, user.vault, addrKRs, updateCh, user.eventCh)
			}); err != nil {
				return fmt.Errorf("failed to sync messages: %w", err)
			}

			if err := user.vault.SetHasMessages(true); err != nil {
				return fmt.Errorf("failed to set has messages: %w", err)
			}
		} else {
			logrus.Info("Messages are already synced, skipping")
		}

		return nil
	})
}

func syncLabels(ctx context.Context, client *liteapi.Client, updateCh ...*queue.QueuedChannel[imap.Update]) error {
	// Sync the system folders.
	system, err := client.GetLabels(ctx, liteapi.LabelTypeSystem)
	if err != nil {
		return fmt.Errorf("failed to get system labels: %w", err)
	}

	for _, label := range xslices.Filter(system, func(label liteapi.Label) bool { return wantLabelID(label.ID) }) {
		for _, updateCh := range updateCh {
			updateCh.Enqueue(newSystemMailboxCreatedUpdate(imap.LabelID(label.ID), label.Name))
		}
	}

	// Create Folders/Labels mailboxes with a random ID and with the \Noselect attribute.
	for _, prefix := range []string{folderPrefix, labelPrefix} {
		for _, updateCh := range updateCh {
			updateCh.Enqueue(newPlaceHolderMailboxCreatedUpdate(prefix))
		}
	}

	// Sync the API folders.
	folders, err := client.GetLabels(ctx, liteapi.LabelTypeFolder)
	if err != nil {
		return fmt.Errorf("failed to get folders: %w", err)
	}

	for _, folder := range folders {
		for _, updateCh := range updateCh {
			updateCh.Enqueue(newMailboxCreatedUpdate(imap.LabelID(folder.ID), getMailboxName(folder)))
		}
	}

	// Sync the API labels.
	labels, err := client.GetLabels(ctx, liteapi.LabelTypeLabel)
	if err != nil {
		return fmt.Errorf("failed to get labels: %w", err)
	}

	for _, label := range labels {
		for _, updateCh := range updateCh {
			updateCh.Enqueue(newMailboxCreatedUpdate(imap.LabelID(label.ID), getMailboxName(label)))
		}
	}

	// Wait for all label updates to be applied.
	for _, updateCh := range updateCh {
		update := imap.NewNoop()
		defer update.WaitContext(ctx)

		updateCh.Enqueue(update)
	}

	return nil
}

func syncMessages( //nolint:funlen
	ctx context.Context,
	userID string,
	client *liteapi.Client,
	vault *vault.User,
	addrKRs map[string]*crypto.KeyRing,
	updateCh map[string]*queue.QueuedChannel[imap.Update],
	eventCh *queue.QueuedChannel[events.Event],
) error {
	// Determine which messages to sync.
	metadata, err := client.GetAllMessageMetadata(ctx, nil)
	if err != nil {
		return fmt.Errorf("get all message metadata: %w", err)
	}

	// Get the message IDs to sync.
	messageIDs := xslices.Map(metadata, func(metadata liteapi.MessageMetadata) string {
		return metadata.ID
	})

	// If possible, begin syncing from one beyond the last synced message.
	if idx := xslices.Index(messageIDs, vault.SyncStatus().LastMessageID); idx >= 0 {
		messageIDs = messageIDs[idx+1:]
	}

	// Fetch and build each message.
	buildCh := stream.Map(
		client.GetFullMessages(ctx, runtime.NumCPU(), runtime.NumCPU(), messageIDs...),
		func(ctx context.Context, full liteapi.FullMessage) (*buildRes, error) {
			return buildRFC822(ctx, full, addrKRs[full.AddressID])
		},
	)
	defer buildCh.Close()

	// Create the flushers, one per update channel.
	flushers := make(map[string]*flusher)

	for addrID, updateCh := range updateCh {
		flusher := newFlusher(updateCh, maxUpdateSize)
		defer flusher.flush(ctx, true)

		flushers[addrID] = flusher
	}

	// Create a reporter to report sync progress updates.
	reporter := newReporter(userID, eventCh, len(messageIDs), time.Second)
	defer reporter.done()

	// Send each update to the appropriate flusher.
	return forEach(ctx, stream.Chunk(buildCh, maxBatchSize), func(batch []*buildRes) error {
		for _, res := range batch {
			flushers[res.addressID].push(ctx, res.update)
		}

		for _, flusher := range flushers {
			flusher.flush(ctx, true)
		}

		if err := vault.SetLastMessageID(batch[len(batch)-1].messageID); err != nil {
			return fmt.Errorf("failed to set last synced message ID: %w", err)
		}

		reporter.add(len(batch))

		return nil
	})
}

func newSystemMailboxCreatedUpdate(labelID imap.LabelID, labelName string) *imap.MailboxCreated {
	if strings.EqualFold(labelName, imap.Inbox) {
		labelName = imap.Inbox
	}

	return imap.NewMailboxCreated(imap.Mailbox{
		ID:             labelID,
		Name:           []string{labelName},
		Flags:          defaultFlags,
		PermanentFlags: defaultPermanentFlags,
		Attributes:     imap.NewFlagSet(imap.AttrNoInferiors),
	})
}

func newPlaceHolderMailboxCreatedUpdate(labelName string) *imap.MailboxCreated {
	return imap.NewMailboxCreated(imap.Mailbox{
		ID:             imap.LabelID(uuid.NewString()),
		Name:           []string{labelName},
		Flags:          defaultFlags,
		PermanentFlags: defaultPermanentFlags,
		Attributes:     imap.NewFlagSet(imap.AttrNoSelect),
	})
}

func newMailboxCreatedUpdate(labelID imap.LabelID, labelName []string) *imap.MailboxCreated {
	return imap.NewMailboxCreated(imap.Mailbox{
		ID:             labelID,
		Name:           labelName,
		Flags:          defaultFlags,
		PermanentFlags: defaultPermanentFlags,
		Attributes:     imap.NewFlagSet(),
	})
}

func wantLabelID(labelID string) bool {
	switch labelID {
	case liteapi.AllDraftsLabel, liteapi.AllSentLabel, liteapi.OutboxLabel:
		return false

	default:
		return true
	}
}

func forEach[T any](ctx context.Context, streamer stream.Stream[T], fn func(T) error) error {
	for {
		res, err := streamer.Next(ctx)
		if errors.Is(err, stream.End) {
			return nil
		} else if err != nil {
			return fmt.Errorf("failed to get next stream item: %w", err)
		}

		if err := fn(res); err != nil {
			return fmt.Errorf("failed to process stream item: %w", err)
		}
	}
}
