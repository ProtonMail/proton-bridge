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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

// Package metrics collects string constants used to report anonymous usage metrics.
package metrics

type (
	Category string
	Action   string
	Label    string
)

// Metric represents a single metric that can be reported and contains the necessary fields
// of category, action and label that the /metrics endpoint expects.
type Metric struct {
	c Category
	a Action
	l Label
}

// New returns a metric struct with the given category, action and label.
// Maybe in future we could perform checks here that the correct category is given for each action.
// That's why the Metric fields are not exported; we don't want users creating broken metrics
// (though for now they still can do that).
func New(c Category, a Action, l Label) Metric {
	return Metric{c: c, a: a, l: l}
}

// Get returns the category, action and label of a metric.
func (m Metric) Get() (Category, Action, Label) {
	return m.c, m.a, m.l
}

// Metrics related to bridge/account setup.
const (
	// Setup is used to group metrics related to bridge setup e.g. first start, new user.
	Setup = Category("setup")

	// FirstStart signifies that the bridge has been started for the first time on a user's
	// machine (or at least, no config directory was found).
	FirstStart = Action("first_start")

	// NewUser signifies a new user account has been added to the bridge.
	NewUser = Action("new_user")
)

// Metrics related to heartbeats of various kinds.
const (
	// Heartbeat is used to group heartbeat metrics e.g. the daily alive signal.
	Heartbeat = Category("heartbeat")

	// Daily is a daily signal that indicates continued bridge usage.
	Daily = Action("daily")
)

// Metrics related to import-export (transfer) process.
const (
	// Import is used to group import metrics.
	Import = Category("import")

	// Export is used to group export metrics.
	Export = Category("export")

	// TransferLoad signifies that the transfer load source.
	// It can be IMAP or local files for import, or PM for export.
	// With this will be reported also label with number of source mailboxes.
	TransferLoad = Action("load")

	// TransferStart signifies started transfer.
	TransferStart = Action("start")

	// TransferComplete signifies completed transfer without crash.
	TransferComplete = Action("complete")

	// TransferCancel signifies cancelled transfer by an user.
	TransferCancel = Action("cancel")

	// TransferFail signifies stopped transfer because of an fatal error.
	TransferFail = Action("fail")
)

const NoLabel = Label("")
