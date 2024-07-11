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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeaderLines(t *testing.T) {
	want := [][]byte{
		[]byte("To: somebody\r\n"),
		[]byte("From: somebody else\r\n"),
		[]byte("Subject: RE: this is\r\n\ta multiline field: with colon\r\n\tor: many: more: colons\r\n"),
		[]byte("X-Special: \r\n\tNothing on the first line\r\n\tbut has something on the other lines\r\n"),
		[]byte("\r\n"),
	}
	var header []byte
	for _, line := range want {
		header = append(header, line...)
	}

	assert.Equal(t, want, HeaderLines(header))
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

func TestReadHeaderBody(t *testing.T) {
	const data = "key: value\r\n\r\nbody\n"

	header, body, err := readHeaderBody([]byte(data))

	assert.NoError(t, err)
	assert.Equal(t, 1, header.Len())
	assert.Equal(t, "value", header.Get("key"))
	assert.Equal(t, []byte("body\n"), body)
}

func TestReadHeaderBodyWithoutHeader(t *testing.T) {
	const data = "body\n"

	header, body, err := readHeaderBody([]byte(data))

	assert.NoError(t, err)
	assert.Equal(t, 0, header.Len())
	assert.Equal(t, []byte(data), body)
}

func TestReadHeaderBodyInvalidHeader(t *testing.T) {
	const data = "value\r\n\r\nbody\n"

	header, body, err := readHeaderBody([]byte(data))

	assert.NoError(t, err)
	assert.Equal(t, 0, header.Len())
	assert.Equal(t, []byte(data), body)
}

func FuzzReadHeaderBody(f *testing.F) {
	header := `Content-Type: application/msword; name="=E5=B8=B6=E6=9C=89=E5=A4=96=E5=9C=8B=E5=AD=97=E7=AC=A6=E7=9A=84=E9=99=84=E4=
	=BB=B6.DOC"
	Content-Transfer-Encoding: base64
	Content-Disposition: attachment; filename="=E5=B8=B6=E6=9C=89=E5=A4=96=E5=9C=8B=E5=AD=97=E7=AC=A6=E7=9A=84=E9=99=84=E4=
	=BB=B6.DOC"
	Content-ID: <>
	`
	data0 := "key: value\r\n\r\nbody\n"
	data1 := "key: value\r\n\r\nbody\n"

	f.Add([]byte(header))
	f.Add([]byte(data0))
	f.Add([]byte(data1))

	f.Fuzz(func(_ *testing.T, b []byte) {
		_, _, _ = readHeaderBody(b)
	})
}
