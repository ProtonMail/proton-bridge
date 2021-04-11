// Copyright (c) 2021 Proton Technologies AG
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
	"fmt"
	"mime"

	"github.com/ProtonMail/gopenpgp/v2/constants"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
	"github.com/emersion/go-message"
)

/* writeCustomTextPart writes an armored-PGP text part for a message body that couldn't be decrypted.
The following is an example of such a message:

Mime-Version: 1.0
Content-Type: multipart/mixed;
 boundary=d8a04ef2e12150946d27f84fc82b9e70a5e314f91a0f2190e38c7aad623466da
Message-Id: <messageID@protonmail.internalid>
Date: Fri, 26 Mar 2021 14:10:00 +0000

--d8a04ef2e12150946d27f84fc82b9e70a5e314f91a0f2190e38c7aad623466da
Content-Type: text/plain

-----BEGIN PGP MESSAGE-----
Version: GopenPGP 2.1.3
Comment: This message could not be decrypted: gopenpgp: error in reading message: openpgp: incorrect key

wcBMA1iOyuyBImpHAQf/boKq75jyV5kGzl8ObcjLuaVrVEw5YbrI1M9gAb/5eq+t
Mzd2qK0owYwBmuvYj3pk2b15miOIPfAYlbQVyIaseq/O4XchPvtJZKxh42eDcYhD
zwqHrYxtCLOOBY6b15Aq8tVLORi4+/VCWJrTHtCdKq71UX7bvBdQRaZqOIJ+bUiP
72IkaWjdVmCgLCmzPN3R3VKsmjTNBt1GRMc6vosQyoAeClyf94wgE1sXGsumXBQs
MX69zK9a36Ya67PpQ9AGVR2peEF07s5O+t/PxkStWfFvMwnnOOSgdHfN0PSwkhoL
EIwDA6lh7+2UlbX9pFi908PDl0OuoLgrL+NyleUPtdLAugGUa0U1fn6K1gwFGcsj
DKflgAMXixhf83xpmz4fM3zzM988ecQsXozQcUDrxSTxzdfWXSaWQDWHguCiL0xM
aPkf4kUn4RX3AoZVTC4GXvupzAQUqjJhyd0p7lJGOwIw0Ik77PxeCFNvWiZHJ3W+
7PYuehSUgyDJJe7q1GFFcMUspnPmuvhbenKIHlKwGfptnXLRIW15sdcca2Etz5i4
H71rHKwzAPdNHBQVvN0Nmt77RGAIy4Az8fRwz7hD7E92r4BkaaYntrQe9DUXYc50
HVXyeFcvpfbEaK+fObxk1ljig+Io2liPx7pn/YBFytMy1mlwhjroMPjMUXuTXqje
NHnOJVxnnXMwGh34wUIEwOmTskJoGkQjQbjvixkwNRGnLzIHuD6614pXtgTDtBzc
AKuQssqOky1Q7KjdusMWnoVL7gTbSrIUxXoKChear5XgPzFevQhQNEfTI8DDMbYR
La1Cc6aFtMPUShzNc3Wp3dVLM09WygEUsYZFmA==
=00ma
-----END PGP MESSAGE-----
--d8a04ef2e12150946d27f84fc82b9e70a5e314f91a0f2190e38c7aad623466da--

.
*/
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

	// Even if it is HTML, we don't care... we abuse pgp/inline here.
	hdr.SetContentType("text/plain", nil)

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

	var hdr message.Header

	hdr.SetContentType("application/pgp-encrypted", map[string]string{"name": mime.QEncoding.Encode("utf-8", att.Name+".pgp")})

	hdr.SetContentDisposition(att.Disposition, map[string]string{"filename": mime.QEncoding.Encode("utf-8", att.Name+".pgp")})

	part, err := w.CreatePart(hdr)
	if err != nil {
		return err
	}

	if _, err := part.Write([]byte(arm)); err != nil {
		return err
	}

	return part.Close()
}
