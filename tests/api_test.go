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

package tests

import (
	"github.com/Masterminds/semver/v3"
	"gitlab.protontech.ch/go/liteapi/server"
)

type API interface {
	SetMinAppVersion(*semver.Version)

	GetHostURL() string
	AddCallWatcher(func(server.Call), ...string)

	CreateUser(username, address string, password []byte) (string, string, error)
	RevokeUser(userID string) error

	CreateAddress(userID, address string, password []byte) (string, error)
	RemoveAddress(userID, addrID string) error
	GetAddressKeyIDs(userID, addrID string) ([]string, error)
	RemoveAddressKey(userID, addrID, keyID string) error

	Close()
}

type fakeAPI struct {
	*server.Server
}

func newFakeAPI() *fakeAPI {
	return &fakeAPI{
		Server: server.New(),
	}
}
