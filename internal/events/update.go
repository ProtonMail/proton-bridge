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

package events

import (
	"fmt"

	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
)

type UpdateAvailable struct {
	eventBase

	Version updater.VersionInfo

	CanInstall bool
}

func (event UpdateAvailable) String() string {
	return fmt.Sprintf("UpdateAvailable: Version %s, CanInstall %t", event.Version.Version, event.CanInstall)
}

type UpdateNotAvailable struct {
	eventBase
}

func (event UpdateNotAvailable) String() string {
	return "UpdateNotAvailable"
}

type UpdateInstalled struct {
	eventBase

	Version updater.VersionInfo
}

func (event UpdateInstalled) String() string {
	return fmt.Sprintf("UpdateInstalled: Version %s", event.Version.Version)
}

type UpdateForced struct {
	eventBase
}

func (event UpdateForced) String() string {
	return "UpdateForced"
}
