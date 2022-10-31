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

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
)

// Role values.
const (
	FreeUserRole = iota
	PaidMemberRole
	PaidAdminRole
)

// User status.
const (
	DeletedUser  = 0
	DisabledUser = 1
	ActiveUser   = 2
	VPNAdminUser = 3
	AdminUser    = 4
	SuperUser    = 5
)

// Delinquent values.
const (
	CurrentUser = iota
	AvailableUser
	OverdueUser
	DelinquentUser
	NoReceiveUser
)

// PMSignature values.
const (
	PMSignatureDisabled = iota
	PMSignatureEnabled
	PMSignatureLocked
)

// User holds the user details.
type User struct {
	ID         string
	Name       string
	UsedSpace  *int64
	Currency   string
	Credit     int
	MaxSpace   *int64
	MaxUpload  int64
	Role       int
	Private    int
	Subscribed int
	Services   int
	Deliquent  int

	Keys PMKeys

	VPN struct {
		Status         int
		ExpirationTime int
		PlanName       string
		MaxConnect     int
		MaxTier        int
	}
}

func (c *client) GetUser(ctx context.Context) (user *User, err error) {
	var res struct {
		User *User
	}

	if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		return r.SetResult(&res).Get("/users")
	}); err != nil {
		return nil, err
	}

	return res.User, nil
}

// unlockUser unlocks all the client's user keys using the given passphrase.
func (c *client) unlockUser(passphrase []byte) (err error) {
	if c.user == nil {
		return errors.New("user data is not loaded")
	}

	if c.userKeyRing, err = c.user.Keys.UnlockAll(passphrase, nil); err != nil {
		return errors.Wrap(err, "failed to unlock user keys")
	}

	return
}

// UpdateUser retrieves details about user and loads its addresses.
func (c *client) UpdateUser(ctx context.Context) (*User, error) {
	user, err := c.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	addresses, err := c.GetAddresses(ctx)
	if err != nil {
		return nil, err
	}

	c.user = user
	c.addresses = addresses
	c.manager.setSentryUserID(user.ID)

	return user, err
}

// CurrentUser returns currently active user or user will be updated.
func (c *client) CurrentUser(ctx context.Context) (*User, error) {
	if c.user != nil && len(c.addresses) != 0 {
		return c.user, nil
	}

	return c.UpdateUser(ctx)
}

// CurrentUser returns currently active user or user will be updated.
func (c *client) GetUserKeyRing() (*crypto.KeyRing, error) {
	if c.userKeyRing == nil {
		return nil, errors.New("user keyring is not available")
	}

	return c.userKeyRing, nil
}
