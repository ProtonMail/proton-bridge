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

package credentials

import (
	"errors"
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
	secrets *keychain.Access
}

// NewStore creates a new encrypted credentials store.
func NewStore() (*Store, error) {
	secrets, err := keychain.NewAccess("bridge")
	return &Store{
		secrets: secrets,
	}, err
}

func (s *Store) Add(userID, userName, apiToken, mailboxPassword string, emails []string) (creds *Credentials, err error) {
	storeLocker.Lock()
	defer storeLocker.Unlock()

	log.WithFields(logrus.Fields{
		"user":     userID,
		"username": userName,
		"emails":   emails,
	}).Trace("Adding new credentials")

	if err = s.checkKeychain(); err != nil {
		return
	}

	creds = &Credentials{
		UserID:          userID,
		Name:            userName,
		APIToken:        apiToken,
		MailboxPassword: mailboxPassword,
		IsHidden:        false,
	}

	creds.SetEmailList(emails)

	var has bool
	if has, err = s.has(userID); err != nil {
		log.WithField("userID", userID).WithError(err).Error("Could not check if user credentials already exist")
		return
	}

	if has {
		log.Info("Updating credentials of existing user")
		currentCredentials, err := s.get(userID)
		if err != nil {
			return nil, err
		}
		creds.BridgePassword = currentCredentials.BridgePassword
		creds.IsCombinedAddressMode = currentCredentials.IsCombinedAddressMode
		creds.Timestamp = currentCredentials.Timestamp
	} else {
		log.Info("Generating credentials for new user")
		creds.BridgePassword = generatePassword()
		creds.IsCombinedAddressMode = true
		creds.Timestamp = time.Now().Unix()
	}

	if err = s.saveCredentials(creds); err != nil {
		return
	}

	return creds, err
}

func (s *Store) SwitchAddressMode(userID string) error {
	storeLocker.Lock()
	defer storeLocker.Unlock()

	credentials, err := s.get(userID)
	if err != nil {
		return err
	}

	credentials.IsCombinedAddressMode = !credentials.IsCombinedAddressMode
	credentials.BridgePassword = generatePassword()

	return s.saveCredentials(credentials)
}

func (s *Store) UpdateEmails(userID string, emails []string) error {
	storeLocker.Lock()
	defer storeLocker.Unlock()

	credentials, err := s.get(userID)
	if err != nil {
		return err
	}

	credentials.SetEmailList(emails)

	return s.saveCredentials(credentials)
}

func (s *Store) UpdatePassword(userID, password string) error {
	storeLocker.Lock()
	defer storeLocker.Unlock()

	credentials, err := s.get(userID)
	if err != nil {
		return err
	}

	credentials.MailboxPassword = password

	return s.saveCredentials(credentials)
}

func (s *Store) UpdateToken(userID, apiToken string) error {
	storeLocker.Lock()
	defer storeLocker.Unlock()

	credentials, err := s.get(userID)
	if err != nil {
		return err
	}

	credentials.APIToken = apiToken

	return s.saveCredentials(credentials)
}

func (s *Store) Logout(userID string) error {
	storeLocker.Lock()
	defer storeLocker.Unlock()

	credentials, err := s.get(userID)
	if err != nil {
		return err
	}

	credentials.Logout()

	return s.saveCredentials(credentials)
}

// List returns a list of usernames that have credentials stored.
func (s *Store) List() (userIDs []string, err error) {
	storeLocker.RLock()
	defer storeLocker.RUnlock()

	log.Trace("Listing credentials in credentials store")

	if err = s.checkKeychain(); err != nil {
		return
	}

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

	var has bool
	if has, err = s.has(userID); err != nil {
		log.WithError(err).Error("Could not check for credentials")
		return
	}

	if !has {
		err = errors.New("no credentials found for given userID")
		return
	}

	return s.get(userID)
}

func (s *Store) has(userID string) (has bool, err error) {
	if err = s.checkKeychain(); err != nil {
		return
	}

	var ids []string
	if ids, err = s.secrets.List(); err != nil {
		log.WithError(err).Error("Could not list credentials")
		return
	}

	for _, id := range ids {
		if id == userID {
			has = true
		}
	}

	return
}

func (s *Store) get(userID string) (creds *Credentials, err error) {
	log := log.WithField("user", userID)

	if err = s.checkKeychain(); err != nil {
		return
	}

	secret, err := s.secrets.Get(userID)
	if err != nil {
		log.WithError(err).Error("Could not get credentials from native keychain")
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
func (s *Store) saveCredentials(credentials *Credentials) (err error) {
	if err = s.checkKeychain(); err != nil {
		return
	}

	credentials.Version = keychain.KeychainVersion

	return s.secrets.Put(credentials.UserID, credentials.Marshal())
}

func (s *Store) checkKeychain() (err error) {
	if s.secrets == nil {
		err = keychain.ErrNoKeychainInstalled
		log.WithError(err).Error("Store is unusable")
	}

	return
}

// Delete removes credentials from the store.
func (s *Store) Delete(userID string) (err error) {
	storeLocker.Lock()
	defer storeLocker.Unlock()

	if err = s.checkKeychain(); err != nil {
		return
	}

	return s.secrets.Delete(userID)
}
