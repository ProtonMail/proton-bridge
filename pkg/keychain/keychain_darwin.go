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

package keychain

import (
	"strings"

	"github.com/docker/docker-credential-helpers/credentials"
	mackeychain "github.com/keybase/go-keychain"
)

func (s *Access) KeychainName(userID string) string {
	return s.KeychainMacURL + "/" + userID
}

func (s *Access) KeychainOldName(userID string) string {
	return s.KeychainOldMacURL + "/" + userID
}

type osxkeychain struct {
}

func newKeychain() (credentials.Helper, error) {
	log.Debug("Creating osckeychain")
	return &osxkeychain{}, nil
}

func newQuery(serviceName, username string) mackeychain.Item {
	query := mackeychain.NewItem()
	query.SetSecClass(mackeychain.SecClassGenericPassword)
	query.SetService(serviceName)
	query.SetAccount(username)
	return query
}

func parseError(original error) error {
	if original != nil && strings.Contains(original.Error(), "25293") {
		return ErrMacKeychainRebuild
	}
	return original
}

// Add appends credentials to the store (assuming old record with same ID is already deleted).
func (s *osxkeychain) Add(cred *credentials.Credentials) error {
	serviceName, userID, err := splitServiceAndID(cred.ServerURL)
	if err != nil {
		return err
	}

	query := newQuery(serviceName, userID)
	query.SetData([]byte(cred.Secret))
	err = mackeychain.AddItem(query)
	return parseError(err)
}

// Delete removes credentials from the store.
func (s *osxkeychain) Delete(serverURL string) error {
	serviceName, userID, err := splitServiceAndID(serverURL)
	if err != nil {
		return err
	}

	query := newQuery(serviceName, userID)
	err = mackeychain.DeleteItem(query)
	if err != nil && !strings.Contains(err.Error(), "25300") { // Missing item is not error.
		return err
	}
	return nil
}

// Get retrieves credentials from the store.
// It returns username and secret as strings.
func (s *osxkeychain) Get(serverURL string) (userID string, secret string, err error) {
	serviceName, userID, err := splitServiceAndID(serverURL)
	if err != nil {
		return
	}

	query := newQuery(serviceName, userID)
	query.SetMatchLimit(mackeychain.MatchLimitOne)
	query.SetReturnData(true)
	results, err := mackeychain.QueryItem(query)
	if err != nil {
		return "", "", parseError(err)
	}

	if len(results) == 1 {
		secret = string(results[0].Data)
	}

	return
}

// ListKeychain lists items in our services.
func (s *Access) ListKeychain() (userIDByURL map[string]string, err error) {
	// Pick up correct service name and trim '/'.
	serviceName, _, err := splitServiceAndID(s.KeychainOldName("not-id"))
	if err != nil {
		return
	}

	userIDByURL = make(map[string]string)

	if oldIDs, err := mackeychain.GetGenericPasswordAccounts(serviceName); err == nil {
		for _, userIDold := range oldIDs {
			userIDByURL[s.KeychainOldName(userIDold)] = userIDold
		}
	}

	serviceName, _, _ = splitServiceAndID(s.KeychainName("not-id"))
	if userIDs, err := mackeychain.GetGenericPasswordAccounts(serviceName); err == nil {
		for _, userID := range userIDs {
			userIDByURL[s.KeychainName(userID)] = userID
		}
	}

	return
}

// List returns the stored serverURLs and their associated usernames.
// NOTE: This is not valid for go-keychain. Use ListKeychain instead.
func (s *osxkeychain) List() (userIDByURL map[string]string, err error) {
	err = ErrMacKeychainList
	return
}
