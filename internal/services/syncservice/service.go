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
	"github.com/ProtonMail/gluon/reporter"
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
	panicHandler  async.PanicHandler
}

func NewService(reporter reporter.Reporter,
	panicHandler async.PanicHandler,
) *Service {
	limits := newSyncLimits(2 * Gigabyte)

	metaCh := NewChannelConsumerProducer[*Job]()
	downloadCh := NewChannelConsumerProducer[DownloadRequest]()
	buildCh := NewChannelConsumerProducer[BuildRequest]()
	applyCh := NewChannelConsumerProducer[ApplyRequest]()

	return &Service{
		limits:        limits,
		metadataStage: NewMetadataStage(metaCh, downloadCh, limits.DownloadRequestMem, panicHandler),
		downloadStage: NewDownloadStage(downloadCh, buildCh, 20, panicHandler),
		buildStage:    NewBuildStage(buildCh, applyCh, limits.MessageBuildMem, panicHandler, reporter),
		applyStage:    NewApplyStage(applyCh),
		metaCh:        metaCh,
		panicHandler:  panicHandler,
	}
}

func (s *Service) Run(group *async.Group) {
	group.Once(func(ctx context.Context) {
		syncGroup := async.NewGroup(ctx, s.panicHandler)

		s.metadataStage.Run(syncGroup)
		s.downloadStage.Run(syncGroup)
		s.buildStage.Run(syncGroup)
		s.applyStage.Run(syncGroup)

		defer s.metaCh.Close()
		defer syncGroup.CancelAndWait()

		<-ctx.Done()
	})
}

func (s *Service) Sync(ctx context.Context, stage *Job) {
	s.metaCh.Produce(ctx, stage)
}
