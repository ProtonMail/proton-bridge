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

package store

import (
	"context"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
)

type PanicHandler interface {
	HandlePanic()
}

// BridgeUser is subset of bridge.User for use by the Store.
type BridgeUser interface {
	ID() string
	GetAddressID(address string) (string, error)
	IsConnected() bool
	IsCombinedAddressMode() bool
	GetPrimaryAddress() string
	GetStoreAddresses() []string
	GetClient() pmapi.Client
	UpdateUser(context.Context) error
	CloseAllConnections()
	CloseConnection(string)
	Logout() error
}
