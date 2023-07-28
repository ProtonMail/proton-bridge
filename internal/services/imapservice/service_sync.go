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
	"context"
	"fmt"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gluon/imap"
)

type syncUpdate struct {
	addrID string
	update imap.Update
}

type syncUpdater struct {
	ch chan syncUpdate
}

type syncUpdatePublisher struct {
	addrID  string
	updater *syncUpdater
}

func (s *syncUpdatePublisher) publishUpdate(ctx context.Context, update imap.Update) {
	select {
	case <-ctx.Done():
		update.Done(fmt.Errorf("not applied: %w", ctx.Err()))
		return

	case s.updater.ch <- syncUpdate{addrID: s.addrID, update: update}:
	}
}

func newSyncUpdater() *syncUpdater {
	return &syncUpdater{ch: make(chan syncUpdate)}
}

func (s *syncUpdater) createPublisher(addrID string) *syncUpdatePublisher {
	return &syncUpdatePublisher{updater: s, addrID: addrID}
}

func (s *syncUpdater) Close() {
	close(s.ch)
}

type syncHandler struct {
	group          *async.Group
	updater        *syncUpdater
	syncFinishedCh chan error
}

func newSyncHandler(ctx context.Context, handler async.PanicHandler) *syncHandler {
	return &syncHandler{
		group:          async.NewGroup(ctx, handler),
		updater:        newSyncUpdater(),
		syncFinishedCh: make(chan error, 2),
	}
}

func (s *syncHandler) Close() {
	s.group.CancelAndWait()
	close(s.syncFinishedCh)
}

func (s *syncHandler) CancelAndWait() {
	s.group.CancelAndWait()
}

func (s *syncHandler) Cancel() {
	s.group.Cancel()
}

func (s *syncHandler) OnSyncFinishedCH() <-chan error {
	return s.syncFinishedCh
}

func (s *syncHandler) launch(service *Service) {
	service.eventProvider.Pause()

	labels := service.labels.GetLabelMap()

	updaters := make(map[string]updatePublisher, len(service.connectors))

	for _, c := range service.connectors {
		updaters[c.addrID] = s.updater.createPublisher(c.addrID)
	}

	state := &syncJob{
		client:         service.client,
		userID:         service.identityState.UserID(),
		labels:         labels,
		updaters:       updaters,
		addressMode:    service.addressMode,
		syncState:      service.syncStateProvider,
		eventPublisher: service.eventPublisher,
		log:            service.log,
		// We make a copy of the identity state to avoid holding on to locks for a very long time.
		identityState: service.identityState.Clone(),
		panicHandler:  service.panicHandler,
		reporter:      service.reporter,
		maxSyncMemory: service.maxSyncMemory,
	}

	s.group.Once(func(ctx context.Context) {
		err := state.run(ctx)
		s.syncFinishedCh <- err
	})
}
