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
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/logging"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/safe"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/bradenaw/juniper/parallel"
	"github.com/bradenaw/juniper/xslices"
	"github.com/pbnjay/memory"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// syncSystemLabels ensures that system labels are all known to gluon.
func (user *User) syncSystemLabels(ctx context.Context) error {
	return safe.RLockRet(func() error {
		var updates []imap.Update

		for _, label := range xslices.Filter(maps.Values(user.apiLabels), func(label proton.Label) bool { return label.Type == proton.LabelTypeSystem }) {
			if !wantLabel(label) {
				continue
			}

			for _, updateCh := range xslices.Unique(maps.Values(user.updateCh)) {
				update := newSystemMailboxCreatedUpdate(imap.MailboxID(label.ID), label.Name)
				updateCh.Enqueue(update)
				updates = append(updates, update)
			}
		}

		if err := waitOnIMAPUpdates(ctx, updates); err != nil {
			return fmt.Errorf("could not sync system labels: %w", err)
		}

		return nil
	}, user.apiUserLock, user.apiAddrsLock, user.apiLabelsLock, user.updateChLock)
}

// doSync begins syncing the user's data.
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
				if err := user.syncMessages(
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
					user.maxSyncMemory,
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
func syncLabels(ctx context.Context, apiLabels map[string]proton.Label, updateCh ...*async.QueuedChannel[imap.Update]) error {
	var updates []imap.Update

	// Create placeholder Folders/Labels mailboxes with the \Noselect attribute.
	for _, prefix := range []string{folderPrefix, labelPrefix} {
		for _, updateCh := range updateCh {
			update := newPlaceHolderMailboxCreatedUpdate(prefix)
			updateCh.Enqueue(update)
			updates = append(updates, update)
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
				update := newSystemMailboxCreatedUpdate(imap.MailboxID(label.ID), label.Name)
				updateCh.Enqueue(update)
				updates = append(updates, update)
			}

		case proton.LabelTypeFolder, proton.LabelTypeLabel:
			for _, updateCh := range updateCh {
				update := newMailboxCreatedUpdate(imap.MailboxID(labelID), getMailboxName(label))
				updateCh.Enqueue(update)
				updates = append(updates, update)
			}

		default:
			return fmt.Errorf("unknown label type: %d", label.Type)
		}
	}

	// Wait for all label updates to be applied.
	for _, update := range updates {
		err, ok := update.WaitContext(ctx)
		if ok && err != nil {
			return fmt.Errorf("failed to apply label create update in gluon %v: %w", update.String(), err)
		}
	}

	return nil
}

const Kilobyte = uint64(1024)
const Megabyte = 1024 * Kilobyte
const Gigabyte = 1024 * Megabyte

func toMB(v uint64) float64 {
	return float64(v) / float64(Megabyte)
}

// nolint:gocyclo
func (user *User) syncMessages(
	ctx context.Context,
	userID string,
	messageIDs []string,
	client *proton.Client,
	sentry reporter.Reporter,
	vault *vault.User,
	apiLabels map[string]proton.Label,
	addrKRs map[string]*crypto.KeyRing,
	updateCh map[string]*async.QueuedChannel[imap.Update],
	eventCh *async.QueuedChannel[events.Event],
	maxSyncMemory uint64,
) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Track the amount of time to process all the messages.
	syncStartTime := time.Now()
	defer func() { logrus.WithField("duration", time.Since(syncStartTime)).Info("Message sync completed") }()

	user.log.WithFields(logrus.Fields{
		"messages": len(messageIDs),
		"numCPU":   runtime.NumCPU(),
	}).Info("Starting message sync")

	// Create the flushers, one per update channel.

	// Create a reporter to report sync progress updates.
	syncReporter := newSyncReporter(userID, eventCh, len(messageIDs), time.Second)
	defer syncReporter.done()

	// Expected mem usage for this whole process should be the sum of MaxMessageBuildingMem and MaxDownloadRequestMem
	// times x due to pipeline and all additional memory used by network requests and compression+io.

	// There's no point in using more than 128MB of download data per stage, after that we reach a point of diminishing
	// returns as we can't keep the pipeline fed fast enough.
	const MaxDownloadRequestMem = 128 * Megabyte

	// Any lower than this and we may fail to download messages.
	const MinDownloadRequestMem = 40 * Megabyte

	// This value can be increased to your hearts content. The more system memory the user has, the more messages
	// we can build in parallel.
	const MaxMessageBuildingMem = 128 * Megabyte
	const MinMessageBuildingMem = 64 * Megabyte

	// Maximum recommend value for parallel downloads by the API team.
	const maxParallelDownloads = 20

	totalMemory := memory.TotalMemory()

	if maxSyncMemory >= totalMemory/2 {
		logrus.Warnf("Requested max sync memory of %v MB is greater than half of system memory (%v MB), forcing to half of system memory",
			maxSyncMemory, toMB(totalMemory/2))
		maxSyncMemory = totalMemory / 2
	}

	if maxSyncMemory < 800*Megabyte {
		logrus.Warnf("Requested max sync memory of %v MB, but minimum recommended is 800 MB, forcing max syncMemory to 800MB", toMB(maxSyncMemory))
		maxSyncMemory = 800 * Megabyte
	}

	logrus.Debugf("Total System Memory: %v", toMB(totalMemory))

	syncMaxDownloadRequestMem := MaxDownloadRequestMem
	syncMaxMessageBuildingMem := MaxMessageBuildingMem

	// If less than 2GB available try and limit max memory to 512 MB
	switch {
	case maxSyncMemory < 2*Gigabyte:
		if maxSyncMemory < 800*Megabyte {
			logrus.Warnf("System has less than 800MB of memory, you may experience issues sycing large mailboxes")
		}
		syncMaxDownloadRequestMem = MinDownloadRequestMem
		syncMaxMessageBuildingMem = MinMessageBuildingMem
	case maxSyncMemory == 2*Gigabyte:
		// Increasing the max download capacity has very little effect on sync speed. We could increase the download
		// memory but the user would see less sync notifications. A smaller value here leads to more frequent
		// updates. Additionally, most of ot sync time is spent in the message building.
		syncMaxDownloadRequestMem = MaxDownloadRequestMem
		// Currently limited so that if a user has multiple accounts active it also doesn't cause excessive memory usage.
		syncMaxMessageBuildingMem = MaxMessageBuildingMem
	default:
		// Divide by 8 as download stage and build stage will use aprox. 4x the specified memory.
		remainingMemory := (maxSyncMemory - 2*Gigabyte) / 8
		syncMaxDownloadRequestMem = MaxDownloadRequestMem + remainingMemory
		syncMaxMessageBuildingMem = MaxMessageBuildingMem + remainingMemory
	}

	logrus.Debugf("Max memory usage for sync Download=%vMB Building=%vMB Predicted Max Total=%vMB",
		toMB(syncMaxDownloadRequestMem),
		toMB(syncMaxMessageBuildingMem),
		toMB((syncMaxMessageBuildingMem*4)+(syncMaxDownloadRequestMem*4)),
	)

	type flushUpdate struct {
		messageID string
		err       error
		batchLen  int
	}

	type downloadRequest struct {
		ids          []string
		expectedSize uint64
		err          error
	}

	type downloadedMessageBatch struct {
		batch []proton.FullMessage
	}

	type builtMessageBatch struct {
		batch []*buildRes
	}

	downloadCh := make(chan downloadRequest)

	buildCh := make(chan downloadedMessageBatch)

	// The higher this value, the longer we can continue our download iteration before being blocked on channel writes
	// to the update flushing goroutine.
	flushCh := make(chan builtMessageBatch)

	flushUpdateCh := make(chan flushUpdate)

	errorCh := make(chan error, maxParallelDownloads*4)

	// Go routine in charge of downloading message metadata
	async.GoAnnotated(ctx, user.panicHandler, func(ctx context.Context) {
		defer close(downloadCh)
		const MetadataDataPageSize = 150

		var downloadReq downloadRequest
		downloadReq.ids = make([]string, 0, MetadataDataPageSize)

		metadataChunks := xslices.Chunk(messageIDs, MetadataDataPageSize)
		for i, metadataChunk := range metadataChunks {
			logrus.Debugf("Metadata Request (%v of %v), previous: %v", i, len(metadataChunks), len(downloadReq.ids))
			metadata, err := client.GetMessageMetadataPage(ctx, 0, len(metadataChunk), proton.MessageFilter{ID: metadataChunk})
			if err != nil {
				downloadReq.err = err
				select {
				case downloadCh <- downloadReq:
				case <-ctx.Done():
					return
				}
				return
			}

			if ctx.Err() != nil {
				return
			}

			// Build look up table so that messages are processed in the same order.
			metadataMap := make(map[string]int, len(metadata))
			for i, v := range metadata {
				metadataMap[v.ID] = i
			}

			for i, id := range metadataChunk {
				m := &metadata[metadataMap[id]]
				nextSize := downloadReq.expectedSize + uint64(m.Size)
				if nextSize >= syncMaxDownloadRequestMem || len(downloadReq.ids) >= 256 {
					logrus.Debugf("Download Request Sent at %v of %v", i, len(metadata))
					select {
					case downloadCh <- downloadReq:

					case <-ctx.Done():
						return
					}
					downloadReq.expectedSize = 0
					downloadReq.ids = make([]string, 0, MetadataDataPageSize)
					nextSize = uint64(m.Size)
				}
				downloadReq.ids = append(downloadReq.ids, id)
				downloadReq.expectedSize = nextSize
			}
		}

		if len(downloadReq.ids) != 0 {
			logrus.Debugf("Sending remaining download request")
			select {
			case downloadCh <- downloadReq:

			case <-ctx.Done():
				return
			}
		}
	}, logging.Labels{"sync-stage": "meta-data"})

	// Goroutine in charge of downloading and building messages in maxBatchSize batches.
	async.GoAnnotated(ctx, user.panicHandler, func(ctx context.Context) {
		defer close(buildCh)
		defer close(errorCh)
		defer func() {
			logrus.Debugf("sync downloader exit")
		}()

		attachmentDownloader := user.newAttachmentDownloader(ctx, client, maxParallelDownloads)
		defer attachmentDownloader.close()

		for request := range downloadCh {
			logrus.Debugf("Download request: %v MB:%v", len(request.ids), toMB(request.expectedSize))
			if request.err != nil {
				errorCh <- request.err
				return
			}

			if ctx.Err() != nil {
				errorCh <- ctx.Err()
				return
			}

			result, err := parallel.MapContext(ctx, maxParallelDownloads, request.ids, func(ctx context.Context, id string) (proton.FullMessage, error) {
				defer async.HandlePanic(user.panicHandler)

				var result proton.FullMessage

				msg, err := client.GetMessage(ctx, id)
				if err != nil {
					return proton.FullMessage{}, err
				}

				attachments, err := attachmentDownloader.getAttachments(ctx, msg.Attachments)
				if err != nil {
					return proton.FullMessage{}, err
				}

				result.Message = msg
				result.AttData = attachments

				return result, nil
			})
			if err != nil {
				errorCh <- err
				return
			}

			select {
			case buildCh <- downloadedMessageBatch{
				batch: result,
			}:

			case <-ctx.Done():
				return
			}
		}
	}, logging.Labels{"sync-stage": "download"})

	// Goroutine which builds messages after they have been downloaded
	async.GoAnnotated(ctx, user.panicHandler, func(ctx context.Context) {
		defer close(flushCh)
		defer func() {
			logrus.Debugf("sync builder exit")
		}()

		maxMessagesInParallel := runtime.NumCPU()

		for buildBatch := range buildCh {
			if ctx.Err() != nil {
				return
			}

			chunks := chunkSyncBuilderBatch(buildBatch.batch, syncMaxMessageBuildingMem)

			for index, chunk := range chunks {
				logrus.Debugf("Build request: %v of %v count=%v", index, len(chunks), len(chunk))

				result, err := parallel.MapContext(ctx, maxMessagesInParallel, chunk, func(ctx context.Context, msg proton.FullMessage) (*buildRes, error) {
					defer async.HandlePanic(user.panicHandler)

					kr, ok := addrKRs[msg.AddressID]
					if !ok {
						logrus.Errorf("Address '%v' on message '%v' does not have an unlocked kerying", msg.AddressID, msg.ID)
						return &buildRes{
							messageID: msg.ID,
							addressID: msg.AddressID,
							err:       fmt.Errorf("address does not have an unlocked keyring"),
						}, nil
					}

					res := buildRFC822(apiLabels, msg, kr, new(bytes.Buffer))
					if res.err != nil {
						logrus.WithError(res.err).WithField("msgID", msg.ID).Error("Failed to build message (syn)")
					}

					return res, nil
				})
				if err != nil {
					return
				}

				select {
				case flushCh <- builtMessageBatch{result}:

				case <-ctx.Done():
					return
				}
			}
		}
	}, logging.Labels{"sync-stage": "builder"})

	// Goroutine which converts the messages into updates and builds a waitable structure for progress tracking.
	async.GoAnnotated(ctx, user.panicHandler, func(ctx context.Context) {
		defer close(flushUpdateCh)
		defer func() {
			logrus.Debugf("sync flush exit")
		}()

		type updateTargetInfo struct {
			queueIndex int
			ch         *async.QueuedChannel[imap.Update]
		}

		pendingUpdates := make([][]*imap.MessageCreated, len(updateCh))
		addressToIndex := make(map[string]updateTargetInfo)

		{
			i := 0
			for addrID, updateCh := range updateCh {
				addressToIndex[addrID] = updateTargetInfo{
					ch:         updateCh,
					queueIndex: i,
				}
				i++
			}
		}

		for downloadBatch := range flushCh {
			logrus.Debugf("Flush batch: %v", len(downloadBatch.batch))
			for _, res := range downloadBatch.batch {
				if res.err != nil {
					if err := vault.AddFailedMessageID(res.messageID); err != nil {
						logrus.WithError(err).Error("Failed to add failed message ID")
					}

					if err := sentry.ReportMessageWithContext("Failed to build message (sync)", reporter.Context{
						"messageID": res.messageID,
						"error":     res.err,
					}); err != nil {
						logrus.WithError(err).Error("Failed to report message build error")
					}

					// We could sync a placeholder message here, but for now we skip it entirely.
					continue
				}

				if err := vault.RemFailedMessageID(res.messageID); err != nil {
					logrus.WithError(err).Error("Failed to remove failed message ID")
				}

				targetInfo := addressToIndex[res.addressID]
				pendingUpdates[targetInfo.queueIndex] = append(pendingUpdates[targetInfo.queueIndex], res.update)
			}

			for _, info := range addressToIndex {
				up := imap.NewMessagesCreated(true, pendingUpdates[info.queueIndex]...)
				info.ch.Enqueue(up)

				err, ok := up.WaitContext(ctx)
				if ok && err != nil {
					flushUpdateCh <- flushUpdate{
						err: fmt.Errorf("failed to apply sync update to gluon %v: %w", up.String(), err),
					}
					return
				}

				pendingUpdates[info.queueIndex] = pendingUpdates[info.queueIndex][:0]
			}

			select {
			case flushUpdateCh <- flushUpdate{
				messageID: downloadBatch.batch[0].messageID,
				err:       nil,
				batchLen:  len(downloadBatch.batch),
			}:
			case <-ctx.Done():
				return
			}
		}
	}, logging.Labels{"sync-stage": "flush"})

	for flushUpdate := range flushUpdateCh {
		if flushUpdate.err != nil {
			return flushUpdate.err
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

	case proton.AllScheduledLabel:
		labelName = "Scheduled" // API actual name is "All Scheduled"
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
		ID:             imap.MailboxID(labelName),
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

	case proton.AllScheduledLabel:
		return true

	default:
		return false
	}
}

func wantLabels(apiLabels map[string]proton.Label, labelIDs []string) []string {
	return xslices.Filter(labelIDs, func(labelID string) bool {
		apiLabel, ok := apiLabels[labelID]
		if !ok {
			return false
		}

		return wantLabel(apiLabel)
	})
}

type attachmentResult struct {
	attachment []byte
	err        error
}

type attachmentJob struct {
	id     string
	size   int64
	result chan attachmentResult
}

type attachmentDownloader struct {
	workerCh chan attachmentJob
	cancel   context.CancelFunc
}

func attachmentWorker(ctx context.Context, client *proton.Client, work <-chan attachmentJob) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-work:
			if !ok {
				return
			}
			var b bytes.Buffer
			b.Grow(int(job.size))
			err := client.GetAttachmentInto(ctx, job.id, &b)
			select {
			case <-ctx.Done():
				close(job.result)
				return
			case job.result <- attachmentResult{attachment: b.Bytes(), err: err}:
				close(job.result)
			}
		}
	}
}

func (user *User) newAttachmentDownloader(ctx context.Context, client *proton.Client, workerCount int) *attachmentDownloader {
	workerCh := make(chan attachmentJob, (workerCount+2)*workerCount)
	ctx, cancel := context.WithCancel(ctx)
	for i := 0; i < workerCount; i++ {
		workerCh = make(chan attachmentJob)
		async.GoAnnotated(ctx, user.panicHandler, func(ctx context.Context) { attachmentWorker(ctx, client, workerCh) }, logging.Labels{
			"sync": fmt.Sprintf("att-downloader %v", i),
		})
	}

	return &attachmentDownloader{
		workerCh: workerCh,
		cancel:   cancel,
	}
}

func (a *attachmentDownloader) getAttachments(ctx context.Context, attachments []proton.Attachment) ([][]byte, error) {
	resultChs := make([]chan attachmentResult, len(attachments))
	for i, id := range attachments {
		resultChs[i] = make(chan attachmentResult, 1)
		select {
		case a.workerCh <- attachmentJob{id: id.ID, result: resultChs[i], size: id.Size}:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	result := make([][]byte, len(attachments))
	var err error
	for i := 0; i < len(attachments); i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case r := <-resultChs[i]:
			if r.err != nil {
				err = fmt.Errorf("failed to get attachment %v: %w", attachments[i], r.err)
			}
			result[i] = r.attachment
		}
	}

	return result, err
}

func (a *attachmentDownloader) close() {
	a.cancel()
}

func chunkSyncBuilderBatch(batch []proton.FullMessage, maxMemory uint64) [][]proton.FullMessage {
	var expectedMemUsage uint64
	var chunks [][]proton.FullMessage
	var lastIndex int
	var index int

	for _, v := range batch {
		var dataSize uint64
		for _, a := range v.Attachments {
			dataSize += uint64(a.Size)
		}

		// 2x increase for attachment due to extra memory needed for decrypting and writing
		// in memory buffer.
		dataSize *= 2
		dataSize += uint64(len(v.Body))

		nextMemSize := expectedMemUsage + dataSize
		if nextMemSize >= maxMemory {
			chunks = append(chunks, batch[lastIndex:index])
			lastIndex = index
			expectedMemUsage = dataSize
		} else {
			expectedMemUsage = nextMemSize
		}

		index++
	}

	if lastIndex < len(batch) {
		chunks = append(chunks, batch[lastIndex:])
	}

	return chunks
}
