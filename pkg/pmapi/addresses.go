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
	"strings"

	pmcrypto "github.com/ProtonMail/gopenpgp/crypto"
)

// Address statuses.
const (
	DisabledAddress = iota
	EnabledAddress
)

// Address receive values.
const (
	CannotReceive = iota
	CanReceive
)

// Address HasKeys values.
const (
	MissingKeys = iota
	KeysPresent
)

// Address types.
const (
	_ = iota // Skip first.
	OriginalAddress
	AliasAddress
	CustomAddress
	PremiumAddress
)

// Address Send values.
const (
	NoSendAddress = iota
	MainSendAddress
	SecondarySendAddress
)

// Address represents a user's address.
type Address struct {
	ID          string
	DomainID    string
	Email       string
	Send        int
	Receive     int
	Status      int
	Order       int `json:",omitempty"`
	Type        int
	DisplayName string
	Signature   string
	MemberID    string `json:",omitempty"`
	MemberName  string `json:",omitempty"`

	HasKeys int
	Keys    PMKeys
}

// AddressList is a list of addresses.
type AddressList []*Address

type AddressesRes struct {
	Res
	Addresses AddressList
}

// KeyRing returns the (possibly unlocked) PMKeys KeyRing.
func (a *Address) KeyRing() *pmcrypto.KeyRing {
	return a.Keys.KeyRing
}

// ByID returns an address by id. Returns nil if no address is found.
func (l AddressList) ByID(id string) *Address {
	for _, addr := range l {
		if addr.ID == id {
			return addr
		}
	}
	return nil
}

func (l AddressList) ActiveEmails() (addresses []string) {
	for _, a := range l {
		if a.Receive == CanReceive {
			addresses = append(addresses, a.Email)
		}
	}
	return
}

// Main gets the main address.
func (l AddressList) Main() *Address {
	for _, addr := range l {
		if addr.Order == 1 {
			return addr
		}
	}
	return l[0] // Should not happen.
}

// ByEmail gets an address by email. Returns nil if no address is found.
func (l AddressList) ByEmail(email string) *Address {
	email = SanitizeEmail(email)
	for _, addr := range l {
		if strings.EqualFold(addr.Email, email) {
			return addr
		}
	}
	return nil
}

func SanitizeEmail(email string) string {
	splitAt := strings.Split(email, "@")
	if len(splitAt) != 2 {
		return email
	}
	splitPlus := strings.Split(splitAt[0], "+")
	email = splitPlus[0] + "@" + splitAt[1]
	return email
}

func ConstructAddress(headerEmail string, addressEmail string) string {
	splitAtHeader := strings.Split(headerEmail, "@")
	if len(splitAtHeader) != 2 {
		return addressEmail
	}

	splitPlus := strings.Split(splitAtHeader[0], "+")
	if len(splitPlus) != 2 {
		return addressEmail
	}

	splitAtAddress := strings.Split(addressEmail, "@")
	if len(splitAtAddress) != 2 {
		return addressEmail
	}

	return splitAtAddress[0] + "+" + splitPlus[1] + "@" + splitAtAddress[1]
}

// GetAddresses requests all of current user addresses (without pagination).
func (c *Client) GetAddresses() (addresses AddressList, err error) {
	req, err := NewRequest("GET", "/addresses", nil)
	if err != nil {
		return
	}

	var res AddressesRes
	if err = c.DoJSON(req, &res); err != nil {
		return
	}

	return res.Addresses, res.Err()
}

func (c *Client) Addresses() AddressList {
	return c.addresses
}

// UnlockAddresses unlocks all keys for all addresses of current user.
func (c *Client) UnlockAddresses(passphrase []byte) (err error) {
	for _, a := range c.addresses {
		if a.HasKeys == MissingKeys {
			continue
		}

		// Unlock the address token using the UserKey, use the unlocked token to unlock the keyring.
		if err = a.Keys.unlockKeyRing(c.kr, passphrase, c.keyLocker); err != nil {
			err = fmt.Errorf("pmapi: cannot unlock private key of address %v: %v", a.Email, err)
			return
		}
	}

	return
}

func (c *Client) KeyRingForAddressID(addrID string) *pmcrypto.KeyRing {
	addr := c.addresses.ByID(addrID)
	if addr == nil {
		addr = c.addresses.Main()
	}
	return addr.KeyRing()
}
