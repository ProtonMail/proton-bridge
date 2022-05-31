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

package pmapi

import (
	"io/ioutil"
	"strings"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

const (
	testMailboxPassword       = "apple"
	testMailboxPasswordLegacy = "123"
)

var (
	testPrivateKeyRing *crypto.KeyRing
	testPublicKeyRing  *crypto.KeyRing
)

func init() {
	testPrivateKey := readTestFile("testPrivateKey", false)
	testPublicKey := readTestFile("testPublicKey", false)

	var err error

	privKey, err := crypto.NewKeyFromArmored(testPrivateKey)
	if err != nil {
		panic(err)
	}

	privKeyUnlocked, err := privKey.Unlock([]byte(testMailboxPassword))
	if err != nil {
		panic(err)
	}

	pubKey, err := crypto.NewKeyFromArmored(testPublicKey)
	if err != nil {
		panic(err)
	}

	if testPrivateKeyRing, err = crypto.NewKeyRing(privKeyUnlocked); err != nil {
		panic(err)
	}

	if testPublicKeyRing, err = crypto.NewKeyRing(pubKey); err != nil {
		panic(err)
	}
}

func readTestFile(name string, trimNewlines bool) string { //nolint:unparam
	data, err := ioutil.ReadFile("testdata/" + name)
	if err != nil {
		panic(err)
	}
	if trimNewlines {
		return strings.TrimRight(string(data), "\n")
	}
	return string(data)
}
