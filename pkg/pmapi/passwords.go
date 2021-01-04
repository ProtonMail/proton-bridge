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

package pmapi

import (
	"encoding/base64"
	"errors"

	"github.com/jameskeane/bcrypt"
)

func HashMailboxPassword(password, salt string) (hashedPassword string, err error) {
	if salt == "" {
		hashedPassword = password
		return
	}
	decodedSalt, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return
	}
	encodedSalt := base64.NewEncoding("./ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789").WithPadding(base64.NoPadding).EncodeToString(decodedSalt)
	hashResult, err := bcrypt.Hash(password, "$2y$10$"+encodedSalt)
	if err != nil {
		return
	}
	if len(hashResult) != 60 {
		err = errors.New("pmapi: invalid mailbox password hash")
		return
	}
	hashedPassword = hashResult[len(hashResult)-31:]
	return
}
