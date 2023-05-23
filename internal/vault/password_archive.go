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

package vault

import (
	"crypto/sha256"
	"fmt"
)

// set archives the password for an email address, overwriting any existing archived value.
func (p *PasswordArchive) set(emailAddress string, password []byte) {
	if p.Archive == nil {
		p.Archive = make(map[string][]byte)
	}

	p.Archive[emailHashString(emailAddress)] = password
}

// get retrieves the archived password for an email address, or nil if not found.
func (p *PasswordArchive) get(emailAddress string) []byte {
	if p.Archive == nil {
		return nil
	}

	return p.Archive[emailHashString(emailAddress)]
}

// emailHashString returns a hash string for an email address as a hexadecimal string.
func emailHashString(emailAddress string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(emailAddress)))
}
