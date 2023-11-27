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
	"sync"
	"time"

	"github.com/ProtonMail/proton-bridge/v3/internal/events"
)

type syncData struct {
	start time.Time
	total int64
	count int64

	last time.Time
	freq time.Duration
}

type syncReporter struct {
	userID         string
	eventPublisher events.EventPublisher

	dataLock sync.Mutex
	data     syncData
}

func (rep *syncReporter) withData(f func(s *syncData)) {
	rep.dataLock.Lock()
	defer rep.dataLock.Unlock()

	f(&rep.data)
}

func (rep *syncReporter) OnStart(ctx context.Context) {
	rep.withData(func(s *syncData) {
		s.start = time.Now()
	})
	rep.eventPublisher.PublishEvent(ctx, events.SyncStarted{UserID: rep.userID})
}

func (rep *syncReporter) OnFinished(ctx context.Context) {
	rep.eventPublisher.PublishEvent(ctx, events.SyncFinished{
		UserID: rep.userID,
	})
}

func (rep *syncReporter) OnError(ctx context.Context, err error) {
	rep.eventPublisher.PublishEvent(ctx, events.SyncFailed{
		UserID: rep.userID,
		Error:  err,
	})
}

func (rep *syncReporter) OnProgress(ctx context.Context, delta int64) {
	rep.withData(func(s *syncData) {
		s.count += delta
		var progress float64
		var remaining time.Duration

		// It's possible for count to be bigger or smaller than total depending on when the sync begins and whether new
		// messages are added/removed during this period. When this happens just limited the progress to 100%.
		if s.count > s.total {
			progress = 1
		} else {
			progress = float64(s.count) / float64(s.total)
			remaining = time.Since(s.start) * time.Duration(s.total-(s.count+1)) / time.Duration(s.count+1)
		}

		if time.Since(s.last) > s.freq {
			rep.eventPublisher.PublishEvent(ctx, events.SyncProgress{
				UserID:    rep.userID,
				Progress:  progress,
				Elapsed:   time.Since(s.start),
				Remaining: remaining,
			})

			s.last = time.Now()
		}
	})
}

func (rep *syncReporter) InitializeProgressCounter(_ context.Context, current int64, total int64) {
	rep.withData(func(s *syncData) {
		s.count = current
		s.total = total
	})
}

func newSyncReporter(userID string, eventsPublisher events.EventPublisher, freq time.Duration) *syncReporter {
	return &syncReporter{
		userID:         userID,
		eventPublisher: eventsPublisher,

		data: syncData{
			start: time.Now(),
			freq:  freq,
		},
	}
}
