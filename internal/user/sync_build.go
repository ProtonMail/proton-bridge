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
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/pkg/message"
)

type result[T any] struct {
	v T
	e error
}

func resOk[T any](v T) result[T] {
	return result[T]{v: v}
}

func resErr[T any](e error) result[T] {
	return result[T]{e: e}
}

func (r *result[T]) unwrap() T {
	if r.e != nil {
		panic(r.err)
	}

	return r.v
}

func (r *result[T]) unpack() (T, error) {
	return r.v, r.e
}

func (r *result[T]) err() error {
	return r.e
}

type buildRes struct {
	messageID string
	addressID string
	update    result[*imap.MessageCreated]
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

func buildRFC822(apiLabels map[string]proton.Label, full proton.FullMessage, addrKR *crypto.KeyRing) *buildRes {
	var update result[*imap.MessageCreated]

	if literal, err := message.BuildRFC822(addrKR, full.Message, full.AttData, defaultJobOpts()); err != nil {
		update = resErr[*imap.MessageCreated](fmt.Errorf("failed to build RFC822 for message %s: %w", full.ID, err))
	} else if created, err := newMessageCreatedUpdate(apiLabels, full.MessageMetadata, literal); err != nil {
		update = resErr[*imap.MessageCreated](fmt.Errorf("failed to create IMAP update for message %s: %w", full.ID, err))
	} else {
		update = resOk(created)
	}

	return &buildRes{
		messageID: full.ID,
		addressID: full.AddressID,
		update:    update,
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
		MailboxIDs:    mapTo[string, imap.MailboxID](wantLabels(apiLabels, message.LabelIDs)),
		ParsedMessage: parsedMessage,
	}, nil
}
