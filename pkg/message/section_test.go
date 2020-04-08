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

package message

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func debug(msg string, v ...interface{}) {
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
		contentType := sec.header.Get("Content-Type")
		debug("%10s: %-50s %5d %5d %5d %5d", path, contentType, sec.start, sec.size, sec.bsize, sec.lines)
		require.Equal(t, expectedStructure[path], contentType)
	}

	require.True(t, len(*bs) == len(expectedStructure), "Wrong number of sections expected %d but have %d", len(expectedStructure), len(*bs))
}

func TestGetSection(t *testing.T) {
	structReader := strings.NewReader(sampleMail)
	bs, err := NewBodyStructure(structReader)
	require.NoError(t, err)
	// Whole section.
	for _, try := range testPaths {
		mailReader := strings.NewReader(sampleMail)
		info, err := bs.getInfo(try.path)
		require.NoError(t, err)
		section, err := bs.GetSection(mailReader, try.path)
		require.NoError(t, err)

		debug("section %v: %d %d\n___\n%s\n‾‾‾\n", try.path, info.start, info.size, string(section))

		require.True(t, string(section) == try.expectedSection, "not same as expected:\n___\n%s\n‾‾‾", try.expectedSection)
	}
	// Body content.
	for _, try := range testPaths {
		mailReader := strings.NewReader(sampleMail)
		info, err := bs.getInfo(try.path)
		require.NoError(t, err)
		section, err := bs.GetSectionContent(mailReader, try.path)
		require.NoError(t, err)

		debug("content %v: %d %d\n___\n%s\n‾‾‾\n", try.path, info.start+info.size-info.bsize, info.bsize, string(section))

		require.True(t, string(section) == try.expectedBody, "not same as expected:\n___\n%s\n‾‾‾", try.expectedBody)
	}
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
	{[]int{},
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

	{[]int{1},
		`Content-Type: text/plain

1. main message


`,
		`1. main message


`,
	},
	{[]int{3},
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
	{[]int{3, 1},
		`Content-Type: text/plain

3.1 message text

`,
		`3.1 message text

`,
	},
	{[]int{3, 2},
		`Content-Type: application/octet-stream
Content-Disposition: attachment; filename="msg_3_signature.sig"
Content-Transfer-Encoding: base64

3/2/MessageOctestStream/==

`,
		`3/2/MessageOctestStream/==

`,
	},
	{[]int{4, 2, 2, 1},
		`Content-Type: text/plain

4.2.2.1 plain text

`,
		`4.2.2.1 plain text

`,
	},
	{[]int{4, 2, 2, 2},
		`Content-Type: text/html

<h1>4.2.2.2 html text</h1>

`,
		`<h1>4.2.2.2 html text</h1>

`,
	},
}
