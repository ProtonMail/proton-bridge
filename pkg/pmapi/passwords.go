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

import (
	"encoding/base64"

	"github.com/ProtonMail/go-srp"
	"github.com/pkg/errors"
)

// HashMailboxPassword expectects 128bit long salt encoded by standard base64.
func HashMailboxPassword(password []byte, salt string) ([]byte, error) {
	if salt == "" {
		return password, nil
	}

	decodedSalt, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode salt")
	}

	hash, err := srp.MailboxPassword(password, decodedSalt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to hash password")
	}

	return hash[len(hash)-31:], nil
}
