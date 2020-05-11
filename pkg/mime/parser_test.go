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

package pmmime

import (
	"bytes"
	"fmt"

	"io/ioutil"
	"net/mail"

	"net/textproto"
	"strings"
	"testing"
)

func minimalParse(mimeBody string) (readBody string, plainContents string, err error) {
	mm, err := mail.ReadMessage(strings.NewReader(mimeBody))
	if err != nil {
		return
	}

	h := textproto.MIMEHeader(mm.Header)
	mmBodyData, err := ioutil.ReadAll(mm.Body)
	if err != nil {
		return
	}

	printAccepter := NewMIMEPrinter()
	plainTextCollector := NewPlainTextCollector(printAccepter)
	visitor := NewMimeVisitor(plainTextCollector)
	err = VisitAll(bytes.NewReader(mmBodyData), h, visitor)

	readBody = printAccepter.String()
	plainContents = plainTextCollector.GetPlainText()

	return readBody, plainContents, err
}

func androidParse(mimeBody string) (body, headers string, atts, attHeaders []string, err error) {
	mm, err := mail.ReadMessage(strings.NewReader(mimeBody))
	if err != nil {
		return
	}

	h := textproto.MIMEHeader(mm.Header)
	mmBodyData, err := ioutil.ReadAll(mm.Body)

	printAccepter := NewMIMEPrinter()
	bodyCollector := NewBodyCollector(printAccepter)
	attachmentsCollector := NewAttachmentsCollector(bodyCollector)
	mimeVisitor := NewMimeVisitor(attachmentsCollector)
	err = VisitAll(bytes.NewReader(mmBodyData), h, mimeVisitor)

	body, _ = bodyCollector.GetBody()
	headers = bodyCollector.GetHeaders()
	atts = attachmentsCollector.GetAttachments()
	attHeaders = attachmentsCollector.GetAttHeaders()

	return
}

func TestParseBoundaryIsEmpty(t *testing.T) {
	testMessage :=
		`Date: Sun, 10 Mar 2019 11:10:06 -0600
In-Reply-To: <abcbase64@protonmail.com>
X-Original-To: enterprise@protonmail.com
References: <abc64@unicoderns.com> <abc63@protonmail.com> <abc64@protonmail.com> <abc65@mail.gmail.com> <abc66@protonmail.com>
To: "ProtonMail" <enterprise@protonmail.com>
X-Pm-Origin: external
Delivered-To: enterprise@protonmail.com
Content-Type: multipart/mixed; boundary=ac7e36bd45425e70b4dab2128f34172e4dc3f9ff2eeb47e909267d4252794ec7
Reply-To: XYZ <xyz@xyz.com>
Mime-Version: 1.0
Subject: Encrypted Message
Return-Path: <xyz@xyz.com>
From: XYZ <xyz@xyz.com>
X-Pm-ConversationID-Id: gNX9bDPLmBgFZ-C3Tdlb628cas1Xl0m4dql5nsWzQAEI-WQv0ytfwPR4-PWELEK0_87XuFOgetc239Y0pjPYHQ==
X-Pm-Date: Sun, 10 Mar 2019 18:10:06 +0100
Message-Id: <68c11e46-e611-d9e4-edc1-5ec96bac77cc@unicoderns.com>
X-Pm-Transfer-Encryption: TLSv1.2 with cipher ECDHE-RSA-AES256-GCM-SHA384 (256/256 bits)
X-Pm-External-Id: <68c11e46-e611-d9e4-edc1-5ec96bac77cc@unicoderns.com>
X-Pm-Internal-Id: _iJ8ETxcqXTSK8IzCn0qFpMUTwvRf-xJUtldRA1f6yHdmXjXzKleG3F_NLjZL3FvIWVHoItTxOuuVXcukwwW3g==
Openpgp: preference=signencrypt
User-Agent: Mozilla/5.0 (X11; Linux x86_64; rv:60.0) Gecko/20100101 Thunderbird/60.4.0
X-Pm-Content-Encryption: end-to-end

--ac7e36bd45425e70b4dab2128f34172e4dc3f9ff2eeb47e909267d4252794ec7
Content-Disposition: inline
Content-Transfer-Encoding: quoted-printable
Content-Type: multipart/mixed; charset=utf-8

Content-Type: multipart/mixed; boundary="xnAIW3Turb9YQZ2rXc2ZGZH45WepHIZyy";
 protected-headers="v1"
From: XYZ <xyz@xyz.com>
To: "ProtonMail" <enterprise@protonmail.com>
Subject: Encrypted Message
Message-ID: <68c11e46-e611-d9e4-edc1-5ec96bac77cc@unicoderns.com>
References: <abc64@unicoderns.com> <abc63@protonmail.com> <abc64@protonmail.com> <abc65@mail.gmail.com> <abc66@protonmail.com>
In-Reply-To: <abcbase64@protonmail.com>

--xnAIW3Turb9YQZ2rXc2ZGZH45WepHIZyy
Content-Type: text/rfc822-headers; protected-headers="v1"
Content-Disposition: inline

From: XYZ <xyz@xyz.com>
To: ProtonMail <enterprise@protonmail.com>
Subject: Re: Encrypted Message

--xnAIW3Turb9YQZ2rXc2ZGZH45WepHIZyy
Content-Type: multipart/alternative;
 boundary="------------F9E5AA6D49692F51484075E3"
Content-Language: en-US

This is a multi-part message in MIME format.
--------------F9E5AA6D49692F51484075E3
Content-Type: text/plain; charset=utf-8
Content-Transfer-Encoding: quoted-printable

Hi ...

--------------F9E5AA6D49692F51484075E3
Content-Type: text/html; charset=utf-8
Content-Transfer-Encoding: quoted-printable

<html>
  <head>
  </head>
  <body text=3D"#000000" bgcolor=3D"#FFFFFF">
    <p>Hi ..  </p>
  </body>
</html>

--------------F9E5AA6D49692F51484075E3--

--xnAIW3Turb9YQZ2rXc2ZGZH45WepHIZyy--

--ac7e36bd45425e70b4dab2128f34172e4dc3f9ff2eeb47e909267d4252794ec7--


`

	body, content, err := minimalParse(testMessage)
	if err == nil {
		t.Fatal("should have error but is", err)
	}
	t.Log("==BODY==")
	t.Log(body)
	t.Log("==CONTENT==")
	t.Log(content)
}

func TestParse(t *testing.T) {
	testMessage :=
		`From: John Doe <example@example.com>
MIME-Version: 1.0
Content-Type: multipart/mixed;
        boundary="XXXXboundary text"

This is a multipart message in MIME format.

--XXXXboundary text
Content-Type: text/plain; charset=utf-8

this is the body text

--XXXXboundary text
Content-Type: text/html; charset=utf-8

<html><body>this is the html body text</body></html>

--XXXXboundary text
Content-Type: text/plain; charset=utf-8
Content-Disposition: attachment;
        filename="test.txt"

this is the attachment text

--XXXXboundary text--


`
	body, heads, att, attHeads, err := androidParse(testMessage)
	if err != nil {
		t.Error("parse error", err)
	}

	fmt.Println("==BODY:")
	fmt.Println(body)
	fmt.Println("==BODY HEADERS:")
	fmt.Println(heads)

	fmt.Println("==ATTACHMENTS:")
	fmt.Println(att)
	fmt.Println("==ATTACHMENT HEADERS:")
	fmt.Println(attHeads)
}

func TestParseAddressComment(t *testing.T) {
	parsingExamples := map[string]string{
		"":                          "",
		"(Only Comment) here@pm.me": "\"Only Comment\" <here@pm.me>",
		"Normal Name (With Comment) <here@pm.me>":          "\"Normal Name\" <here@pm.me>",
		"<Muhammed.(I am the greatest)Ali@(the)Vegas.WBA>": "\"I am the greatest the\" <Muhammed.Ali@Vegas.WBA>",
	}

	for raw, expected := range parsingExamples {
		parsed := parseAddressComment(raw)
		if expected != parsed {
			t.Errorf("When parsing %q expected %q but have %q", raw, expected, parsed)
		}
	}
}
