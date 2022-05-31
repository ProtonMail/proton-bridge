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
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var enableDebug = false //nolint:global

func debug(msg string, v ...interface{}) {
	if !enableDebug {
		return
	}

	_, file, line, _ := runtime.Caller(1)
	fmt.Printf("%s:%d: \033[2;33m"+msg+"\033[0;39m\n", append([]interface{}{filepath.Base(file), line}, v...)...)
}

func TestParseBodyStructure(t *testing.T) {
	expectedStructure := map[string]string{
		"":        "multipart/mixed; boundary=\"0000MAIN\"",
		"1":       "text/plain",
		"2":       "application/octet-stream",
		"3":       "message/rfc822; boundary=\"0003MSG\"",
		"3.1":     "text/plain",
		"3.2":     "application/octet-stream",
		"4":       "multipart/mixed; boundary=\"0004ATTACH\"",
		"4.1":     "image/gif",
		"4.2":     "message/rfc822; boundary=\"0042MSG\"",
		"4.2.1":   "text/plain",
		"4.2.2":   "multipart/alternative; boundary=\"0422ALTER\"",
		"4.2.2.1": "text/plain",
		"4.2.2.2": "text/html",
	}
	mailReader := strings.NewReader(sampleMail)
	bs, err := NewBodyStructure(mailReader)
	require.NoError(t, err)

	paths := []string{}
	for path := range *bs {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	debug("%10s: %-50s %5s %5s %5s %5s", "section", "type", "start", "size", "bsize", "lines")
	for _, path := range paths {
		sec := (*bs)[path]
		header, err := sec.GetMIMEHeader()
		require.NoError(t, err)
		contentType := header.Get("Content-Type")
		debug("%10s: %-50s %5d %5d %5d %5d", path, contentType, sec.Start, sec.Size, sec.BSize, sec.Lines)
		require.Equal(t, expectedStructure[path], contentType)
	}

	require.True(t, len(*bs) == len(expectedStructure), "Wrong number of sections expected %d but have %d", len(expectedStructure), len(*bs))
}

func TestParseBodyStructurePGP(t *testing.T) {
	expectedStructure := map[string]string{
		"":        "multipart/signed; micalg=pgp-sha256; protocol=\"application/pgp-signature\"; boundary=\"MHEDFShwcX18dyE3X7RXujo5fjpgdjHNM\"",
		"1":       "multipart/mixed; boundary=\"FBBl2LNv76z8UkvHhSkT9vLwVwxqV8378\"; protected-headers=\"v1\"",
		"1.1":     "multipart/mixed; boundary=\"------------F97C8ED4878E94675762AE43\"",
		"1.1.1":   "multipart/alternative; boundary=\"------------041318B15DD3FA540FED32C6\"",
		"1.1.1.1": "text/plain; charset=utf-8; format=flowed",
		"1.1.1.2": "text/html; charset=utf-8",
		"1.1.2":   "application/pdf; name=\"minimal.pdf\"",
		"1.1.3":   "application/pgp-keys; name=\"OpenPGP_0x161C0875822359F7.asc\"",
		"2":       "application/pgp-signature; name=\"OpenPGP_signature.asc\"",
	}

	b, err := ioutil.ReadFile("testdata/enc-body-structure.eml")
	require.NoError(t, err)

	bs, err := NewBodyStructure(bytes.NewReader(b))
	require.NoError(t, err)

	haveStructure := map[string]string{}
	for path := range *bs {
		header, err := (*bs)[path].GetMIMEHeader()
		require.NoError(t, err)
		haveStructure[path] = header.Get("Content-Type")
	}

	require.Equal(t, expectedStructure, haveStructure)
}

func TestGetSection(t *testing.T) {
	structReader := strings.NewReader(sampleMail)
	bs, err := NewBodyStructure(structReader)
	require.NoError(t, err)

	// Bad paths
	wantPaths := [][]int{{0}, {-1}, {3, 2, 3}}
	for _, wantPath := range wantPaths {
		_, err = bs.getInfo(wantPath)
		require.Error(t, err, "path %v", wantPath)
	}

	// Whole section.
	for _, try := range testPaths {
		mailReader := strings.NewReader(sampleMail)
		info, err := bs.getInfo(try.path)
		require.NoError(t, err)
		section, err := bs.GetSection(mailReader, try.path)
		require.NoError(t, err)

		debug("section %v: %d %d\n___\n%s\n‾‾‾\n", try.path, info.Start, info.Size, string(section))

		require.True(t, string(section) == try.expectedSection, "not same as expected:\n___\n%s\n‾‾‾", try.expectedSection)
	}
	// Body content.
	for _, try := range testPaths {
		mailReader := strings.NewReader(sampleMail)
		info, err := bs.getInfo(try.path)
		require.NoError(t, err)
		section, err := bs.GetSectionContent(mailReader, try.path)
		require.NoError(t, err)

		debug("content %v: %d %d\n___\n%s\n‾‾‾\n", try.path, info.Start+info.Size-info.BSize, info.BSize, string(section))

		require.True(t, string(section) == try.expectedBody, "not same as expected:\n___\n%s\n‾‾‾", try.expectedBody)
	}
}

func TestGetSecionNoMIMEParts(t *testing.T) {
	wantBody := "This is just a simple mail with no multipart structure.\n"
	wantHeader := `Subject: Sample mail
From: John Doe <jdoe@machine.example>
To: Mary Smith <mary@example.net>
Date: Fri, 21 Nov 1997 09:55:06 -0600
Content-Type: plain/text

`
	wantMail := wantHeader + wantBody

	r := require.New(t)
	bs, err := NewBodyStructure(strings.NewReader(wantMail))
	r.NoError(err)

	// Bad parts
	wantPaths := [][]int{{0}, {2}, {1, 2, 3}}
	for _, wantPath := range wantPaths {
		_, err = bs.getInfoCheckSection(wantPath)
		r.Error(err, "path %v: %d %d\n__\n%s\n", wantPath)
	}

	debug := func(wantPath []int, info *SectionInfo, section []byte) string {
		if info == nil {
			info = &SectionInfo{}
		}
		return fmt.Sprintf("path %v %q: %d %d\n___\n%s\n‾‾‾\n",
			wantPath, stringPathFromInts(wantPath), info.Start, info.Size,
			string(section),
		)
	}

	// Ok Parts
	wantPaths = [][]int{{}, {1}}
	for _, p := range wantPaths {
		wantPath := append([]int{}, p...)

		info, err := bs.getInfoCheckSection(wantPath)
		r.NoError(err, debug(wantPath, info, []byte{}))

		section, err := bs.GetSection(strings.NewReader(wantMail), wantPath)
		r.NoError(err, debug(wantPath, info, section))
		r.Equal(wantMail, string(section), debug(wantPath, info, section))

		haveBody, err := bs.GetSectionContent(strings.NewReader(wantMail), wantPath)
		r.NoError(err, debug(wantPath, info, haveBody))
		r.Equal(wantBody, string(haveBody), debug(wantPath, info, haveBody))

		haveHeader, err := bs.GetSectionHeaderBytes(wantPath)
		r.NoError(err, debug(wantPath, info, haveHeader))
		r.Equal(wantHeader, string(haveHeader), debug(wantPath, info, haveHeader))
	}
}

func TestGetMainHeaderBytes(t *testing.T) {
	wantHeader := []byte(`Subject: Sample mail
From: John Doe <jdoe@machine.example>
To: Mary Smith <mary@example.net>
Date: Fri, 21 Nov 1997 09:55:06 -0600
Content-Type: multipart/mixed; boundary="0000MAIN"

`)

	structReader := strings.NewReader(sampleMail)
	bs, err := NewBodyStructure(structReader)
	require.NoError(t, err)

	haveHeader, err := bs.GetMailHeaderBytes()
	require.NoError(t, err)
	require.Equal(t, wantHeader, haveHeader)
}

/* Structure example:
HEADER     ([RFC-2822] header of the message)
TEXT       ([RFC-2822] text body of the message) MULTIPART/MIXED
1          TEXT/PLAIN
2          APPLICATION/OCTET-STREAM
3          MESSAGE/RFC822
3.HEADER   ([RFC-2822] header of the message)
3.TEXT     ([RFC-2822] text body of the message) MULTIPART/MIXED
3.1        TEXT/PLAIN
3.2        APPLICATION/OCTET-STREAM
4          MULTIPART/MIXED
4.1        IMAGE/GIF
4.1.MIME   ([MIME-IMB] header for the IMAGE/GIF)
4.2        MESSAGE/RFC822
4.2.HEADER ([RFC-2822] header of the message)
4.2.TEXT   ([RFC-2822] text body of the message) MULTIPART/MIXED
4.2.1      TEXT/PLAIN
4.2.2      MULTIPART/ALTERNATIVE
4.2.2.1    TEXT/PLAIN
4.2.2.2    TEXT/RICHTEXT
*/

var sampleMail = `Subject: Sample mail
From: John Doe <jdoe@machine.example>
To: Mary Smith <mary@example.net>
Date: Fri, 21 Nov 1997 09:55:06 -0600
Content-Type: multipart/mixed; boundary="0000MAIN"

main summary

--0000MAIN
Content-Type: text/plain

1. main message


--0000MAIN
Content-Type: application/octet-stream
Content-Disposition: inline; filename="main_signature.sig"
Content-Transfer-Encoding: base64

2/MainOctetStream

--0000MAIN
Subject: Inside mail 3
From: Mary Smith <mary@example.net>
To: John Doe <jdoe@machine.example>
Date: Fri, 20 Nov 1997 09:55:06 -0600
Content-Type: message/rfc822; boundary="0003MSG"

3. message summary

--0003MSG
Content-Type: text/plain

3.1 message text

--0003MSG
Content-Type: application/octet-stream
Content-Disposition: attachment; filename="msg_3_signature.sig"
Content-Transfer-Encoding: base64

3/2/MessageOctestStream/==

--0003MSG--

--0000MAIN
Content-Type: multipart/mixed; boundary="0004ATTACH"

4 attach summary

--0004ATTACH
Content-Type: image/gif
Content-Disposition: attachment; filename="att4.1_gif.sig"
Content-Transfer-Encoding: base64

4/1/Gif=

--0004ATTACH
Subject: Inside mail 4.2
From: Mary Smith <mary@example.net>
To: John Doe <jdoe@machine.example>
Date: Fri, 10 Nov 1997 09:55:06 -0600
Content-Type: message/rfc822; boundary="0042MSG"

4.2 message summary

--0042MSG
Content-Type: text/plain

4.2.1 message text

--0042MSG
Content-Type: multipart/alternative; boundary="0422ALTER"

4.2.2 alternative summary

--0422ALTER
Content-Type: text/plain

4.2.2.1 plain text

--0422ALTER
Content-Type: text/html

<h1>4.2.2.2 html text</h1>

--0422ALTER--

--0042MSG--

--0004ATTACH--

--0000MAIN--


`

var testPaths = []struct {
	path                          []int
	expectedSection, expectedBody string
}{
	{
		[]int{},
		sampleMail,
		`main summary

--0000MAIN
Content-Type: text/plain

1. main message


--0000MAIN
Content-Type: application/octet-stream
Content-Disposition: inline; filename="main_signature.sig"
Content-Transfer-Encoding: base64

2/MainOctetStream

--0000MAIN
Subject: Inside mail 3
From: Mary Smith <mary@example.net>
To: John Doe <jdoe@machine.example>
Date: Fri, 20 Nov 1997 09:55:06 -0600
Content-Type: message/rfc822; boundary="0003MSG"

3. message summary

--0003MSG
Content-Type: text/plain

3.1 message text

--0003MSG
Content-Type: application/octet-stream
Content-Disposition: attachment; filename="msg_3_signature.sig"
Content-Transfer-Encoding: base64

3/2/MessageOctestStream/==

--0003MSG--

--0000MAIN
Content-Type: multipart/mixed; boundary="0004ATTACH"

4 attach summary

--0004ATTACH
Content-Type: image/gif
Content-Disposition: attachment; filename="att4.1_gif.sig"
Content-Transfer-Encoding: base64

4/1/Gif=

--0004ATTACH
Subject: Inside mail 4.2
From: Mary Smith <mary@example.net>
To: John Doe <jdoe@machine.example>
Date: Fri, 10 Nov 1997 09:55:06 -0600
Content-Type: message/rfc822; boundary="0042MSG"

4.2 message summary

--0042MSG
Content-Type: text/plain

4.2.1 message text

--0042MSG
Content-Type: multipart/alternative; boundary="0422ALTER"

4.2.2 alternative summary

--0422ALTER
Content-Type: text/plain

4.2.2.1 plain text

--0422ALTER
Content-Type: text/html

<h1>4.2.2.2 html text</h1>

--0422ALTER--

--0042MSG--

--0004ATTACH--

--0000MAIN--


`,
	},

	{
		[]int{1},
		`Content-Type: text/plain

1. main message


`,
		`1. main message


`,
	},
	{
		[]int{3},
		`Subject: Inside mail 3
From: Mary Smith <mary@example.net>
To: John Doe <jdoe@machine.example>
Date: Fri, 20 Nov 1997 09:55:06 -0600
Content-Type: message/rfc822; boundary="0003MSG"

3. message summary

--0003MSG
Content-Type: text/plain

3.1 message text

--0003MSG
Content-Type: application/octet-stream
Content-Disposition: attachment; filename="msg_3_signature.sig"
Content-Transfer-Encoding: base64

3/2/MessageOctestStream/==

--0003MSG--

`,
		`3. message summary

--0003MSG
Content-Type: text/plain

3.1 message text

--0003MSG
Content-Type: application/octet-stream
Content-Disposition: attachment; filename="msg_3_signature.sig"
Content-Transfer-Encoding: base64

3/2/MessageOctestStream/==

--0003MSG--

`,
	},
	{
		[]int{3, 1},
		`Content-Type: text/plain

3.1 message text

`,
		`3.1 message text

`,
	},
	{
		[]int{3, 2},
		`Content-Type: application/octet-stream
Content-Disposition: attachment; filename="msg_3_signature.sig"
Content-Transfer-Encoding: base64

3/2/MessageOctestStream/==

`,
		`3/2/MessageOctestStream/==

`,
	},
	{
		[]int{4, 2, 2, 1},
		`Content-Type: text/plain

4.2.2.1 plain text

`,
		`4.2.2.1 plain text

`,
	},
	{
		[]int{4, 2, 2, 2},
		`Content-Type: text/html

<h1>4.2.2.2 html text</h1>

`,
		`<h1>4.2.2.2 html text</h1>

`,
	},
}

func TestBodyStructureSerialize(t *testing.T) {
	r := require.New(t)
	want := &BodyStructure{
		"1": {
			Header: []byte("Content: type"),
			Start:  1,
			Size:   2,
			BSize:  3,
			Lines:  4,
		},
		"1.1.1": {
			Header: []byte("X-Pm-Key: id"),
			Start:  11,
			Size:   12,
			BSize:  13,
			Lines:  14,
			reader: bytes.NewBuffer([]byte("this should not be serialized")),
		},
	}

	raw, err := want.Serialize()
	r.NoError(err)
	have, err := DeserializeBodyStructure(raw)
	r.NoError(err)

	// Before compare remove reader (should not be serialized)
	(*want)["1.1.1"].reader = nil
	r.Equal(want, have)
}

func TestSectionInfoReadHeader(t *testing.T) {
	r := require.New(t)

	testData := []struct {
		wantHeader, mail string
	}{
		{
			"key1: val1\nkey2: val2\n\n",
			"key1: val1\nkey2: val2\n\nbody is here\n\nand it is not confused",
		},
		{
			"key1:\n val1\n\n",
			"key1:\n val1\n\nbody is here",
		},
		{
			"key1: val1\r\nkey2: val2\r\n\r\n",
			"key1: val1\r\nkey2: val2\r\n\r\nbody is here\r\n\r\nand it is not confused",
		},
	}

	for _, td := range testData {
		bs, err := NewBodyStructure(strings.NewReader(td.mail))
		r.NoError(err, "case %q", td.mail)
		haveHeader, err := bs.GetMailHeaderBytes()
		r.NoError(err, "case %q", td.mail)
		r.Equal(td.wantHeader, string(haveHeader), "case %q", td.mail)
	}
}
