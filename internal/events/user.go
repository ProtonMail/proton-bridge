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

package events

import (
	"fmt"

	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
)

// AllUsersLoaded is emitted when all users have been loaded.
type AllUsersLoaded struct {
	eventBase
}

func (event AllUsersLoaded) String() string {
	return "AllUsersLoaded"
}

// UserLoading is emitted when a user is being loaded.
type UserLoading struct {
	eventBase

	UserID string
}

func (event UserLoading) String() string {
	return fmt.Sprintf("UserLoading: UserID: %s", event.UserID)
}

// UserLoadSuccess is emitted when a user has been loaded successfully.
type UserLoadSuccess struct {
	eventBase

	UserID string
}

func (event UserLoadSuccess) String() string {
	return fmt.Sprintf("UserLoadSuccess: UserID: %s", event.UserID)
}

// UserLoadFail is emitted when a user has failed to load.
type UserLoadFail struct {
	eventBase

	UserID string
	Error  error
}

func (event UserLoadFail) String() string {
	return fmt.Sprintf("UserLoadFail: UserID: %s, Error: %s", event.UserID, event.Error)
}

// UserLoggedIn is emitted when a user has logged in.
type UserLoggedIn struct {
	eventBase

	UserID string
}

func (event UserLoggedIn) String() string {
	return fmt.Sprintf("UserLoggedIn: UserID: %s", event.UserID)
}

// UserLoggedOut is emitted when a user has logged out.
type UserLoggedOut struct {
	eventBase

	UserID string
}

func (event UserLoggedOut) String() string {
	return fmt.Sprintf("UserLoggedOut: UserID: %s", event.UserID)
}

// UserDeauth is emitted when a user has lost its API authentication.
type UserDeauth struct {
	eventBase

	UserID string
}

func (event UserDeauth) String() string {
	return fmt.Sprintf("UserDeauth: UserID: %s", event.UserID)
}

// UserBadEvent is emitted when a user cannot apply an event.
type UserBadEvent struct {
	eventBase

	UserID     string
	OldEventID string
	NewEventID string
	EventInfo  string

	Error error
}

func (event UserBadEvent) String() string {
	return fmt.Sprintf(
		"UserBadEvent: UserID: %s, OldEventID: %s, NewEventID: %s, EventInfo: %v, Error: %s",
		event.UserID,
		event.OldEventID,
		event.NewEventID,
		event.EventInfo,
		event.Error,
	)
}

// UserDeleted is emitted when a user has been deleted.
type UserDeleted struct {
	eventBase

	UserID string
}

func (event UserDeleted) String() string {
	return fmt.Sprintf("UserDeleted: UserID: %s", event.UserID)
}

// UserChanged is emitted when a user's data has changed (name, email, etc.).
type UserChanged struct {
	eventBase

	UserID string
}

func (event UserChanged) String() string {
	return fmt.Sprintf("UserChanged: UserID: %s", event.UserID)
}

// UserRefreshed is emitted when an API refresh was issued for a user.
type UserRefreshed struct {
	eventBase

	UserID          string
	CancelEventPool bool
}

func (event UserRefreshed) String() string {
	return fmt.Sprintf("UserRefreshed: UserID: %s", event.UserID)
}

// AddressModeChanged is emitted when a user's address mode has changed.
type AddressModeChanged struct {
	eventBase

	UserID string

	AddressMode vault.AddressMode
}

func (event AddressModeChanged) String() string {
	return fmt.Sprintf("AddressModeChanged: UserID: %s, AddressMode: %s", event.UserID, event.AddressMode)
}

// UsedSpaceChanged is emitted when the storage space used by the user has changed.
type UsedSpaceChanged struct {
	eventBase

	UserID string

	UsedSpace uint64
}

func (event UsedSpaceChanged) String() string {
	return fmt.Sprintf("UsedSpaceChanged: UserID: %s, UsedSpace: %v", event.UserID, event.UsedSpace)
}

type IMAPLoginFailed struct {
	eventBase

	Username string
}

func (event IMAPLoginFailed) String() string {
	return fmt.Sprintf("IMAPLoginFailed: Username: %s", event.Username)
}

type UncategorizedEventError struct {
	eventBase

	UserID string
	Error  error
}

func (event UncategorizedEventError) String() string {
	return fmt.Sprintf("UncategorizedEventError: UserID: %s, Source:%T, Error: %s", event.UserID, event.Error, event.Error)
}

type UserLoadedCheckResync struct {
	eventBase

	UserID string
}

func (event UserLoadedCheckResync) String() string {
	return fmt.Sprintf("UserLoadedCheckResync: UserID: %s", event.UserID)
}
