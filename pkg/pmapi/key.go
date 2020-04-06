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
	"fmt"
	"net/http"
	"net/url"
	"strings"

	pmcrypto "github.com/ProtonMail/gopenpgp/crypto"
)

// Flags
const (
	UseToVerifyFlag = 1 << iota
	UseToEncryptFlag
)

type PublicKeyRes struct {
	Res

	RecipientType int
	MIMEType      string
	Keys          []PublicKey
}

type PublicKey struct {
	Flags     int
	PublicKey string
}

// PublicKeys returns the public keys of the given email addresses.
func (c *client) PublicKeys(emails []string) (keys map[string]*pmcrypto.KeyRing, err error) {
	if len(emails) == 0 {
		err = fmt.Errorf("pmapi: cannot get public keys: no email address provided")
		return
	}
	keys = map[string]*pmcrypto.KeyRing{}

	for _, email := range emails {
		email = url.QueryEscape(email)

		var req *http.Request
		if req, err = c.NewRequest("GET", "/keys?Email="+email, nil); err != nil {
			return
		}

		var res PublicKeyRes
		if err = c.DoJSON(req, &res); err != nil {
			return
		}

		for _, key := range res.Keys {
			if key.Flags&UseToEncryptFlag == UseToEncryptFlag {
				var kr *pmcrypto.KeyRing
				if kr, err = pmcrypto.ReadArmoredKeyRing(strings.NewReader(key.PublicKey)); err != nil {
					return
				}
				keys[email] = kr
			}
		}
	}

	return keys, err
}

const (
	RecipientInternal = 1
	RecipientExternal = 2
)

// GetPublicKeysForEmail returns all sending public keys for the given email address.
func (c *client) GetPublicKeysForEmail(email string) (keys []PublicKey, internal bool, err error) {
	email = url.QueryEscape(email)

	var req *http.Request
	if req, err = c.NewRequest("GET", "/keys?Email="+email, nil); err != nil {
		return
	}

	var res PublicKeyRes
	if err = c.DoJSON(req, &res); err != nil {
		return
	}

	internal = res.RecipientType == RecipientInternal

	for _, key := range res.Keys {
		if key.Flags&UseToEncryptFlag == UseToEncryptFlag {
			keys = append(keys, key)
		}
	}
	return
}

// KeySalt contains id and salt for key.
type KeySalt struct {
	ID, KeySalt string
}

// KeySaltRes is used to unmarshal API response.
type KeySaltRes struct {
	Res
	KeySalts []KeySalt
}

// GetKeySalts sends request to get list of key salts (n.b. locked route).
func (c *client) GetKeySalts() (keySalts []KeySalt, err error) {
	var req *http.Request
	if req, err = c.NewRequest("GET", "/keys/salts", nil); err != nil {
		return
	}

	var res KeySaltRes
	if err = c.DoJSON(req, &res); err != nil {
		return
	}

	keySalts = res.KeySalts

	return
}
