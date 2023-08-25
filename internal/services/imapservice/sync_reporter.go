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
	"time"

	"github.com/ProtonMail/proton-bridge/v3/internal/events"
)

type syncReporter struct {
	userID         string
	eventPublisher events.EventPublisher

	start time.Time
	total int64
	count int64

	last time.Time
	freq time.Duration
}

func (rep *syncReporter) OnStart(ctx context.Context) {
	rep.start = time.Now()
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
	rep.count += delta

	if time.Since(rep.last) > rep.freq {
		rep.eventPublisher.PublishEvent(ctx, events.SyncProgress{
			UserID:    rep.userID,
			Progress:  float64(rep.count) / float64(rep.total),
			Elapsed:   time.Since(rep.start),
			Remaining: time.Since(rep.start) * time.Duration(rep.total-(rep.count+1)) / time.Duration(rep.count+1),
		})

		rep.last = time.Now()
	}
}

func (rep *syncReporter) InitializeProgressCounter(_ context.Context, current int64, total int64) {
	rep.count = current
	rep.total = total
}

func newSyncReporter(userID string, eventsPublisher events.EventPublisher, freq time.Duration) *syncReporter {
	return &syncReporter{
		userID:         userID,
		eventPublisher: eventsPublisher,

		start: time.Now(),
		freq:  freq,
	}
}
