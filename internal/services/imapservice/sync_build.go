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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package imapservice

import (
	"bytes"
	"html/template"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/internal/usertypes"
	"github.com/ProtonMail/proton-bridge/v3/pkg/algo"
	"github.com/ProtonMail/proton-bridge/v3/pkg/message"
	"github.com/bradenaw/juniper/xslices"
)

type buildRes struct {
	messageID string
	addressID string
	update    *imap.MessageCreated
	err       error
}

func defaultMessageJobOpts() message.JobOptions {
	return message.JobOptions{
		IgnoreDecryptionErrors: true, // Whether to ignore decryption errors and create a "custom message" instead.
		SanitizeDate:           true, // Whether to replace all dates before 1970 with RFC822's birthdate.
		AddInternalID:          true, // Whether to include MessageID as X-Pm-Internal-Id.
		AddExternalID:          true, // Whether to include ExternalID as X-Pm-External-Id.
		AddMessageDate:         true, // Whether to include message time as X-Pm-Date.
		AddMessageIDReference:  true, // Whether to include the MessageID in References.
	}
}

func buildRFC822(apiLabels map[string]proton.Label, full proton.FullMessage, addrKR *crypto.KeyRing, buffer *bytes.Buffer) *buildRes {
	var (
		update *imap.MessageCreated
		err    error
	)

	buffer.Grow(full.Size)

	if buildErr := message.DecryptAndBuildRFC822Into(addrKR, full.Message, full.AttData, defaultMessageJobOpts(), buffer); buildErr != nil {
		update = newMessageCreatedFailedUpdate(apiLabels, full.MessageMetadata, buildErr)
		err = buildErr
	} else if created, parseErr := newMessageCreatedUpdate(apiLabels, full.MessageMetadata, buffer.Bytes()); parseErr != nil {
		update = newMessageCreatedFailedUpdate(apiLabels, full.MessageMetadata, parseErr)
		err = parseErr
	} else {
		update = created
	}

	return &buildRes{
		messageID: full.ID,
		addressID: full.AddressID,
		update:    update,
		err:       err,
	}
}

func newMessageCreatedUpdate(
	apiLabels map[string]proton.Label,
	message proton.MessageMetadata,
	literal []byte,
) (*imap.MessageCreated, error) {
	parsedMessage, err := imap.NewParsedMessage(literal)
	if err != nil {
		return nil, err
	}

	return &imap.MessageCreated{
		Message:       toIMAPMessage(message),
		Literal:       literal,
		MailboxIDs:    usertypes.MapTo[string, imap.MailboxID](wantLabels(apiLabels, message.LabelIDs)),
		ParsedMessage: parsedMessage,
	}, nil
}

func newMessageCreatedFailedUpdate(
	apiLabels map[string]proton.Label,
	message proton.MessageMetadata,
	err error,
) *imap.MessageCreated {
	literal := newFailedMessageLiteral(message.ID, time.Unix(message.Time, 0), message.Subject, err)

	parsedMessage, err := imap.NewParsedMessage(literal)
	if err != nil {
		panic(err)
	}

	return &imap.MessageCreated{
		Message:       toIMAPMessage(message),
		MailboxIDs:    usertypes.MapTo[string, imap.MailboxID](wantLabels(apiLabels, message.LabelIDs)),
		Literal:       literal,
		ParsedMessage: parsedMessage,
	}
}

func newFailedMessageLiteral(
	messageID string,
	date time.Time,
	subject string,
	syncErr error,
) []byte {
	var buf bytes.Buffer

	if tmpl, err := template.New("header").Parse(failedMessageHeaderTemplate); err != nil {
		panic(err)
	} else if b, err := tmplExec(tmpl, map[string]any{
		"Date": date.In(time.UTC).Format(time.RFC822),
	}); err != nil {
		panic(err)
	} else if _, err := buf.Write(b); err != nil {
		panic(err)
	}

	if tmpl, err := template.New("body").Parse(failedMessageBodyTemplate); err != nil {
		panic(err)
	} else if b, err := tmplExec(tmpl, map[string]any{
		"MessageID": messageID,
		"Subject":   subject,
		"Error":     syncErr.Error(),
	}); err != nil {
		panic(err)
	} else if _, err := buf.Write(lineWrap(algo.B64Encode(b))); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func tmplExec(template *template.Template, data any) ([]byte, error) {
	var buf bytes.Buffer

	if err := template.Execute(&buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func lineWrap(b []byte) []byte {
	return bytes.Join(xslices.Chunk(b, 76), []byte{'\r', '\n'})
}

const failedMessageHeaderTemplate = `Date: {{.Date}}
Subject: Message failed to build
Content-Type: text/plain
Content-Transfer-Encoding: base64

`

const failedMessageBodyTemplate = `Failed to build message: 
Subject:   {{.Subject}}
Error:     {{.Error}}
MessageID: {{.MessageID}}
`
