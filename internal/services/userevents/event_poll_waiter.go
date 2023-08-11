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

package userevents

import (
	"context"
	"time"
)

// EventPollWaiter is meant to be used to wait for the event loop to finish processing the current events after
// being paused.
type EventPollWaiter struct {
	ch chan struct{}
}

func newEventPollWaiter() *EventPollWaiter {
	return &EventPollWaiter{ch: make(chan struct{})}
}

func (e *EventPollWaiter) WaitPollFinished(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-e.ch:
		return nil
	}
}

func (e *EventPollWaiter) WaitPollFinishedWithDeadline(ctx context.Context, t time.Time) error {
	ctx, cancel := context.WithDeadline(ctx, t)
	defer cancel()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-e.ch:
		return nil
	}
}

func (e *EventPollWaiter) close() {
	close(e.ch)
}
