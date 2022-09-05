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

// Package types provides interfaces used in frontend packages.
package types

import (
	"crypto/tls"

	"github.com/ProtonMail/proton-bridge/v2/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
	"github.com/ProtonMail/proton-bridge/v2/internal/users"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
)

// PanicHandler is an interface of a type that can be used to gracefully handle panics which occur.
type PanicHandler interface {
	HandlePanic()
}

// Restarter allows the app to set itself to restart next time it is closed.
type Restarter interface {
	SetToRestart()
	ForceLauncher(string)
	SetMainExecutable(string)
}

type Updater interface {
	Check() (updater.VersionInfo, error)
	InstallUpdate(updater.VersionInfo) error
	IsUpdateApplicable(updater.VersionInfo) bool
	CanInstall(updater.VersionInfo) bool
}

// Bridger is an interface of bridge needed by frontend.
type Bridger interface {
	Login(username string, password []byte) (pmapi.Client, *pmapi.Auth, error)
	FinishLogin(client pmapi.Client, auth *pmapi.Auth, mailboxPassword []byte) (string, error)

	GetUserIDs() []string
	GetUserInfo(string) (users.UserInfo, error)
	LogoutUser(userID string) error
	DeleteUser(userID string, clearCache bool) error
	SetAddressMode(userID string, split users.AddressMode) error

	ClearData() error
	ClearUsers() error
	FactoryReset()

	GetTLSConfig() (*tls.Config, error)
	ProvideLogsPath() (string, error)
	GetLicenseFilePath() string
	GetDependencyLicensesLink() string

	GetCurrentUserAgent() string
	SetCurrentPlatform(string)

	Get(settings.Key) string
	Set(settings.Key, string)
	GetBool(settings.Key) bool
	SetBool(settings.Key, bool)
	GetInt(settings.Key) int
	SetInt(settings.Key, int)

	ConfigureAppleMail(userID, address string) (bool, error)

	// -- old --

	ReportBug(osType, osVersion, description, accountName, address, emailClient string, attachLogs bool) error
	SetProxyAllowed(bool)
	GetProxyAllowed() bool
	EnableCache() error
	DisableCache() error
	MigrateCache(from, to string) error
	GetUpdateChannel() updater.UpdateChannel
	SetUpdateChannel(updater.UpdateChannel)
	GetKeychainApp() string
	SetKeychainApp(keychain string)
	HasError(err error) bool
	IsAutostartEnabled() bool
	EnableAutostart() error
	DisableAutostart() error
	GetLastVersion() string
	IsFirstStart() bool
	IsAllMailVisible() bool
	SetIsAllMailVisible(bool)
}
