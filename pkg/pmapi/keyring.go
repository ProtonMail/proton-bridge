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

package pmapi

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type PMKey struct {
	ID          string
	Version     int
	Flags       int
	Fingerprint string
	PrivateKey  *crypto.Key
	Primary     int
	Token       string
	Active      Boolean
	Signature   string
}

type clearable []byte

func (c *clearable) UnmarshalJSON(b []byte) error {
	b = bytes.Trim(b, "\"")
	b = bytes.ReplaceAll(b, []byte("\\n"), []byte("\n"))
	b = bytes.ReplaceAll(b, []byte("\\r"), []byte("\r"))
	*c = b
	return nil
}

func (c *clearable) clear() {
	for i := range *c {
		(*c)[i] = 0
	}
}

func (key *PMKey) UnmarshalJSON(b []byte) (err error) {
	type _PMKey PMKey

	rawKey := struct {
		_PMKey
		PrivateKey clearable
	}{}

	defer rawKey.PrivateKey.clear()

	if err = json.Unmarshal(b, &rawKey); err != nil {
		return
	}

	*key = PMKey(rawKey._PMKey)

	if key.PrivateKey, err = crypto.NewKeyFromArmoredReader(bytes.NewReader(rawKey.PrivateKey)); err != nil {
		return errors.Wrap(err, "failed to create crypto key from armored private key")
	}

	return
}

func (key PMKey) getPassphraseFromToken(kr *crypto.KeyRing) (passphrase []byte, err error) {
	if kr == nil {
		return nil, errors.New("no user key was provided")
	}

	msg, err := crypto.NewPGPMessageFromArmored(key.Token)
	if err != nil {
		return
	}

	sig, err := crypto.NewPGPSignatureFromArmored(key.Signature)
	if err != nil {
		return
	}

	token, err := kr.Decrypt(msg, nil, 0)
	if err != nil {
		return
	}

	if err = kr.VerifyDetached(token, sig, 0); err != nil {
		return
	}

	return token.GetBinary(), nil
}

func (key PMKey) unlock(passphrase []byte) (unlockedKey *crypto.Key, err error) {
	if unlockedKey, err = key.PrivateKey.Unlock(passphrase); err != nil {
		return
	}

	ok, err := unlockedKey.Check()
	if err != nil {
		return
	}
	if !ok {
		err = errors.New("private and public keys do not match")
		return
	}

	return
}

type PMKeys []PMKey

// UnlockAll goes through each key and unlocks it, returning a keyring containing all unlocked keys,
// or an error if no keys could be unlocked.
// The passphrase is used to unlock the key unless the key's token and signature are both non-nil,
// in which case the given userkey is used to deduce the passphrase.
func (keys *PMKeys) UnlockAll(passphrase []byte, userKey *crypto.KeyRing) (kr *crypto.KeyRing, err error) {
	if kr, err = crypto.NewKeyRing(nil); err != nil {
		return
	}

	for _, key := range *keys {
		if !key.Active {
			logrus.WithField("fingerprint", key.Fingerprint).Warn("Skipping inactive key")
			continue
		}

		var secret []byte

		if key.Token == "" || key.Signature == "" {
			secret = passphrase
		} else if secret, err = key.getPassphraseFromToken(userKey); err != nil {
			return
		}

		k, unlockErr := key.unlock(secret)
		if unlockErr != nil {
			logrus.WithError(unlockErr).WithField("fingerprint", key.Fingerprint).Warn("Failed to unlock key")
			continue
		}

		if addKeyErr := kr.AddKey(k); addKeyErr != nil {
			logrus.WithError(addKeyErr).Warn("Failed to add key to keyring")
			continue
		}
	}

	if kr.CountEntities() == 0 {
		err = errors.New("no keys could be unlocked")
		return
	}

	return kr, err
}

// ErrNoKeyringAvailable represents an error caused by a keyring being nil or having no entities.
var ErrNoKeyringAvailable = errors.New("no keyring available")

func encrypt(encrypter *crypto.KeyRing, plain string, signer *crypto.KeyRing) (armored string, err error) {
	if encrypter == nil {
		return "", ErrNoKeyringAvailable
	}

	firstKey, err := encrypter.FirstKey()
	if err != nil {
		return "", err
	}

	plainMessage := crypto.NewPlainMessageFromString(plain)

	// We use only primary key to encrypt the message. Our keyring contains all keys (primary, old and deacivated ones).
	pgpMessage, err := firstKey.Encrypt(plainMessage, signer)
	if err != nil {
		return
	}
	return pgpMessage.GetArmored()
}

func (c *client) decrypt(armored string) (plain []byte, err error) {
	return decrypt(c.userKeyRing, armored)
}

func decrypt(decrypter *crypto.KeyRing, armored string) (plainBody []byte, err error) {
	if decrypter == nil {
		return nil, ErrNoKeyringAvailable
	}
	pgpMessage, err := crypto.NewPGPMessageFromArmored(armored)
	if err != nil {
		return
	}
	plainMessage, err := decrypter.Decrypt(pgpMessage, nil, 0)
	if err != nil {
		return
	}
	return plainMessage.GetBinary(), nil
}

func (c *client) verify(plain, amroredSignature string) (err error) {
	plainMessage := crypto.NewPlainMessageFromString(plain)
	pgpSignature, err := crypto.NewPGPSignatureFromArmored(amroredSignature)
	if err != nil {
		return
	}
	verifyTime := int64(0) // By default it will use current timestamp.
	return c.userKeyRing.VerifyDetached(plainMessage, pgpSignature, verifyTime)
}

func encryptAttachment(kr *crypto.KeyRing, data io.Reader, filename string) (encrypted io.Reader, err error) {
	if kr == nil {
		return nil, ErrNoKeyringAvailable
	}

	firstKey, err := kr.FirstKey()
	if err != nil {
		return nil, err
	}

	dataBytes, err := ioutil.ReadAll(data)
	if err != nil {
		return
	}

	plainMessage := crypto.NewPlainMessage(dataBytes)

	// We use only primary key to encrypt the message. Our keyring contains all keys (primary, old and deacivated ones).
	pgpSplitMessage, err := firstKey.EncryptAttachment(plainMessage, filename)
	if err != nil {
		return
	}

	packets := append(pgpSplitMessage.KeyPacket, pgpSplitMessage.DataPacket...) //nolint:gocritic

	return bytes.NewReader(packets), nil
}

func decryptAttachment(kr *crypto.KeyRing, keyPackets []byte, data io.Reader) (decrypted io.Reader, err error) {
	if kr == nil {
		return nil, ErrNoKeyringAvailable
	}
	dataBytes, err := ioutil.ReadAll(data)
	if err != nil {
		return
	}
	pgpSplitMessage := crypto.NewPGPSplitMessage(keyPackets, dataBytes)
	plainMessage, err := kr.DecryptAttachment(pgpSplitMessage)
	if err != nil {
		return
	}
	return plainMessage.NewReader(), nil
}

func signAttachment(encrypter *crypto.KeyRing, data io.Reader) (signature io.Reader, err error) {
	if encrypter == nil {
		return nil, ErrNoKeyringAvailable
	}
	dataBytes, err := ioutil.ReadAll(data)
	if err != nil {
		return
	}
	plainMessage := crypto.NewPlainMessage(dataBytes)
	sig, err := encrypter.SignDetached(plainMessage)
	if err != nil {
		return
	}
	return bytes.NewReader(sig.GetBinary()), nil
}

func encryptAndEncodeSessionKeys(
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

func encryptSymmDecryptKey(
	kr *crypto.KeyRing,
	textToEncrypt string,
) (decryptedKey *crypto.SessionKey, symEncryptedData []byte, err error) {
	// We use only primary key to encrypt the message. Our keyring contains all keys (primary, old and deacivated ones).
	firstKey, err := kr.FirstKey()
	if err != nil {
		return
	}

	pgpMessage, err := firstKey.Encrypt(crypto.NewPlainMessageFromString(textToEncrypt), kr)
	if err != nil {
		return
	}

	pgpSplitMessage, err := pgpMessage.SplitMessage()
	if err != nil {
		return
	}

	decryptedKey, err = kr.DecryptSessionKey(pgpSplitMessage.GetBinaryKeyPacket())
	if err != nil {
		return
	}

	symEncryptedData = pgpSplitMessage.GetBinaryDataPacket()

	return
}
