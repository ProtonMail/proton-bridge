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
	"bytes"
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/logging"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/bradenaw/juniper/parallel"
	"github.com/bradenaw/juniper/xslices"
	"github.com/pbnjay/memory"
	"github.com/sirupsen/logrus"
)

func (s *syncJob) syncMessages(ctx context.Context, messageIDs []string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Track the amount of time to process all the messages.
	syncStartTime := time.Now()
	defer func() { logrus.WithField("duration", time.Since(syncStartTime)).Info("Message sync completed") }()

	s.log.WithFields(logrus.Fields{
		"messages": len(messageIDs),
		"numCPU":   runtime.NumCPU(),
	}).Info("Starting message sync")

	// Create the flushers, one per update channel.

	// Create a reporter to report sync progress updates.
	syncReporter := newSyncReporter(s.userID, s.eventPublisher, len(messageIDs), time.Second)
	defer syncReporter.done(ctx)

	// Expected mem usage for this whole process should be the sum of MaxMessageBuildingMem and MaxDownloadRequestMem
	// times x due to pipeline and all additional memory used by network requests and compression+io.

	totalMemory := memory.TotalMemory()

	syncLimits := newSyncLimits(s.maxSyncMemory)

	if syncLimits.MaxSyncMemory >= totalMemory/2 {
		logrus.Warnf("Requested max sync memory of %v MB is greater than half of system memory (%v MB), forcing to half of system memory",
			toMB(syncLimits.MaxSyncMemory), toMB(totalMemory/2))
		syncLimits.MaxSyncMemory = totalMemory / 2
	}

	if syncLimits.MaxSyncMemory < 800*Megabyte {
		logrus.Warnf("Requested max sync memory of %v MB, but minimum recommended is 800 MB, forcing max syncMemory to 800MB", toMB(syncLimits.MaxSyncMemory))
		syncLimits.MaxSyncMemory = 800 * Megabyte
	}

	logrus.Debugf("Total System Memory: %v", toMB(totalMemory))

	// Linter says it's not used. This is a lie.
	var syncMaxDownloadRequestMem uint64

	// Linter says it's not used. This is a lie.
	var syncMaxMessageBuildingMem uint64

	// If less than 2GB available try and limit max memory to 512 MB
	switch {
	case syncLimits.MaxSyncMemory < 2*Gigabyte:
		if syncLimits.MaxSyncMemory < 800*Megabyte {
			logrus.Warnf("System has less than 800MB of memory, you may experience issues sycing large mailboxes")
		}
		syncMaxDownloadRequestMem = syncLimits.MinDownloadRequestMem
		syncMaxMessageBuildingMem = syncLimits.MinMessageBuildingMem
	case syncLimits.MaxSyncMemory == 2*Gigabyte:
		// Increasing the max download capacity has very little effect on sync speed. We could increase the download
		// memory but the user would see less sync notifications. A smaller value here leads to more frequent
		// updates. Additionally, most of sync time is spent in the message building.
		syncMaxDownloadRequestMem = syncLimits.MaxDownloadRequestMem
		// Currently limited so that if a user has multiple accounts active it also doesn't cause excessive memory usage.
		syncMaxMessageBuildingMem = syncLimits.MaxMessageBuildingMem
	default:
		// Divide by 8 as download stage and build stage will use aprox. 4x the specified memory.
		remainingMemory := (syncLimits.MaxSyncMemory - 2*Gigabyte) / 8
		syncMaxDownloadRequestMem = syncLimits.MaxDownloadRequestMem + remainingMemory
		syncMaxMessageBuildingMem = syncLimits.MaxMessageBuildingMem + remainingMemory
	}

	logrus.Debugf("Max memory usage for sync Download=%vMB Building=%vMB Predicted Max Total=%vMB",
		toMB(syncMaxDownloadRequestMem),
		toMB(syncMaxMessageBuildingMem),
		toMB((syncMaxMessageBuildingMem*4)+(syncMaxDownloadRequestMem*4)),
	)

	downloadCh := startMetadataDownloader(ctx, s, messageIDs, syncMaxDownloadRequestMem)
	buildCh, errorCh := startMessageDownloader(ctx, s, syncLimits, downloadCh)
	flushCh := startMessageBuilder(ctx, s, buildCh, syncMaxMessageBuildingMem)
	flushUpdateCh := startMessageFlusher(ctx, s, flushCh)

	for flushUpdate := range flushUpdateCh {
		if flushUpdate.err != nil {
			return flushUpdate.err
		}

		if err := s.syncState.SetLastMessageID(flushUpdate.messageID); err != nil {
			return fmt.Errorf("failed to set last synced message ID: %w", err)
		}

		syncReporter.add(ctx, flushUpdate.batchLen)
	}

	return <-errorCh
}

const Kilobyte = uint64(1024)
const Megabyte = 1024 * Kilobyte
const Gigabyte = 1024 * Megabyte

func toMB(v uint64) float64 {
	return float64(v) / float64(Megabyte)
}

type syncLimits struct {
	MaxDownloadRequestMem uint64
	MinDownloadRequestMem uint64
	MaxMessageBuildingMem uint64
	MinMessageBuildingMem uint64
	MaxSyncMemory         uint64
	MaxParallelDownloads  int
}

func newSyncLimits(maxSyncMemory uint64) syncLimits {
	limits := syncLimits{
		// There's no point in using more than 128MB of download data per stage, after that we reach a point of diminishing
		// returns as we can't keep the pipeline fed fast enough.
		MaxDownloadRequestMem: 128 * Megabyte,

		// Any lower than this and we may fail to download messages.
		MinDownloadRequestMem: 40 * Megabyte,

		// This value can be increased to your hearts content. The more system memory the user has, the more messages
		// we can build in parallel.
		MaxMessageBuildingMem: 128 * Megabyte,
		MinMessageBuildingMem: 64 * Megabyte,

		// Maximum recommend value for parallel downloads by the API team.
		MaxParallelDownloads: 20,

		MaxSyncMemory: maxSyncMemory,
	}

	if _, ok := os.LookupEnv("BRIDGE_SYNC_FORCE_MINIMUM_SPEC"); ok {
		logrus.Warn("Sync specs forced to minimum")
		limits.MaxDownloadRequestMem = 50 * Megabyte
		limits.MaxMessageBuildingMem = 80 * Megabyte
		limits.MaxParallelDownloads = 2
		limits.MaxSyncMemory = 800 * Megabyte
	}

	return limits
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

func startMetadataDownloader(ctx context.Context, s *syncJob, messageIDs []string, syncMaxDownloadRequestMem uint64) <-chan downloadRequest {
	downloadCh := make(chan downloadRequest)
	// Go routine in charge of downloading message metadata
	async.GoAnnotated(ctx, s.panicHandler, func(ctx context.Context) {
		defer close(downloadCh)
		const MetadataDataPageSize = 150

		var downloadReq downloadRequest
		downloadReq.ids = make([]string, 0, MetadataDataPageSize)

		metadataChunks := xslices.Chunk(messageIDs, MetadataDataPageSize)
		for i, metadataChunk := range metadataChunks {
			logrus.Debugf("Metadata Request (%v of %v), previous: %v", i, len(metadataChunks), len(downloadReq.ids))
			metadata, err := s.client.GetMessageMetadataPage(ctx, 0, len(metadataChunk), proton.MessageFilter{ID: metadataChunk})
			if err != nil {
				logrus.WithError(err).Errorf("Failed to download message metadata for chunk %v", i)
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

	return downloadCh
}

func startMessageDownloader(ctx context.Context, s *syncJob, syncLimits syncLimits, downloadCh <-chan downloadRequest) (<-chan downloadedMessageBatch, <-chan error) {
	buildCh := make(chan downloadedMessageBatch)
	errorCh := make(chan error, syncLimits.MaxParallelDownloads*4)

	// Goroutine in charge of downloading and building messages in maxBatchSize batches.
	async.GoAnnotated(ctx, s.panicHandler, func(ctx context.Context) {
		defer close(buildCh)
		defer close(errorCh)
		defer func() {
			logrus.Debugf("sync downloader exit")
		}()

		attachmentDownloader := s.newAttachmentDownloader(ctx, s.client, syncLimits.MaxParallelDownloads)
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

			result, err := parallel.MapContext(ctx, syncLimits.MaxParallelDownloads, request.ids, func(ctx context.Context, id string) (proton.FullMessage, error) {
				defer async.HandlePanic(s.panicHandler)

				var result proton.FullMessage

				msg, err := s.client.GetMessage(ctx, id)
				if err != nil {
					logrus.WithError(err).WithField("msgID", msg.ID).Error("Failed to download message")
					return proton.FullMessage{}, err
				}

				attachments, err := attachmentDownloader.getAttachments(ctx, msg.Attachments)
				if err != nil {
					logrus.WithError(err).WithField("msgID", msg.ID).Error("Failed to download message attachments")
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

	return buildCh, errorCh
}

func startMessageBuilder(ctx context.Context, s *syncJob, buildCh <-chan downloadedMessageBatch, syncMaxMessageBuildingMem uint64) <-chan builtMessageBatch {
	flushCh := make(chan builtMessageBatch)

	// Goroutine which builds messages after they have been downloaded
	async.GoAnnotated(ctx, s.panicHandler, func(ctx context.Context) {
		defer close(flushCh)
		defer func() {
			logrus.Debugf("sync builder exit")
		}()

		if err := s.identityState.WithAddrKRs(nil, func(_ *crypto.KeyRing, addrKRs map[string]*crypto.KeyRing) error {
			maxMessagesInParallel := runtime.NumCPU()

			for buildBatch := range buildCh {
				if ctx.Err() != nil {
					return ctx.Err()
				}

				chunks := chunkSyncBuilderBatch(buildBatch.batch, syncMaxMessageBuildingMem)

				for index, chunk := range chunks {
					logrus.Debugf("Build request: %v of %v count=%v", index, len(chunks), len(chunk))

					result, err := parallel.MapContext(ctx, maxMessagesInParallel, chunk, func(ctx context.Context, msg proton.FullMessage) (*buildRes, error) {
						defer async.HandlePanic(s.panicHandler)

						kr, ok := addrKRs[msg.AddressID]
						if !ok {
							logrus.Errorf("Address '%v' on message '%v' does not have an unlocked kerying", msg.AddressID, msg.ID)
							return &buildRes{
								messageID: msg.ID,
								addressID: msg.AddressID,
								err:       fmt.Errorf("address does not have an unlocked keyring"),
							}, nil
						}

						res := buildRFC822(s.labels, msg, kr, new(bytes.Buffer))
						if res.err != nil {
							s.log.WithError(res.err).WithField("msgID", msg.ID).Error("Failed to build message (syn)")
						}

						return res, nil
					})
					if err != nil {
						return err
					}

					select {
					case flushCh <- builtMessageBatch{result}:

					case <-ctx.Done():
						return nil
					}
				}
			}

			return nil
		}); err != nil {
			s.log.WithError(err).Error("Sync message builder exited with error")
		}
	}, logging.Labels{"sync-stage": "builder"})

	return flushCh
}

func startMessageFlusher(ctx context.Context, s *syncJob, messageBatchCH <-chan builtMessageBatch) <-chan flushUpdate {
	flushUpdateCh := make(chan flushUpdate)

	// Goroutine which converts the messages into updates and builds a waitable structure for progress tracking.
	async.GoAnnotated(ctx, s.panicHandler, func(ctx context.Context) {
		defer close(flushUpdateCh)
		defer func() {
			logrus.Debugf("sync flush exit")
		}()

		type updateTargetInfo struct {
			queueIndex int
			ch         updatePublisher
		}

		pendingUpdates := make([][]*imap.MessageCreated, len(s.updaters))
		addressToIndex := make(map[string]updateTargetInfo)

		{
			i := 0
			for addrID, updateCh := range s.updaters {
				addressToIndex[addrID] = updateTargetInfo{
					ch:         updateCh,
					queueIndex: i,
				}
				i++
			}
		}

		for downloadBatch := range messageBatchCH {
			logrus.Debugf("Flush batch: %v", len(downloadBatch.batch))
			for _, res := range downloadBatch.batch {
				if res.err != nil {
					if err := s.syncState.AddFailedMessageID(res.messageID); err != nil {
						logrus.WithError(err).Error("Failed to add failed message ID")
					}

					if err := s.reporter.ReportMessageWithContext("Failed to build message (sync)", reporter.Context{
						"messageID": res.messageID,
						"error":     res.err,
					}); err != nil {
						s.log.WithError(err).Error("Failed to report message build error")
					}

					// We could sync a placeholder message here, but for now we skip it entirely.
					continue
				}

				if err := s.syncState.RemFailedMessageID(res.messageID); err != nil {
					logrus.WithError(err).Error("Failed to remove failed message ID")
				}

				targetInfo := addressToIndex[res.addressID]
				pendingUpdates[targetInfo.queueIndex] = append(pendingUpdates[targetInfo.queueIndex], res.update)
			}

			for _, info := range addressToIndex {
				up := imap.NewMessagesCreated(true, pendingUpdates[info.queueIndex]...)
				info.ch.publishUpdate(ctx, up)

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

	return flushUpdateCh
}
