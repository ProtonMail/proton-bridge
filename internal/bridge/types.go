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

package bridge

import (
	"context"

	"github.com/ProtonMail/proton-bridge/v3/internal/updater"
)

type Locator interface {
	ProvideSettingsPath() (string, error)
	ProvideLogsPath() (string, error)
	ProvideGluonCachePath() (string, error)
	ProvideGluonDataPath() (string, error)
	ProvideStatsPath() (string, error)
	GetLicenseFilePath() string
	GetDependencyLicensesLink() string
	Clear(...string) error
	ProvideIMAPSyncConfigPath() (string, error)
	ProvideUnleashCachePath() (string, error)
	ProvideNotificationsCachePath() (string, error)
}

type ProxyController interface {
	AllowProxy()
	DisallowProxy()
}

type TLSReporter interface {
	GetTLSIssueCh() <-chan struct{}
}

type Autostarter interface {
	Enable() error
	Disable() error
	IsEnabled() bool
}

type Updater interface {
	GetVersionInfo(context.Context, updater.Downloader, updater.Channel) (updater.VersionInfo, error)
	InstallUpdate(context.Context, updater.Downloader, updater.VersionInfo) error
	RemoveOldUpdates() error
}
