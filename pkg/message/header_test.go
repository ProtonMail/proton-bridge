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
	"regexp"
	"strings"
	"testing"

	gomessage "github.com/emersion/go-message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestHeaderOrder(t *testing.T) {
	literal := []byte(`X-Pm-Content-Encryption: end-to-end
X-Pm-Origin: internal
Subject: header test
To: Test Proton <test@proton.me>
From: Dummy Recipient <dummy@proton.me>
Date: Tue, 15 Oct 2024 07:54:39 +0000
Mime-Version: 1.0
Content-Type: multipart/mixed;boundary=---------------------a136fc3851075ca3f022f5c3ec6bf8f5
Message-Id: <1rYR51zNVZdyCXVvAZ8C9N8OaBg4wO_wg6VlSoLK_Mv-2AaiF5UL-vE_tIZ6FdYP8ylsuV3fpaKUpVwuUcnQ6ql_83aEgZvfC5QcZbind1k=@proton.me>
X-Pm-Spamscore: 0
Received: from mail.protonmail.ch by mail.protonmail.ch; Tue, 15 Oct 2024 07:54:43 +0000
X-Original-To: test@proton.me
Return-Path: <dummy@proton.me>
Delivered-To: test@proton.me

lorem`)

	// build a proton message
	message := newTestMessageFromRFC822(t, literal)
	options := JobOptions{
		IgnoreDecryptionErrors: true,
		SanitizeDate:           true,
		AddInternalID:          true,
		AddExternalID:          true,
		AddMessageDate:         true,
		AddMessageIDReference:  true,
		SanitizeMBOXHeaderLine: true,
	}

	// Rebuild the headers using bridge's algorithm, sanitizing fields.
	hdr := getTextPartHeader(getMessageHeader(message, options), []byte(message.Body), message.MIMEType)
	var b bytes.Buffer
	w, err := gomessage.CreateWriter(&b, hdr)
	require.NoError(t, err)
	_ = w.Close()

	// split the header
	str := string(regexp.MustCompile(`\r\n(\s+)`).ReplaceAll(b.Bytes(), nil)) // join multi
	lines := strings.Split(str, "\r\n")

	// Check we have the expected order
	require.Equal(t, len(lines), 17)

	// The fields added or modified are at the top
	require.True(t, strings.HasPrefix(lines[0], "Content-Type: multipart/mixed;boundary=")) // we changed the boundary
	require.True(t, strings.HasPrefix(lines[1], "References: "))                            // Reference was added
	require.True(t, strings.HasPrefix(lines[2], "X-Pm-Date: "))                             // X-Pm-Date was added
	require.True(t, strings.HasPrefix(lines[3], "X-Pm-Internal-Id: "))                      // X-Pm-Internal-Id was added
	require.Equal(t, `To: "Test Proton" <test@proton.me>`, lines[4])                        // Name was double quoted
	require.Equal(t, `From: "Dummy Recipient" <dummy@proton.me>`, lines[5])                 // Name was double quoted

	// all other fields appear in their original order
	require.Equal(t, `X-Pm-Content-Encryption: end-to-end`, lines[6])
	require.Equal(t, `X-Pm-Origin: internal`, lines[7])
	require.Equal(t, `Subject: header test`, lines[8])
	require.Equal(t, `Date: Tue, 15 Oct 2024 07:54:39 +0000`, lines[9])
	require.Equal(t, `Mime-Version: 1.0`, lines[10])
	require.Equal(t, `Message-Id: <1rYR51zNVZdyCXVvAZ8C9N8OaBg4wO_wg6VlSoLK_Mv-2AaiF5UL-vE_tIZ6FdYP8ylsuV3fpaKUpVwuUcnQ6ql_83aEgZvfC5QcZbind1k=@proton.me>`, lines[11])
	require.Equal(t, `X-Pm-Spamscore: 0`, lines[12])
	require.Equal(t, `Received: from mail.protonmail.ch by mail.protonmail.ch; Tue, 15 Oct 2024 07:54:43 +0000`, lines[13])
	require.Equal(t, `X-Original-To: test@proton.me`, lines[14])
	require.Equal(t, `Return-Path: <dummy@proton.me>`, lines[15])
	require.Equal(t, `Delivered-To: test@proton.me`, lines[16])
}
