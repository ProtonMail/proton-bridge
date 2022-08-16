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

package context

import (
	"strings"

	"github.com/ProtonMail/proton-bridge/v2/internal/users/credentials"
	"github.com/pkg/errors"
)

// bridgePassword is password to be used for IMAP or SMTP under tests.
const bridgePassword = "bridgepassword"

type fakeCredStore struct {
	credentials map[string]*credentials.Credentials

	failOnWrite bool
}

// newFakeCredStore returns a fake credentials store (optionally configured with the given credentials).
func newFakeCredStore(initCreds ...*credentials.Credentials) (c *fakeCredStore) {
	c = &fakeCredStore{credentials: map[string]*credentials.Credentials{}, failOnWrite: false}
	for _, creds := range initCreds {
		if creds == nil {
			continue
		}
		c.credentials[creds.UserID] = &credentials.Credentials{}
		*c.credentials[creds.UserID] = *creds
	}
	return
}

func (c *fakeCredStore) List() (userIDs []string, err error) {
	keys := []string{}
	for key := range c.credentials {
		keys = append(keys, key)
	}
	return keys, nil
}

func (c *fakeCredStore) Add(userID, userName, uid, ref string, mailboxPassword []byte, emails []string) (*credentials.Credentials, error) {
	if c.failOnWrite {
		return nil, errors.New("An invalid attempt to change the owner of an item. (-25244)")
	}
	bridgePassword := bridgePassword
	if c, ok := c.credentials[userID]; ok {
		bridgePassword = c.BridgePassword
	}
	c.credentials[userID] = &credentials.Credentials{
		UserID:                userID,
		Name:                  userName,
		Emails:                strings.Join(emails, ";"),
		APIToken:              uid + ":" + ref,
		MailboxPassword:       mailboxPassword,
		BridgePassword:        bridgePassword,
		IsCombinedAddressMode: true, // otherwise by default starts in split mode
	}
	return c.Get(userID)
}

func (c *fakeCredStore) Get(userID string) (*credentials.Credentials, error) {
	return c.credentials[userID], nil
}

func (c *fakeCredStore) SwitchAddressMode(userID string) (*credentials.Credentials, error) {
	if c.failOnWrite {
		return nil, errors.New("An invalid attempt to change the owner of an item. (-25244)")
	}
	return c.credentials[userID], nil
}

func (c *fakeCredStore) UpdateEmails(userID string, emails []string) (*credentials.Credentials, error) {
	if c.failOnWrite {
		return nil, errors.New("An invalid attempt to change the owner of an item. (-25244)")
	}
	return c.credentials[userID], nil
}

func (c *fakeCredStore) UpdatePassword(userID string, password []byte) (*credentials.Credentials, error) {
	if c.failOnWrite {
		return nil, errors.New("An invalid attempt to change the owner of an item. (-25244)")
	}
	creds, err := c.Get(userID)
	if err != nil {
		return nil, err
	}
	creds.MailboxPassword = password
	return creds, nil
}

func (c *fakeCredStore) UpdateToken(userID, uid, ref string) (*credentials.Credentials, error) {
	if c.failOnWrite {
		return nil, errors.New("An invalid attempt to change the owner of an item. (-25244)")
	}
	creds, err := c.Get(userID)
	if err != nil {
		return nil, err
	}
	creds.APIToken = uid + ":" + ref
	return creds, nil
}

func (c *fakeCredStore) Logout(userID string) (*credentials.Credentials, error) {
	if c.failOnWrite {
		return nil, errors.New("An invalid attempt to change the owner of an item. (-25244)")
	}
	c.credentials[userID].APIToken = ""
	c.credentials[userID].MailboxPassword = []byte{}
	return c.credentials[userID], nil
}

func (c *fakeCredStore) Delete(userID string) error {
	if c.failOnWrite {
		return errors.New("An invalid attempt to change the owner of an item. (-25244)")
	}
	delete(c.credentials, userID)
	return nil
}
