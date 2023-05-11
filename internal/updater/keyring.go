// Copyright (c) 2023 Proton AG
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

package updater

import (
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/sirupsen/logrus"
)

func GetDefaultKeyring() (*crypto.KeyRing, error) {
	l := logrus.WithField("pkg", "updater")

	key, err := crypto.NewKeyFromArmored(DefaultPublicKey)
	if err != nil {
		l.WithError(err).Error("Failed to create new verification key")
		return nil, err
	}

	kr, err := crypto.NewKeyRing(key)
	if err != nil {
		l.WithError(err).Fatal("Failed to create new verification keyring")
	}

	return kr, nil
}
