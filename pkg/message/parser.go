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

func Parse(r io.Reader, key, keyName string) (m *pmapi.Message, mimeMessage, plainBody string, attReaders []io.Reader, err error) {
	p, err := parser.New(r)
	if err != nil {
		return
	}

	m = pmapi.NewMessage()

	if err = parseHeader(m, p.Root().Header); err != nil {
		return
	}

	atts, attReaders, err := collectAttachments(p)
	if err != nil {
		return
	}
	m.Attachments = atts

	richBody, plainBody, err := collectBodyParts(p)
	if err != nil {
		return
	}
	m.Body = richBody

	mimeType, err := determineMIMEType(p)
	if err != nil {
		return
	}
	m.MIMEType = mimeType

	if key != "" {
		attachPublicKey(p.Root(), key, keyName)
	}

	if mimeMessage, err = writeMIMEMessage(p); err != nil {
		return
	}

	return
}

func collectAttachments(p *parser.Parser) (atts []*pmapi.Attachment, data []io.Reader, err error) {
	w := p.NewWalker()

	w.RegisterContentDispositionHandler("attachment").
		OnEnter(func(p *parser.Part, _ parser.PartHandlerFunc) (err error) {
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

// collectBodyParts returns a richtext body (used for normal sending)
// and a plaintext body (used for sending to recipients that prefer plaintext).
func collectBodyParts(p *parser.Parser) (richBody, plainBody string, err error) {
	var richParts, plainParts []string

	w := p.NewWalker()

	w.RegisterContentTypeHandler("text/plain").
		OnEnter(func(p *parser.Part) error {
			plainParts = append(plainParts, string(p.Body))

			if !isAlternative(p) {
				richParts = append(richParts, string(p.Body))
			}

			return nil
		})

	w.RegisterContentTypeHandler("text/html").
		OnEnter(func(p *parser.Part) error {
			richParts = append(richParts, string(p.Body))

			if !isAlternative(p) {
				plain, htmlErr := html2text.FromString(string(p.Body))
				if htmlErr != nil {
					plain = string(p.Body)
				}
				plainParts = append(plainParts, plain)
			}

			return nil
		})

	if err = w.Walk(); err != nil {
		return
	}

	return strings.Join(richParts, "\r\n"), strings.Join(plainParts, "\r\n"), nil
}

func isAlternative(p *parser.Part) bool {
	parent := p.Parent()
	if parent == nil {
		return false
	}

	t, _, err := parent.Header.ContentType()
	if err != nil {
		return false
	}

	return t == "multipart/alternative"
}

func determineMIMEType(p *parser.Parser) (string, error) {
	w := p.NewWalker()

	var isHTML bool

	w.RegisterContentTypeHandler("text/html").
		OnEnter(func(p *parser.Part) (err error) {
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
