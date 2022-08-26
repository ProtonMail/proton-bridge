package vault_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVault_TLSCerts(t *testing.T) {
	// create a new test vault.
	s := newVault(t)

	// Check the default bridge TLS certs.
	require.NotEmpty(t, s.GetBridgeTLSCert())
	require.NotEmpty(t, s.GetBridgeTLSKey())

	// Check the certificates are not installed.
	require.False(t, s.GetCertsInstalled())

	// Install the certificates.
	require.NoError(t, s.SetCertsInstalled(true))

	// Check the certificates are installed.
	require.True(t, s.GetCertsInstalled())
}
