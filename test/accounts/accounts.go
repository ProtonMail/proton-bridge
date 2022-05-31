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

	"github.com/ProtonMail/proton-bridge/v2/pkg/pmapi"
	"github.com/pkg/errors"
)

// BridgePassword is password to be used for IMAP or SMTP under tests.
const BridgePassword = "bridgepassword"

type TestAccounts struct {
	Users            map[string]*pmapi.User               // Key is user ID used in BDD.
	Addresses        map[string]map[string]*pmapi.Address // Key is real user ID, second key is address ID used in BDD.
	Passwords        map[string]string                    // Key is real user ID.
	MailboxPasswords map[string]string                    // Key is real user ID.
	TwoFAs           map[string]bool                      // Key is real user ID.
}

func Load(path string) (*TestAccounts, error) {
	data, err := ioutil.ReadFile(path) //nolint:gosec
	if err != nil {
		return nil, errors.Wrap(err, "failed to load JSON")
	}

	var testAccounts TestAccounts
	err = json.Unmarshal(data, &testAccounts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal JSON")
	}

	return &testAccounts, nil
}

func (a *TestAccounts) GetTestAccount(username string) *TestAccount {
	return a.GetTestAccountWithAddress(username, "")
}

// GetTestAccountWithAddress returns the test account with the given username configured to use the given bddAddressID.
func (a *TestAccounts) GetTestAccountWithAddress(username, bddAddressID string) *TestAccount {
	// Do lookup by full address and convert to name in tests.
	// Used by getting real data to ensure correct address or address ID.
	for key, user := range a.Users {
		if user.Name == username {
			username = key
			break
		}
	}
	user, ok := a.Users[username]
	if !ok {
		return nil
	}
	return newTestAccount(
		user,
		a.Addresses[user.Name],
		bddAddressID,
		a.Passwords[user.Name],
		a.MailboxPasswords[user.Name],
		a.TwoFAs[user.Name],
	)
}
