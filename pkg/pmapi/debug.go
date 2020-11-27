// Copyright (c) 2020 Proton Technologies AG
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

import "unicode/utf8"

func printBytes(body []byte) string {
	if utf8.Valid(body) {
		return string(body)
	}
	enc := []rune{}
	for _, b := range body {
		switch {
		case b == 9:
			enc = append(enc, rune('⟼'))
		case b == 13:
			enc = append(enc, rune('↵'))
		case b < 32, b == 127:
			enc = append(enc, '◡')
		case b > 31 && b < 127, b == 10:
			enc = append(enc, rune(b))
		default:
			enc = append(enc, 9728+rune(b))
		}
	}

	return string(enc)
}
