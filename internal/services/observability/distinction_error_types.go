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

package observability

import "time"

// DistinctionErrorTypeEnum - maps to the specific error schema for which we
// want to send a user update.
type DistinctionErrorTypeEnum int

const (
	SyncError DistinctionErrorTypeEnum = iota
	GluonImapError
	GluonMessageError
	GluonOtherError
	SMTPError
	EventLoopError // EventLoopError - should always be kept last when inserting new keys.
)

// errorSchemaMap - maps between the DistinctionErrorTypeEnum and the relevant schema name.
var errorSchemaMap = map[DistinctionErrorTypeEnum]string{ //nolint:gochecknoglobals
	SyncError:         "bridge_sync_errors_users_total",
	EventLoopError:    "bridge_event_loop_events_errors_users_total",
	GluonImapError:    "bridge_gluon_imap_errors_users_total",
	GluonMessageError: "bridge_gluon_message_errors_users_total",
	SMTPError:         "bridge_smtp_errors_users_total",
	GluonOtherError:   "bridge_gluon_other_errors_users_total",
}

// createLastSentMap - needs to be updated whenever we make changes to the enum.
func createLastSentMap() map[DistinctionErrorTypeEnum]time.Time {
	registerTime := time.Now().Add(-updateInterval)
	lastSentMap := make(map[DistinctionErrorTypeEnum]time.Time)

	for errType := SyncError; errType <= EventLoopError; errType++ {
		lastSentMap[errType] = registerTime
	}

	return lastSentMap
}
