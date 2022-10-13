// Copyright (c) 2022 Proton AG
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
	"time"

	"github.com/ProtonMail/gluon/queue"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
)

type reporter struct {
	userID  string
	eventCh *queue.QueuedChannel[events.Event]

	start time.Time
	total int
	count int

	last time.Time
	freq time.Duration
}

func newReporter(userID string, eventCh *queue.QueuedChannel[events.Event], total int, freq time.Duration) *reporter {
	return &reporter{
		userID:  userID,
		eventCh: eventCh,

		start: time.Now(),
		total: total,
		freq:  freq,
	}
}

func (rep *reporter) add(delta int) {
	rep.count += delta

	if time.Since(rep.last) > rep.freq {
		rep.eventCh.Enqueue(events.SyncProgress{
			UserID:    rep.userID,
			Progress:  float64(rep.count) / float64(rep.total),
			Elapsed:   time.Since(rep.start),
			Remaining: time.Since(rep.start) * time.Duration(rep.total-(rep.count+1)) / time.Duration(rep.count+1),
		})

		rep.last = time.Now()
	}
}

func (rep *reporter) done() {
	rep.eventCh.Enqueue(events.SyncProgress{
		UserID:    rep.userID,
		Progress:  1,
		Elapsed:   time.Since(rep.start),
		Remaining: 0,
	})
}
