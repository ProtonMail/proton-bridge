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

package syncservice

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/go-proton-api"
	"github.com/sirupsen/logrus"
)

// Job represents a unit of work that will travel down the sync pipeline. The job will be split up into child jobs
// for each batch. The parent job (this) will then wait until all the children have finished executing. Execution can
// terminate by either:
// * Completing the pipeline successfully
// * Context Cancellation
// * Errors
// On error, or context cancellation all child jobs are cancelled.
type Job struct {
	ctx    context.Context
	cancel func()

	client APIClient
	state  StateProvider

	userID         string
	labels         LabelMap
	messageBuilder MessageBuilder
	updateApplier  UpdateApplier
	syncReporter   Reporter

	log     *logrus.Entry
	errorCh *async.QueuedChannel[error]
	wg      sync.WaitGroup
	once    sync.Once

	panicHandler  async.PanicHandler
	downloadCache *DownloadCache

	metadataFetched   int64
	totalMessageCount int64
}

func NewJob(ctx context.Context,
	client APIClient,
	userID string,
	labels LabelMap,
	messageBuilder MessageBuilder,
	updateApplier UpdateApplier,
	syncReporter Reporter,
	state StateProvider,
	panicHandler async.PanicHandler,
	cache *DownloadCache,
	log *logrus.Entry,
) *Job {
	ctx, cancel := context.WithCancel(ctx)

	return &Job{
		ctx:            ctx,
		client:         client,
		userID:         userID,
		cancel:         cancel,
		state:          state,
		log:            log,
		labels:         labels,
		messageBuilder: messageBuilder,
		updateApplier:  updateApplier,
		syncReporter:   syncReporter,
		errorCh:        async.NewQueuedChannel[error](4, 8, panicHandler, fmt.Sprintf("sync-job-error-%v", userID)),
		panicHandler:   panicHandler,
		downloadCache:  cache,
	}
}

func (j *Job) Close() {
	j.errorCh.CloseAndDiscardQueued()
	j.wg.Wait()
}

func (j *Job) onError(err error) {
	defer j.wg.Done()

	// context cancelled is caught & handled in a different location.
	if errors.Is(err, context.Canceled) {
		return
	}

	j.errorCh.Enqueue(err)
	j.cancel()
}

func (j *Job) onStageCompleted(ctx context.Context, count int64) {
	j.syncReporter.OnProgress(ctx, count)
}

func (j *Job) onJobFinished(ctx context.Context, lastMessageID string, count int64) {
	if err := j.state.SetLastMessageID(ctx, lastMessageID, count); err != nil {
		j.log.WithError(err).Error("Failed to store last synced message id")
		j.onError(err)
		return
	}

	// j.onError() also calls j.wg.Done().
	j.wg.Done()
	j.syncReporter.OnProgress(ctx, count)
}

// begin is expected to be called once the job enters the pipeline.
func (j *Job) begin() {
	j.log.Info("Job started")
	j.wg.Add(1)
	j.startChildWaiter()
}

// end is expected to be called once the job has no further work left.
func (j *Job) end() {
	j.log.Info("Job finished")
	j.wg.Done()
}

// wait waits until the job has finished, the context got cancelled or an error occurred.
func (j *Job) wait(ctx context.Context) error {
	defer j.wg.Wait()

	select {
	case <-ctx.Done():
		j.cancel()
		return ctx.Err()
	case err := <-j.errorCh.GetChannel():
		return err
	}
}

func (j *Job) newChildJob(messageID string, messageCount int64) childJob {
	j.log.Infof("Creating new child job")
	j.wg.Add(1)
	return childJob{job: j, lastMessageID: messageID, messageCount: messageCount}
}

func (j *Job) startChildWaiter() {
	j.once.Do(func() {
		go func() {
			defer async.HandlePanic(j.panicHandler)

			j.wg.Wait()
			j.log.Info("All child jobs succeeded")
			j.errorCh.Enqueue(j.ctx.Err())
		}()
	})
}

// childJob represents a batch of work that goes down the pipeline. It keeps track of the message ID that is in the
// batch and the number of messages in the batch.
type childJob struct {
	job                 *Job
	lastMessageID       string
	messageCount        int64
	cachedMessageIDs    []string
	cachedAttachmentIDs []string
}

func (s *childJob) onError(err error) {
	s.job.log.WithError(err).Info("Child job ran into error")
	s.job.onError(err)
}

func (s *childJob) userID() string {
	return s.job.userID
}

func (s *childJob) chunkDivide(chunks [][]proton.FullMessage) []childJob {
	numChunks := len(chunks)

	if numChunks == 1 {
		return []childJob{*s}
	}

	result := make([]childJob, numChunks)
	for i := 0; i < numChunks-1; i++ {
		result[i] = s.job.newChildJob(chunks[i][len(chunks[i])-1].ID, int64(len(chunks[i])))
		collectIDs(&result[i], chunks[i])
	}

	result[numChunks-1] = *s
	collectIDs(&result[numChunks-1], chunks[numChunks-1])

	return result
}

func collectIDs(j *childJob, msgs []proton.FullMessage) {
	j.cachedAttachmentIDs = make([]string, 0, len(msgs))
	j.cachedMessageIDs = make([]string, 0, len(msgs))
	for _, msg := range msgs {
		j.cachedMessageIDs = append(j.cachedMessageIDs, msg.ID)
		for _, attach := range msg.Attachments {
			j.cachedAttachmentIDs = append(j.cachedAttachmentIDs, attach.ID)
		}
	}
}

func (s *childJob) onFinished(ctx context.Context) {
	s.job.log.Infof("Child job finished")
	s.job.onJobFinished(ctx, s.lastMessageID, s.messageCount)
	s.job.downloadCache.DeleteMessages(s.cachedMessageIDs...)
	s.job.downloadCache.DeleteAttachments(s.cachedAttachmentIDs...)
}

func (s *childJob) onStageCompleted(ctx context.Context) {
	s.job.onStageCompleted(ctx, s.messageCount)
}

func (s *childJob) checkCancelled() bool {
	err := s.job.ctx.Err()
	if err != nil {
		s.job.log.Infof("Child job exit due to context cancelled")
		s.job.wg.Done()
		return true
	}

	return false
}

func (s *childJob) getContext() context.Context {
	return s.job.ctx
}
