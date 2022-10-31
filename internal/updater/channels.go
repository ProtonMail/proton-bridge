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

package updater

// UpdateChannel represents an update channel users can be subscribed to.
type UpdateChannel string

const (
	// StableChannel is the channel all users are subscribed to by default.
	StableChannel UpdateChannel = "stable"

	// EarlyChannel is the channel users subscribe to when they enable "Early Access".
	EarlyChannel UpdateChannel = "early"
)

// DefaultUpdateChannel is the default update channel to subscribe to.
// It is set to the stable channel by default, unless overridden at build time.
var DefaultUpdateChannel = StableChannel //nolint:gochecknoglobals
