package vault_test

import (
	"encoding/hex"
	"testing"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
	"github.com/stretchr/testify/require"
)

func TestUser(t *testing.T) {
	// Replace the token generator with a dummy one.
	vault.RandomToken = func(size int) ([]byte, error) {
		return []byte("token"), nil
	}

	// create a new test vault.
	s := newVault(t)

	// Set auth information for user 1 and 2.
	user1, err := s.AddUser("userID1", "user1", "authUID1", "authRef1", []byte("keyPass1"))
	require.NoError(t, err)
	user2, err := s.AddUser("userID2", "user2", "authUID2", "authRef2", []byte("keyPass2"))
	require.NoError(t, err)

	// Set event IDs for user 1 and 2.
	require.NoError(t, user1.SetEventID("eventID1"))
	require.NoError(t, user2.SetEventID("eventID2"))

	// Set sync state for user 1 and 2.
	require.NoError(t, user1.SetSync(true))
	require.NoError(t, user2.SetSync(false))

	// Set gluon data for user 1 and 2.
	require.NoError(t, user1.SetGluonID("addrID1", "gluonID1"))
	require.NoError(t, user2.SetGluonID("addrID2", "gluonID2"))
	require.NoError(t, user1.SetUIDValidity("addrID1", imap.UID(1)))
	require.NoError(t, user2.SetUIDValidity("addrID2", imap.UID(2)))

	// List available users.
	require.ElementsMatch(t, []string{"userID1", "userID2"}, s.GetUserIDs())

	// Check gluon information for user 1.
	gluonID1, ok := user1.GetGluonIDs()["addrID1"]
	require.True(t, ok)
	require.Equal(t, "gluonID1", gluonID1)
	uidValidity1, ok := user1.GetUIDValidity("addrID1")
	require.True(t, ok)
	require.Equal(t, imap.UID(1), uidValidity1)
	require.NotEmpty(t, user1.GluonKey())

	// Get auth information for user 1.
	require.Equal(t, "userID1", user1.UserID())
	require.Equal(t, "user1", user1.Username())
	require.Equal(t, hex.EncodeToString([]byte("token")), user1.BridgePass())
	require.Equal(t, vault.CombinedMode, user1.AddressMode())
	require.Equal(t, "authUID1", user1.AuthUID())
	require.Equal(t, "authRef1", user1.AuthRef())
	require.Equal(t, []byte("keyPass1"), user1.KeyPass())
	require.Equal(t, "eventID1", user1.EventID())
	require.Equal(t, true, user1.HasSync())

	// Check gluon information for user 1.
	gluonID2, ok := user2.GetGluonIDs()["addrID2"]
	require.True(t, ok)
	require.Equal(t, "gluonID2", gluonID2)
	uidValidity2, ok := user2.GetUIDValidity("addrID2")
	require.True(t, ok)
	require.Equal(t, imap.UID(2), uidValidity2)
	require.NotEmpty(t, user2.GluonKey())

	// Get auth information for user 2.
	require.Equal(t, "userID2", user2.UserID())
	require.Equal(t, "user2", user2.Username())
	require.Equal(t, hex.EncodeToString([]byte("token")), user2.BridgePass())
	require.Equal(t, vault.CombinedMode, user2.AddressMode())
	require.Equal(t, "authUID2", user2.AuthUID())
	require.Equal(t, "authRef2", user2.AuthRef())
	require.Equal(t, []byte("keyPass2"), user2.KeyPass())
	require.Equal(t, "eventID2", user2.EventID())
	require.Equal(t, false, user2.HasSync())

	// Clear the users.
	require.NoError(t, s.ClearUser("userID1"))
	require.NoError(t, s.ClearUser("userID2"))

	// Their secrets should now be cleared.
	require.Equal(t, "", user1.AuthUID())
	require.Equal(t, "", user1.AuthRef())
	require.Empty(t, user1.KeyPass())

	// Get auth information for user 2.
	require.Equal(t, "", user2.AuthUID())
	require.Equal(t, "", user2.AuthRef())
	require.Empty(t, user2.KeyPass())

	// Delete auth information for user 1.
	require.NoError(t, s.DeleteUser("userID1"))

	// List available userIDs. User 1 should be gone.
	require.ElementsMatch(t, []string{"userID2"}, s.GetUserIDs())
}
