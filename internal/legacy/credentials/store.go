// Copyright (c) 2023 Proton AG
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

package credentials

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

var storeLocker = sync.RWMutex{} //nolint:gochecknoglobals

// Store is an encrypted credentials store.
type Store struct {
	secrets Keychain
}

type Keychain interface {
	List() ([]string, error)
	Get(string) (string, string, error)
	Put(string, string) error
	Delete(string) error
}

// NewStore creates a new encrypted credentials store.
func NewStore(keychain Keychain) *Store {
	return &Store{secrets: keychain}
}

// List returns a list of usernames that have credentials stored.
func (s *Store) List() (userIDs []string, err error) {
	storeLocker.RLock()
	defer storeLocker.RUnlock()

	log.Trace("Listing credentials in credentials store")

	var allUserIDs []string
	if allUserIDs, err = s.secrets.List(); err != nil {
		log.WithError(err).Error("Could not list credentials")
		return nil, err
	}

	credentialList := []*Credentials{}
	for _, userID := range allUserIDs {
		creds, getErr := s.get(userID)
		if getErr != nil {
			log.WithField("userID", userID).WithError(getErr).Warn("Failed to get credentials")
			continue
		}

		// Disabled credentials
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

func (s *Store) Get(userID string) (creds *Credentials, err error) {
	storeLocker.RLock()
	defer storeLocker.RUnlock()

	return s.get(userID)
}

func (s *Store) get(userID string) (*Credentials, error) {
	log := log.WithField("user", userID)

	_, secret, err := s.secrets.Get(userID)
	if err != nil {
		return nil, err
	}

	if secret == "" {
		return nil, errors.New("secret is empty")
	}

	credentials := &Credentials{UserID: userID}

	if err := credentials.Unmarshal(secret); err != nil {
		log.WithError(fmt.Errorf("malformed secret: %w", err)).Error("Could not unmarshal secret")

		if err := s.secrets.Delete(userID); err != nil {
			log.WithError(err).Error("Failed to remove malformed secret")
		}

		return nil, err
	}

	return credentials, nil
}
