// Copyright (c) 2022 Proton AG
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

// Package keychain implements a native secure password store for each platform.
package keychain

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ProtonMail/proton-bridge/v2/internal/config/settings"
	"github.com/docker/docker-credential-helpers/credentials"
)

// helperConstructor constructs a keychain helperConstructor.
type helperConstructor func(string) (credentials.Helper, error)

// Version is the keychain data version.
const Version = "k11"

var (
	// ErrNoKeychain indicates that no suitable keychain implementation could be loaded.
	ErrNoKeychain = errors.New("no keychain") //nolint:gochecknoglobals

	// ErrMacKeychainRebuild is returned on macOS with blocked or corrupted keychain.
	ErrMacKeychainRebuild = errors.New("keychain error -25293")

	// Helpers holds all discovered keychain helpers. It is populated in init().
	Helpers map[string]helperConstructor //nolint:gochecknoglobals

	// defaultHelper is the default helper to use if the user hasn't yet set a preference.
	defaultHelper string //nolint:gochecknoglobals
)

// NewKeychain creates a new native keychain.
func NewKeychain(s *settings.Settings, keychainName string) (*Keychain, error) {
	// There must be at least one keychain helper available.
	if len(Helpers) < 1 {
		return nil, ErrNoKeychain
	}

	// If the preferred keychain is unsupported, fallback to the default one.
	// NOTE: Maybe we want to error out here and show something in the GUI instead?
	if _, ok := Helpers[s.Get(settings.PreferredKeychainKey)]; !ok {
		s.Set(settings.PreferredKeychainKey, defaultHelper)
	}

	// Load the user's preferred keychain helper.
	helperConstructor, ok := Helpers[s.Get(settings.PreferredKeychainKey)]
	if !ok {
		return nil, ErrNoKeychain
	}

	// Construct the keychain helper.
	helper, err := helperConstructor(hostURL(keychainName))
	if err != nil {
		return nil, err
	}

	return newKeychain(helper, hostURL(keychainName)), nil
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

	var userIDs []string //nolint:prealloc

	for url, userID := range userIDsByURL {
		if url != kc.secretURL(userID) {
			continue
		}

		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

func (kc *Keychain) Delete(userID string) error {
	kc.locker.Lock()
	defer kc.locker.Unlock()

	userIDsByURL, err := kc.helper.List()
	if err != nil {
		return err
	}

	if _, ok := userIDsByURL[kc.secretURL(userID)]; !ok {
		return nil
	}

	return kc.helper.Delete(kc.secretURL(userID))
}

// Get returns the username and secret for the given userID.
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
