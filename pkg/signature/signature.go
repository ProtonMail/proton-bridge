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

// Package signature implements functions to verify files by their detached signatures.
package signature

import (
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/pkg/errors"
)

// Verify verifies the given file by its signature using the given armored public key.
func Verify(fileBytes, sigBytes []byte, pubKey string) error {
	key, err := crypto.NewKeyFromArmored(pubKey)
	if err != nil {
		return errors.Wrap(err, "failed to load key")
	}

	kr, err := crypto.NewKeyRing(key)
	if err != nil {
		return errors.Wrap(err, "failed to create keyring")
	}

	return kr.VerifyDetached(
		crypto.NewPlainMessage(fileBytes),
		crypto.NewPGPSignature(sigBytes),
		crypto.GetUnixTime(),
	)
}
