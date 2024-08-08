// Copyright (c) 2024 Proton AG
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

package keychain

import (
	"errors"
	"fmt"
	"strings"

	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/keybase/go-keychain"
	"github.com/sirupsen/logrus"
)

const (
	MacOSKeychain = "macos-keychain"
)

func listHelpers(skipKeychainTest bool) (Helpers, string) {
	helpers := make(Helpers)

	// MacOS always provides a keychain.
	if skipKeychainTest {
		logrus.WithField("pkg", "keychain").Info("Skipping macOS keychain test")
		helpers[MacOSKeychain] = newMacOSHelper
	} else {
		if isUsable(newMacOSHelper("")) {
			helpers[MacOSKeychain] = newMacOSHelper
			logrus.WithField("keychain", "MacOSKeychain").Info("Keychain is usable.")
		} else {
			logrus.WithField("keychain", "MacOSKeychain").Debug("Keychain is not available.")
		}
	}

	// Use MacOSKeychain by default.
	return helpers, MacOSKeychain
}

func parseError(original error) error {
	if original == nil {
		return nil
	}
	if strings.Contains(original.Error(), "25293") {
		return ErrMacKeychainRebuild
	}
	return original
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
	return parseError(keychain.AddItem(query))
}

func (h *macOSHelper) Delete(secretURL string) error {
	hostURL, userID, err := splitSecretURL(secretURL)
	if err != nil {
		return err
	}

	query := newQuery(hostURL, userID)

	return parseError(keychain.DeleteItem(query))
}

func (h *macOSHelper) Get(secretURL string) (string, string, error) {
	hostURL, userID, err := splitSecretURL(secretURL)
	if err != nil {
		return "", "", err
	}

	l := logrus.WithField("pkg", "keychain/darwin").WithField("h.url", h.url).WithField("userID", userID)

	query := newQuery(hostURL, userID)
	query.SetMatchLimit(keychain.MatchLimitOne)
	query.SetReturnData(true)

	results, err := keychain.QueryItem(query)
	if err != nil {
		l.WithError(err).Error("Query item failed")
		return "", "", parseError(err)
	}

	if len(results) == 0 {
		return "", "", ErrKeychainNoItem
	}

	if len(results) != 1 {
		return "", "", errors.New("ambiguous results")
	}

	return userID, string(results[0].Data), nil
}

func (h *macOSHelper) List() (map[string]string, error) {
	userIDByURL := make(map[string]string)

	l := logrus.WithField("pkg", "keychain/darwin").WithField("h.url", h.url)

	userIDs, err := keychain.GetGenericPasswordAccounts(h.url)
	if err != nil {
		l.WithError(err).Warn("Get generic password accounts failed")
		return nil, parseError(err)
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
