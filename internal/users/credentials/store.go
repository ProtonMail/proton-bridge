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

package credentials

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/ProtonMail/proton-bridge/pkg/keychain"
	"github.com/sirupsen/logrus"
)

var storeLocker = sync.RWMutex{} //nolint[gochecknoglobals]

// Store is an encrypted credentials store.
type Store struct {
	secrets *keychain.Keychain
}

// NewStore creates a new encrypted credentials store.
func NewStore(keychain *keychain.Keychain) *Store {
	return &Store{secrets: keychain}
}

func (s *Store) Add(userID, userName, uid, ref, mailboxPassword string, emails []string) (*Credentials, error) {
	storeLocker.Lock()
	defer storeLocker.Unlock()

	log.WithFields(logrus.Fields{
		"user":     userID,
		"username": userName,
		"emails":   emails,
	}).Trace("Adding new credentials")

	creds := &Credentials{
		UserID:          userID,
		Name:            userName,
		APIToken:        uid + ":" + ref,
		MailboxPassword: mailboxPassword,
		IsHidden:        false,
	}

	creds.SetEmailList(emails)

	currentCredentials, err := s.get(userID)
	if err == nil {
		log.Info("Updating credentials of existing user")
		creds.BridgePassword = currentCredentials.BridgePassword
		creds.IsCombinedAddressMode = currentCredentials.IsCombinedAddressMode
		creds.Timestamp = currentCredentials.Timestamp
	} else {
		log.Info("Generating credentials for new user")
		creds.BridgePassword = generatePassword()
		creds.IsCombinedAddressMode = true
		creds.Timestamp = time.Now().Unix()
	}

	if err := s.saveCredentials(creds); err != nil {
		return nil, err
	}

	return creds, nil
}

func (s *Store) SwitchAddressMode(userID string) (*Credentials, error) {
	storeLocker.Lock()
	defer storeLocker.Unlock()

	credentials, err := s.get(userID)
	if err != nil {
		return nil, err
	}

	credentials.IsCombinedAddressMode = !credentials.IsCombinedAddressMode
	credentials.BridgePassword = generatePassword()

	return credentials, s.saveCredentials(credentials)
}

func (s *Store) UpdateEmails(userID string, emails []string) (*Credentials, error) {
	storeLocker.Lock()
	defer storeLocker.Unlock()

	credentials, err := s.get(userID)
	if err != nil {
		return nil, err
	}

	credentials.SetEmailList(emails)

	return credentials, s.saveCredentials(credentials)
}

func (s *Store) UpdatePassword(userID, password string) (*Credentials, error) {
	storeLocker.Lock()
	defer storeLocker.Unlock()

	credentials, err := s.get(userID)
	if err != nil {
		return nil, err
	}

	credentials.MailboxPassword = password

	return credentials, s.saveCredentials(credentials)
}

func (s *Store) UpdateToken(userID, uid, ref string) (*Credentials, error) {
	storeLocker.Lock()
	defer storeLocker.Unlock()

	credentials, err := s.get(userID)
	if err != nil {
		return nil, err
	}

	credentials.APIToken = uid + ":" + ref

	return credentials, s.saveCredentials(credentials)
}

func (s *Store) Logout(userID string) (*Credentials, error) {
	storeLocker.Lock()
	defer storeLocker.Unlock()

	credentials, err := s.get(userID)
	if err != nil {
		return nil, err
	}

	credentials.Logout()

	return credentials, s.saveCredentials(credentials)
}

// List returns a list of usernames that have credentials stored.
func (s *Store) List() (userIDs []string, err error) {
	storeLocker.RLock()
	defer storeLocker.RUnlock()

	log.Trace("Listing credentials in credentials store")

	var allUserIDs []string
	if allUserIDs, err = s.secrets.List(); err != nil {
		log.WithError(err).Error("Could not list credentials")
		return
	}

	credentialList := []*Credentials{}
	for _, userID := range allUserIDs {
		creds, getErr := s.get(userID)
		if getErr != nil {
			log.WithField("userID", userID).WithError(getErr).Warn("Failed to get credentials")
			continue
		}

		if creds.Timestamp == 0 {
			continue
		}

		// Old credentials using username as a key does not work with new code.
		// We need to ask user to login again to get ID from API and migrate creds.
		if creds.UserID == creds.Name && creds.APIToken != "" {
			creds.Logout()
			_ = s.saveCredentials(creds)
		}

		credentialList = append(credentialList, creds)
	}

	sort.Slice(credentialList, func(i, j int) bool {
		return credentialList[i].Timestamp < credentialList[j].Timestamp
	})

	for _, credentials := range credentialList {
		userIDs = append(userIDs, credentials.UserID)
	}

	return userIDs, err
}

func (s *Store) GetAndCheckPassword(userID, password string) (creds *Credentials, err error) {
	storeLocker.RLock()
	defer storeLocker.RUnlock()

	log.WithFields(logrus.Fields{
		"userID": userID,
	}).Debug("Checking bridge password")

	credentials, err := s.Get(userID)
	if err != nil {
		return nil, err
	}

	if err := credentials.CheckPassword(password); err != nil {
		log.WithFields(logrus.Fields{
			"userID": userID,
			"err":    err,
		}).Debug("Incorrect bridge password")

		return nil, err
	}

	return credentials, nil
}

func (s *Store) Get(userID string) (creds *Credentials, err error) {
	storeLocker.RLock()
	defer storeLocker.RUnlock()

	return s.get(userID)
}

func (s *Store) get(userID string) (creds *Credentials, err error) {
	log := log.WithField("user", userID)

	_, secret, err := s.secrets.Get(userID)
	if err != nil {
		log.WithError(err).Warn("Could not get credentials from native keychain")
		return
	}

	credentials := &Credentials{UserID: userID}
	if err = credentials.Unmarshal(secret); err != nil {
		err = fmt.Errorf("backend/credentials: malformed secret: %v", err)
		_ = s.secrets.Delete(userID)
		log.WithError(err).Error("Could not unmarshal secret")
		return
	}

	return credentials, nil
}

// saveCredentials encrypts and saves password to the keychain store.
func (s *Store) saveCredentials(credentials *Credentials) error {
	credentials.Version = keychain.Version

	return s.secrets.Put(credentials.UserID, credentials.Marshal())
}

// Delete removes credentials from the store.
func (s *Store) Delete(userID string) (err error) {
	storeLocker.Lock()
	defer storeLocker.Unlock()

	return s.secrets.Delete(userID)
}
