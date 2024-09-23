// Copyright (c) 2024 Proton AG
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

	log *logrus.Entry
	jw  *jobWaiter

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

	j := &Job{
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
		panicHandler:   panicHandler,
		downloadCache:  cache,
		jw:             newJobWaiter(log.WithField("sync-job", "waiter"), panicHandler),
	}

	j.jw.begin()

	return j
}

func (j *Job) close() {
	j.jw.close()
}

func (j *Job) onError(err error) {
	defer j.jw.onTaskFinished(err)

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

	// j.onError() also calls j.jw.onTaskFinished().
	defer j.jw.onTaskFinished(nil)
	j.syncReporter.OnProgress(ctx, count)
}

// begin is expected to be called once the job enters the pipeline.
func (j *Job) begin() {
	j.log.Info("Job started")
}

// end is expected to be called once the job has no further work left.
func (j *Job) end() {
	j.log.Info("Job finished")
	j.jw.onTaskFinished(nil)
}

// waitAndClose waits until the job has finished, the context got cancelled or an error occurred.
func (j *Job) waitAndClose(ctx context.Context) error {
	defer j.close()
	select {
	case <-ctx.Done():
		<-j.jw.doneCh
		return ctx.Err()
	case e := <-j.jw.doneCh:
		return e
	}
}

func (j *Job) newChildJob(messageID string, messageCount int64) childJob {
	j.log.Infof("Creating new child job")
	j.jw.onTaskCreated()
	return childJob{job: j, lastMessageID: messageID, messageCount: messageCount}
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
		s.job.jw.onTaskFinished(err)
		return true
	}

	return false
}

func (s *childJob) getContext() context.Context {
	return s.job.ctx
}

type JobWaiterMessage int

const (
	JobWaiterMessageCreated JobWaiterMessage = iota
	JobWaiterMessageFinished
)

type jobWaiterMessagePair struct {
	m   JobWaiterMessage
	err error
}

// jobWaiter is meant to be used to track ongoing sync batches. Once all the child jobs
// have completed, the first recorded error (if any) will be written to doneCh and then this
// channel will be closed.
type jobWaiter struct {
	ch           chan jobWaiterMessagePair
	doneCh       chan error
	log          *logrus.Entry
	panicHandler async.PanicHandler
}

func newJobWaiter(log *logrus.Entry, panicHandler async.PanicHandler) *jobWaiter {
	return &jobWaiter{
		ch:           make(chan jobWaiterMessagePair),
		doneCh:       make(chan error, 2),
		log:          log,
		panicHandler: panicHandler,
	}
}

func (j *jobWaiter) close() {
	close(j.ch)
}

func (j *jobWaiter) sendMessage(m JobWaiterMessage, err error) {
	j.ch <- jobWaiterMessagePair{
		m:   m,
		err: err,
	}
}

func (j *jobWaiter) onTaskFinished(err error) {
	j.sendMessage(JobWaiterMessageFinished, err)
}

func (j *jobWaiter) onTaskCreated() {
	j.sendMessage(JobWaiterMessageCreated, nil)
}

func (j *jobWaiter) begin() {
	go func() {
		defer async.HandlePanic(j.panicHandler)

		total := 1
		var err error

		defer func() {
			j.doneCh <- err
			close(j.doneCh)
		}()

		for {
			m, ok := <-j.ch
			if !ok {
				return
			}

			switch m.m {
			case JobWaiterMessageCreated:
				total++
			case JobWaiterMessageFinished:
				total--
				if m.err != nil && err == nil {
					err = m.err
				}
			default:
				j.log.Errorf("Unknown message type: %v", m.m)
				continue
			}

			if total <= 0 {
				if total < 0 {
					logrus.Errorf("Child count less than 0, shouldn't happen...")
				}
				j.log.Info("All child jobs completed")
				return
			}
		}
	}()
}
