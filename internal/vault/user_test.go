package vault_test

import (
	"encoding/hex"
	"testing"

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
	require.NoError(t, user1.UpdateEventID("eventID1"))
	require.NoError(t, user2.UpdateEventID("eventID2"))

	// Set sync state for user 1 and 2.
	require.NoError(t, user1.UpdateSync(true))
	require.NoError(t, user2.UpdateSync(false))

	// Set gluon data for user 1 and 2.
	require.NoError(t, user1.UpdateGluonData("gluonID1", []byte("gluonKey1")))
	require.NoError(t, user2.UpdateGluonData("gluonID2", []byte("gluonKey2")))

	// List available users.
	require.ElementsMatch(t, []string{"userID1", "userID2"}, s.GetUserIDs())

	// Get auth information for user 1.
	require.Equal(t, "userID1", user1.UserID())
	require.Equal(t, "user1", user1.Username())
	require.Equal(t, "gluonID1", user1.GluonID())
	require.Equal(t, []byte("gluonKey1"), user1.GluonKey())
	require.Equal(t, hex.EncodeToString([]byte("token")), user1.BridgePass())
	require.Equal(t, "authUID1", user1.AuthUID())
	require.Equal(t, "authRef1", user1.AuthRef())
	require.Equal(t, []byte("keyPass1"), user1.KeyPass())
	require.Equal(t, "eventID1", user1.EventID())
	require.Equal(t, true, user1.HasSync())

	// Get auth information for user 2.
	require.Equal(t, "userID2", user2.UserID())
	require.Equal(t, "user2", user2.Username())
	require.Equal(t, "gluonID2", user2.GluonID())
	require.Equal(t, []byte("gluonKey2"), user2.GluonKey())
	require.Equal(t, hex.EncodeToString([]byte("token")), user2.BridgePass())
	require.Equal(t, "authUID2", user2.AuthUID())
	require.Equal(t, "authRef2", user2.AuthRef())
	require.Equal(t, []byte("keyPass2"), user2.KeyPass())
	require.Equal(t, "eventID2", user2.EventID())
	require.Equal(t, false, user2.HasSync())

	// Clear the users.
	require.NoError(t, user1.Clear())
	require.NoError(t, user2.Clear())

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
