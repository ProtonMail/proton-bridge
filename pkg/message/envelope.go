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
	"time"

	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/emersion/go-imap"
	mail "github.com/ProtonMail/proton-bridge/pkg/net/mail"
)

func GetEnvelope(m *pmapi.Message) *imap.Envelope {
	messageID := m.ExternalID
	if messageID == "" {
		messageID = m.Header.Get("Message-Id")
	} else {
		messageID = "<" + messageID + ">"
	}

	return &imap.Envelope{
		Date:      time.Unix(m.Time, 0),
		Subject:   m.Subject,
		From:      getAddresses([]*mail.Address{m.Sender}),
		Sender:    getAddresses([]*mail.Address{m.Sender}),
		ReplyTo:   getAddresses(m.ReplyTos),
		To:        getAddresses(m.ToList),
		Cc:        getAddresses(m.CCList),
		Bcc:       getAddresses(m.BCCList),
		InReplyTo: m.Header.Get("In-Reply-To"),
		MessageId: messageID,
	}
}
