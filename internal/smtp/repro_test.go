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

package smtp

import (
	"testing"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/stretchr/testify/assert"
)

func TestKeyRingsAreEqualAfterFiltering(t *testing.T) {
	// Load the key.
	key, err := crypto.NewKeyFromArmored(testPublicKey)
	if err != nil {
		panic(err)
	}

	// Put it in a keyring.
	keyRing, err := crypto.NewKeyRing(key)
	if err != nil {
		panic(err)
	}

	// Filter out expired ones.
	validKeyRings, err := crypto.FilterExpiredKeys([]*crypto.KeyRing{keyRing})
	if err != nil {
		panic(err)
	}

	// Filtering shouldn't make them unequal.
	assert.True(t, isEqual(t, keyRing, validKeyRings[0]))
}

func isEqual(t *testing.T, a, b *crypto.KeyRing) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil && b != nil || a != nil && b == nil {
		return false
	}

	aKeys, bKeys := a.GetKeys(), b.GetKeys()

	if len(aKeys) != len(bKeys) {
		return false
	}

	for i := range aKeys {
		aFPs := aKeys[i].GetSHA256Fingerprints()
		bFPs := bKeys[i].GetSHA256Fingerprints()

		if !assert.Equal(t, aFPs, bFPs) {
			return false
		}
	}

	return true
}
