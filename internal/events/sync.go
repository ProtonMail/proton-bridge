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

package events

import (
	"fmt"
	"time"
)

type SyncStarted struct {
	eventBase

	UserID string
}

func (event SyncStarted) String() string {
	return fmt.Sprintf("SyncStarted: UserID: %s", event.UserID)
}

type SyncProgress struct {
	eventBase

	UserID    string
	Progress  float64
	Elapsed   time.Duration
	Remaining time.Duration
}

func (event SyncProgress) String() string {
	return fmt.Sprintf(
		"SyncProgress: UserID: %s, Progress: %f, Elapsed: %0.1fs, Remaining: %0.1fs",
		event.UserID,
		event.Progress,
		event.Elapsed.Seconds(),
		event.Remaining.Seconds(),
	)
}

type SyncFinished struct {
	eventBase

	UserID string
}

func (event SyncFinished) String() string {
	return fmt.Sprintf("SyncFinished: UserID: %s", event.UserID)
}

type SyncFailed struct {
	eventBase

	UserID string
	Error  error
}

func (event SyncFailed) String() string {
	return fmt.Sprintf("SyncFailed: UserID: %s, Err: %s", event.UserID, event.Error)
}
