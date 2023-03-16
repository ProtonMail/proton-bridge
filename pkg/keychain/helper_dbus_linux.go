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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package keychain

import (
	"strings"

	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/bradenaw/juniper/xslices"
	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/godbus/dbus"
	"github.com/keybase/go-keychain/secretservice"
)

const (
	serverAtt   = "server"
	labelAtt    = "label"
	usernameAtt = "username"

	defaultLabel = "Proton Mail Bridge Credentials"
)

func getDomain() string {
	return hostURL(constants.KeyChainName)
}

func getSession() (*secretservice.SecretService, *secretservice.Session, error) {
	service, err := secretservice.NewService()
	if err != nil {
		return nil, nil, err
	}

	session, err := service.OpenSession(secretservice.AuthenticationDHAES)
	if err != nil {
		return nil, nil, err
	}

	return service, session, nil
}

func handleTimeout(f func() error) error {
	err := f()
	if err == secretservice.ErrPromptTimedOut {
		return f()
	}
	return err
}

func getItems(service *secretservice.SecretService, attributes map[string]string) ([]dbus.ObjectPath, error) {
	if err := unlock(service); err != nil {
		return nil, err
	}

	var items []dbus.ObjectPath
	err := handleTimeout(func() error {
		var err error
		items, err = service.SearchCollection(
			secretservice.DefaultCollection,
			attributes,
		)
		return err
	})
	if err != nil {
		return nil, err
	}
	return xslices.Filter(items, func(t dbus.ObjectPath) bool {
		return strings.HasPrefix(string(t), "/org/freedesktop/secrets")
	}), err
}

func unlock(service *secretservice.SecretService) error {
	return handleTimeout(func() error {
		return service.Unlock([]dbus.ObjectPath{secretservice.DefaultCollection})
	})
}

// SecretServiceDBusHelper is wrapper around keybase/go-keychain/secretservice
// library.
type SecretServiceDBusHelper struct{}

// Add appends credentials to the store.
func (s *SecretServiceDBusHelper) Add(creds *credentials.Credentials) error {
	service, session, err := getSession()
	if err != nil {
		return err
	}
	defer service.CloseSession(session)

	if err := unlock(service); err != nil {
		return err
	}

	secret, err := session.NewSecret([]byte(creds.Secret))
	if err != nil {
		return err
	}

	attributes := map[string]string{
		usernameAtt: creds.Username,
		serverAtt:   creds.ServerURL,
		labelAtt:    defaultLabel,
	}

	return handleTimeout(func() error {
		_, err = service.CreateItem(
			secretservice.DefaultCollection,
			secretservice.NewSecretProperties(creds.ServerURL, attributes),
			secret,
			secretservice.ReplaceBehaviorReplace,
		)

		return err
	})
}

// Delete removes credentials from the store.
func (s *SecretServiceDBusHelper) Delete(serverURL string) error {
	service, session, err := getSession()
	if err != nil {
		return err
	}
	defer service.CloseSession(session)

	items, err := getItems(service, map[string]string{
		labelAtt:  defaultLabel,
		serverAtt: serverURL,
	})

	if len(items) == 0 || err != nil {
		return err
	}

	return handleTimeout(func() error {
		return service.DeleteItem(items[0])
	})
}

// Get retrieves credentials from the store.
// It returns username and secret as strings.
func (s *SecretServiceDBusHelper) Get(serverURL string) (string, string, error) {
	service, session, err := getSession()
	if err != nil {
		return "", "", err
	}
	defer service.CloseSession(session)

	if err := unlock(service); err != nil {
		return "", "", err
	}

	items, err := getItems(service, map[string]string{
		labelAtt:  defaultLabel,
		serverAtt: serverURL,
	})

	if len(items) == 0 || err != nil {
		return "", "", err
	}

	item := items[0]

	attributes, err := service.GetAttributes(item)
	if err != nil {
		return "", "", err
	}

	var secretPlaintext []byte
	err = handleTimeout(func() error {
		var err error
		secretPlaintext, err = service.GetSecret(item, *session)
		return err
	})
	if err != nil {
		return "", "", err
	}

	return attributes[usernameAtt], string(secretPlaintext), nil
}

// List returns the stored serverURLs and their associated usernames.
func (s *SecretServiceDBusHelper) List() (map[string]string, error) {
	userIDByURL := make(map[string]string)

	service, session, err := getSession()
	if err != nil {
		return nil, err
	}
	defer service.CloseSession(session)

	items, err := getItems(service, map[string]string{labelAtt: defaultLabel})
	if err != nil {
		return nil, err
	}

	defaultDomain := getDomain()

	for _, it := range items {
		attributes, err := service.GetAttributes(it)
		if err != nil {
			return nil, err
		}

		if !strings.HasPrefix(attributes[serverAtt], defaultDomain) {
			continue
		}

		userIDByURL[attributes[serverAtt]] = attributes[usernameAtt]
	}

	return userIDByURL, nil
}
