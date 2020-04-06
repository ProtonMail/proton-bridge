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
	pmcrypto "github.com/ProtonMail/gopenpgp/crypto"
	"github.com/getsentry/raven-go"
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
	VPN        struct {
		Status         int
		ExpirationTime int
		PlanName       string
		MaxConnect     int
		MaxTier        int
	}
	Deliquent int
	Keys      PMKeys
}

// UserRes holds structure of JSON response.
type UserRes struct {
	Res

	User *User
}

// KeyRing returns the (possibly unlocked) PMKeys KeyRing.
func (u *User) KeyRing() *pmcrypto.KeyRing {
	return u.Keys.KeyRing
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
	raven.SetUserContext(&raven.User{ID: user.ID})

	var tmpList AddressList
	if tmpList, err = c.GetAddresses(); err == nil {
		c.addresses = tmpList
	}

	c.log.WithField("userID", user.ID).Info("Updated user")

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
