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

// Package keychain implements a native secure password store for each platform.
package keychain

import (
	"errors"
	"strings"
	"sync"

	"github.com/ProtonMail/proton-bridge/pkg/config"
	"github.com/docker/docker-credential-helpers/credentials"
)

const (
	KeychainVersion = "k11" //nolint[golint]
)

var (
	log = config.GetLogEntry("bridgeUtils/keychain") //nolint[gochecknoglobals]

	ErrWrongKeychainURL    = errors.New("wrong keychain base URL")
	ErrMacKeychainRebuild  = errors.New("keychain error -25293")
	ErrMacKeychainList     = errors.New("function `osxkeychain.List()` is not valid function for mac keychain. Use `Access.ListKeychain()` instead")
	ErrNoKeychainInstalled = errors.New("no keychain management installed on this system")
	accessLocker           = &sync.Mutex{} //nolint[gochecknoglobals]
)

// NewAccess creates a new native keychain.
func NewAccess(appName string) (*Access, error) {
	newHelper, err := newKeychain()
	if err != nil {
		return nil, err
	}
	return &Access{
		helper:            newHelper,
		KeychainURL:       "protonmail/" + appName + "/users",
		KeychainOldURL:    "protonmail/users",
		KeychainMacURL:    "ProtonMail" + strings.Title(appName) + "Service",
		KeychainOldMacURL: "ProtonMailService",
	}, nil
}

type Access struct {
	helper credentials.Helper
	KeychainURL,
	KeychainOldURL,
	KeychainMacURL,
	KeychainOldMacURL string
}

func (s *Access) List() (userIDs []string, err error) {
	accessLocker.Lock()
	defer accessLocker.Unlock()

	var userIDByURL map[string]string
	userIDByURL, err = s.ListKeychain()

	if err != nil {
		return
	}

	for itemURL, userID := range userIDByURL {
		if itemURL == s.KeychainName(userID) {
			userIDs = append(userIDs, userID)
		}

		// Clean up old keychain name.
		if itemURL == s.KeychainOldName(userID) {
			_ = s.helper.Delete(s.KeychainOldName(userID))
		}
	}

	return
}

func (s *Access) Delete(userID string) (err error) {
	accessLocker.Lock()
	defer accessLocker.Unlock()
	return s.helper.Delete(s.KeychainName(userID))
}

func (s *Access) Get(userID string) (secret string, err error) {
	accessLocker.Lock()
	defer accessLocker.Unlock()
	_, secret, err = s.helper.Get(s.KeychainName(userID))
	return
}

func (s *Access) Put(userID, secret string) error {
	accessLocker.Lock()
	defer accessLocker.Unlock()

	// On macOS, adding a credential that already exists does not update it and returns an error.
	// So let's remove it first.
	_ = s.helper.Delete(s.KeychainName(userID))

	cred := &credentials.Credentials{
		ServerURL: s.KeychainName(userID),
		Username:  userID,
		Secret:    secret,
	}

	return s.helper.Add(cred)
}

func splitServiceAndID(keychainName string) (serviceName string, userID string, err error) { //nolint[unused]
	splitted := strings.FieldsFunc(keychainName, func(c rune) bool { return c == '/' })
	n := len(splitted)
	if n <= 1 {
		return "", "", ErrWrongKeychainURL
	}
	userID = splitted[len(splitted)-1]
	serviceName = strings.Join(splitted[:len(splitted)-1], "/")
	return
}
