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
	"io"
	"mime"
	"net/mail"
	"net/textproto"
	"regexp"
	"strings"

	"github.com/ProtonMail/go-rfc5322"
	"github.com/ProtonMail/proton-bridge/v2/pkg/message/parser"
	pmmime "github.com/ProtonMail/proton-bridge/v2/pkg/mime"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/emersion/go-message"
	"github.com/jaytaylor/html2text"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Parse parses RAW message.
func Parse(r io.Reader) (m *pmapi.Message, mimeBody, plainBody string, attReaders []io.Reader, err error) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}

		err = fmt.Errorf("panic while parsing message: %v", r)
	}()

	p, err := parser.New(r)
	if err != nil {
		return nil, "", "", nil, errors.Wrap(err, "failed to create new parser")
	}

	m, plainBody, attReaders, err = ParserWithParser(p)
	if err != nil {
		return nil, "", "", nil, errors.Wrap(err, "failed to parse the message")
	}

	mimeBody, err = BuildMIMEBody(p)
	if err != nil {
		return nil, "", "", nil, errors.Wrap(err, "failed to build mime body")
	}

	return m, mimeBody, plainBody, attReaders, nil
}

// ParserWithParser parses message from Parser without building MIME body.
func ParserWithParser(p *parser.Parser) (m *pmapi.Message, plainBody string, attReaders []io.Reader, err error) {
	logrus.Trace("Parsing message")

	if err = convertEncodedTransferEncoding(p); err != nil {
		err = errors.Wrap(err, "failed to convert encoded transfer encodings")
		return
	}

	if err = convertForeignEncodings(p); err != nil {
		err = errors.Wrap(err, "failed to convert foreign encodings")
		return
	}

	m = pmapi.NewMessage()

	if err = parseMessageHeader(m, p.Root().Header); err != nil {
		err = errors.Wrap(err, "failed to parse message header")
		return
	}

	if m.Attachments, attReaders, err = collectAttachments(p); err != nil {
		err = errors.Wrap(err, "failed to collect attachments")
		return
	}

	if m.Body, plainBody, err = buildBodies(p); err != nil {
		err = errors.Wrap(err, "failed to build bodies")
		return
	}

	if m.MIMEType, err = determineMIMEType(p); err != nil {
		err = errors.Wrap(err, "failed to determine mime type")
		return
	}

	return m, plainBody, attReaders, nil
}

// BuildMIMEBody builds mime body from the parser returned by NewParser.
func BuildMIMEBody(p *parser.Parser) (mimeBody string, err error) {
	mimeBodyBuffer := new(bytes.Buffer)

	if err = p.NewWriter().Write(mimeBodyBuffer); err != nil {
		err = errors.Wrap(err, "failed to write out mime message")
		return
	}

	return mimeBodyBuffer.String(), nil
}

// convertEncodedTransferEncoding decodes any RFC2047-encoded content transfer encodings.
// Such content transfer encodings go against RFC but still exist in the wild anyway.
func convertEncodedTransferEncoding(p *parser.Parser) error {
	logrus.Trace("Converting encoded transfer encoding")

	return p.NewWalker().
		RegisterDefaultHandler(func(p *parser.Part) error {
			encoding := p.Header.Get("Content-Transfer-Encoding")
			if encoding == "" {
				return nil
			}

			dec, err := pmmime.WordDec.DecodeHeader(encoding)
			if err != nil {
				return err
			}

			p.Header.Set("Content-Transfer-Encoding", dec)

			return nil
		}).
		Walk()
}

func convertForeignEncodings(p *parser.Parser) error {
	logrus.Trace("Converting foreign encodings")

	return p.NewWalker().
		RegisterContentTypeHandler("text/html", func(p *parser.Part) error {
			if err := p.ConvertToUTF8(); err != nil {
				return err
			}

			return p.ConvertMetaCharset()
		}).
		RegisterContentTypeHandler("text/.*", func(p *parser.Part) error {
			return p.ConvertToUTF8()
		}).
		Walk()
}

func collectAttachments(p *parser.Parser) ([]*pmapi.Attachment, []io.Reader, error) {
	var (
		atts []*pmapi.Attachment
		data []io.Reader
		err  error
	)

	w := p.NewWalker().
		RegisterContentDispositionHandler("attachment", func(p *parser.Part) error {
			att, err := parseAttachment(p.Header)
			if err != nil {
				return err
			}

			atts = append(atts, att)
			data = append(data, bytes.NewReader(p.Body))

			return nil
		}).
		RegisterContentTypeHandler("text/calendar", func(p *parser.Part) error {
			att, err := parseAttachment(p.Header)
			if err != nil {
				return err
			}

			atts = append(atts, att)
			data = append(data, bytes.NewReader(p.Body))

			return nil
		}).
		RegisterContentTypeHandler("text/.*", func(p *parser.Part) error {
			return nil
		}).
		RegisterDefaultHandler(func(p *parser.Part) error {
			if len(p.Children()) > 0 {
				return nil
			}

			att, err := parseAttachment(p.Header)
			if err != nil {
				return err
			}

			atts = append(atts, att)
			data = append(data, bytes.NewReader(p.Body))

			return nil
		})

	if err = w.Walk(); err != nil {
		return nil, nil, err
	}

	return atts, data, nil
}

// buildBodies collects all text/html and text/plain parts and returns two bodies,
//  - a rich text body (in which html is allowed), and
//  - a plaintext body (in which html is converted to plaintext).
//
// text/html parts are converted to plaintext in order to build the plaintext body,
// unless there is already a plaintext part provided via multipart/alternative,
// in which case the provided alternative is chosen.
func buildBodies(p *parser.Parser) (richBody, plainBody string, err error) {
	richParts, err := collectBodyParts(p, "text/html")
	if err != nil {
		return
	}

	plainParts, err := collectBodyParts(p, "text/plain")
	if err != nil {
		return
	}

	richBuilder, plainBuilder := strings.Builder{}, strings.Builder{}

	for _, richPart := range richParts {
		_, _ = richBuilder.Write(richPart.Body)
	}

	for _, plainPart := range plainParts {
		_, _ = plainBuilder.Write(getPlainBody(plainPart))
	}

	return richBuilder.String(), plainBuilder.String(), nil
}

// collectBodyParts collects all body parts in the parse tree, preferring
// parts of the given content type if alternatives exist.
func collectBodyParts(p *parser.Parser, preferredContentType string) (parser.Parts, error) {
	v := p.
		NewVisitor(func(p *parser.Part, visit parser.Visit) (interface{}, error) {
			childParts, err := collectChildParts(p, visit)
			if err != nil {
				return nil, err
			}

			return joinChildParts(childParts), nil
		}).
		RegisterRule("multipart/alternative", func(p *parser.Part, visit parser.Visit) (interface{}, error) {
			childParts, err := collectChildParts(p, visit)
			if err != nil {
				return nil, err
			}

			return bestChoice(childParts, preferredContentType), nil
		}).
		RegisterRule("text/plain", func(p *parser.Part, visit parser.Visit) (interface{}, error) {
			disp, _, err := p.Header.ContentDisposition()
			if err != nil {
				disp = ""
			}

			if disp == "attachment" {
				return parser.Parts{}, nil
			}

			return parser.Parts{p}, nil
		}).
		RegisterRule("text/html", func(p *parser.Part, visit parser.Visit) (interface{}, error) {
			disp, _, err := p.Header.ContentDisposition()
			if err != nil {
				disp = ""
			}

			if disp == "attachment" {
				return parser.Parts{}, nil
			}

			return parser.Parts{p}, nil
		})

	res, err := v.Visit()
	if err != nil {
		return nil, err
	}

	return res.(parser.Parts), nil //nolint:forcetypeassert
}

func collectChildParts(p *parser.Part, visit parser.Visit) ([]parser.Parts, error) {
	childParts := []parser.Parts{}

	for _, child := range p.Children() {
		res, err := visit(child)
		if err != nil {
			return nil, err
		}

		childParts = append(childParts, res.(parser.Parts)) //nolint:forcetypeassert
	}

	return childParts, nil
}

func joinChildParts(childParts []parser.Parts) parser.Parts {
	res := parser.Parts{}

	for _, parts := range childParts {
		res = append(res, parts...)
	}

	return res
}

func bestChoice(childParts []parser.Parts, preferredContentType string) parser.Parts {
	// If one of the parts has preferred content type, use that.
	for i := len(childParts) - 1; i >= 0; i-- {
		if allPartsHaveContentType(childParts[i], preferredContentType) {
			return childParts[i]
		}
	}

	// Otherwise, choose the last one, if it exists.
	if len(childParts) > 0 {
		return childParts[len(childParts)-1]
	}

	return parser.Parts{}
}

func allPartsHaveContentType(parts parser.Parts, contentType string) bool {
	if len(parts) == 0 {
		return false
	}

	for _, part := range parts {
		t, _, err := part.ContentType()
		if err != nil {
			return false
		}

		if t != contentType {
			return false
		}
	}

	return true
}

func determineMIMEType(p *parser.Parser) (string, error) {
	var isHTML bool

	w := p.NewWalker().
		RegisterContentTypeHandler("text/html", func(p *parser.Part) (err error) {
			isHTML = true
			return
		})

	if err := w.Walk(); err != nil {
		return "", err
	}

	if isHTML {
		return "text/html", nil
	}

	return "text/plain", nil
}

// getPlainBody returns the body of the given part, converting html to
// plaintext where possible.
func getPlainBody(part *parser.Part) []byte {
	contentType, _, err := part.ContentType()
	if err != nil {
		return part.Body
	}

	switch contentType {
	case "text/html":
		text, err := html2text.FromReader(bytes.NewReader(part.Body))
		if err != nil {
			return part.Body
		}

		return []byte(text)

	default:
		return part.Body
	}
}

func AttachPublicKey(p *parser.Parser, key, keyName string) {
	h := message.Header{}

	h.Set("Content-Type", fmt.Sprintf(`application/pgp-keys; name="%v.asc"; filename="%v.asc"`, keyName, keyName))
	h.Set("Content-Disposition", fmt.Sprintf(`attachment; name="%v.asc"; filename="%v.asc"`, keyName, keyName))
	h.Set("Content-Transfer-Encoding", "base64")

	p.Root().AddChild(&parser.Part{
		Header: h,
		Body:   []byte(key),
	})
}

func parseMessageHeader(m *pmapi.Message, h message.Header) error { //nolint:funlen
	mimeHeader, err := toMailHeader(h)
	if err != nil {
		return err
	}
	m.Header = mimeHeader

	fields := h.Fields()

	for fields.Next() {
		switch strings.ToLower(fields.Key()) {
		case "subject":
			s, err := fields.Text()
			if err != nil {
				if s, err = pmmime.DecodeHeader(fields.Value()); err != nil {
					return errors.Wrap(err, "failed to parse subject")
				}
			}

			m.Subject = s

		case "from":
			sender, err := rfc5322.ParseAddressList(fields.Value())
			if err != nil {
				return errors.Wrap(err, "failed to parse from")
			}
			if len(sender) > 0 {
				m.Sender = sender[0]
			}

		case "to":
			toList, err := rfc5322.ParseAddressList(fields.Value())
			if err != nil {
				return errors.Wrap(err, "failed to parse to")
			}
			m.ToList = toList

		case "reply-to":
			replyTos, err := rfc5322.ParseAddressList(fields.Value())
			if err != nil {
				return errors.Wrap(err, "failed to parse reply-to")
			}
			m.ReplyTos = replyTos

		case "cc":
			ccList, err := rfc5322.ParseAddressList(fields.Value())
			if err != nil {
				return errors.Wrap(err, "failed to parse cc")
			}
			m.CCList = ccList

		case "bcc":
			bccList, err := rfc5322.ParseAddressList(fields.Value())
			if err != nil {
				return errors.Wrap(err, "failed to parse bcc")
			}
			m.BCCList = bccList

		case "date":
			date, err := rfc5322.ParseDateTime(fields.Value())
			if err != nil {
				return errors.Wrap(err, "failed to parse date")
			}
			m.Time = date.Unix()

		case "message-id":
			m.ExternalID = regexp.MustCompile("<(.*)>").ReplaceAllString(fields.Value(), "$1")
		}
	}

	return nil
}

func parseAttachment(h message.Header) (*pmapi.Attachment, error) {
	att := &pmapi.Attachment{}

	mimeHeader, err := toMIMEHeader(h)
	if err != nil {
		return nil, err
	}
	att.Header = mimeHeader

	mimeType, mimeTypeParams, err := h.ContentType()
	if err != nil {
		return nil, err
	}
	att.MIMEType = mimeType

	// Prefer attachment name from filename param in content disposition.
	// If not available, try to get it from name param in content type.
	// Otherwise fallback to attachment.bin.
	_, dispParams, dispErr := h.ContentDisposition()
	if dispErr != nil {
		ext, err := mime.ExtensionsByType(att.MIMEType)
		if err != nil {
			return nil, err
		}

		if len(ext) > 0 {
			att.Name = "attachment" + ext[0]
		}
	} else {
		att.Name = dispParams["filename"]
	}
	if att.Name == "" {
		att.Name = mimeTypeParams["name"]
	}
	if att.Name == "" && mimeType == rfc822Message {
		att.Name = "message.eml"
	}
	if att.Name == "" {
		att.Name = "attachment.bin"
	}

	// Only set ContentID if it should be inline;
	// API infers content disposition based on whether ContentID is present.
	// If Content-Disposition is present, we base our decision on that.
	// Otherwise, if Content-Disposition is missing but there is a ContentID, set it.
	// (This is necessary because some clients don't set Content-Disposition at all,
	// so we need to rely on other information to deduce if it's inline or attachment.)
	if h.Has("Content-Disposition") {
		if disp, _, err := h.ContentDisposition(); err != nil {
			return nil, err
		} else if disp == pmapi.DispositionInline {
			att.ContentID = strings.Trim(h.Get("Content-Id"), " <>")
		}
	} else if h.Has("Content-Id") {
		att.ContentID = strings.Trim(h.Get("Content-Id"), " <>")
	}

	return att, nil
}

func toMailHeader(h message.Header) (mail.Header, error) {
	mimeHeader := make(mail.Header)

	if err := forEachDecodedHeaderField(h, func(key, val string) error {
		mimeHeader[key] = []string{val}
		return nil
	}); err != nil {
		return nil, err
	}

	return mimeHeader, nil
}

func toMIMEHeader(h message.Header) (textproto.MIMEHeader, error) {
	mimeHeader := make(textproto.MIMEHeader)

	if err := forEachDecodedHeaderField(h, func(key, val string) error {
		mimeHeader[key] = []string{val}
		return nil
	}); err != nil {
		return nil, err
	}

	return mimeHeader, nil
}

func forEachDecodedHeaderField(h message.Header, fn func(string, string) error) error {
	fields := h.Fields()

	for fields.Next() {
		text, err := fields.Text()
		if err != nil {
			if !message.IsUnknownCharset(err) {
				return err
			}

			if text, err = pmmime.DecodeHeader(fields.Value()); err != nil {
				return err
			}
		}

		if err := fn(fields.Key(), text); err != nil {
			return err
		}
	}

	return nil
}
