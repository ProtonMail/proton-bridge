// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
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

package accounts

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
)

const (
	testUserKey    = "user_key.json"
	testAddressKey = "address_key.json"
)

type TestAccount struct {
	user                    *pmapi.User
	addressToBeUsed         *pmapi.Address
	addressesByBDDAddressID map[string]*pmapi.Address
	password                string
	mailboxPassword         string
	twoFAEnabled            bool
}

func newTestAccount(
	user *pmapi.User,
	addressesByBDDAddressID map[string]*pmapi.Address,
	addressIDToBeUsed string,
	password,
	mailboxPassword string,
	twoFAEnabled bool,
) *TestAccount {
	account := &TestAccount{
		user:                    user,
		addressesByBDDAddressID: addressesByBDDAddressID,
		password:                password,
		mailboxPassword:         mailboxPassword,
		twoFAEnabled:            twoFAEnabled,
	}

	if addressIDToBeUsed == "" {
		account.addressToBeUsed = account.Addresses().Main()
	} else {
		for addressID, address := range addressesByBDDAddressID {
			if addressID == addressIDToBeUsed {
				account.addressToBeUsed = address
			}
		}
	}
	if account.addressToBeUsed == nil {
		// Return nothing which will be interpreted as not implemented the same way the whole account.
		return nil
	}

	account.initKeys()
	return account
}

func (a *TestAccount) initKeys() {
	userKeys := loadPMKeys(readTestFile(testUserKey))

	addressKeys := loadPMKeys(readTestFile(testAddressKey))

	a.user.Keys = *userKeys
	for _, addressEmail := range a.Addresses().ActiveEmails() {
		a.Addresses().ByEmail(addressEmail).Keys = *addressKeys
	}
}

func readTestFile(fileName string) []byte {
	testDataFolder := os.Getenv("TEST_DATA")
	path := filepath.Join(testDataFolder, fileName)
	data, err := ioutil.ReadFile(path) //nolint:gosec
	if err != nil {
		panic(err)
	}
	return data
}

func loadPMKeys(jsonKeys []byte) (keys *pmapi.PMKeys) {
	_ = json.Unmarshal(jsonKeys, &keys)
	return
}

func (a *TestAccount) User() *pmapi.User {
	return a.user
}

func (a *TestAccount) UserID() string {
	return a.user.ID
}

func (a *TestAccount) Username() string {
	return a.user.Name
}

func (a *TestAccount) Addresses() *pmapi.AddressList {
	addressArray := []*pmapi.Address{}
	for _, address := range a.addressesByBDDAddressID {
		addressArray = append(addressArray, address)
	}
	// The order of addresses is important in PMAPI because the primary
	// address is always the first in array. We are using map to define
	// testing addresses which can cause random re-schuffle between tests
	sort.SliceStable(
		addressArray,
		func(i, j int) bool {
			return addressArray[i].Order < addressArray[j].Order
		},
	)
	addresses := pmapi.AddressList(addressArray)
	return &addresses
}

func (a *TestAccount) Address() string {
	return a.addressToBeUsed.Email
}

func (a *TestAccount) AddressID() string {
	return a.addressToBeUsed.ID
}

func (a *TestAccount) GetAddressID(addressTestID string) string {
	return a.addressesByBDDAddressID[addressTestID].ID
}

// EnsureAddressID accepts address (simply the address) or bddAddressID used
// in tests (in format [bddAddressID]) and returns always the real address ID.
// If the address is not found, the ID of main address is returned.
func (a *TestAccount) EnsureAddressID(addressOrAddressTestID string) string {
	if strings.HasPrefix(addressOrAddressTestID, "[") {
		addressTestID := addressOrAddressTestID[1 : len(addressOrAddressTestID)-1]
		address := a.addressesByBDDAddressID[addressTestID]
		return address.ID
	}
	for _, address := range a.addressesByBDDAddressID {
		if address.Email == addressOrAddressTestID {
			return address.ID
		}
	}
	return a.AddressID()
}

func (a *TestAccount) GetAddress(addressTestID string) string {
	return a.addressesByBDDAddressID[addressTestID].Email
}

// EnsureAddress accepts address (simply the address) or bddAddressID used
// in tests (in format [bddAddressID]) and returns always the address.
// If the address ID cannot be found, the original value is returned.
func (a *TestAccount) EnsureAddress(addressOrAddressTestID string) string {
	if strings.HasPrefix(addressOrAddressTestID, "[") {
		addressTestID := addressOrAddressTestID[1 : len(addressOrAddressTestID)-1]
		address := a.addressesByBDDAddressID[addressTestID]
		return address.Email
	}
	return addressOrAddressTestID
}

func (a *TestAccount) Password() []byte {
	return []byte(a.password)
}

func (a *TestAccount) MailboxPassword() []byte {
	return []byte(a.mailboxPassword)
}

func (a *TestAccount) IsTwoFAEnabled() bool {
	return a.twoFAEnabled
}

func (a *TestAccount) BridgePassword() string {
	return BridgePassword
}
