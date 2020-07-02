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
	"mime"
	"net/textproto"
	"strings"
	"time"

	pmmime "github.com/ProtonMail/proton-bridge/pkg/mime"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
)

// GetHeader builds the header for the message.
func GetHeader(msg *pmapi.Message) textproto.MIMEHeader { //nolint[funlen]
	h := make(textproto.MIMEHeader)

	// Copy the custom header fields if there are some.
	if msg.Header != nil {
		h = textproto.MIMEHeader(msg.Header)
	}

	// Add or rewrite fields.
	h.Set("Subject", pmmime.EncodeHeader(msg.Subject))
	if msg.Sender != nil {
		h.Set("From", pmmime.EncodeHeader(msg.Sender.String()))
	}
	if len(msg.ReplyTos) > 0 {
		h.Set("Reply-To", pmmime.EncodeHeader(formatAddressList(msg.ReplyTos)))
	}
	if len(msg.ToList) > 0 {
		h.Set("To", pmmime.EncodeHeader(formatAddressList(msg.ToList)))
	}
	if len(msg.CCList) > 0 {
		h.Set("Cc", pmmime.EncodeHeader(formatAddressList(msg.CCList)))
	}
	if len(msg.BCCList) > 0 {
		h.Set("Bcc", pmmime.EncodeHeader(formatAddressList(msg.BCCList)))
	}

	// Add or rewrite date related fields.
	if msg.Time > 0 {
		h.Set("X-Pm-Date", time.Unix(msg.Time, 0).Format(time.RFC1123Z))
		if d, err := msg.Header.Date(); err != nil || d.IsZero() { // Fix date if needed.
			h.Set("Date", time.Unix(msg.Time, 0).Format(time.RFC1123Z))
		}
	}

	// Use External-Id if available to ensure email clients:
	// * build the conversations threads correctly (Thunderbird, Mac Outlook, Apple Mail)
	// * do not think the message is lost (Apple Mail)
	if msg.ExternalID != "" {
		h.Set("X-Pm-External-Id", "<"+msg.ExternalID+">")
		if h.Get("Message-Id") == "" {
			h.Set("Message-Id", "<"+msg.ExternalID+">")
		}
	}
	if msg.ID != "" {
		if h.Get("Message-Id") == "" {
			h.Set("Message-Id", "<"+msg.ID+"@"+pmapi.InternalIDDomain+">")
		}
		h.Set("X-Pm-Internal-Id", msg.ID)
		// Forward References, and include the message ID here (to improve outlook support).
		if references := h.Get("References"); !strings.Contains(references, msg.ID) {
			references += " <" + msg.ID + "@" + pmapi.InternalIDDomain + ">"
			h.Set("References", references)
		}
	}
	if msg.ConversationID != "" {
		h.Set("X-Pm-ConversationID-Id", msg.ConversationID)
		if references := h.Get("References"); !strings.Contains(references, msg.ConversationID) {
			references += " <" + msg.ConversationID + "@" + pmapi.ConversationIDDomain + ">"
			h.Set("References", references)
		}
	}

	return h
}

func SetBodyContentFields(h *textproto.MIMEHeader, m *pmapi.Message) {
	h.Set("Content-Type", m.MIMEType+"; charset=utf-8")
	h.Set("Content-Disposition", "inline")
	h.Set("Content-Transfer-Encoding", "quoted-printable")
}

func GetBodyHeader(m *pmapi.Message) textproto.MIMEHeader {
	h := make(textproto.MIMEHeader)
	SetBodyContentFields(&h, m)
	return h
}

func GetRelatedHeader(m *pmapi.Message) textproto.MIMEHeader {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Type", "multipart/related; boundary="+GetRelatedBoundary(m))
	return h
}

func GetAttachmentHeader(att *pmapi.Attachment) textproto.MIMEHeader {
	mediaType := att.MIMEType
	if mediaType == "application/pgp-encrypted" {
		mediaType = "application/octet-stream"
	}

	encodedName := pmmime.EncodeHeader(att.Name)
	disposition := "attachment" //nolint[goconst]
	if strings.Contains(att.Header.Get("Content-Disposition"), "inline") {
		disposition = "inline"
	}

	h := make(textproto.MIMEHeader)
	h.Set("Content-Type", mime.FormatMediaType(mediaType, map[string]string{"name": encodedName}))
	h.Set("Content-Transfer-Encoding", "base64")
	h.Set("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": encodedName}))

	// Forward some original header lines.
	forward := []string{"Content-Id", "Content-Description", "Content-Location"}
	for _, k := range forward {
		v := att.Header.Get(k)
		if v != "" {
			h.Set(k, v)
		}
	}

	return h
}
