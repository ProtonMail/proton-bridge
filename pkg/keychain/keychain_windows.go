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
	"github.com/docker/docker-credential-helpers/wincred"
)

func newKeychain() (credentials.Helper, error) {
	log.Debug("creating wincred")
	return &wincred.Wincred{}, nil
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
