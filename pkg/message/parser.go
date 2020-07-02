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
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/mail"
	"strings"

	"github.com/ProtonMail/proton-bridge/pkg/message/parser"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/emersion/go-message"
	"github.com/jaytaylor/html2text"
)

func Parse(r io.Reader, key, keyName string) (m *pmapi.Message, mime, plain string, atts []io.Reader, err error) {
	p, err := parser.New(r)
	if err != nil {
		return
	}

	m = pmapi.NewMessage()

	if err = parseHeader(m, p.Root().Header); err != nil {
		return
	}

	if m.Attachments, atts, err = collectAttachments(p); err != nil {
		return
	}

	if m.Body, plain, err = collectBodyParts(p); err != nil {
		return
	}

	if key != "" {
		attachPublicKey(p.Root(), key, keyName)
	}

	if mime, err = writeMIMEMessage(p); err != nil {
		return
	}

	return
}

func collectAttachments(p *parser.Parser) (atts []*pmapi.Attachment, data []io.Reader, err error) {
	w := p.
		NewWalker().
		WithContentDispositionHandler("attachment", func(p *parser.Part, _ parser.PartHandler) (err error) {
			att, err := parseAttachment(p.Header)
			if err != nil {
				return
			}

			atts = append(atts, att)
			data = append(data, bytes.NewReader(p.Body))

			return
		})

	if err = w.Walk(); err != nil {
		return
	}

	return
}

func collectBodyParts(p *parser.Parser) (body, plain string, err error) {
	var parts, plainParts []string

	w := p.
		NewWalker().
		WithContentTypeHandler("text/plain", func(p *parser.Part) (err error) {
			parts = append(parts, string(p.Body))
			plainParts = append(plainParts, string(p.Body))
			return
		}).
		WithContentTypeHandler("text/html", func(p *parser.Part) (err error) {
			parts = append(parts, string(p.Body))

			text, err := html2text.FromString(string(p.Body))
			if err != nil {
				text = string(p.Body)
			}
			plainParts = append(plainParts, text)

			return
		})

	if err = w.Walk(); err != nil {
		return
	}

	return strings.Join(parts, "\r\n"), strings.Join(plainParts, "\r\n"), nil
}

func writeMIMEMessage(p *parser.Parser) (mime string, err error) {
	writer := p.
		NewWriter().
		WithCondition(func(p *parser.Part) (keep bool) {
			disp, _, err := p.Header.ContentDisposition()
			return err != nil || disp != "attachment"
		})

	buf := new(bytes.Buffer)

	if err = writer.Write(buf); err != nil {
		return
	}

	return buf.String(), nil
}

func attachPublicKey(p *parser.Part, key, keyName string) {
	h := message.Header{}

	h.Set("Content-Type", fmt.Sprintf(`application/pgp-key; name="%v"`, keyName))
	h.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%v.asc.pgp"`, keyName))
	h.Set("Content-Transfer-Encoding", "base64")

	// TODO: Split body at col width 72.

	p.AddChild(&parser.Part{
		Header: h,
		Body:   []byte(key),
	})
}

func parseHeader(m *pmapi.Message, h message.Header) (err error) {
	m.Header = make(mail.Header)

	fields := h.Fields()

	for fields.Next() {
		var text string

		if text, err = fields.Text(); err != nil {
			return
		}

		switch strings.ToLower(fields.Key()) {
		case "subject":
			m.Subject = text

		case "from":
			if m.Sender, err = mail.ParseAddress(text); err != nil {
				return
			}

		case "to":
			if m.ToList, err = mail.ParseAddressList(text); err != nil {
				return
			}

		case "reply-to":
			if m.ReplyTos, err = mail.ParseAddressList(text); err != nil {
				return
			}

		case "cc":
			if m.CCList, err = mail.ParseAddressList(text); err != nil {
				return
			}

		case "bcc":
			if m.BCCList, err = mail.ParseAddressList(text); err != nil {
				return
			}

		case "date":
			// TODO
		}
	}

	return
}

func parseAttachment(h message.Header) (att *pmapi.Attachment, err error) {
	att = &pmapi.Attachment{}

	if att.MIMEType, _, err = h.ContentType(); err != nil {
		return
	}

	if _, dispParams, dispErr := h.ContentDisposition(); dispErr != nil {
		var ext []string

		if ext, err = mime.ExtensionsByType(att.MIMEType); err != nil {
			return
		}

		if len(ext) > 0 {
			att.Name = "attachment" + ext[0]
		}
	} else {
		att.Name = dispParams["filename"]
	}

	// TODO: Set att.Header

	return
}
