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

package message

import (
	"bytes"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

func DecryptAndBuildRFC822(kr *crypto.KeyRing, msg proton.Message, attData [][]byte, opts JobOptions) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := DecryptAndBuildRFC822Into(kr, msg, attData, opts, buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func DecryptAndBuildRFC822Into(kr *crypto.KeyRing, msg proton.Message, attData [][]byte, opts JobOptions, buf *bytes.Buffer) error {
	decrypted := DecryptMessage(kr, msg, attData)

	return BuildRFC822Into(kr, &decrypted, opts, buf)
}
