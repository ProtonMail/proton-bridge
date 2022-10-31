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

package keychain

import (
	"reflect"

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

func init() { //nolint:gochecknoinits
	Helpers = make(map[string]helperConstructor)

	if isUsable(newDBusHelper("")) {
		Helpers[SecretServiceDBus] = newDBusHelper
	}

	if _, err := execabs.LookPath("gnome-keyring"); err == nil && isUsable(newSecretServiceHelper("")) {
		Helpers[SecretService] = newSecretServiceHelper
	}

	if _, err := execabs.LookPath("pass"); err == nil && isUsable(newPassHelper("")) {
		Helpers[Pass] = newPassHelper
	}

	defaultHelper = SecretServiceDBus

	// If Pass is available, use it by default.
	// Otherwise, if SecretService is available, use it by default.
	if _, ok := Helpers[Pass]; ok {
		defaultHelper = Pass
	} else if _, ok := Helpers[SecretService]; ok {
		defaultHelper = SecretService
	}
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

// isUsable returns whether the credentials helper is usable.
func isUsable(helper credentials.Helper, err error) bool {
	l := logrus.WithField("helper", reflect.TypeOf(helper))

	if err != nil {
		l.WithError(err).Warn("Keychain helper couldn't be created")
		return false
	}

	creds := &credentials.Credentials{
		ServerURL: "bridge/check",
		Username:  "check",
		Secret:    "check",
	}

	if err := helper.Add(creds); err != nil {
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
