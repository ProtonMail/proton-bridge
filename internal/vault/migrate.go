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

import "fmt"

type Version int

const (
	v2_3_x Version = iota
	v2_4_x
	v2_5_x

	Current = v2_5_x
)

// upgrade migrates the vault from the given version to the next version.
func upgrade(v Version, b []byte) ([]byte, error) {
	switch v {
	case v2_3_x:
		return upgrade_2_3_x(b)

	case v2_4_x:
		return upgrade_2_4_x(b)

	case Current:
		return nil, fmt.Errorf("already at current version %d", Current)

	default:
		return nil, fmt.Errorf("unknown version %d", v)
	}
}
