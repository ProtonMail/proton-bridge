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
	"github.com/ProtonMail/proton-bridge/v3/internal/services/observability"
)

// Service which mediates IMAP syncing in Bridge.
// IMPORTANT: Be sure to cancel all ongoing sync Handlers before cancelling this service's Group.
type Service struct {
	metadataStage *MetadataStage
	downloadStage *DownloadStage
	buildStage    *BuildStage
	applyStage    *ApplyStage
	limits        syncLimits
	metaCh        *ChannelConsumerProducer[*Job]
	group         *async.Group
}

func NewService(
	panicHandler async.PanicHandler,
	observabilitySender observability.Sender,
) *Service {
	limits := newSyncLimits(2 * Gigabyte)

	metaCh := NewChannelConsumerProducer[*Job]()
	downloadCh := NewChannelConsumerProducer[DownloadRequest]()
	buildCh := NewChannelConsumerProducer[BuildRequest]()
	applyCh := NewChannelConsumerProducer[ApplyRequest]()

	return &Service{
		limits:        limits,
		metadataStage: NewMetadataStage(metaCh, downloadCh, limits.DownloadRequestMem, panicHandler),
		downloadStage: NewDownloadStage(downloadCh, buildCh, limits.MaxParallelDownloads, panicHandler),
		buildStage:    NewBuildStage(buildCh, applyCh, limits.MessageBuildMem, panicHandler, observabilitySender),
		applyStage:    NewApplyStage(applyCh),
		metaCh:        metaCh,
		group:         async.NewGroup(context.Background(), panicHandler),
	}
}

func (s *Service) Run() {
	s.metadataStage.Run(s.group)
	s.downloadStage.Run(s.group)
	s.buildStage.Run(s.group)
	s.applyStage.Run(s.group)
}

func (s *Service) Sync(ctx context.Context, stage *Job) error {
	return s.metaCh.Produce(ctx, stage)
}

func (s *Service) Close() {
	s.group.CancelAndWait()
	s.metaCh.Close()
}
