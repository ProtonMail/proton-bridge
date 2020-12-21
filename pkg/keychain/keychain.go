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

// Package keychain implements a native secure password store for each platform.
package keychain

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/ProtonMail/proton-bridge/internal/config/settings"
	"github.com/docker/docker-credential-helpers/credentials"
)

// helper constructs a keychain helper.
type helper func(string) (credentials.Helper, error)

// Version is the keychain data version.
const Version = "k11"

var (
	// ErrNoKeychain indicates that no suitable keychain implementation could be loaded.
	ErrNoKeychain = errors.New("no keychain") // nolint[noglobals]

	// Helpers holds all discovered keychain helpers. It is populated in init().
	Helpers map[string]helper // nolint[noglobals]
)

// NewKeychain creates a new native keychain.
func NewKeychain(s *settings.Settings, keychainName string) (*Keychain, error) {
	// There must be at least one keychain helper available.
	if len(Helpers) < 1 {
		return nil, ErrNoKeychain
	}

	// hostURL uniquely identifies the app's keychain items within the system keychain.
	hostURL := fmt.Sprintf("protonmail/%v/users", keychainName)

	// If the preferred keychain is unsupported, set a default one.
	if _, ok := Helpers[s.Get(settings.PreferredKeychainKey)]; !ok {
		s.Set(settings.PreferredKeychainKey, reflect.ValueOf(Helpers).MapKeys()[0].Interface().(string))
	}

	// Load the user's preferred keychain helper.
	helper, err := Helpers[s.Get(settings.PreferredKeychainKey)](hostURL)
	if err != nil {
		return nil, err
	}

	return newKeychain(helper, hostURL), nil
}

func newKeychain(helper credentials.Helper, url string) *Keychain {
	return &Keychain{
		helper: helper,
		url:    url,
		locker: &sync.Mutex{},
	}
}

type Keychain struct {
	helper credentials.Helper
	url    string
	locker sync.Locker
}

func (kc *Keychain) List() ([]string, error) {
	kc.locker.Lock()
	defer kc.locker.Unlock()

	userIDsByURL, err := kc.helper.List()
	if err != nil {
		return nil, err
	}

	var userIDs []string // nolint[prealloc]

	for url, userID := range userIDsByURL {
		if url != kc.secretURL(userID) {
			continue
		}

		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

func (kc *Keychain) Delete(userID string) (err error) {
	kc.locker.Lock()
	defer kc.locker.Unlock()

	return kc.helper.Delete(kc.secretURL(userID))
}

func (kc *Keychain) Get(userID string) (string, string, error) {
	kc.locker.Lock()
	defer kc.locker.Unlock()

	return kc.helper.Get(kc.secretURL(userID))
}

func (kc *Keychain) Put(userID, secret string) error {
	kc.locker.Lock()
	defer kc.locker.Unlock()

	return kc.helper.Add(&credentials.Credentials{
		ServerURL: kc.secretURL(userID),
		Username:  userID,
		Secret:    secret,
	})
}

// secretURL returns the URL referring to a userID's secrets.
func (kc *Keychain) secretURL(userID string) string {
	return fmt.Sprintf("%v/%v", kc.url, userID)
}
