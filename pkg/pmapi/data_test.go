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

package pmapi

import "github.com/ProtonMail/gopenpgp/v2/crypto"

var testIdentity = &crypto.Identity{
	Name:  "UserID",
	Email: "",
}

const (
	testUID          = "729ad6012421d67ad26950dc898bebe3a6e3caa2" //nolint:gosec
	testAccessToken  = "de0423049b44243afeec7d9c1d99be7b46da1e8a" //nolint:gosec
	testRefreshToken = "a49b98256745bb497bec20e9b55f5de16f01fb52" //nolint:gosec
)
