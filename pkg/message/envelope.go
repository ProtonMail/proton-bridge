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
	"net/mail"
	"net/textproto"
	"strings"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/emersion/go-imap"
)

// GetEnvelope will prepare envelope from pmapi message and cached header.
func GetEnvelope(msg *pmapi.Message, header textproto.MIMEHeader) *imap.Envelope {
	hdr := toMessageHeader(mail.Header(header))
	setMessageIDIfNeeded(msg, &hdr)

	return &imap.Envelope{
		Date:      SanitizeMessageDate(msg.Time),
		Subject:   msg.Subject,
		From:      getAddresses([]*mail.Address{msg.Sender}),
		Sender:    getAddresses([]*mail.Address{msg.Sender}),
		ReplyTo:   getAddresses(msg.ReplyTos),
		To:        getAddresses(msg.ToList),
		Cc:        getAddresses(msg.CCList),
		Bcc:       getAddresses(msg.BCCList),
		InReplyTo: hdr.Get("In-Reply-To"),
		MessageId: hdr.Get("Message-Id"),
	}
}

func getAddresses(addrs []*mail.Address) (imapAddrs []*imap.Address) {
	for _, a := range addrs {
		if a == nil {
			continue
		}

		parts := strings.SplitN(a.Address, "@", 2)
		if len(parts) != 2 {
			continue
		}

		imapAddrs = append(imapAddrs, &imap.Address{
			PersonalName: a.Name,
			MailboxName:  parts[0],
			HostName:     parts[1],
		})
	}

	return
}
