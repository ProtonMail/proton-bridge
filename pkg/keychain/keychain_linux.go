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

package keychain

import (
	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/docker/docker-credential-helpers/pass"
	"github.com/docker/docker-credential-helpers/secretservice"
)

func newKeychain() (credentials.Helper, error) {
	log.Debug("creating pass")
	passHelper := &pass.Pass{}
	passErr := checkPassIsUsable(passHelper)
	if passErr == nil {
		return passHelper, nil
	}

	log.Debug("creating secretservice")
	sserviceHelper := &secretservice.Secretservice{}
	_, sserviceErr := sserviceHelper.List()
	if sserviceErr == nil {
		return sserviceHelper, nil
	}

	log.Error("No keychain! Pass: ", passErr, ", secretService: ", sserviceErr)
	return nil, ErrNoKeychainInstalled
}

func checkPassIsUsable(passHelper *pass.Pass) (err error) {
	creds := &credentials.Credentials{
		ServerURL: "initCheck/pass",
		Username:  "pass",
		Secret:    "pass",
	}

	if err = passHelper.Add(creds); err != nil {
		return
	}
	// Pass is not asked about unlock until you try to decrypt.
	if _, _, err = passHelper.Get(creds.ServerURL); err != nil {
		return
	}
	_ = passHelper.Delete(creds.ServerURL) // Doesn't matter if you are able to clear.
	return
}

func (s *Access) KeychainName(userID string) string {
	return s.KeychainURL + "/" + userID
}

func (s *Access) KeychainOldName(userID string) string {
	return s.KeychainOldURL + "/" + userID
}

func (s *Access) ListKeychain() (map[string]string, error) {
	return s.helper.List()
}
