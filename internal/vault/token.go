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

package vault

import (
	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

// RandomToken is a function that returns a random token.
// By default, we use crypto.RandomToken to generate tokens.
var RandomToken = crypto.RandomToken // nolint:gochecknoglobals

func newRandomToken(size int) []byte {
	token, err := RandomToken(size)
	if err != nil {
		panic(err)
	}

	return token
}
