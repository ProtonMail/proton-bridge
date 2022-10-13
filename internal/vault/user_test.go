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

func TestUser_New(t *testing.T) {
	// Replace the token generator with a dummy one.
	vault.RandomToken = func(size int) ([]byte, error) {
		return []byte("token"), nil
	}

	// Create a new test vault.
	s := newVault(t)

	// There should be no users in the store.
	require.Empty(t, s.GetUserIDs())

	// Create a new user.
	user, err := s.AddUser("userID", "username", "authUID", "authRef", []byte("keyPass"))
	require.NoError(t, err)

	// The user should be listed in the store.
	require.ElementsMatch(t, []string{"userID"}, s.GetUserIDs())

	// Check the user's default user information.
	require.Equal(t, "userID", user.UserID())
	require.Equal(t, "username", user.Username())

	// Check the user's default auth information.
	require.Equal(t, "authUID", user.AuthUID())
	require.Equal(t, "authRef", user.AuthRef())
	require.Equal(t, "keyPass", string(user.KeyPass()))

	// Check the user has a random bridge password and gluon key.
	require.Equal(t, "token", string(user.BridgePass()))
	require.Equal(t, "token", string(user.GluonKey()))

	// Check the user's initial sync status.
	require.False(t, user.SyncStatus().HasLabels)
	require.False(t, user.SyncStatus().HasMessages)
}

func TestUser_Clear(t *testing.T) {
	// Create a new test vault.
	s := newVault(t)

	// Create a new user.
	user, err := s.AddUser("userID", "username", "authUID", "authRef", []byte("keyPass"))
	require.NoError(t, err)

	// Check the user's default auth information.
	require.Equal(t, "authUID", user.AuthUID())
	require.Equal(t, "authRef", user.AuthRef())
	require.Equal(t, "keyPass", string(user.KeyPass()))

	// Clear the user's auth information.
	require.NoError(t, user.Clear())

	// Check the user's cleared auth information.
	require.Empty(t, user.AuthUID())
	require.Empty(t, user.AuthRef())
	require.Empty(t, user.KeyPass())
}

func TestUser_Delete(t *testing.T) {
	// Create a new test vault.
	s := newVault(t)

	// The store should have no users.
	require.Empty(t, s.GetUserIDs())

	// Create a new user.
	user, err := s.AddUser("userID", "username", "authUID", "authRef", []byte("keyPass"))
	require.NoError(t, err)

	// The user should be listed in the store.
	require.ElementsMatch(t, []string{"userID"}, s.GetUserIDs())

	// Try to delete the user; it should fail because it is still in use.
	require.Error(t, s.DeleteUser("userID"))

	// Close the user; it should now be deletable.
	require.NoError(t, user.Close())
	require.NoError(t, s.DeleteUser("userID"))

	// The store should have no users again.
	require.Empty(t, s.GetUserIDs())

	// Attempting to use the user should return an error.
	require.Panics(t, func() { _ = user.AddressMode() })
}

func TestUser_SyncStatus(t *testing.T) {
	// Create a new test vault.
	s := newVault(t)

	// Create a new user.
	user, err := s.AddUser("userID", "username", "authUID", "authRef", []byte("keyPass"))
	require.NoError(t, err)

	// Check the user's initial sync status.
	require.False(t, user.SyncStatus().HasLabels)
	require.False(t, user.SyncStatus().HasMessages)
	require.Empty(t, user.SyncStatus().LastMessageID)

	// Simulate having synced a message.
	require.NoError(t, user.SetLastMessageID("test"))
	require.Equal(t, "test", user.SyncStatus().LastMessageID)

	// Simulate finishing the sync.
	require.NoError(t, user.SetHasLabels(true))
	require.NoError(t, user.SetHasMessages(true))
	require.True(t, user.SyncStatus().HasLabels)
	require.True(t, user.SyncStatus().HasMessages)

	// Clear the sync status.
	require.NoError(t, user.ClearSyncStatus())

	// Check the user's cleared sync status.
	require.False(t, user.SyncStatus().HasLabels)
	require.False(t, user.SyncStatus().HasMessages)
	require.Empty(t, user.SyncStatus().LastMessageID)
}

func TestUser_ForEach(t *testing.T) {
	// Create a new test vault.
	s := newVault(t)

	// Create some new users.
	user1, err := s.AddUser("userID1", "username1", "authUID1", "authRef1", []byte("keyPass1"))
	require.NoError(t, err)
	user2, err := s.AddUser("userID2", "username2", "authUID2", "authRef2", []byte("keyPass2"))
	require.NoError(t, err)

	// Iterate through the users.
	s.ForUser(func(user *vault.User) error {
		switch user.UserID() {
		case "userID1":
			require.Equal(t, "username1", user.Username())
			require.Equal(t, "authUID1", user.AuthUID())
			require.Equal(t, "authRef1", user.AuthRef())
			require.Equal(t, "keyPass1", string(user.KeyPass()))

		case "userID2":
			require.Equal(t, "username2", user.Username())
			require.Equal(t, "authUID2", user.AuthUID())
			require.Equal(t, "authRef2", user.AuthRef())
			require.Equal(t, "keyPass2", string(user.KeyPass()))

		default:
			t.Fatalf("unexpected user %q", user.UserID())
		}

		return nil
	})

	// Try to delete the first user; it should fail because it is still in use.
	require.Error(t, s.DeleteUser("userID1"))

	// Close the first user; it should now be deletable.
	require.NoError(t, user1.Close())
	require.NoError(t, s.DeleteUser("userID1"))

	// Try to delete the second user; it should fail because it is still in use.
	require.Error(t, s.DeleteUser("userID2"))

	// Close the second user; it should now be deletable.
	require.NoError(t, user2.Close())
	require.NoError(t, s.DeleteUser("userID2"))

	// The store should have no users again.
	require.Empty(t, s.GetUserIDs())
}

/*
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
	require.Equal(t, hex.EncodeToString([]byte("token")), string(user1.BridgePass()))
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
	require.Equal(t, hex.EncodeToString([]byte("token")), string(user2.BridgePass()))
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

*/
