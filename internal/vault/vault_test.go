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
