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
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/queue"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/safe"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/bradenaw/juniper/parallel"
	"github.com/bradenaw/juniper/xslices"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

const (
	maxUpdateSize = 1 << 27 // 128 MiB
	maxBatchSize  = 1 << 8  // 256
)

// doSync begins syncing the users data.
// It first ensures the latest event ID is known; if not, it fetches it.
// It sends a SyncStarted event and then either SyncFinished or SyncFailed
// depending on whether the sync was successful.
func (user *User) doSync(ctx context.Context) error {
	if user.vault.EventID() == "" {
		eventID, err := user.client.GetLatestEventID(ctx)
		if err != nil {
			return fmt.Errorf("failed to get latest event ID: %w", err)
		}

		if err := user.vault.SetEventID(eventID); err != nil {
			return fmt.Errorf("failed to set latest event ID: %w", err)
		}
	}

	start := time.Now()

	user.log.WithField("start", start).Info("Beginning user sync")

	user.eventCh.Enqueue(events.SyncStarted{
		UserID: user.ID(),
	})

	if err := user.sync(ctx); err != nil {
		user.log.WithError(err).Warn("Failed to sync user")

		user.eventCh.Enqueue(events.SyncFailed{
			UserID: user.ID(),
			Error:  err,
		})

		return fmt.Errorf("failed to sync: %w", err)
	}

	user.log.WithField("duration", time.Since(start)).Info("Finished user sync")

	user.eventCh.Enqueue(events.SyncFinished{
		UserID: user.ID(),
	})

	return nil
}

// nolint:funlen
func (user *User) sync(ctx context.Context) error {
	return safe.RLockRet(func() error {
		return withAddrKRs(user.apiUser, user.apiAddrs, user.vault.KeyPass(), func(_ *crypto.KeyRing, addrKRs map[string]*crypto.KeyRing) error {
			if !user.vault.SyncStatus().HasLabels {
				user.log.Info("Syncing labels")

				if err := syncLabels(ctx, user.apiLabels, xslices.Unique(maps.Values(user.updateCh))...); err != nil {
					return fmt.Errorf("failed to sync labels: %w", err)
				}

				if err := user.vault.SetHasLabels(true); err != nil {
					return fmt.Errorf("failed to set has labels: %w", err)
				}

				user.log.Info("Synced labels")
			} else {
				user.log.Info("Labels are already synced, skipping")
			}

			if !user.vault.SyncStatus().HasMessages {
				user.log.Info("Syncing messages")

				// Determine which messages to sync.
				messageIDs, err := user.client.GetMessageIDs(ctx, "")
				if err != nil {
					return fmt.Errorf("failed to get message IDs to sync: %w", err)
				}

				// Remove any messages that have already failed to sync.
				messageIDs = xslices.Filter(messageIDs, func(messageID string) bool {
					return !slices.Contains(user.vault.SyncStatus().FailedMessageIDs, messageID)
				})

				// Reverse the order of the message IDs so that the newest messages are synced first.
				xslices.Reverse(messageIDs)

				// If we have a message ID that we've already synced, then we can skip all messages before it.
				if idx := xslices.Index(messageIDs, user.vault.SyncStatus().LastMessageID); idx >= 0 {
					messageIDs = messageIDs[idx+1:]
				}

				// Sync the messages.
				if err := syncMessages(
					ctx,
					user.ID(),
					messageIDs,
					user.client,
					user.reporter,
					user.vault,
					user.apiLabels,
					addrKRs,
					user.updateCh,
					user.eventCh,
					user.syncWorkers,
				); err != nil {
					return fmt.Errorf("failed to sync messages: %w", err)
				}

				if err := user.vault.SetHasMessages(true); err != nil {
					return fmt.Errorf("failed to set has messages: %w", err)
				}

				user.log.Info("Synced messages")
			} else {
				user.log.Info("Messages are already synced, skipping")
			}

			return nil
		})
	}, user.apiUserLock, user.apiAddrsLock, user.apiLabelsLock, user.updateChLock)
}

// nolint:exhaustive
func syncLabels(ctx context.Context, apiLabels map[string]proton.Label, updateCh ...*queue.QueuedChannel[imap.Update]) error {
	// Create placeholder Folders/Labels mailboxes with a random ID and with the \Noselect attribute.
	for _, prefix := range []string{folderPrefix, labelPrefix} {
		for _, updateCh := range updateCh {
			updateCh.Enqueue(newPlaceHolderMailboxCreatedUpdate(prefix))
		}
	}

	// Sync the user's labels.
	for labelID, label := range apiLabels {
		if !wantLabel(label) {
			continue
		}

		switch label.Type {
		case proton.LabelTypeSystem:
			for _, updateCh := range updateCh {
				updateCh.Enqueue(newSystemMailboxCreatedUpdate(imap.MailboxID(label.ID), label.Name))
			}

		case proton.LabelTypeFolder, proton.LabelTypeLabel:
			for _, updateCh := range updateCh {
				updateCh.Enqueue(newMailboxCreatedUpdate(imap.MailboxID(labelID), getMailboxName(label)))
			}

		default:
			return fmt.Errorf("unknown label type: %d", label.Type)
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

// nolint:funlen
func syncMessages(
	ctx context.Context,
	userID string,
	messageIDs []string,
	client *proton.Client,
	sentry reporter.Reporter,
	vault *vault.User,
	apiLabels map[string]proton.Label,
	addrKRs map[string]*crypto.KeyRing,
	updateCh map[string]*queue.QueuedChannel[imap.Update],
	eventCh *queue.QueuedChannel[events.Event],
	syncWorkers int,
) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Track the amount of time to process all the messages.
	syncStartTime := time.Now()
	defer func() { logrus.WithField("duration", time.Since(syncStartTime)).Info("Message sync completed") }()

	logrus.WithFields(logrus.Fields{
		"messages": len(messageIDs),
		"workers":  syncWorkers,
		"numCPU":   runtime.NumCPU(),
	}).Info("Starting message sync")

	// Create the flushers, one per update channel.
	flushers := make(map[string]*flusher, len(updateCh))

	for addrID, updateCh := range updateCh {
		flushers[addrID] = newFlusher(updateCh, maxUpdateSize)
	}

	// Create a reporter to report sync progress updates.
	syncReporter := newSyncReporter(userID, eventCh, len(messageIDs), time.Second)
	defer syncReporter.done()

	type flushUpdate struct {
		messageID string
		noOps     []*imap.Noop
		batchLen  int
	}

	// The higher this value, the longer we can continue our download iteration before being blocked on channel writes
	// to the update flushing goroutine.
	flushCh := make(chan []*buildRes, 2)

	// Allow up to 4 batched wait requests.
	flushUpdateCh := make(chan flushUpdate, 4)

	errorCh := make(chan error, syncWorkers)

	// Goroutine in charge of downloading and building messages in maxBatchSize batches.
	go func() {
		defer close(flushCh)
		defer close(errorCh)

		for _, batch := range xslices.Chunk(messageIDs, maxBatchSize) {
			if ctx.Err() != nil {
				errorCh <- ctx.Err()
				return
			}

			result, err := parallel.MapContext(ctx, syncWorkers, batch, func(ctx context.Context, id string) (*buildRes, error) {
				msg, err := client.GetFullMessage(ctx, id)
				if err != nil {
					return nil, err
				}

				if ctx.Err() != nil {
					return nil, ctx.Err()
				}

				return buildRFC822(apiLabels, msg, addrKRs[msg.AddressID]), nil
			})
			if err != nil {
				errorCh <- err
				return
			}

			if ctx.Err() != nil {
				errorCh <- ctx.Err()
				return
			}

			flushCh <- result
		}
	}()

	// Goroutine in charge of converting the messages into updates and building a waitable structure for progress
	// tracking.
	go func() {
		defer close(flushUpdateCh)
		for batch := range flushCh {
			for _, res := range batch {
				if err := res.update.err(); err != nil {
					if err := vault.AddFailedMessageID(res.messageID); err != nil {
						logrus.WithError(err).Error("Failed to add failed message ID")
					}

					if err := sentry.ReportMessageWithContext("Failed to sync message", reporter.Context{
						"messageID": res.messageID,
						"error":     err,
					}); err != nil {
						logrus.WithError(err).Error("Failed to report message sync error")
					}

					continue
				} else {
					if err := vault.RemFailedMessageID(res.messageID); err != nil {
						logrus.WithError(err).Error("Failed to remove failed message ID")
					}

					flushers[res.addressID].push(res.update.unwrap())
				}
			}

			for _, flusher := range flushers {
				flusher.flush()
			}

			noopUpdates := make([]*imap.Noop, len(updateCh))
			index := 0
			for _, updateCh := range updateCh {
				noopUpdates[index] = imap.NewNoop()
				updateCh.Enqueue(noopUpdates[index])
				index++
			}

			flushUpdateCh <- flushUpdate{
				messageID: batch[0].messageID,
				noOps:     noopUpdates,
				batchLen:  len(batch),
			}
		}
	}()

	for flushUpdate := range flushUpdateCh {
		for _, up := range flushUpdate.noOps {
			up.WaitContext(ctx)
		}

		if err := vault.SetLastMessageID(flushUpdate.messageID); err != nil {
			return fmt.Errorf("failed to set last synced message ID: %w", err)
		}

		syncReporter.add(flushUpdate.batchLen)
	}

	return <-errorCh
}

func newSystemMailboxCreatedUpdate(labelID imap.MailboxID, labelName string) *imap.MailboxCreated {
	if strings.EqualFold(labelName, imap.Inbox) {
		labelName = imap.Inbox
	}

	attrs := imap.NewFlagSet(imap.AttrNoInferiors)
	permanentFlags := defaultPermanentFlags
	flags := defaultFlags

	switch labelID {
	case proton.TrashLabel:
		attrs = attrs.Add(imap.AttrTrash)

	case proton.SpamLabel:
		attrs = attrs.Add(imap.AttrJunk)

	case proton.AllMailLabel:
		attrs = attrs.Add(imap.AttrAll)
		flags = imap.NewFlagSet(imap.FlagSeen, imap.FlagFlagged)
		permanentFlags = imap.NewFlagSet(imap.FlagSeen, imap.FlagFlagged)

	case proton.ArchiveLabel:
		attrs = attrs.Add(imap.AttrArchive)

	case proton.SentLabel:
		attrs = attrs.Add(imap.AttrSent)

	case proton.DraftsLabel:
		attrs = attrs.Add(imap.AttrDrafts)

	case proton.StarredLabel:
		attrs = attrs.Add(imap.AttrFlagged)
	}

	return imap.NewMailboxCreated(imap.Mailbox{
		ID:             labelID,
		Name:           []string{labelName},
		Flags:          flags,
		PermanentFlags: permanentFlags,
		Attributes:     attrs,
	})
}

func newPlaceHolderMailboxCreatedUpdate(labelName string) *imap.MailboxCreated {
	return imap.NewMailboxCreated(imap.Mailbox{
		ID:             imap.MailboxID(uuid.NewString()),
		Name:           []string{labelName},
		Flags:          defaultFlags,
		PermanentFlags: defaultPermanentFlags,
		Attributes:     imap.NewFlagSet(imap.AttrNoSelect),
	})
}

func newMailboxCreatedUpdate(labelID imap.MailboxID, labelName []string) *imap.MailboxCreated {
	return imap.NewMailboxCreated(imap.Mailbox{
		ID:             labelID,
		Name:           labelName,
		Flags:          defaultFlags,
		PermanentFlags: defaultPermanentFlags,
		Attributes:     imap.NewFlagSet(),
	})
}

func wantLabel(label proton.Label) bool {
	if label.Type != proton.LabelTypeSystem {
		return true
	}

	// nolint:exhaustive
	switch label.ID {
	case proton.InboxLabel:
		return true

	case proton.TrashLabel:
		return true

	case proton.SpamLabel:
		return true

	case proton.AllMailLabel:
		return true

	case proton.ArchiveLabel:
		return true

	case proton.SentLabel:
		return true

	case proton.DraftsLabel:
		return true

	case proton.StarredLabel:
		return true

	default:
		return false
	}
}

func wantLabels(apiLabels map[string]proton.Label, labelIDs []string) []string {
	return xslices.Filter(labelIDs, func(labelID string) bool {
		return wantLabel(apiLabels[labelID])
	})
}
