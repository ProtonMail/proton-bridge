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

	"github.com/ProtonMail/proton-bridge/v3/internal/updater"
)

// UpdateLatest is published when the latest version of bridge is known.
type UpdateLatest struct {
	eventBase

	Version updater.VersionInfo
}

func (event UpdateLatest) String() string {
	return fmt.Sprintf("UpdateLatest: Version: %s", event.Version.Version)
}

// UpdateAvailable is published when an update is available.
// If the update is compatible (can be installed automatically), Compatible is true.
// If the update will be installed silently (without user interaction), Silent is true.
type UpdateAvailable struct {
	eventBase

	Version updater.VersionInfo

	// Compatible is true if the update can be installed automatically.
	Compatible bool

	// Silent is true if the update will be installed silently.
	Silent bool
}

func (event UpdateAvailable) String() string {
	return fmt.Sprintf("UpdateAvailable: Version %s, Compatible: %t, Silent: %t", event.Version.Version, event.Compatible, event.Silent)
}

// UpdateNotAvailable is published when no update is available.
type UpdateNotAvailable struct {
	eventBase
}

func (event UpdateNotAvailable) String() string {
	return "UpdateNotAvailable"
}

// UpdateInstalling is published when bridge begins installing an update.
type UpdateInstalling struct {
	eventBase

	Version updater.VersionInfo

	Silent bool
}

func (event UpdateInstalling) String() string {
	return fmt.Sprintf("UpdateInstalling: Version %s, Silent: %t", event.Version.Version, event.Silent)
}

// UpdateInstalled is published when an update has been installed.
type UpdateInstalled struct {
	eventBase

	Version updater.VersionInfo

	Silent bool
}

func (event UpdateInstalled) String() string {
	return fmt.Sprintf("UpdateInstalled: Version %s, Silent: %t", event.Version.Version, event.Silent)
}

// UpdateFailed is published when an update fails to be installed.
type UpdateFailed struct {
	eventBase

	Version updater.VersionInfo

	Silent bool

	Error error
}

func (event UpdateFailed) String() string {
	return fmt.Sprintf("UpdateFailed: Version %s, Silent: %t, Error: %s", event.Version.Version, event.Silent, event.Error)
}

// UpdateForced is published when the bridge version is too old and must be updated.
type UpdateForced struct {
	eventBase
}

func (event UpdateForced) String() string {
	return "UpdateForced"
}

// UpdateCheckFailed is published when the update check fails.
type UpdateCheckFailed struct {
	eventBase

	Error error
}

func (event UpdateCheckFailed) String() string {
	return fmt.Sprintf("UpdateCheckFailed: Error: %s", event.Error)
}
