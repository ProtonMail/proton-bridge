package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/queue"
	"github.com/bradenaw/juniper/stream"
	"github.com/bradenaw/juniper/xslices"
	"github.com/google/uuid"
	"gitlab.protontech.ch/go/liteapi"
	"golang.org/x/exp/maps"
)

const (
	maxUpdateSize = 1 << 25
	maxBatchSize  = 1 << 8
)

func (user *User) sync(ctx context.Context) error {
	if !user.vault.SyncStatus().HasLabels {
		if err := syncLabels(ctx, user.client, maps.Values(user.updateCh)...); err != nil {
			return fmt.Errorf("failed to sync labels: %w", err)
		}

		if err := user.vault.SetHasLabels(true); err != nil {
			return fmt.Errorf("failed to set has labels: %w", err)
		}
	}

	if !user.vault.SyncStatus().HasMessages {
		if err := user.syncMessages(ctx); err != nil {
			return fmt.Errorf("failed to sync messages: %w", err)
		}

		if err := user.vault.SetHasMessages(true); err != nil {
			return fmt.Errorf("failed to set has messages: %w", err)
		}
	}

	return nil
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

func (user *User) syncMessages(ctx context.Context) error {
	// Determine which messages to sync.
	allMetadata, err := user.client.GetAllMessageMetadata(ctx, nil)
	if err != nil {
		return fmt.Errorf("get all message metadata: %w", err)
	}

	metadata := allMetadata

	// If possible, begin syncing from one beyond the last synced message.
	if beginID := user.vault.SyncStatus().LastMessageID; beginID != "" {
		if idx := xslices.IndexFunc(metadata, func(metadata liteapi.MessageMetadata) bool {
			return metadata.ID == beginID
		}); idx >= 0 {
			metadata = metadata[idx+1:]
		}
	}

	// Process the metadata, building the messages.
	buildCh := stream.Chunk(stream.Map(
		user.client.GetFullMessages(ctx, xslices.Map(metadata, func(metadata liteapi.MessageMetadata) string {
			return metadata.ID
		})...),
		func(ctx context.Context, full liteapi.FullMessage) (*buildRes, error) {
			return buildRFC822(ctx, full, user.addrKRs)
		},
	), maxBatchSize)
	defer buildCh.Close()

	// Create the flushers, one per update channel.
	flushers := make(map[string]*flusher)

	for addrID, updateCh := range user.updateCh {
		flusher := newFlusher(updateCh, maxUpdateSize)
		defer flusher.flush(ctx, true)

		flushers[addrID] = flusher
	}

	// Create a reporter to report sync progress updates.
	reporter := newReporter(user.ID(), user.eventCh, len(metadata), time.Second)
	defer reporter.done()

	var count int

	// Send each update to the appropriate flusher.
	for {
		batch, err := buildCh.Next(ctx)
		if errors.Is(err, stream.End) {
			return nil
		} else if err != nil {
			return fmt.Errorf("failed to get next sync batch: %w", err)
		}

		user.apiAddrs.Get(func(apiAddrs []liteapi.Address) {
			for _, res := range batch {
				if len(flushers) > 1 {
					flushers[res.addressID].push(ctx, res.update)
				} else {
					flushers[apiAddrs[0].ID].push(ctx, res.update)
				}
			}
		})

		for _, flusher := range flushers {
			flusher.flush(ctx, true)
		}

		if err := user.vault.SetLastMessageID(batch[len(batch)-1].messageID); err != nil {
			return fmt.Errorf("failed to set last synced message ID: %w", err)
		}

		reporter.add(len(batch))

		count += len(batch)
	}
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
