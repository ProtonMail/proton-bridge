// Copyright (c) 2024 Proton AG
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
	"encoding/base64"
	"io"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/pkg/errors"
)

type DecryptedAttachment struct {
	Packet    []byte
	Encrypted []byte
	Data      bytes.Buffer
	Err       error
}

type DecryptedMessage struct {
	Msg         proton.Message
	Body        bytes.Buffer
	BodyErr     error
	Attachments []DecryptedAttachment
}

var ErrInvalidAttachmentPacket = errors.New("invalid attachment packet")

func DecryptMessage(kr *crypto.KeyRing, msg proton.Message, attData [][]byte) DecryptedMessage {
	result := DecryptedMessage{
		Msg: msg,
	}

	result.Body.Grow(len(msg.Body))

	if err := msg.DecryptInto(kr, &result.Body); err != nil {
		result.BodyErr = errors.Wrap(ErrDecryptionFailed, err.Error())
	}

	result.Attachments = make([]DecryptedAttachment, len(msg.Attachments))

	for i, attachment := range msg.Attachments {
		result.Attachments[i].Encrypted = attData[i]

		kps, err := base64.StdEncoding.DecodeString(attachment.KeyPackets)
		if err != nil {
			result.Attachments[i].Err = errors.Wrap(ErrInvalidAttachmentPacket, err.Error())
			continue
		}

		result.Attachments[i].Packet = kps

		// Use io.Multi
		attachmentReader := io.MultiReader(bytes.NewReader(kps), bytes.NewReader(attData[i]))

		stream, err := kr.DecryptStream(attachmentReader, nil, crypto.GetUnixTime())
		if err != nil {
			result.Attachments[i].Err = errors.Wrap(ErrDecryptionFailed, err.Error())
			continue
		}

		result.Attachments[i].Data.Grow(len(kps) + len(attData))

		if _, err := result.Attachments[i].Data.ReadFrom(stream); err != nil {
			result.Attachments[i].Err = errors.Wrap(ErrDecryptionFailed, err.Error())
			continue
		}
	}

	return result
}
