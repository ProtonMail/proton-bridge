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

import "errors"

var (
	ErrVaultInsecure = errors.New("the vault is insecure")
	ErrVaultCorrupt  = errors.New("the vault is corrupt")
	ErrWatchUpdates  = errors.New("failed to watch for updates")

	ErrNoSuchUser          = errors.New("no such user")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrUserAlreadyLoggedIn = errors.New("the user is already logged in")
	ErrNotImplemented      = errors.New("not implemented")

	ErrSizeTooLarge = errors.New("file is too big")
)
