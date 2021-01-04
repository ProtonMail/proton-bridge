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

package pmapi

import (
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
)

// Role values.
const (
	FreeUserRole = iota
	PaidMemberRole
	PaidAdminRole
)

// User status
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
	UsedSpace  int64
	Currency   string
	Credit     int
	MaxSpace   int64
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

// UserRes holds structure of JSON response.
type UserRes struct {
	Res

	User *User
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
func (c *client) UpdateUser() (user *User, err error) {
	req, err := c.NewRequest("GET", "/users", nil)
	if err != nil {
		return
	}

	var res UserRes
	if err = c.DoJSON(req, &res); err != nil {
		return
	}

	user, err = res.User, res.Err()
	if err != nil {
		return nil, err
	}

	c.user = user
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{
			ID: user.ID,
		})
	})

	var tmpList AddressList
	if tmpList, err = c.GetAddresses(); err == nil {
		c.addresses = tmpList
	}

	return user, err
}

// CurrentUser returns currently active user or user will be updated.
func (c *client) CurrentUser() (user *User, err error) {
	if c.user != nil && len(c.addresses) != 0 {
		user = c.user
		return
	}
	return c.UpdateUser()
}
