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
	"context"

	"github.com/go-resty/resty/v2"
)

// Key flags.
const (
	UseToVerifyFlag = 1 << iota
	UseToEncryptFlag
)

type PublicKey struct {
	Flags     int
	PublicKey string
}

type RecipientType int

const (
	RecipientTypeInternal RecipientType = iota + 1
	RecipientTypeExternal
)

// GetPublicKeysForEmail returns all sending public keys for the given email address.
func (c *client) GetPublicKeysForEmail(ctx context.Context, email string) (keys []PublicKey, internal bool, err error) {
	var res struct {
		Keys          []PublicKey
		RecipientType RecipientType
	}

	if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		return r.SetResult(&res).SetQueryParam("Email", email).Get("/keys")
	}); err != nil {
		return nil, false, err
	}

	return res.Keys, res.RecipientType == RecipientTypeInternal, nil
}

// KeySalt contains id and salt for key.
type KeySalt struct {
	ID, KeySalt string
}

// GetKeySalts sends request to get list of key salts (n.b. locked route).
func (c *client) GetKeySalts(ctx context.Context) (keySalts []KeySalt, err error) {
	var res struct {
		KeySalts []KeySalt
	}

	if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		return r.SetResult(&res).Get("/keys/salts")
	}); err != nil {
		return nil, err
	}

	return res.KeySalts, nil
}
