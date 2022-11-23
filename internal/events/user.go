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

	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
)

type AllUsersLoaded struct {
	eventBase
}

func (event AllUsersLoaded) String() string {
	return "AllUsersLoaded"
}

type UserLoading struct {
	eventBase

	UserID string
}

func (event UserLoading) String() string {
	return fmt.Sprintf("UserLoading: UserID: %s", event.UserID)
}

type UserLoadSuccess struct {
	eventBase

	UserID string
}

func (event UserLoadSuccess) String() string {
	return fmt.Sprintf("UserLoadSuccess: UserID: %s", event.UserID)
}

type UserLoadFail struct {
	eventBase

	UserID string
	Error  error
}

func (event UserLoadFail) String() string {
	return fmt.Sprintf("UserLoadFail: UserID: %s, Error: %s", event.UserID, event.Error)
}

type UserLoggedIn struct {
	eventBase

	UserID string
}

func (event UserLoggedIn) String() string {
	return fmt.Sprintf("UserLoggedIn: UserID: %s", event.UserID)
}

type UserLoggedOut struct {
	eventBase

	UserID string
}

func (event UserLoggedOut) String() string {
	return fmt.Sprintf("UserLoggedOut: UserID: %s", event.UserID)
}

type UserDeauth struct {
	eventBase

	UserID string
}

func (event UserDeauth) String() string {
	return fmt.Sprintf("UserDeauth: UserID: %s", event.UserID)
}

type UserDeleted struct {
	eventBase

	UserID string
}

func (event UserDeleted) String() string {
	return fmt.Sprintf("UserDeleted: UserID: %s", event.UserID)
}

type UserChanged struct {
	eventBase

	UserID string
}

func (event UserChanged) String() string {
	return fmt.Sprintf("UserChanged: UserID: %s", event.UserID)
}

type UserRefreshed struct {
	eventBase

	UserID string
}

func (event UserRefreshed) String() string {
	return fmt.Sprintf("UserRefreshed: UserID: %s", event.UserID)
}

type AddressModeChanged struct {
	eventBase

	UserID string

	AddressMode vault.AddressMode
}

func (event AddressModeChanged) String() string {
	return fmt.Sprintf("AddressModeChanged: UserID: %s, AddressMode: %s", event.UserID, event.AddressMode)
}
