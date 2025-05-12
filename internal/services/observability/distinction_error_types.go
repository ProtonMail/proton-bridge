// Copyright (c) 2025 Proton AG
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

// DistinctionMetricTypeEnum - used to distinct specific metrics which we want to limit over some interval.
// Most enums are tied to a specific error schema for which we also send a specific distinction user update.
type DistinctionMetricTypeEnum int

const (
	SyncError DistinctionMetricTypeEnum = iota
	GluonImapError
	GluonMessageError
	GluonOtherError
	SMTPError
	EventLoopError // EventLoopError - should always be kept last when inserting new keys.
	NewIMAPConnectionsExceedThreshold
)

// errorSchemaMap - maps between some DistinctionMetricTypeEnum and the relevant schema name.
var errorSchemaMap = map[DistinctionMetricTypeEnum]string{ //nolint:gochecknoglobals
	SyncError:         "bridge_sync_errors_users_total",
	EventLoopError:    "bridge_event_loop_events_errors_users_total",
	GluonImapError:    "bridge_gluon_imap_errors_users_total",
	GluonMessageError: "bridge_gluon_message_errors_users_total",
	SMTPError:         "bridge_smtp_errors_users_total",
	GluonOtherError:   "bridge_gluon_other_errors_users_total",
}

// createLastSentMap - needs to be updated whenever we make changes to the enum.
func createLastSentMap() map[DistinctionMetricTypeEnum]time.Time {
	registerTime := time.Now().Add(-updateInterval)
	lastSentMap := make(map[DistinctionMetricTypeEnum]time.Time)

	for errType := SyncError; errType <= EventLoopError; errType++ {
		lastSentMap[errType] = registerTime
	}

	return lastSentMap
}
