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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package user

import (
	"fmt"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/pkg/message"
	"github.com/bradenaw/juniper/xslices"
	"gitlab.protontech.ch/go/liteapi"
)

type buildRes struct {
	messageID string
	addressID string
	update    *imap.MessageCreated
}

func defaultJobOpts() message.JobOptions {
	return message.JobOptions{
		IgnoreDecryptionErrors: true, // Whether to ignore decryption errors and create a "custom message" instead.
		SanitizeDate:           true, // Whether to replace all dates before 1970 with RFC822's birthdate.
		AddInternalID:          true, // Whether to include MessageID as X-Pm-Internal-Id.
		AddExternalID:          true, // Whether to include ExternalID as X-Pm-External-Id.
		AddMessageDate:         true, // Whether to include message time as X-Pm-Date.
		AddMessageIDReference:  true, // Whether to include the MessageID in References.
	}
}

func buildRFC822(full liteapi.FullMessage, addrKR *crypto.KeyRing) (*buildRes, error) {
	literal, err := message.BuildRFC822(addrKR, full.Message, full.AttData, defaultJobOpts())
	if err != nil {
		return nil, fmt.Errorf("failed to build message %s: %w", full.ID, err)
	}

	update, err := newMessageCreatedUpdate(full.MessageMetadata, literal)
	if err != nil {
		return nil, fmt.Errorf("failed to create IMAP update for message %s: %w", full.ID, err)
	}

	return &buildRes{
		messageID: full.ID,
		addressID: full.AddressID,
		update:    update,
	}, nil
}

func newMessageCreatedUpdate(message liteapi.MessageMetadata, literal []byte) (*imap.MessageCreated, error) {
	parsedMessage, err := imap.NewParsedMessage(literal)
	if err != nil {
		return nil, err
	}

	return &imap.MessageCreated{
		Message:       toIMAPMessage(message),
		Literal:       literal,
		MailboxIDs:    mapTo[string, imap.MailboxID](xslices.Filter(message.LabelIDs, wantLabelID)),
		ParsedMessage: parsedMessage,
	}, nil
}
