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

package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeaderLines(t *testing.T) {
	const header = "To: somebody\r\nFrom: somebody else\r\nSubject: this is\r\n\ta multiline field\r\n\r\n"

	assert.Equal(t, [][]byte{
		[]byte("To: somebody\r\n"),
		[]byte("From: somebody else\r\n"),
		[]byte("Subject: this is\r\n\ta multiline field\r\n"),
		[]byte("\r\n"),
	}, HeaderLines([]byte(header)))
}

func TestHeaderLinesMultilineFilename(t *testing.T) {
	const header = "Content-Type: application/msword; name=\"this is a very long\nfilename.doc\""

	assert.Equal(t, [][]byte{
		[]byte("Content-Type: application/msword; name=\"this is a very long\nfilename.doc\""),
	}, HeaderLines([]byte(header)))
}

func TestHeaderLinesMultilineFilenameWithColon(t *testing.T) {
	const header = "Content-Type: application/msword; name=\"this is a very long\nfilename: too long.doc\""

	assert.Equal(t, [][]byte{
		[]byte("Content-Type: application/msword; name=\"this is a very long\nfilename: too long.doc\""),
	}, HeaderLines([]byte(header)))
}

func TestHeaderLinesMultilineFilenameWithColonAndNewline(t *testing.T) {
	const header = "Content-Type: application/msword; name=\"this is a very long\nfilename: too long.doc\"\n"

	assert.Equal(t, [][]byte{
		[]byte("Content-Type: application/msword; name=\"this is a very long\nfilename: too long.doc\"\n"),
	}, HeaderLines([]byte(header)))
}

func TestHeaderLinesMultipleMultilineFilenames(t *testing.T) {
	const header = `Content-Type: application/msword; name="=E5=B8=B6=E6=9C=89=E5=A4=96=E5=9C=8B=E5=AD=97=E7=AC=A6=E7=9A=84=E9=99=84=E4=
=BB=B6.DOC"
Content-Transfer-Encoding: base64
Content-Disposition: attachment; filename="=E5=B8=B6=E6=9C=89=E5=A4=96=E5=9C=8B=E5=AD=97=E7=AC=A6=E7=9A=84=E9=99=84=E4=
=BB=B6.DOC"
Content-ID: <>
`

	assert.Equal(t, [][]byte{
		[]byte("Content-Type: application/msword; name=\"=E5=B8=B6=E6=9C=89=E5=A4=96=E5=9C=8B=E5=AD=97=E7=AC=A6=E7=9A=84=E9=99=84=E4=\n=BB=B6.DOC\"\n"),
		[]byte("Content-Transfer-Encoding: base64\n"),
		[]byte("Content-Disposition: attachment; filename=\"=E5=B8=B6=E6=9C=89=E5=A4=96=E5=9C=8B=E5=AD=97=E7=AC=A6=E7=9A=84=E9=99=84=E4=\n=BB=B6.DOC\"\n"),
		[]byte("Content-ID: <>\n"),
	}, HeaderLines([]byte(header)))
}
