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

package transfer

import (
	"fmt"
	"mime"
	"time"

	mail "github.com/ProtonMail/proton-bridge/pkg/net/mail"
)

// Message is data holder passed between import and export.
type Message struct {
	ID      string
	Unread  bool
	Body    []byte
	Source  Mailbox
	Targets []Mailbox
}

// MessageStatus holds status for message used by progress manager.
type MessageStatus struct {
	eventTime time.Time // Time of adding message to the process.
	rule      *Rule     // Rule with source and target mailboxes.
	SourceID  string    // Message ID at the source.
	targetID  string    // Message ID at the target (if any).
	bodyHash  string    // Hash of the message body.

	exported  bool
	imported  bool
	exportErr error
	importErr error

	// Info about message displayed to user.
	// This is needed only for failed messages, but we cannot know in advance
	// which message will fail. We could clear it once the message passed
	// without any error. However, if we say one message takes about 100 bytes
	// in average, it's about 100 MB per million of messages, which is fine.
	Subject string
	From    string
	Time    time.Time
}

func (status *MessageStatus) String() string {
	return fmt.Sprintf("%s (%s, %s, %s): %s", status.SourceID, status.Subject, status.From, status.Time, status.GetErrorMessage())
}

func (status *MessageStatus) setDetailsFromHeader(header mail.Header) {
	dec := &mime.WordDecoder{}

	status.Subject = header.Get("subject")
	if subject, err := dec.Decode(status.Subject); err == nil {
		status.Subject = subject
	}

	status.From = header.Get("from")
	if from, err := dec.Decode(status.From); err == nil {
		status.From = from
	}

	if msgTime, err := header.Date(); err == nil {
		status.Time = msgTime
	}
}

func (status *MessageStatus) hasError(includeMissing bool) bool {
	return status.exportErr != nil || status.importErr != nil || (includeMissing && !status.imported)
}

// GetErrorMessage returns error message.
func (status *MessageStatus) GetErrorMessage() string {
	return status.getErrorMessage(true)
}

func (status *MessageStatus) getErrorMessage(includeMissing bool) string {
	if status.exportErr != nil {
		return fmt.Sprintf("failed to export: %s", status.exportErr)
	}
	if status.importErr != nil {
		return fmt.Sprintf("failed to import: %s", status.importErr)
	}
	if includeMissing && !status.imported {
		if !status.exported {
			return "failed to import: lost before read"
		}
		return "failed to import: lost in the process"
	}
	return ""
}
