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
	"fmt"
	"io"
	"mime"
	"net/mail"
	"regexp"
	"strings"

	"github.com/ProtonMail/gluon/rfc822"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/go-rfc5322"
	"github.com/ProtonMail/proton-bridge/v3/pkg/message/parser"
	pmmime "github.com/ProtonMail/proton-bridge/v3/pkg/mime"
	"github.com/emersion/go-message"
	"github.com/jaytaylor/html2text"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type MIMEBody string

type Body string

type Message struct {
	MIMEBody    MIMEBody
	RichBody    Body
	PlainBody   Body
	Attachments []Attachment
	MIMEType    rfc822.MIMEType
	IsReply     bool

	Subject  string
	Sender   *mail.Address
	ToList   []*mail.Address
	CCList   []*mail.Address
	BCCList  []*mail.Address
	ReplyTos []*mail.Address

	References []string
	ExternalID string
	InReplyTo  string
}

type Attachment struct {
	Header      mail.Header
	Name        string
	ContentID   string
	MIMEType    string
	Disposition string
	Data        []byte
}

// Parse parses an RFC822 message.
func Parse(r io.Reader) (m Message, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic while parsing message: %v", r)
		}
	}()

	p, err := parser.New(r)
	if err != nil {
		return Message{}, errors.Wrap(err, "failed to create new parser")
	}

	return parse(p)
}

// ParseWithParser parses an RFC822 message using an existing parser.
func ParseWithParser(p *parser.Parser) (m Message, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic while parsing message: %v", r)
		}
	}()

	return parse(p)
}

func parse(p *parser.Parser) (Message, error) {
	if err := convertEncodedTransferEncoding(p); err != nil {
		return Message{}, errors.Wrap(err, "failed to convert encoded transfer encoding")
	}

	if err := convertForeignEncodings(p); err != nil {
		return Message{}, errors.Wrap(err, "failed to convert foreign encodings")
	}

	m, err := parseMessageHeader(p.Root().Header)
	if err != nil {
		return Message{}, errors.Wrap(err, "failed to parse message header")
	}

	atts, err := collectAttachments(p)
	if err != nil {
		return Message{}, errors.Wrap(err, "failed to collect attachments")
	}

	m.Attachments = atts

	richBody, plainBody, err := buildBodies(p)
	if err != nil {
		return Message{}, errors.Wrap(err, "failed to build bodies")
	}

	mimeBody, err := buildMIMEBody(p)
	if err != nil {
		return Message{}, errors.Wrap(err, "failed to build mime body")
	}

	m.RichBody = Body(richBody)
	m.PlainBody = Body(plainBody)
	m.MIMEBody = MIMEBody(mimeBody)

	mimeType, err := determineMIMEType(p)
	if err != nil {
		return Message{}, errors.Wrap(err, "failed to get mime type")
	}

	m.MIMEType = rfc822.MIMEType(mimeType)

	return m, nil
}

// buildMIMEBody builds mime body from the parser returned by NewParser.
func buildMIMEBody(p *parser.Parser) (mimeBody string, err error) {
	buf := new(bytes.Buffer)

	if err := p.NewWriter().Write(buf); err != nil {
		return "", fmt.Errorf("failed to write message: %w", err)
	}

	return buf.String(), nil
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

func collectAttachments(p *parser.Parser) ([]Attachment, error) {
	var (
		atts []Attachment
		err  error
	)

	w := p.NewWalker().
		RegisterContentDispositionHandler("attachment", func(p *parser.Part) error {
			att, err := parseAttachment(p.Header, p.Body)
			if err != nil {
				return err
			}

			atts = append(atts, att)

			return nil
		}).
		RegisterContentTypeHandler("text/calendar", func(p *parser.Part) error {
			att, err := parseAttachment(p.Header, p.Body)
			if err != nil {
				return err
			}

			atts = append(atts, att)

			return nil
		}).
		RegisterContentTypeHandler("text/.*", func(p *parser.Part) error {
			return nil
		}).
		RegisterDefaultHandler(func(p *parser.Part) error {
			if len(p.Children()) > 0 {
				return nil
			}

			att, err := parseAttachment(p.Header, p.Body)
			if err != nil {
				return err
			}

			atts = append(atts, att)

			return nil
		})

	if err = w.Walk(); err != nil {
		return nil, err
	}

	return atts, nil
}

// buildBodies collects all text/html and text/plain parts and returns two bodies,
//   - a rich text body (in which html is allowed), and
//   - a plaintext body (in which html is converted to plaintext).
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

func parseMessageHeader(h message.Header) (Message, error) { //nolint:funlen
	var m Message

	for fields := h.Fields(); fields.Next(); {
		switch strings.ToLower(fields.Key()) {
		case "subject":
			s, err := fields.Text()
			if err != nil {
				if s, err = pmmime.DecodeHeader(fields.Value()); err != nil {
					return Message{}, errors.Wrap(err, "failed to parse subject")
				}
			}

			m.Subject = s

		case "from":
			sender, err := rfc5322.ParseAddressList(fields.Value())
			if err != nil {
				return Message{}, errors.Wrap(err, "failed to parse from")
			}

			if len(sender) > 0 {
				m.Sender = sender[0]
			}

		case "to":
			toList, err := rfc5322.ParseAddressList(fields.Value())
			if err != nil {
				return Message{}, errors.Wrap(err, "failed to parse to")
			}

			m.ToList = toList

		case "reply-to":
			replyTos, err := rfc5322.ParseAddressList(fields.Value())
			if err != nil {
				return Message{}, errors.Wrap(err, "failed to parse reply-to")
			}

			m.ReplyTos = replyTos

		case "cc":
			ccList, err := rfc5322.ParseAddressList(fields.Value())
			if err != nil {
				return Message{}, errors.Wrap(err, "failed to parse cc")
			}

			m.CCList = ccList

		case "bcc":
			bccList, err := rfc5322.ParseAddressList(fields.Value())
			if err != nil {
				return Message{}, errors.Wrap(err, "failed to parse bcc")
			}

			m.BCCList = bccList

		case "message-id":
			m.ExternalID = regexp.MustCompile("<(.*)>").ReplaceAllString(fields.Value(), "$1")

		case "in-reply-to":
			m.InReplyTo = regexp.MustCompile("<(.*)>").ReplaceAllString(fields.Value(), "$1")

		case "references":
			for _, ref := range strings.Fields(fields.Value()) {
				for _, ref := range strings.Split(ref, ",") {
					m.References = append(m.References, strings.Trim(ref, "<>"))
				}
			}
		}
	}

	return m, nil
}

func parseAttachment(h message.Header, body []byte) (Attachment, error) {
	att := Attachment{
		Data: body,
	}

	mimeHeader, err := toMailHeader(h)
	if err != nil {
		return Attachment{}, err
	}
	att.Header = mimeHeader

	mimeType, mimeTypeParams, err := h.ContentType()
	if err != nil {
		return Attachment{}, err
	}
	att.MIMEType = mimeType

	// Prefer attachment name from filename param in content disposition.
	// If not available, try to get it from name param in content type.
	// Otherwise fallback to attachment.bin.
	if disp, dispParams, err := h.ContentDisposition(); err == nil {
		att.Disposition = disp

		if filename, ok := dispParams["filename"]; ok {
			att.Name = filename
		}
	}

	if att.Name == "" {
		if filename, ok := mimeTypeParams["name"]; ok {
			att.Name = filename
		} else if mimeType == string(rfc822.MessageRFC822) {
			att.Name = "message.eml"
		} else if ext, err := mime.ExtensionsByType(att.MIMEType); err == nil && len(ext) > 0 {
			att.Name = "attachment" + ext[0]
		} else {
			att.Name = "attachment.bin"
		}
	}

	// Only set ContentID if it should be inline;
	// API infers content disposition based on whether ContentID is present.
	// If Content-Disposition is present, we base our decision on that.
	// Otherwise, if Content-Disposition is missing but there is a ContentID, set it.
	// (This is necessary because some clients don't set Content-Disposition at all,
	// so we need to rely on other information to deduce if it's inline or attachment.)
	if h.Has("Content-Disposition") {
		disp, _, err := h.ContentDisposition()
		if err != nil {
			return Attachment{}, err
		}

		if disp == string(proton.InlineDisposition) {
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
