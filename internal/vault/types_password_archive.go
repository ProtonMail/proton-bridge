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

// PasswordArchive maps a list email address hashes to passwords.
// The type is not defined as a map alias to prevent having to handle nil default values when vault was created by an older version of the application.
type PasswordArchive struct {
	// we store the SHA-256 sum as string for readability and JSON marshalling of map[[32]byte][]byte will not be allowed, thus breaking vault-editor.
	Archive map[string][]byte
}
