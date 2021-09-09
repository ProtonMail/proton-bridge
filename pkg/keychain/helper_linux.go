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
	"os/exec"
	"reflect"

	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/docker/docker-credential-helpers/pass"
	"github.com/docker/docker-credential-helpers/secretservice"
	"github.com/sirupsen/logrus"
)

const (
	Pass         = "pass-app"
	SecretService = "secret-service"
)

func init() { // nolint[noinit]
	Helpers = make(map[string]helperConstructor)

	if _, err := exec.LookPath("pass"); err == nil {
		Helpers[Pass] = newPassHelper
	}

	Helpers[SecretService] = newSecretServiceHelper

	// If Pass is available, use it by default.
	// Otherwise, if SecretService is available, use it by default.
	if _, ok := Helpers[Pass]; ok && isUsable(newPassHelper("")) {
		defaultHelper = Pass
	} else if _, ok := Helpers[SecretService]; ok && isUsable(newSecretServiceHelper("")) {
		defaultHelper = SecretService
	}
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
