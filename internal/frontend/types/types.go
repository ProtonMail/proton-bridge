// Copyright (c) 2020 Proton Technologies AG
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
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/ProtonMail/proton-bridge/pkg/updates"
)

// PanicHandler is an interface of a type that can be used to gracefully handle panics which occur.
type PanicHandler interface {
	HandlePanic()
}

// Updater is an interface for handling Bridge upgrades.
type Updater interface {
	CheckIsBridgeUpToDate() (isUpToDate bool, latestVersion updates.VersionInfo, err error)
	GetDownloadLink() string
	GetLocalVersion() updates.VersionInfo
	StartUpgrade(currentStatus chan<- updates.Progress)
}

type NoEncConfirmator interface {
	ConfirmNoEncryption(string, bool)
}

// Bridger is an interface of bridge needed by frontend.
type Bridger interface {
	GetCurrentClient() string
	SetCurrentOS(os string)
	Login(username, password string) (pmapi.Client, *pmapi.Auth, error)
	FinishLogin(client pmapi.Client, auth *pmapi.Auth, mailboxPassword string) (BridgeUser, error)
	GetUsers() []BridgeUser
	GetUser(query string) (BridgeUser, error)
	DeleteUser(userID string, clearCache bool) error
	ReportBug(osType, osVersion, description, accountName, address, emailClient string) error
	ClearData() error
	AllowProxy()
	DisallowProxy()
	CheckConnection() error
}

// BridgeUser is an interface of user needed by frontend.
type BridgeUser interface {
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

type bridgeWrap struct {
	*bridge.Bridge
}

// NewBridgeWrap wraps bridge struct into local bridgeWrap to implement local interface.
// The problem is that Bridge returns the bridge package's User type.
// Every method which returns User therefore has to be overridden to fulfill the interface.
func NewBridgeWrap(bridge *bridge.Bridge) *bridgeWrap { //nolint[golint]
	return &bridgeWrap{Bridge: bridge}
}

func (b *bridgeWrap) FinishLogin(client pmapi.Client, auth *pmapi.Auth, mailboxPassword string) (BridgeUser, error) {
	return b.Bridge.FinishLogin(client, auth, mailboxPassword)
}

func (b *bridgeWrap) GetUsers() (users []BridgeUser) {
	for _, user := range b.Bridge.GetUsers() {
		users = append(users, user)
	}
	return
}

func (b *bridgeWrap) GetUser(query string) (BridgeUser, error) {
	return b.Bridge.GetUser(query)
}
