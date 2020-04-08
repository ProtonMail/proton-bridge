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

package pmapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"sync"

	pmcrypto "github.com/ProtonMail/gopenpgp/crypto"
)

// clearableKey is a region of memory intended to hold a private key and which can be securely
// cleared by calling clear().
type clearableKey []byte

// UnmarshalJSON Removes quotation and unescapes CR, LF.
func (pk *clearableKey) UnmarshalJSON(b []byte) (err error) {
	b = bytes.Trim(b, "\"")
	b = bytes.ReplaceAll(b, []byte("\\n"), []byte("\n"))
	b = bytes.ReplaceAll(b, []byte("\\r"), []byte("\r"))
	*pk = b
	return
}

// clear irreversibly destroys the full range of `clearableKey` by filling it with zeros to ensure
// nobody can see what was in there (e.g. while waiting for the garbage collector to clean it up).
func (pk *clearableKey) clear() {
	for b := range *pk {
		(*pk)[b] = 0
	}
}

type PMKey struct {
	ID          string
	Version     int
	Flags       int
	Fingerprint string
	Primary     int
	Token       *string `json:",omitempty"`
	Signature   *string `json:",omitempty"`
}

type PMKeys struct {
	Keys    []PMKey
	KeyRing *pmcrypto.KeyRing
}

func (k *PMKeys) UnmarshalJSON(b []byte) (err error) {
	var rawKeys []struct {
		PMKey
		PrivateKey clearableKey
	}
	if err = json.Unmarshal(b, &rawKeys); err != nil {
		return
	}

	k.KeyRing = &pmcrypto.KeyRing{}
	for _, rawKey := range rawKeys {
		err = k.KeyRing.ReadFrom(bytes.NewReader(rawKey.PrivateKey), true)
		rawKey.PrivateKey.clear()
		if err != nil {
			return
		}
		k.Keys = append(k.Keys, rawKey.PMKey)
	}
	if len(k.Keys) > 0 {
		k.KeyRing.FirstKeyID = k.Keys[0].ID
	}
	return
}

// unlockKeyRing tries to unlock them with the provided keyRing using the token
// and if the token is not available it will use passphrase. It will not fail
// if keyring contains at least one unlocked private key.
func (k *PMKeys) unlockKeyRing(userKeyring *pmcrypto.KeyRing, passphrase []byte, locker sync.Locker) (err error) {
	locker.Lock()
	defer locker.Unlock()

	if k == nil {
		err = errors.New("keys is a nil object")
		return
	}

	for _, key := range k.Keys {
		if key.Token == nil || key.Signature == nil {
			if err = unlockKeyRingNoErrorWhenAlreadyUnlocked(k.KeyRing, passphrase); err != nil {
				return
			}
			continue
		}

		message, err := pmcrypto.NewPGPMessageFromArmored(*key.Token)
		if err != nil {
			return err
		}

		signature, err := pmcrypto.NewPGPSignatureFromArmored(*key.Signature)
		if err != nil {
			return err
		}

		if userKeyring == nil {
			return errors.New("userkey required to decrypt tokens but wasn't provided")
		}
		token, err := userKeyring.Decrypt(message, nil, 0)
		if err != nil {
			return err
		}

		err = userKeyring.VerifyDetached(token, signature, 0)
		if err != nil {
			return err
		}

		err = unlockKeyRingNoErrorWhenAlreadyUnlocked(k.KeyRing, token.GetBinary())
		if err != nil {
			return fmt.Errorf("wrong token: %v", err)
		}
	}

	return nil
}

type unlockError struct {
	error
}

func (err *unlockError) Error() string {
	return "Invalid mailbox password (" + err.error.Error() + ")"
}

// IsUnlockError checks whether the error is due to failure to unlock (which is represented by an unexported type).
func IsUnlockError(err error) bool {
	_, ok := err.(*unlockError)
	return ok
}

func unlockKeyRingNoErrorWhenAlreadyUnlocked(kr *pmcrypto.KeyRing, passphrase []byte) (err error) {
	if err = kr.Unlock(passphrase); err != nil {
		// Do not fail if it has already unlocked keys.
		hasUnlockedKey := false
		for _, e := range kr.GetEntities() {
			if e.PrivateKey != nil && !e.PrivateKey.Encrypted {
				hasUnlockedKey = true
				break
			}
			for _, se := range e.Subkeys {
				if se.PrivateKey != nil && (!se.Sig.FlagsValid || se.Sig.FlagEncryptStorage || se.Sig.FlagEncryptCommunications) && !e.PrivateKey.Encrypted {
					hasUnlockedKey = true
					break
				}
			}
			if hasUnlockedKey {
				break
			}
		}
		if !hasUnlockedKey {
			err = &unlockError{err}
			return
		}
		err = nil
	}
	return
}

// ErrNoKeyringAvailable represents an error caused by a keyring being nil or having no entities.
var ErrNoKeyringAvailable = errors.New("no keyring available")

func (c *Client) encrypt(plain string, signer *pmcrypto.KeyRing) (armored string, err error) {
	return encrypt(c.kr, plain, signer)
}

func encrypt(encrypter *pmcrypto.KeyRing, plain string, signer *pmcrypto.KeyRing) (armored string, err error) {
	if encrypter == nil || encrypter.FirstKey() == nil {
		return "", ErrNoKeyringAvailable
	}
	plainMessage := pmcrypto.NewPlainMessageFromString(plain)
	// We use only primary key to encrypt the message. Our keyring contains all keys (primary, old and deacivated ones).
	pgpMessage, err := encrypter.FirstKey().Encrypt(plainMessage, signer)
	if err != nil {
		return
	}
	return pgpMessage.GetArmored()
}

func (c *Client) decrypt(armored string) (plain string, err error) {
	return decrypt(c.kr, armored)
}

func decrypt(decrypter *pmcrypto.KeyRing, armored string) (plainBody string, err error) {
	if decrypter == nil {
		return "", ErrNoKeyringAvailable
	}
	pgpMessage, err := pmcrypto.NewPGPMessageFromArmored(armored)
	if err != nil {
		return
	}
	plainMessage, err := decrypter.Decrypt(pgpMessage, nil, 0)
	if err != nil {
		return
	}
	return plainMessage.GetString(), nil
}

func (c *Client) sign(plain string) (armoredSignature string, err error) {
	if c.kr == nil {
		return "", ErrNoKeyringAvailable
	}
	plainMessage := pmcrypto.NewPlainMessageFromString(plain)
	pgpSignature, err := c.kr.SignDetached(plainMessage)
	if err != nil {
		return
	}
	return pgpSignature.GetArmored()
}

func (c *Client) verify(plain, amroredSignature string) (err error) {
	plainMessage := pmcrypto.NewPlainMessageFromString(plain)
	pgpSignature, err := pmcrypto.NewPGPSignatureFromArmored(amroredSignature)
	if err != nil {
		return
	}
	verifyTime := int64(0) // By default it will use current timestamp.
	return c.kr.VerifyDetached(plainMessage, pgpSignature, verifyTime)
}

func encryptAttachment(kr *pmcrypto.KeyRing, data io.Reader, filename string) (encrypted io.Reader, err error) {
	if kr == nil || kr.FirstKey() == nil {
		return nil, ErrNoKeyringAvailable
	}
	dataBytes, err := ioutil.ReadAll(data)
	if err != nil {
		return
	}
	plainMessage := pmcrypto.NewPlainMessage(dataBytes)
	// We use only primary key to encrypt the message. Our keyring contains all keys (primary, old and deacivated ones).
	pgpSplitMessage, err := kr.FirstKey().EncryptAttachment(plainMessage, filename)
	if err != nil {
		return
	}
	packets := append(pgpSplitMessage.KeyPacket, pgpSplitMessage.DataPacket...)
	return bytes.NewReader(packets), nil
}

func decryptAttachment(kr *pmcrypto.KeyRing, keyPackets []byte, data io.Reader) (decrypted io.Reader, err error) {
	if kr == nil {
		return nil, ErrNoKeyringAvailable
	}
	dataBytes, err := ioutil.ReadAll(data)
	if err != nil {
		return
	}
	pgpSplitMessage := pmcrypto.NewPGPSplitMessage(keyPackets, dataBytes)
	plainMessage, err := kr.DecryptAttachment(pgpSplitMessage)
	if err != nil {
		return
	}
	return plainMessage.NewReader(), nil
}

func signAttachment(encrypter *pmcrypto.KeyRing, data io.Reader) (signature io.Reader, err error) {
	if encrypter == nil {
		return nil, ErrNoKeyringAvailable
	}
	dataBytes, err := ioutil.ReadAll(data)
	if err != nil {
		return
	}
	plainMessage := pmcrypto.NewPlainMessage(dataBytes)
	sig, err := encrypter.SignDetached(plainMessage)
	if err != nil {
		return
	}
	return bytes.NewReader(sig.GetBinary()), nil
}
