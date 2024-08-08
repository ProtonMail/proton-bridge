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
	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/docker/docker-credential-helpers/pass"
	"github.com/docker/docker-credential-helpers/secretservice"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/execabs"
)

const (
	Pass              = "pass-app"
	SecretService     = "secret-service"
	SecretServiceDBus = "secret-service-dbus"
)

func listHelpers(_ bool) (Helpers, string) {
	helpers := make(Helpers)

	if isUsable(newDBusHelper("")) {
		helpers[SecretServiceDBus] = newDBusHelper
		logrus.WithField("keychain", "SecretServiceDBus").Info("Keychain is usable.")
	} else {
		logrus.WithField("keychain", "SecretServiceDBus").Debug("Keychain is not available.")
	}

	if _, err := execabs.LookPath("gnome-keyring"); err == nil && isUsable(newSecretServiceHelper("")) {
		helpers[SecretService] = newSecretServiceHelper
		logrus.WithField("keychain", "SecretService").Info("Keychain is usable.")
	} else {
		logrus.WithField("keychain", "SecretService").Debug("Keychain is not available.")
	}

	if _, err := execabs.LookPath("pass"); err == nil && isUsable(newPassHelper("")) {
		helpers[Pass] = newPassHelper
		logrus.WithField("keychain", "Pass").Info("Keychain is usable.")
	} else {
		logrus.WithField("keychain", "Pass").Debug("Keychain is not available.")
	}

	defaultHelper := SecretServiceDBus

	// If Pass is available, use it by default.
	// Otherwise, if SecretService is available, use it by default.
	if _, ok := helpers[Pass]; ok {
		defaultHelper = Pass
	} else if _, ok := helpers[SecretService]; ok {
		defaultHelper = SecretService
	}
	return helpers, defaultHelper
}

func newDBusHelper(string) (credentials.Helper, error) {
	return &SecretServiceDBusHelper{}, nil
}

func newPassHelper(string) (credentials.Helper, error) {
	return &pass.Pass{}, nil
}

func newSecretServiceHelper(string) (credentials.Helper, error) {
	return &secretservice.Secretservice{}, nil
}
