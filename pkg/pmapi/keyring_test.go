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
	"encoding/json"
	"strings"
	"sync"
	"testing"

	pmcrypto "github.com/ProtonMail/gopenpgp/crypto"
	"github.com/stretchr/testify/assert"
)

func loadPMKeys(jsonKeys string) (keys *PMKeys) {
	_ = json.Unmarshal([]byte(jsonKeys), &keys)
	return
}

func TestPMKeys_GetKeyRingAndUnlock(t *testing.T) {
	addrKeysWithTokens := loadPMKeys(readTestFile("keyring_addressKeysWithTokens_JSON", false))
	addrKeysWithoutTokens := loadPMKeys(readTestFile("keyring_addressKeysWithoutTokens_JSON", false))
	addrKeysPrimaryHasToken := loadPMKeys(readTestFile("keyring_addressKeysPrimaryHasToken_JSON", false))
	addrKeysSecondaryHasToken := loadPMKeys(readTestFile("keyring_addressKeysSecondaryHasToken_JSON", false))

	userKey, err := pmcrypto.ReadArmoredKeyRing(strings.NewReader(readTestFile("keyring_userKey", false)))
	assert.NoError(t, err, "Expected not to receive an error unlocking user key")

	type args struct {
		userKeyring *pmcrypto.KeyRing
		passphrase  []byte
	}
	tests := []struct {
		name string
		keys *PMKeys
		args args
	}{
		{
			name: "AddressKeys locked with tokens",
			keys: addrKeysWithTokens,
			args: args{userKeyring: userKey, passphrase: []byte("testpassphrase")},
		},
		{
			name: "AddressKeys locked with passphrase, not tokens",
			keys: addrKeysWithoutTokens,
			args: args{userKeyring: userKey, passphrase: []byte("testpassphrase")},
		},
		{
			name: "AddressKeys, primary locked with token, secondary with passphrase",
			keys: addrKeysPrimaryHasToken,
			args: args{userKeyring: userKey, passphrase: []byte("testpassphrase")},
		},
		{
			name: "AddressKeys, primary locked with passphrase, secondary with token",
			keys: addrKeysSecondaryHasToken,
			args: args{userKeyring: userKey, passphrase: []byte("testpassphrase")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempLocker := &sync.Mutex{}

			err := tt.keys.unlockKeyRing(tt.args.userKeyring, tt.args.passphrase, tempLocker) // nolint[scopelint]
			if !assert.NoError(t, err) {
				return
			}

			// assert at least one key has been decrypted
			atLeastOneDecrypted := false
			for _, e := range tt.keys.KeyRing.GetEntities() { // nolint[scopelint]
				if !e.PrivateKey.Encrypted {
					atLeastOneDecrypted = true
					break
				}
			}
			assert.True(t, atLeastOneDecrypted)
		})
	}
}
