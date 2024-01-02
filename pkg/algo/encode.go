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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package algo

import "encoding/base64"

// B64Encode returns the base64 encoding of the given byte slice.
func B64Encode(b []byte) []byte {
	enc := make([]byte, base64.StdEncoding.EncodedLen(len(b)))

	base64.StdEncoding.Encode(enc, b)

	return enc
}

// B64RawEncode returns the base64 encoding of the given byte slice.
func B64RawEncode(b []byte) []byte {
	enc := make([]byte, base64.RawURLEncoding.EncodedLen(len(b)))

	base64.RawURLEncoding.Encode(enc, b)

	return enc
}

// B64RawDecode returns the bytes represented by the base64 encoding of the given byte slice.
func B64RawDecode(b []byte) ([]byte, error) {
	dec := make([]byte, base64.RawURLEncoding.DecodedLen(len(b)))

	n, err := base64.RawURLEncoding.Decode(dec, b)
	if err != nil {
		return nil, err
	}

	return dec[:n], nil
}
