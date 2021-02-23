// Copyright (c) 2021 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

// Package types provides interfaces used in frontend packages.
package types

import (
	"github.com/ProtonMail/proton-bridge/internal/bridge"
	"github.com/ProtonMail/proton-bridge/internal/importexport"
	"github.com/ProtonMail/proton-bridge/internal/transfer"
	"github.com/ProtonMail/proton-bridge/internal/updater"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
)

// PanicHandler is an interface of a type that can be used to gracefully handle panics which occur.
type PanicHandler interface {
	HandlePanic()
}

// Restarter allows the app to set itself to restart next time it is closed.
type Restarter interface {
	SetToRestart()
}

type NoEncConfirmator interface {
	ConfirmNoEncryption(string, bool)
}

type Updater interface {
	Check() (updater.VersionInfo, error)
	InstallUpdate(updater.VersionInfo) error
	IsUpdateApplicable(updater.VersionInfo) bool
	CanInstall(updater.VersionInfo) bool
}

// UserManager is an interface of users needed by frontend.
type UserManager interface {
	Login(username, password string) (pmapi.Client, *pmapi.Auth, error)
	FinishLogin(client pmapi.Client, auth *pmapi.Auth, mailboxPassword string) (User, error)
	GetUsers() []User
	GetUser(query string) (User, error)
	DeleteUser(userID string, clearCache bool) error
	ClearData() error
	CheckConnection() error
}

// User is an interface of user needed by frontend.
type User interface {
	ID() string
	Username() string
	IsConnected() bool
	IsCombinedAddressMode() bool
	GetPrimaryAddress() string
	GetAddresses() []string
	GetBridgePassword() string
	SwitchAddressMode() error
	Logout() error
}

// Bridger is an interface of bridge needed by frontend.
type Bridger interface {
	UserManager

	ReportBug(osType, osVersion, description, accountName, address, emailClient string) error
	AllowProxy()
	DisallowProxy()
	GetUpdateChannel() updater.UpdateChannel
	SetUpdateChannel(updater.UpdateChannel) (needRestart bool, err error)
	GetKeychainApp() string
	SetKeychainApp(keychain string)
}

type bridgeWrap struct {
	*bridge.Bridge
}

// NewBridgeWrap wraps bridge struct into local bridgeWrap to implement local interface.
// The problem is that Bridge returns the bridge package's User type.
// Every method which returns User therefore has to be overridden to fulfill the interface.
func NewBridgeWrap(bridge *bridge.Bridge) *bridgeWrap { //nolint[golint]
	return &bridgeWrap{Bridge: bridge}
}

func (b *bridgeWrap) FinishLogin(client pmapi.Client, auth *pmapi.Auth, mailboxPassword string) (User, error) {
	return b.Bridge.FinishLogin(client, auth, mailboxPassword)
}

func (b *bridgeWrap) GetUsers() (users []User) {
	for _, user := range b.Bridge.GetUsers() {
		users = append(users, user)
	}
	return
}

func (b *bridgeWrap) GetUser(query string) (User, error) {
	return b.Bridge.GetUser(query)
}

// ImportExporter is an interface of import-export needed by frontend.
type ImportExporter interface {
	UserManager

	GetLocalImporter(string, string) (*transfer.Transfer, error)
	GetRemoteImporter(string, string, string, string, string) (*transfer.Transfer, error)
	GetEMLExporter(string, string) (*transfer.Transfer, error)
	GetMBOXExporter(string, string) (*transfer.Transfer, error)
	ReportBug(osType, osVersion, description, accountName, address, emailClient string) error
	ReportFile(osType, osVersion, accountName, address string, logdata []byte) error
}

type importExportWrap struct {
	*importexport.ImportExport
}

// NewImportExportWrap wraps import-export struct into local importExportWrap
// to implement local interface.
// The problem is that Import-Export returns the importexport package's User
// type. Every method which returns User therefore has to be overridden to
// fulfill the interface.
func NewImportExportWrap(ie *importexport.ImportExport) *importExportWrap { //nolint[golint]
	return &importExportWrap{ImportExport: ie}
}

func (b *importExportWrap) FinishLogin(client pmapi.Client, auth *pmapi.Auth, mailboxPassword string) (User, error) {
	return b.ImportExport.FinishLogin(client, auth, mailboxPassword)
}

func (b *importExportWrap) GetUsers() (users []User) {
	for _, user := range b.ImportExport.GetUsers() {
		users = append(users, user)
	}
	return
}

func (b *importExportWrap) GetUser(query string) (User, error) {
	return b.ImportExport.GetUser(query)
}
