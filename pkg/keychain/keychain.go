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

// Package keychain implements a native secure password store for each platform.
package keychain

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/sirupsen/logrus"
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

	ErrKeychainNoItem = errors.New("no such keychain item")
)

func IsErrKeychainNoItem(err error) bool {
	return errors.Is(err, ErrKeychainNoItem) || credentials.IsErrCredentialsNotFound(err)
}

type Helpers map[string]helperConstructor

type List struct {
	helpers       Helpers
	defaultHelper string
	locker        sync.Locker
}

// NewList checks availability of every keychains detected on the User Operating System
// This will ask the user to unlock keychain(s) to check their usability.
// This should only be called once.
func NewList(skipKeychainTest bool) *List {
	var list = List{locker: &sync.Mutex{}}
	list.helpers, list.defaultHelper = listHelpers(skipKeychainTest)
	return &list
}

func (kcl *List) GetHelpers() Helpers {
	kcl.locker.Lock()
	defer kcl.locker.Unlock()

	return kcl.helpers
}

func (kcl *List) GetDefaultHelper() string {
	kcl.locker.Lock()
	defer kcl.locker.Unlock()

	return kcl.defaultHelper
}

// NewKeychain creates a new native keychain.
func NewKeychain(preferred, keychainName string, helpers Helpers, defaultHelper string) (*Keychain, error) {
	// There must be at least one keychain helper available.
	if len(helpers) < 1 {
		return nil, ErrNoKeychain
	}

	// If the preferred keychain is unsupported, fallback to the default one.
	if _, ok := helpers[preferred]; !ok {
		preferred = defaultHelper
	}

	// Load the user's preferred keychain helper.
	helperConstructor, ok := helpers[preferred]
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

func (kc *Keychain) Clear() error {
	entries, err := kc.List()
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if err := kc.Delete(entry); err != nil {
			return err
		}
	}

	return nil
}

// Get returns the username and secret for the given userID.
func (kc *Keychain) Get(userID string) (string, string, error) {
	kc.locker.Lock()
	defer kc.locker.Unlock()

	id, key, err := kc.helper.Get(kc.secretURL(userID))
	if err != nil {
		return id, key, err
	}

	if key == "" {
		return id, key, ErrKeychainNoItem
	}

	return id, key, err
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

// isUsable returns whether the credentials helper is usable.
func isUsable(helper credentials.Helper, err error) bool {
	l := logrus.WithField("helper", reflect.TypeOf(helper))

	if err != nil {
		l.WithError(err).Warn("Keychain helper couldn't be created")
		return false
	}

	creds := getTestCredentials()

	if err := retry(func() error {
		return helper.Add(creds)
	}); err != nil {
		l.WithError(err).Warn("Failed to add test credentials to keychain")
		return false
	}

	if _, _, err := helper.Get(creds.ServerURL); err != nil {
		l.WithError(err).Warn("Failed to get test credentials from keychain")
		return false
	}

	if err := helper.Delete(creds.ServerURL); err != nil {
		l.WithError(err).Warn("Failed to delete test credentials from keychain")
		return false
	}

	return true
}

func getTestCredentials() *credentials.Credentials {
	// On macOS, a handful of users experience failures of the test credentials.
	if runtime.GOOS == "darwin" {
		return &credentials.Credentials{
			ServerURL: hostURL(constants.KeyChainName) + fmt.Sprintf("/check_%v", time.Now().UTC().UnixMicro()),
			Username:  "", // username is ignored on macOS, it's extracted from splitting the server URL
			Secret:    "check",
		}
	}

	return &credentials.Credentials{
		ServerURL: "bridge/check",
		Username:  "check",
		Secret:    "check",
	}
}

func retry(condition func() error) error {
	var maxRetry = 5
	for r := 0; ; r++ {
		err := condition()
		if err == nil || r >= maxRetry {
			return err
		}
		time.Sleep(200 * time.Millisecond)
	}
}
