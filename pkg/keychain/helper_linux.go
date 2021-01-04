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

	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/docker/docker-credential-helpers/pass"
	"github.com/docker/docker-credential-helpers/secretservice"
)

const (
	Pass         = "pass-app"
	GnomeKeyring = "gnome-keyring"
)

func init() { // nolint[noinit]
	Helpers = make(map[string]helperConstructor)

	if _, err := exec.LookPath("pass"); err == nil {
		Helpers[Pass] = newPassHelper
	}

	if _, err := exec.LookPath("gnome-keyring"); err == nil {
		Helpers[GnomeKeyring] = newGnomeKeyringHelper
	}
}

func newPassHelper(string) (credentials.Helper, error) {
	return &pass.Pass{}, nil
}

func newGnomeKeyringHelper(string) (credentials.Helper, error) {
	return &secretservice.Secretservice{}, nil
}
