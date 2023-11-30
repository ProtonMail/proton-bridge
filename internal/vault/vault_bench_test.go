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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package vault_test

import (
	"bytes"
	"runtime"
	"testing"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func BenchmarkVault(b *testing.B) {
	vaultDir, gluonDir := b.TempDir(), b.TempDir()

	// Create a new vault.
	s, corrupt, err := vault.New(vaultDir, gluonDir, []byte("my secret key"), async.NoopPanicHandler{})
	require.NoError(b, err)
	require.NoError(b, corrupt)

	// Add 10kB of cookies to the vault.
	require.NoError(b, s.SetCookies(bytes.Repeat([]byte("a"), 10_000)))

	// Create 10 vault users.
	for idx := 0; idx < 10; idx++ {
		user, err := s.AddUser(uuid.NewString(), "username", "dummy@proton.me", "authUID", "authRef", []byte("keyPass"))
		require.NoError(b, err)

		require.NoError(b, user.SetKeyPass([]byte("new key pass")))
	}

	b.ResetTimer()

	// Time how quickly we can iterate through the users and get their key pass and bridge pass.
	for i := 0; i < b.N; i++ {
		require.NoError(b, s.ForUser(runtime.NumCPU(), func(user *vault.User) error {
			require.NotEmpty(b, user.KeyPass())
			require.NotEmpty(b, user.BridgePass())
			return nil
		}))
	}
}
