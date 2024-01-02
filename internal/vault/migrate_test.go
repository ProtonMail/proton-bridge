// Copyright (c) 2024 Proton AG
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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v5"
)

func TestMigrate(t *testing.T) {
	dir := t.TempDir()

	// Create a v2.3.x vault.
	b := newLegacyVault(t, []byte("my secret key"), v2_3_x, Data_2_3_x{
		Settings: Settings_2_3_x{
			GluonDir: "v2.3.x-gluon-dir",
			IMAPPort: "1234", // string in v2.3.x, current version uses int
			SMTPPort: "5678", // string in v2.3.x, current version uses int
		},
		Users: []UserData_2_3_x{{
			ID:        "user-id",           // called "ID" in v2.3.x, current version has "UserID"
			Name:      "user-name",         // called "Name" in v2.3.x, current version has "Username"
			GluonKey:  []byte("gluon-key"), // []byte in v2.3.x and current version, string in v2.4.x (intermediate)
			SplitMode: true,                // bool in v2.3.x and v2.4.x, enum in current version
		}},
	})

	// Write the vault to disk.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "vault.enc"), b, 0o600))

	// Migrate the vault.
	s, corrupt, err := New(dir, "default-gluon-dir", []byte("my secret key"), async.NoopPanicHandler{})
	require.NoError(t, err)
	require.NoError(t, corrupt)

	// Check the migrated vault.
	require.Equal(t, "v2.3.x-gluon-dir", s.GetGluonCacheDir())
	require.Equal(t, 1234, s.GetIMAPPort())
	require.Equal(t, 5678, s.GetSMTPPort())

	// The user should be migrated.
	userIDs := s.GetUserIDs()
	require.Len(t, userIDs, 1)

	// The migrated user should be correct.
	require.NoError(t, s.GetUser("user-id", func(user *User) {
		require.Equal(t, "user-id", user.UserID())
		require.Equal(t, "user-name", user.Username())
		require.Equal(t, []byte("gluon-key"), user.GluonKey())
		require.Equal(t, SplitMode, user.AddressMode())
	}))
}

func newLegacyVault[T any](t *testing.T, key []byte, version Version, data T) []byte {
	hash256 := sha256.Sum256(key)

	aes, err := aes.NewCipher(hash256[:])
	require.NoError(t, err)

	gcm, err := cipher.NewGCM(aes)
	require.NoError(t, err)

	nonce, err := crypto.RandomToken(gcm.NonceSize())
	require.NoError(t, err)

	dec, err := msgpack.Marshal(data)
	require.NoError(t, err)

	res, err := msgpack.Marshal(File{
		Version: version,
		Data:    gcm.Seal(nonce, nonce, dec, nil),
	})
	require.NoError(t, err)

	return res
}
