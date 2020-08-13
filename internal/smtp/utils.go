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

package smtp

import (
	"encoding/base64"
	"regexp"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/pkg/pmapi"
)

//nolint:gochecknoglobals // Used like a constant
var mailFormat = regexp.MustCompile(`.+@.+\..+`)

// looksLikeEmail validates whether the string resembles an email.
//
// Notice that it does this naively by simply checking for the existence
// of a DOT and an AT sign.
func looksLikeEmail(e string) bool {
	return mailFormat.MatchString(e)
}

func createPackets(
	pubkey *crypto.KeyRing,
	bodyKey *crypto.SessionKey,
	attkeys map[string]*crypto.SessionKey,
) (bodyPacket string, attachmentPackets map[string]string, err error) {
	// Encrypt message body keys.
	packetBytes, err := pubkey.EncryptSessionKey(bodyKey)
	if err != nil {
		return
	}
	bodyPacket = base64.StdEncoding.EncodeToString(packetBytes)

	// Encrypt attachment keys.
	attachmentPackets = make(map[string]string)
	for id, attkey := range attkeys {
		var packets []byte
		if packets, err = pubkey.EncryptSessionKey(attkey); err != nil {
			return
		}
		attachmentPackets[id] = base64.StdEncoding.EncodeToString(packets)
	}
	return
}

func encryptSymmetric(
	kr *crypto.KeyRing,
	textToEncrypt string,
	canonicalizeText bool, // nolint[unparam]
) (key *crypto.SessionKey, symEncryptedData []byte, err error) {
	// We use only primary key to encrypt the message. Our keyring contains all keys (primary, old and deacivated ones).
	firstKey, err := kr.FirstKey()
	if err != nil {
		return
	}

	pgpMessage, err := firstKey.Encrypt(crypto.NewPlainMessageFromString(textToEncrypt), kr)
	if err != nil {
		return
	}

	pgpSplitMessage, err := pgpMessage.SeparateKeyAndData(len(textToEncrypt), 0)
	if err != nil {
		return
	}

	key, err = kr.DecryptSessionKey(pgpSplitMessage.GetBinaryKeyPacket())
	if err != nil {
		return
	}

	symEncryptedData = pgpSplitMessage.GetBinaryDataPacket()

	return
}

func buildPackage(
	addressMap map[string]*pmapi.MessageAddress,
	sharedScheme int,
	mimeType string,
	bodyData []byte,
	bodyKey *crypto.SessionKey,
	attKeys map[string]pmapi.AlgoKey,
) (pkg *pmapi.MessagePackage) {
	if len(addressMap) == 0 {
		return nil
	}

	pkg = &pmapi.MessagePackage{
		Body:      base64.StdEncoding.EncodeToString(bodyData),
		Addresses: addressMap,
		MIMEType:  mimeType,
		Type:      sharedScheme,
	}

	if sharedScheme|pmapi.ClearPackage > 0 {
		pkg.BodyKey.Key = bodyKey.GetBase64Key()
		pkg.BodyKey.Algorithm = bodyKey.Algo
		pkg.AttachmentKeys = attKeys
	}
	return pkg
}
