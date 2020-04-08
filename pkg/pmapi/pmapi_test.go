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

package pmapi

import (
	"io/ioutil"
	"strings"

	pmcrypto "github.com/ProtonMail/gopenpgp/crypto"
)

const testMailboxPassword = "apple"
const testMailboxPasswordLegacy = "123"

var (
	testPrivateKeyRing *pmcrypto.KeyRing
	testPublicKeyRing  *pmcrypto.KeyRing
)

func init() {
	testPrivateKey := readTestFile("testPrivateKey", false)
	testPublicKey := readTestFile("testPublicKey", false)

	var err error
	if testPrivateKeyRing, err = pmcrypto.ReadArmoredKeyRing(strings.NewReader(testPrivateKey)); err != nil {
		panic(err)
	}

	if testPublicKeyRing, err = pmcrypto.ReadArmoredKeyRing(strings.NewReader(testPublicKey)); err != nil {
		panic(err)
	}

	if err := testPrivateKeyRing.Unlock([]byte(testMailboxPassword)); err != nil {
		panic(err)
	}
}

func readTestFile(name string, trimNewlines bool) string { // nolint[unparam]
	data, err := ioutil.ReadFile("testdata/" + name)
	if err != nil {
		panic(err)
	}
	if trimNewlines {
		return strings.TrimRight(string(data), "\n")
	}
	return string(data)
}
