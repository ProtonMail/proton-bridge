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
	"io"
	"time"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gluon/logging"
	"github.com/ProtonMail/go-proton-api"
	"github.com/bradenaw/juniper/parallel"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
)

type downloadRequest struct {
	ids          []string
	expectedSize uint64
	err          error
}

type downloadedMessageBatch struct {
	batch []proton.FullMessage
}

type MessageDownloader interface {
	GetAttachmentInto(ctx context.Context, attachmentID string, reader io.ReaderFrom) error
	GetMessage(ctx context.Context, messageID string) (proton.Message, error)
}

type downloadState int

const (
	downloadStateZero downloadState = iota
	downloadStateHasMessage
	downloadStateFinished
)

type downloadResult struct {
	ID      string
	State   downloadState
	Message proton.FullMessage
	err     error
}

func startSyncDownloader(
	ctx context.Context,
	panicHandler async.PanicHandler,
	downloader MessageDownloader,
	cache *SyncDownloadCache,
	downloadCh <-chan downloadRequest,
	syncLimits syncLimits,
) (<-chan downloadedMessageBatch, <-chan error) {
	buildCh := make(chan downloadedMessageBatch)
	errorCh := make(chan error, syncLimits.MaxParallelDownloads*4)

	// Goroutine in charge of downloading and building messages in maxBatchSize batches.
	async.GoAnnotated(ctx, panicHandler, func(ctx context.Context) {
		defer close(buildCh)
		defer close(errorCh)
		defer func() {
			logrus.Debugf("sync downloader exit")
		}()

		attachmentDownloader := newAttachmentDownloader(ctx, panicHandler, downloader, cache, syncLimits.MaxParallelDownloads)
		defer attachmentDownloader.close()

		for request := range downloadCh {
			logrus.Debugf("Download request: %v MB:%v", len(request.ids), toMB(request.expectedSize))
			if request.err != nil {
				errorCh <- request.err
				return
			}

			result, err := downloadMessageStage1(ctx, panicHandler, request, downloader, attachmentDownloader, cache, syncLimits.MaxParallelDownloads)
			if err != nil {
				errorCh <- err
				return
			}

			if ctx.Err() != nil {
				errorCh <- ctx.Err()
				return
			}

			batch, err := downloadMessagesStage2(ctx, result, downloader, cache, SyncRetryCooldown)
			if err != nil {
				errorCh <- err
				return
			}

			select {
			case buildCh <- downloadedMessageBatch{
				batch: batch,
			}:

			case <-ctx.Done():
				return
			}
		}
	}, logging.Labels{"sync-stage": "download"})

	return buildCh, errorCh
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

func attachmentWorker(ctx context.Context, downloader MessageDownloader, cache *SyncDownloadCache, work <-chan attachmentJob) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-work:
			if !ok {
				return
			}

			var result attachmentResult
			if data, ok := cache.GetAttachment(job.id); ok {
				result.attachment = data
				result.err = nil
			} else {
				var b bytes.Buffer
				b.Grow(int(job.size))
				err := downloader.GetAttachmentInto(ctx, job.id, &b)
				result.attachment = b.Bytes()
				result.err = err
				if err == nil {
					cache.StoreAttachment(job.id, result.attachment)
				}
			}

			select {
			case <-ctx.Done():
				close(job.result)
				return
			case job.result <- result:
				close(job.result)
			}
		}
	}
}

func newAttachmentDownloader(
	ctx context.Context,
	panicHandler async.PanicHandler,
	downloader MessageDownloader,
	cache *SyncDownloadCache,
	workerCount int,
) *attachmentDownloader {
	workerCh := make(chan attachmentJob, (workerCount+2)*workerCount)
	ctx, cancel := context.WithCancel(ctx)
	for i := 0; i < workerCount; i++ {
		workerCh = make(chan attachmentJob)
		async.GoAnnotated(ctx, panicHandler, func(ctx context.Context) { attachmentWorker(ctx, downloader, cache, workerCh) }, logging.Labels{
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

func downloadMessageStage1(
	ctx context.Context,
	panicHandler async.PanicHandler,
	request downloadRequest,
	downloader MessageDownloader,
	attachmentDownloader *attachmentDownloader,
	cache *SyncDownloadCache,
	parallelDownloads int,
) ([]downloadResult, error) {
	// 1st attempt download everything in parallel
	return parallel.MapContext(ctx, parallelDownloads, request.ids, func(ctx context.Context, id string) (downloadResult, error) {
		defer async.HandlePanic(panicHandler)

		result := downloadResult{ID: id}

		v, ok := cache.GetMessage(id)
		if !ok {
			msg, err := downloader.GetMessage(ctx, id)
			if err != nil {
				logrus.WithError(err).WithField("msgID", msg.ID).Error("Failed to download message")
				result.err = err
				return result, nil
			}

			cache.StoreMessage(msg)
			result.Message.Message = msg
		} else {
			result.Message.Message = v
		}

		result.State = downloadStateHasMessage

		attachments, err := attachmentDownloader.getAttachments(ctx, result.Message.Attachments)
		result.Message.AttData = attachments

		if err != nil {
			logrus.WithError(err).WithField("msgID", id).Error("Failed to download message attachments")
			result.err = err
			return result, nil
		}

		result.State = downloadStateFinished

		return result, nil
	})
}

func downloadMessagesStage2(
	ctx context.Context,
	state []downloadResult,
	downloader MessageDownloader,
	cache *SyncDownloadCache,
	coolDown time.Duration,
) ([]proton.FullMessage, error) {
	logrus.Debug("Entering download stage 2")
	var retryList []int
	var shouldWaitBeforeRetry bool

	for {
		if shouldWaitBeforeRetry {
			time.Sleep(coolDown)
		}

		retryList = nil
		shouldWaitBeforeRetry = false

		for index, s := range state {
			if s.State == downloadStateFinished {
				continue
			}

			if s.err != nil {
				if is429Error(s.err) {
					logrus.WithField("msg-id", s.ID).Debug("Message download failed due to 429, retrying")
					retryList = append(retryList, index)
					continue
				}
				return nil, s.err
			}
		}

		if len(retryList) == 0 {
			break
		}

		for _, i := range retryList {
			st := &state[i]
			if st.State == downloadStateZero {
				message, err := downloader.GetMessage(ctx, st.ID)
				if err != nil {
					logrus.WithField("msg-id", st.ID).WithError(err).Error("failed to download message (429)")
					if is429Error(err) {
						st.err = err
						shouldWaitBeforeRetry = true
						continue
					}

					return nil, err
				}

				cache.StoreMessage(message)
				st.Message.Message = message
				st.State = downloadStateHasMessage
			}

			if st.Message.AttData == nil && st.Message.NumAttachments != 0 {
				st.Message.AttData = make([][]byte, st.Message.NumAttachments)
			}

			hasAllAttachments := true
			for i := 0; i < st.Message.NumAttachments; i++ {
				if st.Message.AttData[i] == nil {
					buffer := bytes.Buffer{}
					if err := downloader.GetAttachmentInto(ctx, st.Message.Attachments[i].ID, &buffer); err != nil {
						logrus.WithField("msg-id", st.ID).WithError(err).Errorf("failed to download attachment %v/%v (429)", i+1, len(st.Message.Attachments))
						if is429Error(err) {
							st.err = err
							shouldWaitBeforeRetry = true
							hasAllAttachments = false
							continue
						}

						return nil, err
					}

					st.Message.AttData[i] = buffer.Bytes()
					cache.StoreAttachment(st.Message.Attachments[i].ID, st.Message.AttData[i])
				}
			}

			if hasAllAttachments {
				st.State = downloadStateFinished
			}
		}
	}

	logrus.Debug("All message downloaded successfully")
	return xslices.Map(state, func(s downloadResult) proton.FullMessage {
		return s.Message
	}), nil
}

func is429Error(err error) bool {
	var apiErr *proton.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Status == 429
	}

	return false
}
