// Copyright (c) 2023 Proton AG
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

	"github.com/ProtonMail/proton-bridge/v3/internal/logging"
)

type UserAddressCreated struct {
	eventBase

	UserID    string
	AddressID string
	Email     string
}

func (event UserAddressCreated) String() string {
	return fmt.Sprintf("UserAddressCreated: UserID: %s, AddressID: %s, Email: %s", event.UserID, event.AddressID, logging.Sensitive(event.Email))
}

type UserAddressEnabled struct {
	eventBase

	UserID    string
	AddressID string
	Email     string
}

func (event UserAddressEnabled) String() string {
	return fmt.Sprintf("UserAddressEnabled: UserID: %s, AddressID: %s, Email: %s", event.UserID, event.AddressID, logging.Sensitive(event.Email))
}

type UserAddressDisabled struct {
	eventBase

	UserID    string
	AddressID string
	Email     string
}

func (event UserAddressDisabled) String() string {
	return fmt.Sprintf("UserAddressDisabled: UserID: %s, AddressID: %s, Email: %s", event.UserID, event.AddressID, logging.Sensitive(event.Email))
}

type UserAddressUpdated struct {
	eventBase

	UserID    string
	AddressID string
	Email     string
}

func (event UserAddressUpdated) String() string {
	return fmt.Sprintf("UserAddressUpdated: UserID: %s, AddressID: %s, Email: %s", event.UserID, event.AddressID, logging.Sensitive(event.Email))
}

type UserAddressDeleted struct {
	eventBase

	UserID    string
	AddressID string
	Email     string
}

func (event UserAddressDeleted) String() string {
	return fmt.Sprintf("UserAddressDeleted: UserID: %s, AddressID: %s, Email: %s", event.UserID, event.AddressID, logging.Sensitive(event.Email))
}
