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

package keychain

import (
	"errors"
	"fmt"
	"strings"

	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/keybase/go-keychain"
)

const (
	MacOSKeychain = "macos-keychain"
)

func init() { // nolint[noinit]
	Helpers = make(map[string]helperConstructor)

	// MacOS always provides a keychain.
	Helpers[MacOSKeychain] = newMacOSHelper

	// Use MacOSKeychain by default.
	defaultHelper = MacOSKeychain
}

func newMacOSHelper(url string) (credentials.Helper, error) {
	return &macOSHelper{url: url}, nil
}

type macOSHelper struct {
	url string
}

func newQuery(service, account string) keychain.Item {
	query := keychain.NewItem()
	query.SetSecClass(keychain.SecClassGenericPassword)
	query.SetService(service)
	query.SetAccount(account)
	return query
}

func (h *macOSHelper) Add(creds *credentials.Credentials) error {
	secrets, err := h.List()
	if err != nil {
		return err
	}

	// Adding a secret that already exists results in an error so we first delete the secret.
	if _, ok := secrets[creds.ServerURL]; ok {
		if err := h.Delete(creds.ServerURL); err != nil {
			return err
		}
	}

	hostURL, userID, err := splitSecretURL(creds.ServerURL)
	if err != nil {
		return err
	}

	query := newQuery(hostURL, userID)
	query.SetData([]byte(creds.Secret))
	return keychain.AddItem(query)
}

func (h *macOSHelper) Delete(secretURL string) error {
	hostURL, userID, err := splitSecretURL(secretURL)
	if err != nil {
		return err
	}

	query := newQuery(hostURL, userID)
	if err := keychain.DeleteItem(query); err != nil {
		return err
	}

	return nil
}

func (h *macOSHelper) Get(secretURL string) (string, string, error) {
	hostURL, userID, err := splitSecretURL(secretURL)
	if err != nil {
		return "", "", err
	}

	query := newQuery(hostURL, userID)
	query.SetMatchLimit(keychain.MatchLimitOne)
	query.SetReturnData(true)

	results, err := keychain.QueryItem(query)
	if err != nil {
		return "", "", err
	}

	if len(results) != 1 {
		return "", "", errors.New("ambiguous results")
	}

	return userID, string(results[0].Data), nil
}

func (h *macOSHelper) List() (map[string]string, error) {
	userIDByURL := make(map[string]string)

	userIDs, err := keychain.GetGenericPasswordAccounts(h.url)
	if err != nil {
		return nil, err
	}

	for _, userID := range userIDs {
		userIDByURL[h.secretURL(userID)] = userID
	}

	return userIDByURL, nil
}

// secretURL returns the URL referring to a userID's secrets.
// NOTE: This is the same as the implementation in type `Keychain`.
// I didn't want to make this type depend on `Keychain` but I also don't like duplicate methods...
func (h *macOSHelper) secretURL(userID string) string {
	return fmt.Sprintf("%v/%v", h.url, userID)
}

func splitSecretURL(secretURL string) (string, string, error) {
	split := strings.Split(secretURL, "/")

	if len(split) < 2 {
		return "", "", errors.New("malformed secret")
	}

	return strings.Join(split[:len(split)-1], "/"), split[len(split)-1], nil
}
