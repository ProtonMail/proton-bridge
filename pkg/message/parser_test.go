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
	"fmt"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/pkg/message/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding/charmap"
)

func TestParseLongHeaderLine(t *testing.T) {
	f := getFileReader("long_header_line.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", string(m.RichBody))
	assert.Equal(t, "body", string(m.PlainBody))

	assert.Len(t, m.Attachments, 0)
}

func TestParseLongHeaderLineMultiline(t *testing.T) {
	f := getFileReader("long_header_line_multiline.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", string(m.RichBody))
	assert.Equal(t, "body", string(m.PlainBody))

	assert.Len(t, m.Attachments, 0)
}

func TestParseTextPlain(t *testing.T) {
	f := getFileReader("text_plain.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", string(m.RichBody))
	assert.Equal(t, "body", string(m.PlainBody))

	assert.Len(t, m.Attachments, 0)
}

func TestParseTextPlainUTF8(t *testing.T) {
	f := getFileReader("text_plain_utf8.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", string(m.RichBody))
	assert.Equal(t, "body", string(m.PlainBody))

	assert.Len(t, m.Attachments, 0)
}

func TestParseTextPlainLatin1(t *testing.T) {
	f := getFileReader("text_plain_latin1.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "ééééééé", string(m.RichBody))
	assert.Equal(t, "ééééééé", string(m.PlainBody))

	assert.Len(t, m.Attachments, 0)
}

func TestParseTextPlainUTF8Subject(t *testing.T) {
	f := getFileReader("text_plain_utf8_subject.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())
	assert.Equal(t, `汉字汉字汉`, m.Subject)

	assert.Equal(t, "body", string(m.RichBody))
	assert.Equal(t, "body", string(m.PlainBody))

	assert.Len(t, m.Attachments, 0)
}

func TestParseTextPlainLatin2Subject(t *testing.T) {
	f := getFileReader("text_plain_latin2_subject.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())
	assert.Equal(t, `If you can read this you understand the example.`, m.Subject)

	assert.Equal(t, "body", string(m.RichBody))
	assert.Equal(t, "body", string(m.PlainBody))

	assert.Len(t, m.Attachments, 0)
}

func TestParseTextPlainUnknownCharsetIsActuallyLatin1(t *testing.T) {
	f := getFileReader("text_plain_unknown_latin1.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "ééééééé", string(m.RichBody))
	assert.Equal(t, "ééééééé", string(m.PlainBody))

	assert.Len(t, m.Attachments, 0)
}

func TestParseTextPlainUnknownCharsetIsActuallyLatin2(t *testing.T) {
	f := getFileReader("text_plain_unknown_latin2.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	// The file contains latin2-encoded text, but we will assume it is latin1
	// and decode it as such. This will lead to corruption.
	latin2, _ := charmap.ISO8859_2.NewEncoder().Bytes([]byte("řšřšřš"))
	expect, _ := charmap.ISO8859_1.NewDecoder().Bytes(latin2)
	assert.NotEqual(t, []byte("řšřšřš"), expect)

	assert.Equal(t, string(expect), string(m.RichBody))
	assert.Equal(t, string(expect), string(m.PlainBody))

	assert.Len(t, m.Attachments, 0)
}

func TestParseTextPlainAlready7Bit(t *testing.T) {
	f := getFileReader("text_plain_7bit.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", string(m.RichBody))
	assert.Equal(t, "body", string(m.PlainBody))

	assert.Len(t, m.Attachments, 0)
}

func TestParseTextPlainWithOctetAttachment(t *testing.T) {
	f := getFileReader("text_plain_octet_attachment.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", string(m.RichBody))
	assert.Equal(t, "body", string(m.PlainBody))

	require.Len(t, m.Attachments, 1)
	assert.Equal(t, string(m.Attachments[0].Data), "if you are reading this, hi!")
}

func TestParseTextPlainWithOctetAttachmentGoodFilename(t *testing.T) {
	f := getFileReader("text_plain_octet_attachment_good_2231_filename.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", string(m.RichBody))
	assert.Equal(t, "body", string(m.PlainBody))

	assert.Len(t, m.Attachments, 1)
	assert.Equal(t, string(m.Attachments[0].Data), "if you are reading this, hi!")
	assert.Equal(t, "😁😂.txt", m.Attachments[0].Name)
}

func TestParseTextPlainWithRFC822Attachment(t *testing.T) {
	f := getFileReader("text_plain_rfc822_attachment.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", string(m.RichBody))
	assert.Equal(t, "body", string(m.PlainBody))

	assert.Len(t, m.Attachments, 1)
	assert.Equal(t, "message.eml", m.Attachments[0].Name)
}

func TestParseTextPlainWithOctetAttachmentBadFilename(t *testing.T) {
	f := getFileReader("text_plain_octet_attachment_bad_2231_filename.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", string(m.RichBody))
	assert.Equal(t, "body", string(m.PlainBody))

	assert.Len(t, m.Attachments, 1)
	assert.Equal(t, string(m.Attachments[0].Data), "if you are reading this, hi!")
	assert.Equal(t, "attachment.bin", m.Attachments[0].Name)
}

func TestParseTextPlainWithOctetAttachmentNameInContentType(t *testing.T) {
	f := getFileReader("text_plain_octet_attachment_name_in_contenttype.eml")

	m, err := Parse(f) //nolint:dogsled
	require.NoError(t, err)

	assert.Equal(t, "attachment-contenttype.txt", m.Attachments[0].Name)
}

func TestParseTextPlainWithOctetAttachmentNameConflict(t *testing.T) {
	f := getFileReader("text_plain_octet_attachment_name_conflict.eml")

	m, err := Parse(f) //nolint:dogsled
	require.NoError(t, err)

	assert.Equal(t, "attachment-disposition.txt", m.Attachments[0].Name)
}

func TestParseTextPlainWithPlainAttachment(t *testing.T) {
	f := getFileReader("text_plain_plain_attachment.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", string(m.RichBody))
	assert.Equal(t, "body", string(m.PlainBody))

	require.Len(t, m.Attachments, 1)
	assert.Equal(t, string(m.Attachments[0].Data), "attachment")
}

func TestParseTextPlainEmptyAddresses(t *testing.T) {
	f := getFileReader("text_plain_empty_addresses.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", string(m.RichBody))
	assert.Equal(t, "body", string(m.PlainBody))

	assert.Len(t, m.Attachments, 0)
}

func TestParseTextPlainWithImageInline(t *testing.T) {
	f := getFileReader("text_plain_image_inline.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	require.NotEmpty(t, m.Attachments[0].ContentID)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", string(m.PlainBody))
	assert.Equal(t, fmt.Sprintf(`<html><body><p>body</p><img src="cid:%v"/></body></html>`, m.Attachments[0].ContentID), string(m.RichBody))

	// The inline image is an 8x8 mic-dropping gopher.
	require.Len(t, m.Attachments, 1)
	img, err := png.DecodeConfig(bytes.NewReader(m.Attachments[0].Data))
	require.NoError(t, err)
	assert.Equal(t, 8, img.Width)
	assert.Equal(t, 8, img.Height)
}

func TestParseTextPlainWithImageInlineWithMoreTextParts(t *testing.T) {
	// Inline image test with text - image - text, ensure all parts are convert to html
	f := getFileReader("text_plain_image_inline2.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	require.NotEmpty(t, m.Attachments[0].ContentID)
	assert.Equal(t, "bodybody2", string(m.PlainBody))
	assert.Equal(t, fmt.Sprintf("<html><body><p>body</p><img src=\"cid:%v\"/></body></html><html><body><p>body2<br/>\n</p></body></html>", m.Attachments[0].ContentID), string(m.RichBody))

	// The inline image is an 8x8 mic-dropping gopher.
	require.Len(t, m.Attachments, 1)
	img, err := png.DecodeConfig(bytes.NewReader(m.Attachments[0].Data))
	require.NoError(t, err)
	assert.Equal(t, 8, img.Width)
	assert.Equal(t, 8, img.Height)
}

func TestParseTextPlainWithImageInlineAfterOtherAttachment(t *testing.T) {
	// Inline image test with text - image - text, ensure all parts are convert to html
	f := getFileReader("text_plain_image_inline2.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	require.NotEmpty(t, m.Attachments[0].ContentID)
	assert.Equal(t, "bodybody2", string(m.PlainBody))
	assert.Equal(t, fmt.Sprintf("<html><body><p>body</p><img src=\"cid:%v\"/></body></html><html><body><p>body2<br/>\n</p></body></html>", m.Attachments[0].ContentID), string(m.RichBody))

	// The inline image is an 8x8 mic-dropping gopher.
	require.Len(t, m.Attachments, 1)
	img, err := png.DecodeConfig(bytes.NewReader(m.Attachments[0].Data))
	require.NoError(t, err)
	assert.Equal(t, 8, img.Width)
	assert.Equal(t, 8, img.Height)
}

func TestParseTextPlainWithImageBetweenAttachments(t *testing.T) {
	// Inline image test with text - pdf - image - text. A new part must be created to be injected.
	f := getFileReader("text_plain_image_inline_between_attachment.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	require.Empty(t, m.Attachments[0].ContentID)
	require.NotEmpty(t, m.Attachments[1].ContentID)
	assert.Equal(t, "bodybody2", string(m.PlainBody))
	assert.Equal(t, fmt.Sprintf("<html><body><p>body</p></body></html><html><body><img src=\"cid:%v\"/></body></html><html><body><p>body2<br/>\n</p></body></html>", m.Attachments[1].ContentID), string(m.RichBody))
}

func TestParseTextPlainWithImageFirst(t *testing.T) {
	// Inline image test with text - pdf - image - text. A new part must be created to be injected.
	f := getFileReader("text_plain_image_inline_attachment_first.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	require.NotEmpty(t, m.Attachments[0].ContentID)
	assert.Equal(t, "body", string(m.PlainBody))
	assert.Equal(t, fmt.Sprintf("<html><body><img src=\"cid:%v\"/></body></html><html><body><p>body</p></body></html>", m.Attachments[0].ContentID), string(m.RichBody))
}

func TestParseTextPlainWithDuplicateCharset(t *testing.T) {
	f := getFileReader("text_plain_duplicate_charset.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", string(m.RichBody))
	assert.Equal(t, "body", string(m.PlainBody))

	assert.Len(t, m.Attachments, 0)
}

func TestParseWithMultipleTextParts(t *testing.T) {
	f := getFileReader("multiple_text_parts.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body\nsome other part of the message", string(m.RichBody))
	assert.Equal(t, "body\nsome other part of the message", string(m.PlainBody))

	assert.Len(t, m.Attachments, 0)
}

func TestParseTextHTML(t *testing.T) {
	f := getFileReader("text_html.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "<html><body>This is body of <b>HTML mail</b> without attachment</body></html>", string(m.RichBody))
	assert.Equal(t, "This is body of *HTML mail* without attachment", string(m.PlainBody))

	assert.Len(t, m.Attachments, 0)
}

func TestParseTextHTMLAlready7Bit(t *testing.T) {
	f := getFileReader("text_html_7bit.eml")

	m, err := Parse(f)
	assert.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "<html><body>This is body of <b>HTML mail</b> without attachment</body></html>", string(m.RichBody))
	assert.Equal(t, "This is body of *HTML mail* without attachment", string(m.PlainBody))

	assert.Len(t, m.Attachments, 0)
}

func TestParseTextHTMLWithOctetAttachment(t *testing.T) {
	f := getFileReader("text_html_octet_attachment.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "<html><body>This is body of <b>HTML mail</b> with attachment</body></html>", string(m.RichBody))
	assert.Equal(t, "This is body of *HTML mail* with attachment", string(m.PlainBody))

	require.Len(t, m.Attachments, 1)
	assert.Equal(t, string(m.Attachments[0].Data), "if you are reading this, hi!")
}

func TestParseTextHTMLWithPlainAttachment(t *testing.T) {
	f := getFileReader("text_html_plain_attachment.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	// BAD: plainBody should not be empty!
	assert.Equal(t, "<html><body>This is body of <b>HTML mail</b> with attachment</body></html>", string(m.RichBody))
	assert.Equal(t, "This is body of *HTML mail* with attachment", string(m.PlainBody))

	require.Len(t, m.Attachments, 1)
	assert.Equal(t, string(m.Attachments[0].Data), "attachment")
}

func TestParseTextHTMLWithImageInline(t *testing.T) {
	f := getFileReader("text_html_image_inline.eml")

	m, err := Parse(f)
	assert.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	require.Len(t, m.Attachments, 1)
	require.Equal(t, m.Attachments[0].Disposition, proton.InlineDisposition)

	assert.Equal(t, fmt.Sprintf(`<html><body>This is body of <b>HTML mail</b> with attachment</body></html><html><body><img src="cid:%v"/></body></html>`, m.Attachments[0].ContentID), string(m.RichBody))
	assert.Equal(t, "This is body of *HTML mail* with attachment", string(m.PlainBody))

	// The inline image is an 8x8 mic-dropping gopher.
	img, err := png.DecodeConfig(bytes.NewReader(m.Attachments[0].Data))
	require.NoError(t, err)
	assert.Equal(t, 8, img.Width)
	assert.Equal(t, 8, img.Height)
}

func TestParseTextHTMLWithImageInlineNoDisposition(t *testing.T) {
	f := getFileReader("text_html_image_inline_no_disposition.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	require.Len(t, m.Attachments, 1)

	assert.Equal(t, `<html><body>This is body of <b>HTML mail</b> with attachment</body></html>`, string(m.RichBody))
	assert.Equal(t, "This is body of *HTML mail* with attachment", string(m.PlainBody))

	// The inline image is an 8x8 mic-dropping gopher.
	img, err := png.DecodeConfig(bytes.NewReader(m.Attachments[0].Data))
	require.NoError(t, err)
	assert.Equal(t, 8, img.Width)
	assert.Equal(t, 8, img.Height)
}

func TestParseWithAttachedPublicKey(t *testing.T) {
	f := getFileReader("text_plain.eml")

	p, err := parser.New(f)
	require.NoError(t, err)

	m, err := ParseWithParser(p, false)
	require.NoError(t, err)

	p.AttachPublicKey("publickey", "publickeyname")

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", string(m.RichBody))
	assert.Equal(t, "body", string(m.PlainBody))

	// The pubkey should not be collected as an attachment.
	// We upload the pubkey when creating the draft.
	require.Len(t, m.Attachments, 0)
}

func TestParseTextHTMLWithEmbeddedForeignEncoding(t *testing.T) {
	f := getFileReader("text_html_embedded_foreign_encoding.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, `<html><head><meta charset="UTF-8"/></head><body>latin2 řšřš</body></html>`, string(m.RichBody))
	assert.Equal(t, `latin2 řšřš`, string(m.PlainBody))

	assert.Len(t, m.Attachments, 0)
}

func TestParseMultipartAlternative(t *testing.T) {
	f := getFileReader("multipart_alternative.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"schizofrenic" <schizofrenic@pm.me>`, m.Sender.String())
	assert.Equal(t, `<pmbridgeietest@outlook.com>`, m.ToList[0].String())

	assert.Equal(t, `<html>
  <head>
    <meta http-equiv="content-type" content="text/html; charset=UTF-8">
  </head>
  <body>
    <b>aoeuaoeu</b>
  </body>
</html>
`, string(m.RichBody))

	assert.Equal(t, "*aoeuaoeu*\n\n", string(m.PlainBody))
}

func TestParseMultipartAlternativeNested(t *testing.T) {
	f := getFileReader("multipart_alternative_nested.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"schizofrenic" <schizofrenic@pm.me>`, m.Sender.String())
	assert.Equal(t, `<pmbridgeietest@outlook.com>`, m.ToList[0].String())

	assert.Equal(t, `<html>
  <head>
    <meta http-equiv="content-type" content="text/html; charset=UTF-8">
  </head>
  <body>
    <b>multipart 2.2</b>
  </body>
</html>
`, string(m.RichBody))

	assert.Equal(t, "*multipart 2.1*\n\n", string(m.PlainBody))
}

func TestParseMultipartAlternativeLatin1(t *testing.T) {
	f := getFileReader("multipart_alternative_latin1.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"schizofrenic" <schizofrenic@pm.me>`, m.Sender.String())
	assert.Equal(t, `<pmbridgeietest@outlook.com>`, m.ToList[0].String())

	assert.Equal(t, `<html><head>
    <meta http-equiv="content-type" content="text/html; charset=UTF-8"/>
  </head>
  <body>
    <b>aoeuaoeu</b>
  

</body></html>`, string(m.RichBody))

	assert.Equal(t, "*aoeuaoeu*\n\n", string(m.PlainBody))
}

func TestParseMultipartAttachmentEncodedButUnquoted(t *testing.T) {
	f := getFileReader("multipart_attachment_encoded_no_quote.eml")

	p, err := parser.New(f)
	require.NoError(t, err)

	m, err := ParseWithParser(p, false)
	require.NoError(t, err)
	assert.Equal(t, `"Bridge Test" <bridgetest@pm.test>`, m.Sender.String())
	assert.Equal(t, `"Internal Bridge" <bridgetest@protonmail.com>`, m.ToList[0].String())
}

func TestParseWithTrailingEndOfMailIndicator(t *testing.T) {
	f := getFileReader("text_html_trailing_end_of_mail.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@sender.com>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@receiver.com>`, m.ToList[0].String())

	assert.Equal(t, "<!DOCTYPE HTML>\n<html><body>boo!</body></html>", string(m.RichBody))
	assert.Equal(t, "boo!", string(m.PlainBody))
}

func TestParseEncodedContentType(t *testing.T) {
	f := getFileReader("rfc2047-content-transfer-encoding.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@sender.com>`, m.Sender.String())
	assert.Equal(t, `<user@somewhere.org>`, m.ToList[0].String())

	assert.Equal(t, "bodybodybody\n", string(m.PlainBody))
}

func TestParseNonEncodedContentType(t *testing.T) {
	f := getFileReader("non-encoded-content-transfer-encoding.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@sender.com>`, m.Sender.String())
	assert.Equal(t, `<user@somewhere.org>`, m.ToList[0].String())

	assert.Equal(t, "bodybodybody\n", string(m.PlainBody))
}

func TestParseEncodedContentTypeBad(t *testing.T) {
	f := getFileReader("rfc2047-content-transfer-encoding-bad.eml")

	_, err := Parse(f) //nolint:dogsled
	require.Error(t, err)
}

func TestParseMessageReferences(t *testing.T) {
	f := getFileReader("references.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.ElementsMatch(t, m.References, []string{
		`PMZV4VZMRM@something.com`,
		`OEUOEUEOUOUOU770B9QNZWFVGM@protonmail.ch`,
	})
}

func TestParseMessageReferencesComma(t *testing.T) {
	f := getFileReader("references-comma.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.ElementsMatch(t, m.References, []string{
		`PMZV4VZMRM@something.com`,
		`OEUOEUEOUOUOU770B9QNZWFVGM@protonmail.ch`,
	})
}

func TestParseMessageReplyToWithoutReferences(t *testing.T) {
	f := getFileReader("reply-to_no_references.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.ElementsMatch(t, m.References, []string{})
	assert.Equal(t, m.InReplyTo, "OEUOEUEOUOUOU770B9QNZWFVGM@protonmail.ch")
}

func TestParseIcsAttachment(t *testing.T) {
	f := getFileReader("ics_attachment.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", string(m.RichBody))
	assert.Equal(t, "body", string(m.PlainBody))

	require.Len(t, m.Attachments, 1)
	assert.Equal(t, m.Attachments[0].MIMEType, "text/calendar")
	assert.Equal(t, m.Attachments[0].Name, "invite.ics")
	assert.Equal(t, m.Attachments[0].ContentID, "")
	assert.Equal(t, m.Attachments[0].Disposition, proton.Disposition("attachment"))
	assert.Equal(t, string(m.Attachments[0].Data), "This is an ics calendar invite")
}

func TestParseAllowInvalidAddress(t *testing.T) {
	const literal = `To: foo
From: bar
BCC: fff
CC: FFF
Reply-To: AAA
Subject: Test
`

	// This will fail as the addresses are not valid.
	{
		_, err := Parse(strings.NewReader(literal))
		require.Error(t, err)
	}

	// This will work as invalid addresses will be ignored.
	m, err := ParseAndAllowInvalidAddressLists(strings.NewReader(literal))
	require.NoError(t, err)

	assert.Empty(t, m.ToList)
	assert.Empty(t, m.Sender)
	assert.Empty(t, m.CCList)
	assert.Empty(t, m.BCCList)
	assert.Empty(t, m.ReplyTos)
}

func TestParsePanic(t *testing.T) {
	var err error

	require.NotPanics(t, func() {
		_, err = Parse(&panicReader{})
	})

	require.Error(t, err)
}

func TestParseTextPlainWithPdfAttachmentCyrillic(t *testing.T) {
	f := getFileReader("text_plain_pdf_attachment_cyrillic.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "Shake that body", string(m.RichBody))
	assert.Equal(t, "Shake that body", string(m.PlainBody))

	require.Len(t, m.Attachments, 1)
	require.Equal(t, "application/pdf", m.Attachments[0].MIMEType)
	assert.Equal(t, "АБВГДЃЕЖЗЅИЈКЛЉМНЊОПРСТЌУФХЧЏЗШ.pdf", m.Attachments[0].Name)
}

func TestParseTextPlainWithDocxAttachmentCyrillic(t *testing.T) {
	f := getFileReader("text_plain_docx_attachment_cyrillic.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "Shake that body", string(m.RichBody))
	assert.Equal(t, "Shake that body", string(m.PlainBody))

	require.Len(t, m.Attachments, 1)
	require.Equal(t, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", m.Attachments[0].MIMEType)
	assert.Equal(t, "АБВГДЃЕЖЗЅИЈКЛЉМНЊОПРСТЌУФХЧЏЗШ.docx", m.Attachments[0].Name)
}

func TestParseInReplyToAndXForward(t *testing.T) {
	f := getFileReader("text_plain_utf8_reply_to_and_x_forward.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	require.Equal(t, "00000@protonmail.com", m.XForward)
	require.Equal(t, "00000@protonmail.com", m.InReplyTo)
}

func TestPatchNewLineWithHtmlBreaks(t *testing.T) {
	{
		input := []byte("\nfoo\nbar\n\n\nzz\nddd")
		expected := []byte("<br/>\nfoo<br/>\nbar<br/>\n<br/>\n<br/>\nzz<br/>\nddd")

		result := patchNewLineWithHTMLBreaks(input)
		require.Equal(t, expected, result)
	}
	{
		input := []byte("\r\nfoo\r\nbar\r\n\r\n\r\nzz\r\nddd")
		expected := []byte("<br/>\r\nfoo<br/>\r\nbar<br/>\r\n<br/>\r\n<br/>\r\nzz<br/>\r\nddd")

		result := patchNewLineWithHTMLBreaks(input)
		require.Equal(t, expected, result)
	}
}

func TestParseCp1250Attachment(t *testing.T) {
	r := require.New(t)
	f := getFileReader("text_plain_xml_attachment_cp1250.eml")

	m, err := Parse(f)
	r.NoError(err)

	r.Len(m.Attachments, 1)
	r.Equal("text/xml; charset=windows-1250; name=\"cp1250.xml\"", m.Attachments[0].Header.Get("Content-Type"))
}

func getFileReader(filename string) io.Reader {
	f, err := os.Open(filepath.Join("testdata", filename))
	if err != nil {
		panic(err)
	}

	return f
}

func TestParseInvalidOriginalBoundary(t *testing.T) {
	f := getFileReader("incorrect_boundary_w_invalid_character_tuta.eml")

	p, err := parser.New(f)
	require.NoError(t, err)

	require.Equal(t, true, p.Root().Header.Get("Content-Type") == `multipart/related; boundary="------------1234567890@tutanota"`)

	m, err := ParseWithParser(p, false)
	require.NoError(t, err)

	require.Equal(t, true, strings.HasPrefix(string(m.MIMEBody), "Content-Type: multipart/related;\r\n boundary="))
	require.Equal(t, false, strings.HasPrefix(string(m.MIMEBody), `Content-Type: multipart/related;\n boundary="------------1234567890@tutanota"`))
	require.Equal(t, false, strings.HasPrefix(string(m.MIMEBody), `Content-Type: multipart/related;\n boundary=------------1234567890@tutanota`))
	require.Equal(t, false, strings.HasPrefix(string(m.MIMEBody), `Content-Type: multipart/related;\n boundary=1234567890@tutanota`))
}

type panicReader struct{}

func (panicReader) Read(_ []byte) (int, error) {
	panic("lol")
}
