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
	"image/png"
	"io"
	"io/ioutil"
	"math/rand"
	"net/mail"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/encoding/charmap"
)

func TestRFC822AddressFormat(t *testing.T) { //nolint[funlen]
	tests := []struct {
		address  string
		expected []string
	}{
		{
			" normal name  <username@server.com>",
			[]string{
				"\"normal name\" <username@server.com>",
			},
		},
		{
			" \"comma, name\"  <username@server.com>",
			[]string{
				"\"comma, name\" <username@server.com>",
			},
		},
		{
			" name  <username@server.com> (ignore comment)",
			[]string{
				"\"name\" <username@server.com>",
			},
		},
		{
			" name (ignore comment)  <username@server.com>,  (Comment as name) username2@server.com",
			[]string{
				"\"name\" <username@server.com>",
				"<username2@server.com>",
			},
		},
		{
			" normal name  <username@server.com>, (comment)All.(around)address@(the)server.com",
			[]string{
				"\"normal name\" <username@server.com>",
				"<All.address@server.com>",
			},
		},
		{
			" normal name  <username@server.com>, All.(\"comma, in comment\")address@(the)server.com",
			[]string{
				"\"normal name\" <username@server.com>",
				"<All.address@server.com>",
			},
		},
		{
			" \"normal name\"  <username@server.com>, \"comma, name\" <address@server.com>",
			[]string{
				"\"normal name\" <username@server.com>",
				"\"comma, name\" <address@server.com>",
			},
		},
		{
			" \"comma, one\"  <username@server.com>, \"comma, two\" <address@server.com>",
			[]string{
				"\"comma, one\" <username@server.com>",
				"\"comma, two\" <address@server.com>",
			},
		},
		{
			" \"comma, name\"  <username@server.com>, another, name <address@server.com>",
			[]string{
				"\"comma, name\" <username@server.com>",
				"\"another, name\" <address@server.com>",
			},
		},
	}

	for _, data := range tests {
		uncommented := parseAddressComment(data.address)
		result, err := mail.ParseAddressList(uncommented)
		if err != nil {
			t.Errorf("Can not parse '%s' created from '%s': %v", uncommented, data.address, err)
		}
		if len(result) != len(data.expected) {
			t.Errorf("Wrong parsing of '%s' created from '%s': expected '%s' but have '%+v'", uncommented, data.address, data.expected, result)
		}
		for i, result := range result {
			if data.expected[i] != result.String() {
				t.Errorf("Wrong parsing\nof %q\ncreated from %q:\nexpected %q\nbut have %q", uncommented, data.address, data.expected, result.String())
			}
		}
	}
}

func f(filename string) io.ReadCloser {
	f, err := os.Open(filepath.Join("testdata", filename))

	if err != nil {
		panic(err)
	}

	return f
}

func s(filename string) string {
	b, err := ioutil.ReadAll(f(filename))

	if err != nil {
		panic(err)
	}

	return string(b)
}

func readerToString(r io.Reader) string {
	b, err := ioutil.ReadAll(r)

	if err != nil {
		panic(err)
	}

	return string(b)
}

func TestParseMessageTextPlain(t *testing.T) {
	f := f("text_plain.eml")
	defer func() { _ = f.Close() }()

	m, mimeBody, plainContents, atts, err := Parse(f, "", "")
	assert.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", m.Body)
	assert.Equal(t, s("text_plain.mime"), mimeBody)
	assert.Equal(t, "body", plainContents)

	assert.Len(t, atts, 0)
}

func TestParseMessageTextPlainUTF8(t *testing.T) {
	f := f("text_plain_utf8.eml")
	defer func() { _ = f.Close() }()

	m, mimeBody, plainContents, atts, err := Parse(f, "", "")
	assert.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", m.Body)
	assert.Equal(t, s("text_plain_utf8.mime"), mimeBody)
	assert.Equal(t, "body", plainContents)

	assert.Len(t, atts, 0)
}

func TestParseMessageTextPlainLatin1(t *testing.T) {
	f := f("text_plain_latin1.eml")
	defer func() { _ = f.Close() }()

	m, mimeBody, plainContents, atts, err := Parse(f, "", "")
	assert.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "ééééééé", m.Body)
	assert.Equal(t, s("text_plain_latin1.mime"), mimeBody)
	assert.Equal(t, "ééééééé", plainContents)

	assert.Len(t, atts, 0)
}

func TestParseMessageTextPlainUnknownCharsetIsActuallyLatin1(t *testing.T) {
	f := f("text_plain_unknown_latin1.eml")
	defer func() { _ = f.Close() }()

	m, mimeBody, plainContents, atts, err := Parse(f, "", "")
	assert.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "ééééééé", m.Body)
	assert.Equal(t, s("text_plain_unknown_latin1.mime"), mimeBody)
	assert.Equal(t, "ééééééé", plainContents)

	assert.Len(t, atts, 0)
}

func TestParseMessageTextPlainUnknownCharsetIsActuallyLatin2(t *testing.T) {
	f := f("text_plain_unknown_latin2.eml")
	defer func() { _ = f.Close() }()

	m, mimeBody, plainContents, atts, err := Parse(f, "", "")
	assert.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	// The file contains latin2-encoded text, but we will assume it is latin1
	// and decode it as such. This will lead to corruption.
	latin2, _ := charmap.ISO8859_2.NewEncoder().Bytes([]byte("řšřšřš"))
	expect, _ := charmap.ISO8859_1.NewDecoder().Bytes(latin2)
	assert.NotEqual(t, []byte("řšřšřš"), expect)

	assert.Equal(t, string(expect), m.Body)
	assert.Equal(t, s("text_plain_unknown_latin2.mime"), mimeBody)
	assert.Equal(t, string(expect), plainContents)

	assert.Len(t, atts, 0)
}

func TestParseMessageTextPlainAlready7Bit(t *testing.T) {
	f := f("text_plain_7bit.eml")
	defer func() { _ = f.Close() }()

	m, mimeBody, plainContents, atts, err := Parse(f, "", "")
	assert.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", m.Body)
	assert.Equal(t, s("text_plain_7bit.mime"), mimeBody)
	assert.Equal(t, "body", plainContents)

	assert.Len(t, atts, 0)
}

func TestParseMessageTextPlainWithOctetAttachment(t *testing.T) {
	f := f("text_plain_octet_attachment.eml")
	defer func() { _ = f.Close() }()

	m, mimeBody, plainContents, atts, err := Parse(f, "", "")
	assert.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", m.Body)
	assert.Equal(t, s("text_plain_octet_attachment.mime"), mimeBody)
	assert.Equal(t, "body", plainContents)

	assert.Len(t, atts, 1)
	assert.Equal(t, readerToString(atts[0]), "if you are reading this, hi!")
}

func TestParseMessageTextPlainWithPlainAttachment(t *testing.T) {
	f := f("text_plain_plain_attachment.eml")
	defer func() { _ = f.Close() }()

	m, mimeBody, plainContents, atts, err := Parse(f, "", "")
	assert.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", m.Body)
	assert.Equal(t, s("text_plain_plain_attachment.mime"), mimeBody)
	assert.Equal(t, "body", plainContents)

	assert.Len(t, atts, 1)
	assert.Equal(t, readerToString(atts[0]), "attachment")
}

func TestParseMessageTextPlainWithImageInline(t *testing.T) {
	f := f("text_plain_image_inline.eml")
	defer func() { _ = f.Close() }()

	m, mimeBody, plainContents, atts, err := Parse(f, "", "")
	assert.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", m.Body)
	assert.Equal(t, s("text_plain_image_inline.mime"), mimeBody)
	assert.Equal(t, "body", plainContents)

	// The inline image is an 8x8 mic-dropping gopher.
	assert.Len(t, atts, 1)
	img, err := png.DecodeConfig(atts[0])
	assert.NoError(t, err)
	assert.Equal(t, 8, img.Width)
	assert.Equal(t, 8, img.Height)
}

func TestParseMessageWithMultipleTextParts(t *testing.T) {
	f := f("multiple_text_parts.eml")
	defer func() { _ = f.Close() }()

	m, mimeBody, plainContents, atts, err := Parse(f, "", "")
	assert.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body\nsome other part of the message", m.Body)
	assert.Equal(t, s("multiple_text_parts.mime"), mimeBody)
	assert.Equal(t, "body\nsome other part of the message", plainContents)

	assert.Len(t, atts, 0)
}

func TestParseMessageTextHTML(t *testing.T) {
	rand.Seed(0)

	f := f("text_html.eml")
	defer func() { _ = f.Close() }()

	m, mimeBody, plainContents, atts, err := Parse(f, "", "")
	assert.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "<html><head></head><body>This is body of <b>HTML mail</b> without attachment</body></html>", m.Body)
	assert.Equal(t, s("text_html.mime"), mimeBody)
	assert.Equal(t, "This is body of *HTML mail* without attachment", plainContents)

	assert.Len(t, atts, 0)
}

func TestParseMessageTextHTMLAlready7Bit(t *testing.T) {
	rand.Seed(0)

	f := f("text_html_7bit.eml")
	defer func() { _ = f.Close() }()

	m, mimeBody, plainContents, atts, err := Parse(f, "", "")
	assert.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "<html><head></head><body>This is body of <b>HTML mail</b> without attachment</body></html>", m.Body)
	assert.Equal(t, s("text_html_7bit.mime"), mimeBody)
	assert.Equal(t, "This is body of *HTML mail* without attachment", plainContents)

	assert.Len(t, atts, 0)
}

func TestParseMessageTextHTMLWithOctetAttachment(t *testing.T) {
	rand.Seed(0)

	f := f("text_html_octet_attachment.eml")
	defer func() { _ = f.Close() }()

	m, mimeBody, plainContents, atts, err := Parse(f, "", "")
	assert.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "<html><head></head><body>This is body of <b>HTML mail</b> with attachment</body></html>", m.Body)
	assert.Equal(t, s("text_html_octet_attachment.mime"), mimeBody)
	assert.Equal(t, "This is body of *HTML mail* with attachment", plainContents)

	assert.Len(t, atts, 1)
	assert.Equal(t, readerToString(atts[0]), "if you are reading this, hi!")
}

// NOTE: Enable when bug is fixed.
func _TestParseMessageTextHTMLWithPlainAttachment(t *testing.T) { // nolint[deadcode]
	rand.Seed(0)

	f := f("text_html_plain_attachment.eml")
	defer func() { _ = f.Close() }()

	m, mimeBody, plainContents, atts, err := Parse(f, "", "")
	assert.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	// BAD: plainContents should not be empty!
	assert.Equal(t, "<html><head></head><body>This is body of <b>HTML mail</b> with attachment</body></html>", m.Body)
	assert.Equal(t, s("text_html_plain_attachment.mime"), mimeBody)
	assert.Equal(t, "This is body of *HTML mail* with attachment", plainContents)

	assert.Len(t, atts, 1)
	assert.Equal(t, readerToString(atts[0]), "attachment")
}

func TestParseMessageTextHTMLWithImageInline(t *testing.T) {
	rand.Seed(0)

	f := f("text_html_image_inline.eml")
	defer func() { _ = f.Close() }()

	m, mimeBody, plainContents, atts, err := Parse(f, "", "")
	assert.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "<html><head></head><body>This is body of <b>HTML mail</b> with attachment</body></html>", m.Body)
	assert.Equal(t, s("text_html_image_inline.mime"), mimeBody)
	assert.Equal(t, "This is body of *HTML mail* with attachment", plainContents)

	// The inline image is an 8x8 mic-dropping gopher.
	assert.Len(t, atts, 1)
	img, err := png.DecodeConfig(atts[0])
	assert.NoError(t, err)
	assert.Equal(t, 8, img.Width)
	assert.Equal(t, 8, img.Height)
}

// NOTE: Enable when bug is fixed.
func _TestParseMessageWithAttachedPublicKey(t *testing.T) { // nolint[deadcode]
	f := f("text_plain.eml")
	defer func() { _ = f.Close() }()

	// BAD: Public Key is not attached unless Content-Type is specified (not required)!
	m, mimeBody, plainContents, atts, err := Parse(f, "publickey", "publickeyname")
	assert.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", m.Body)
	assert.Equal(t, s("text_plain_pubkey.mime"), mimeBody)
	assert.Equal(t, "body", plainContents)

	// BAD: Public key not available as an attachment!
	assert.Len(t, atts, 1)
}

// NOTE: Enable when bug is fixed.
func _TestParseMessageTextHTMLWithEmbeddedForeignEncoding(t *testing.T) { // nolint[deadcode]
	rand.Seed(0)

	f := f("text_html_embedded_foreign_encoding.eml")
	defer func() { _ = f.Close() }()

	m, mimeBody, plainContents, atts, err := Parse(f, "", "")
	assert.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	// BAD: Bridge does not detect the charset specified in the <meta> tag of the html.
	assert.Equal(t, `<html><head><meta charset="ISO-8859-2"></head><body>latin2 řšřš</body></html>`, m.Body)
	assert.Equal(t, s("text_html_embedded_foreign_encoding.mime"), mimeBody)
	assert.Equal(t, `latin2 řšřš`, plainContents)

	assert.Len(t, atts, 0)
}
