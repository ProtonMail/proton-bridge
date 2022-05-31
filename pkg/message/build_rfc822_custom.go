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
	"fmt"
	"mime"

	"github.com/ProtonMail/gopenpgp/v2/constants"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/emersion/go-message"
)

// writeCustomTextPart writes an armored-PGP text part for a message body that couldn't be decrypted.
func writeCustomTextPart(
	w *message.Writer,
	msg *pmapi.Message,
	decError error,
) error {
	enc, err := crypto.NewPGPMessageFromArmored(msg.Body)
	if err != nil {
		return err
	}

	arm, err := enc.GetArmoredWithCustomHeaders(
		fmt.Sprintf("This message could not be decrypted: %v", decError),
		constants.ArmorHeaderVersion,
	)
	if err != nil {
		return err
	}

	var hdr message.Header

	hdr.SetContentType(msg.MIMEType, nil)

	part, err := w.CreatePart(hdr)
	if err != nil {
		return err
	}

	if _, err := part.Write([]byte(arm)); err != nil {
		return err
	}

	return nil
}

// writeCustomAttachmentPart writes an armored-PGP data part for an attachment that couldn't be decrypted.
func writeCustomAttachmentPart(
	w *message.Writer,
	att *pmapi.Attachment,
	msg *crypto.PGPMessage,
	decError error,
) error {
	arm, err := msg.GetArmoredWithCustomHeaders(
		fmt.Sprintf("This attachment could not be decrypted: %v", decError),
		constants.ArmorHeaderVersion,
	)
	if err != nil {
		return err
	}

	filename := mime.QEncoding.Encode("utf-8", att.Name+".pgp")

	var hdr message.Header

	hdr.SetContentType("application/octet-stream", map[string]string{"name": filename})
	hdr.SetContentDisposition(att.Disposition, map[string]string{"filename": filename})

	part, err := w.CreatePart(hdr)
	if err != nil {
		return err
	}

	if _, err := part.Write([]byte(arm)); err != nil {
		return err
	}

	return part.Close()
}
