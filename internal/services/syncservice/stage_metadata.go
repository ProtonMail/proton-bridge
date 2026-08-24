// Copyright (c) 2026 Proton AG
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

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gluon/logging"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/network"
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
	log            *logrus.Entry
	panicHandler   async.PanicHandler
}

func NewMetadataStage(
	input MetadataStageInput,
	output MetadataStageOutput,
	maxDownloadMem uint64,
	panicHandler async.PanicHandler,
) *MetadataStage {
	return &MetadataStage{
		input:          input,
		output:         output,
		maxDownloadMem: maxDownloadMem,
		log:            logrus.WithField("sync-stage", "metadata"),
		panicHandler:   panicHandler,
	}
}

const MetadataPageSize = 128
const MetadataMaxMessages = 64

// MetadataIDPageSize is the Proton /mail/v4/messages/ids page size (API max is 1000).
// The metadata-list endpoint used previously (GetMessageMetadataPage with Sort=ID and
// no LabelID) silently stops after ~255 IDs on large mailboxes.
const MetadataIDPageSize = 1000

func (m *MetadataStage) Run(group *async.Group) {
	group.Once(func(ctx context.Context) {
		logging.DoAnnotated(
			ctx,
			func(ctx context.Context) {
				m.run(ctx, MetadataIDPageSize, MetadataMaxMessages, &network.ExpCoolDown{})
			},
			logging.Labels{"sync-stage": "metadata"},
		)
	})
}

func (m *MetadataStage) run(ctx context.Context, metadataPageSize int, maxMessages int, coolDown network.CoolDownProvider) {
	defer m.output.Close()

	group := async.NewGroup(ctx, m.panicHandler)
	defer group.CancelAndWait()

	for {
		job, err := m.input.Consume(ctx)
		if err != nil {
			if !errors.Is(err, context.Canceled) && !errors.Is(err, ErrNoMoreInput) {
				m.log.WithError(err).Error("Error trying to retrieve more work")
			}
			return
		}

		job.begin()
		state, err := newMetadataIterator(job.ctx, job, metadataPageSize, coolDown)
		if err != nil {
			job.onError(err)
			continue
		}

		group.Once(func(ctx context.Context) {
			for {
				if state.stage.ctx.Err() != nil {
					state.stage.end()
					return
				}

				// Check for more work.
				output, hasMore, err := state.Next(m.maxDownloadMem, metadataPageSize, maxMessages)
				if err != nil {
					state.stage.onError(err)
					return
				}

				// If there is actually more work, push it down the pipeline.
				if len(output.ids) != 0 {
					state.stage.metadataFetched += int64(len(output.ids))
					job.log.Debugf("Metada collected: %v/%v", state.stage.metadataFetched, state.stage.totalMessageCount)

					output.onStageCompleted(ctx)

					if err := m.output.Produce(ctx, output); err != nil {
						job.onError(fmt.Errorf("failed to produce output for next stage: %w", err))
						return
					}
				}

				// If this job has no more work left, signal completion.
				if !hasMore {
					state.stage.end()
					return
				}
			}
		})
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
			ids, err := network.RetryWithClient(m.stage.ctx, m.client, func(ctx context.Context, c APIClient) ([]string, error) {
				return c.GetMessageIDs(ctx, m.lastMessageID, metadataPageSize)
			})
			if err != nil {
				m.stage.log.WithError(err).Errorf("Failed to list message IDs with afterID=%v", m.lastMessageID)
				return DownloadRequest{}, false, err
			}

			// AfterID is exclusive on the Proton API. If a build treats it as inclusive,
			// drop the duplicate so we cannot loop on the same page.
			if len(ids) > 0 && m.lastMessageID != "" && ids[0] == m.lastMessageID {
				ids = ids[1:]
			}

			if len(ids) != 0 {
				m.lastMessageID = ids[len(ids)-1]
				metadata := make([]proton.MessageMetadata, len(ids))
				for i, id := range ids {
					// Size is unknown from the IDs endpoint; batching uses maxMessages.
					metadata[i] = proton.MessageMetadata{ID: id, Size: 1}
				}
				m.remaining = append(m.remaining, metadata...)
			}
		}

		if len(m.remaining) == 0 {
			if len(m.downloadReqIDs) != 0 {
				return DownloadRequest{childJob: m.stage.newChildJob(m.downloadReqIDs[len(m.downloadReqIDs)-1], int64(len(m.downloadReqIDs))), ids: m.downloadReqIDs}, false, nil
			}

			return DownloadRequest{}, false, nil
		}

		for idx, meta := range m.remaining {
			nextSize := m.expectedSize + uint64(meta.Size) //nolint:gosec // disable G115
			if nextSize >= maxDownloadMem || len(m.downloadReqIDs) >= maxMessages {
				m.expectedSize = 0
				downloadReqIDs := m.downloadReqIDs
				m.downloadReqIDs = make([]string, 0, metadataPageSize)

				if len(downloadReqIDs) == 0 {
					downloadReqIDs = []string{meta.ID}
					m.remaining = m.remaining[idx+1:]
				} else {
					m.remaining = m.remaining[idx:]
				}

				return DownloadRequest{childJob: m.stage.newChildJob(downloadReqIDs[len(downloadReqIDs)-1], int64(len(downloadReqIDs))), ids: downloadReqIDs}, true, nil
			}

			m.downloadReqIDs = append(m.downloadReqIDs, meta.ID)
			m.expectedSize = nextSize
		}

		m.remaining = nil
	}
}
