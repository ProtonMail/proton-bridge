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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package vault_test

import (
	"testing"

	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
	"github.com/stretchr/testify/require"
)

func TestVaultCorrupt(t *testing.T) {
	vaultDir, gluonDir := t.TempDir(), t.TempDir()

	{
		_, corrupt, err := vault.New(vaultDir, gluonDir, []byte("my secret key"))
		require.NoError(t, err)
		require.False(t, corrupt)
	}

	{
		_, corrupt, err := vault.New(vaultDir, gluonDir, []byte("my secret key"))
		require.NoError(t, err)
		require.False(t, corrupt)
	}

	{
		_, corrupt, err := vault.New(vaultDir, gluonDir, []byte("bad key"))
		require.NoError(t, err)
		require.True(t, corrupt)
	}
}

func newVault(t *testing.T) *vault.Vault {
	t.Helper()

	s, corrupt, err := vault.New(t.TempDir(), t.TempDir(), []byte("my secret key"))
	require.NoError(t, err)
	require.False(t, corrupt)

	return s
}
