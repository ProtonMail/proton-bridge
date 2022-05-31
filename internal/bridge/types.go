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

package bridge

import (
	"github.com/Masterminds/semver/v3"

	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
)

type Locator interface {
	Clear() error
	ClearUpdates() error
	ProvideLogsPath() (string, error)
}

type CacheProvider interface {
	GetIMAPCachePath() string
	GetDBDir() string
	GetDefaultMessageCacheDir() string
}

type SettingsProvider interface {
	Get(key string) string
	Set(key string, value string)
	GetBool(key string) bool
	SetBool(key string, val bool)
	GetInt(key string) int
}

type Updater interface {
	Check() (updater.VersionInfo, error)
	IsDowngrade(updater.VersionInfo) bool
	InstallUpdate(updater.VersionInfo) error
}

type Versioner interface {
	RemoveOtherVersions(*semver.Version) error
}
