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

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/network"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
)

type MetadataStageOutput = StageOutputProducer[DownloadRequest]
type MetadataStageInput = StageInputConsumer[*Job]

// MetadataStage is responsible for the throttling the sync pipeline by only allowing `MetadataMaxMessages` or up to
// maximum allowed memory usage messages to go through the pipeline. It is also responsible for interleaving
// different sync jobs so all jobs can progress and finish.
type MetadataStage struct {
	output         MetadataStageOutput
	input          MetadataStageInput
	maxDownloadMem uint64
	jobs           []*metadataIterator
	log            *logrus.Entry
}

func NewMetadataStage(input MetadataStageInput, output MetadataStageOutput, maxDownloadMem uint64) *MetadataStage {
	return &MetadataStage{input: input, output: output, maxDownloadMem: maxDownloadMem, log: logrus.WithField("sync-stage", "metadata")}
}

const MetadataPageSize = 150
const MetadataMaxMessages = 250

func (m MetadataStage) Run(group *async.Group) {
	group.Once(func(ctx context.Context) {
		m.run(ctx, MetadataPageSize, MetadataMaxMessages, &network.ExpCoolDown{})
	})
}

func (m MetadataStage) run(ctx context.Context, metadataPageSize int, maxMessages int, coolDown network.CoolDownProvider) {
	defer m.output.Close()

	for {
		if ctx.Err() != nil {
			return
		}

		// Check if new job has been submitted
		job, ok, err := m.input.TryConsume(ctx)
		if err != nil {
			m.log.WithError(err).Error("Error trying to retrieve more work")
			return
		}
		if ok {
			job.begin()
			state, err := newMetadataIterator(job.ctx, job, metadataPageSize, coolDown)
			if err != nil {
				job.onError(err)
				continue
			}
			m.jobs = append(m.jobs, state)
		}

		// Iterate over all jobs and produce work.
		for i := 0; i < len(m.jobs); {
			job := m.jobs[i]

			// If the job's context has been cancelled, remove from the list.
			if job.stage.ctx.Err() != nil {
				m.jobs = xslices.RemoveUnordered(m.jobs, i, 1)
				job.stage.end()
				continue
			}

			// Check for more work.
			output, hasMore, err := job.Next(m.maxDownloadMem, metadataPageSize, maxMessages)
			if err != nil {
				job.stage.onError(err)
				m.jobs = xslices.RemoveUnordered(m.jobs, i, 1)
				continue
			}

			// If there is actually more work, push it down the pipeline.
			if len(output.ids) != 0 {
				m.output.Produce(ctx, output)
			}

			// If this job has no more work left, signal completion.
			if !hasMore {
				m.jobs = xslices.RemoveUnordered(m.jobs, i, 1)
				job.stage.end()
				continue
			}

			i++
		}
	}
}

type metadataIterator struct {
	stage          *Job
	client         *network.ProtonClientRetryWrapper[APIClient]
	lastMessageID  string
	remaining      []proton.MessageMetadata
	downloadReqIDs []string
	expectedSize   uint64
}

func newMetadataIterator(ctx context.Context, stage *Job, metadataPageSize int, coolDown network.CoolDownProvider) (*metadataIterator, error) {
	syncStatus, err := stage.state.GetSyncStatus(ctx)
	if err != nil {
		return nil, err
	}
	return &metadataIterator{
		stage:          stage,
		client:         network.NewClientRetryWrapper(stage.client, coolDown),
		lastMessageID:  syncStatus.LastSyncedMessageID,
		remaining:      nil,
		downloadReqIDs: make([]string, 0, metadataPageSize),
	}, nil
}

func (m *metadataIterator) Next(maxDownloadMem uint64, metadataPageSize int, maxMessages int) (DownloadRequest, bool, error) {
	for {
		if m.stage.ctx.Err() != nil {
			return DownloadRequest{}, false, m.stage.ctx.Err()
		}

		if len(m.remaining) == 0 {
			metadata, err := network.RetryWithClient(m.stage.ctx, m.client, func(ctx context.Context, c APIClient) ([]proton.MessageMetadata, error) {
				// To get the metadata of the messages in batches we need to initialize the state with a call to
				// GetMessageMetadata withe filter{Desc:true}.
				if m.lastMessageID == "" {
					return c.GetMessageMetadataPage(ctx, 0, metadataPageSize, proton.MessageFilter{
						Desc: true,
					})
				}

				// Afterward we perform the same query but set the EndID to the last message of the previous batch.
				// Care must be taken here as the EndID will appear again as the first metadata result if it has not
				// been eliminated.
				meta, err := c.GetMessageMetadataPage(ctx, 0, metadataPageSize, proton.MessageFilter{
					EndID: m.lastMessageID,
					Desc:  true,
				})
				if err != nil {
					return nil, err
				}

				// To break the loop we need to check that either:
				// * There are no messages returned
				if len(meta) == 0 {
					return meta, err
				}

				// * There is only one message returned and it matches the EndID query
				if meta[0].ID == m.lastMessageID {
					return meta[1:], nil
				}

				return meta, nil
			})
			if err != nil {
				m.stage.log.WithError(err).Errorf("Failed to download message metadata with lastMessageID=%v", m.lastMessageID)
				return DownloadRequest{}, false, err
			}

			m.remaining = append(m.remaining, metadata...)

			// Update the last message ID
			if len(m.remaining) != 0 {
				m.lastMessageID = m.remaining[len(m.remaining)-1].ID
			}
		}

		if len(m.remaining) == 0 {
			if len(m.downloadReqIDs) != 0 {
				return DownloadRequest{childJob: m.stage.newChildJob(m.downloadReqIDs[len(m.downloadReqIDs)-1], int64(len(m.downloadReqIDs))), ids: m.downloadReqIDs}, false, nil
			}

			return DownloadRequest{}, false, nil
		}

		for idx, meta := range m.remaining {
			nextSize := m.expectedSize + uint64(meta.Size)
			if nextSize >= maxDownloadMem || len(m.downloadReqIDs) >= maxMessages {
				m.expectedSize = 0
				m.remaining = m.remaining[idx:]
				downloadReqIDs := m.downloadReqIDs
				m.downloadReqIDs = make([]string, 0, metadataPageSize)

				return DownloadRequest{childJob: m.stage.newChildJob(downloadReqIDs[len(downloadReqIDs)-1], int64(len(downloadReqIDs))), ids: downloadReqIDs}, true, nil
			}

			m.downloadReqIDs = append(m.downloadReqIDs, meta.ID)
			m.expectedSize = nextSize
		}

		m.remaining = nil
	}
}
