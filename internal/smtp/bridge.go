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

package smtp

import (
	"github.com/ProtonMail/proton-bridge/internal/bridge"
)

type bridger interface {
	GetUser(query string) (bridgeUser, error)
}

type bridgeUser interface {
	CheckBridgeLogin(password string) error
	IsCombinedAddressMode() bool
	GetAddressID(address string) (string, error)
	GetTemporaryPMAPIClient() bridge.PMAPIProvider
	GetStore() storeUserProvider
}

type bridgeWrap struct {
	*bridge.Bridge
}

// newBridgeWrap wraps bridge struct into local bridgeWrap to implement local
// interface. The problem is that bridge returns package bridge's User type, so
// every method that returns User has to be overridden to fulfill the interface.
func newBridgeWrap(bridge *bridge.Bridge) *bridgeWrap {
	return &bridgeWrap{Bridge: bridge}
}

func (b *bridgeWrap) GetUser(query string) (bridgeUser, error) {
	user, err := b.Bridge.GetUser(query)
	if err != nil {
		return nil, err
	}
	return newBridgeUserWrap(user), nil
}

type bridgeUserWrap struct {
	*bridge.User
}

func newBridgeUserWrap(bridgeUser *bridge.User) *bridgeUserWrap {
	return &bridgeUserWrap{User: bridgeUser}
}

func (u *bridgeUserWrap) GetStore() storeUserProvider {
	return u.User.GetStore()
}
