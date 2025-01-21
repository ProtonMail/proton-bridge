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

package events

import (
	"fmt"

	"github.com/ProtonMail/proton-bridge/v3/internal/updater"
)

// UpdateLatest is published when the latest version of bridge is known.
// It is only used for updating the release notes and landing page URLs.
type UpdateLatest struct {
	eventBase

	// VersionLegacy - holds Update version information; corresponding to the old update structure and logic;
	VersionLegacy updater.VersionInfoLegacy

	// Release - holds Release version data; part of the new update logic as of BRIDGE-309.
	Release updater.Release
}

func (event UpdateLatest) GetLatestVersion() string {
	var latestVersion string
	if !event.VersionLegacy.IsEmpty() {
		latestVersion = event.VersionLegacy.Version.String()
	} else if !event.Release.IsEmpty() {
		latestVersion = event.Release.Version.String()
	}
	return latestVersion
}

func (event UpdateLatest) String() string {
	if !event.VersionLegacy.IsEmpty() {
		return fmt.Sprintf("UpdateLatest: Version: %s", event.VersionLegacy.Version)
	}
	if !event.Release.IsEmpty() {
		return fmt.Sprintf("UpdateLatest: Version: %s", event.Release.Version)
	}
	return ""
}

// UpdateAvailable is published when an update is available.
// If the update is compatible (can be installed automatically), Compatible is true.
// If the update will be installed silently (without user interaction), Silent is true.
type UpdateAvailable struct {
	eventBase

	// VersionLegacy - holds Update version information; corresponding to the old update structure and logic;
	VersionLegacy updater.VersionInfoLegacy

	// Release - holds Release version data; part of the new update logic as of BRIDGE-309.
	Release updater.Release

	// Compatible is true if the update can be installed automatically.
	Compatible bool

	// Silent is true if the update will be installed silently.
	Silent bool
}

func (event UpdateAvailable) GetLatestVersion() string {
	var latestVersion string
	if !event.VersionLegacy.IsEmpty() {
		latestVersion = event.VersionLegacy.Version.String()
	} else if !event.Release.IsEmpty() {
		latestVersion = event.Release.Version.String()
	}
	return latestVersion
}

func (event UpdateAvailable) String() string {
	if !event.Release.IsEmpty() {
		return fmt.Sprintf("UpdateAvailable: Version %s, Compatible: %t, Silent: %t", event.Release.Version, event.Compatible, event.Silent)
	} else if !event.VersionLegacy.IsEmpty() {
		return fmt.Sprintf("UpdateAvailable: Version %s, Compatible: %t, Silent: %t", event.VersionLegacy.Version, event.Compatible, event.Silent)
	}
	return ""
}

// UpdateNotAvailable is published when no update is available.
type UpdateNotAvailable struct {
	eventBase
}

func (event UpdateNotAvailable) String() string {
	return "UpdateNotAvailable"
}

// UpdateInstalled is published when an update has been installed.
type UpdateInstalled struct {
	eventBase

	// VersionLegacy - holds Update version information; corresponding to the old update structure and logic;
	VersionLegacy updater.VersionInfoLegacy

	// Release - holds Release version data; part of the new update logic as of BRIDGE-309.
	Release updater.Release

	Silent bool
}

func (event UpdateInstalled) GetLatestVersion() string {
	var latestVersion string
	if !event.VersionLegacy.IsEmpty() {
		latestVersion = event.VersionLegacy.Version.String()
	} else if !event.Release.IsEmpty() {
		latestVersion = event.Release.Version.String()
	}
	return latestVersion
}

func (event UpdateInstalled) String() string {
	if !event.Release.IsEmpty() {
		return fmt.Sprintf("UpdateInstalled: Version %s, Silent: %t", event.Release.Version, event.Silent)
	} else if !event.VersionLegacy.IsEmpty() {
		return fmt.Sprintf("UpdateInstalled: Version %s, Silent: %t", event.VersionLegacy.Version, event.Silent)
	}
	return ""
}

// UpdateFailed is published when an update fails to be installed.
type UpdateFailed struct {
	eventBase

	// VersionLegacy - holds Update version information; corresponding to the old update structure and logic;
	VersionLegacy updater.VersionInfoLegacy

	// Release - holds Release version data; part of the new update logic as of BRIDGE-309.
	Release updater.Release

	Silent bool

	Error error
}

func (event UpdateFailed) GetLatestVersion() string {
	var latestVersion string
	if !event.VersionLegacy.IsEmpty() {
		latestVersion = event.VersionLegacy.Version.String()
	} else if !event.Release.IsEmpty() {
		latestVersion = event.Release.Version.String()
	}
	return latestVersion
}

func (event UpdateFailed) String() string {
	if !event.Release.IsEmpty() {
		return fmt.Sprintf("UpdateFailed: Version %s, Silent: %t, Error: %s", event.Release.Version, event.Silent, event.Error)
	} else if !event.VersionLegacy.IsEmpty() {
		return fmt.Sprintf("UpdateFailed: Version %s, Silent: %t, Error: %s", event.VersionLegacy.Version, event.Silent, event.Error)
	}
	return ""
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
