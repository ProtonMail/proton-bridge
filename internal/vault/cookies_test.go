package vault_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVault_Cookies(t *testing.T) {
	// create a new test vault.
	s := newVault(t)

	// Check the default cookies are empty.
	cookies, err := s.GetCookies()
	require.NoError(t, err)
	require.Empty(t, cookies)

	// Set some cookies.
	require.NoError(t, s.SetCookies([]byte("something")))

	// Check the cookies are as set.
	newCookies, err := s.GetCookies()
	require.NoError(t, err)
	require.Equal(t, []byte("something"), newCookies)
}
