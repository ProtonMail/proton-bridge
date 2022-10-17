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

package message

import (
	"bytes"
	"os"
	"testing"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/stretchr/testify/require"
)

func TestEncryptRFC822(t *testing.T) {
	literal, err := os.ReadFile("testdata/text_plain_latin1.eml")
	require.NoError(t, err)

	key, err := crypto.GenerateKey("name", "email", "rsa", 2048)
	require.NoError(t, err)

	kr, err := crypto.NewKeyRing(key)
	require.NoError(t, err)

	enc, err := EncryptRFC822(kr, bytes.NewReader(literal))
	require.NoError(t, err)

	section(t, enc).
		expectContentType(is(`text/plain`)).
		expectContentTypeParam(`charset`, is(`utf-8`)).
		expectBody(decryptsTo(kr, `ééééééé`))
}

func TestEncryptRFC822Multipart(t *testing.T) {
	literal, err := os.ReadFile("testdata/multipart_alternative_nested.eml")
	require.NoError(t, err)

	key, err := crypto.GenerateKey("name", "email", "rsa", 2048)
	require.NoError(t, err)

	kr, err := crypto.NewKeyRing(key)
	require.NoError(t, err)

	enc, err := EncryptRFC822(kr, bytes.NewReader(literal))
	require.NoError(t, err)

	section(t, enc).
		expectContentType(is(`multipart/alternative`))

	section(t, enc, 1).
		expectContentType(is(`multipart/alternative`))

	section(t, enc, 1, 1).
		expectContentType(is(`text/plain`)).
		expectBody(decryptsTo(kr, "*multipart 1.1*\n\n"))

	section(t, enc, 1, 2).
		expectContentType(is(`text/html`)).
		expectBody(decryptsTo(kr, `<html>
  <head>
    <meta http-equiv="content-type" content="text/html; charset=UTF-8">
  </head>
  <body>
    <b>multipart 1.2</b>
  </body>
</html>
`))

	section(t, enc, 2).
		expectContentType(is(`multipart/alternative`))

	section(t, enc, 2, 1).
		expectContentType(is(`text/plain`)).
		expectBody(decryptsTo(kr, "*multipart 2.1*\n\n"))

	section(t, enc, 2, 2).
		expectContentType(is(`text/html`)).
		expectBody(decryptsTo(kr, `<html>
  <head>
    <meta http-equiv="content-type" content="text/html; charset=UTF-8">
  </head>
  <body>
    <b>multipart 2.2</b>
  </body>
</html>
`))
}
