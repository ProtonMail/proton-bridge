// Copyright (c) 2023 Proton AG
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
	"image/png"
	"io"
	"os"
	"path/filepath"
	"testing"

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

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", string(m.RichBody))
	assert.Equal(t, "body", string(m.PlainBody))

	// The inline image is an 8x8 mic-dropping gopher.
	require.Len(t, m.Attachments, 1)
	img, err := png.DecodeConfig(bytes.NewReader(m.Attachments[0].Data))
	require.NoError(t, err)
	assert.Equal(t, 8, img.Width)
	assert.Equal(t, 8, img.Height)
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

	assert.Equal(t, "<html><head></head><body>This is body of <b>HTML mail</b> without attachment</body></html>", string(m.RichBody))
	assert.Equal(t, "This is body of *HTML mail* without attachment", string(m.PlainBody))

	assert.Len(t, m.Attachments, 0)
}

func TestParseTextHTMLAlready7Bit(t *testing.T) {
	f := getFileReader("text_html_7bit.eml")

	m, err := Parse(f)
	assert.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "<html><head></head><body>This is body of <b>HTML mail</b> without attachment</body></html>", string(m.RichBody))
	assert.Equal(t, "This is body of *HTML mail* without attachment", string(m.PlainBody))

	assert.Len(t, m.Attachments, 0)
}

func TestParseTextHTMLWithOctetAttachment(t *testing.T) {
	f := getFileReader("text_html_octet_attachment.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "<html><head></head><body>This is body of <b>HTML mail</b> with attachment</body></html>", string(m.RichBody))
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
	assert.Equal(t, "<html><head></head><body>This is body of <b>HTML mail</b> with attachment</body></html>", string(m.RichBody))
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

	assert.Equal(t, "<html><head></head><body>This is body of <b>HTML mail</b> with attachment</body></html>", string(m.RichBody))
	assert.Equal(t, "This is body of *HTML mail* with attachment", string(m.PlainBody))

	// The inline image is an 8x8 mic-dropping gopher.
	require.Len(t, m.Attachments, 1)
	img, err := png.DecodeConfig(bytes.NewReader(m.Attachments[0].Data))
	require.NoError(t, err)
	assert.Equal(t, 8, img.Width)
	assert.Equal(t, 8, img.Height)
}

func TestParseWithAttachedPublicKey(t *testing.T) {
	f := getFileReader("text_plain.eml")

	p, err := parser.New(f)
	require.NoError(t, err)

	m, err := ParseWithParser(p)
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

	assert.Equal(t, `<html><head>
    <meta http-equiv="content-type" content="text/html; charset=UTF-8"/>
  </head>
  <body>
    <b>aoeuaoeu</b>
  

</body></html>`, string(m.RichBody))

	assert.Equal(t, "*aoeuaoeu*\n\n", string(m.PlainBody))
}

func TestParseMultipartAlternativeNested(t *testing.T) {
	f := getFileReader("multipart_alternative_nested.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"schizofrenic" <schizofrenic@pm.me>`, m.Sender.String())
	assert.Equal(t, `<pmbridgeietest@outlook.com>`, m.ToList[0].String())

	assert.Equal(t, `<html><head>
    <meta http-equiv="content-type" content="text/html; charset=UTF-8"/>
  </head>
  <body>
    <b>multipart 2.2</b>
  

</body></html>`, string(m.RichBody))

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

func TestParseWithTrailingEndOfMailIndicator(t *testing.T) {
	f := getFileReader("text_html_trailing_end_of_mail.eml")

	m, err := Parse(f)
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@sender.com>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@receiver.com>`, m.ToList[0].String())

	assert.Equal(t, "<!DOCTYPE html><html><head></head><body>boo!</body></html>", string(m.RichBody))
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
	assert.Equal(t, m.Attachments[0].Disposition, "attachment")
	assert.Equal(t, string(m.Attachments[0].Data), "This is an ics calendar invite")
}

func TestParsePanic(t *testing.T) {
	var err error
	require.NotPanics(t, func() {
		_, err = Parse(&panicReader{})
	})
	require.Error(t, err)
}

func getFileReader(filename string) io.Reader {
	f, err := os.Open(filepath.Join("testdata", filename))
	if err != nil {
		panic(err)
	}

	return f
}

type panicReader struct{}

func (panicReader) Read(p []byte) (int, error) {
	panic("lol")
}
