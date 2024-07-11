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

package vault_test

import (
	"runtime"
	"testing"

	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/stretchr/testify/require"
)

func TestUser_New(t *testing.T) {
	// Replace the token generator with a dummy one.
	vault.RandomToken = func(_ int) ([]byte, error) {
		return []byte("token"), nil
	}

	// Create a new test vault.
	s := newVault(t)

	// There should be no users in the store.
	require.Empty(t, s.GetUserIDs())

	// Create a new user.
	user, err := s.AddUser("userID", "username", "username@pm.me", "authUID", "authRef", []byte("keyPass"))
	require.NoError(t, err)

	// The user should be listed in the store.
	require.ElementsMatch(t, []string{"userID"}, s.GetUserIDs())

	// Check the user's default user information.
	require.Equal(t, "userID", user.UserID())
	require.Equal(t, "username", user.Username())
	require.Equal(t, "username@pm.me", user.PrimaryEmail())

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
	require.False(t, user.GetShouldResync())
}

func TestUser_Clear(t *testing.T) {
	// Create a new test vault.
	s := newVault(t)

	// Create a new user.
	user, err := s.AddUser("userID", "username", "username@pm.me", "authUID", "authRef", []byte("keyPass"))
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
	user, err := s.AddUser("userID", "username", "username@pm.me", "authUID", "authRef", []byte("keyPass"))
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
	user, err := s.AddUser("userID", "username", "username@pm.me", "authUID", "authRef", []byte("keyPass"))
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
	require.NoError(t, user.ClearSyncStatusDeprecated())

	// Check the user's cleared sync status.
	require.False(t, user.SyncStatus().HasLabels)
	require.False(t, user.SyncStatus().HasMessages)
	require.Empty(t, user.SyncStatus().LastMessageID)
}

func TestUser_ClearSyncStatusWithoutEventID(t *testing.T) {
	// Create a new test vault.
	s := newVault(t)

	// Create a new user.
	user, err := s.AddUser("userID", "username", "username@pm.me", "authUID", "authRef", []byte("keyPass"))
	require.NoError(t, err)

	// Simulate finishing the sync.
	require.NoError(t, user.SetHasLabels(true))
	require.NoError(t, user.SetHasMessages(true))
	require.True(t, user.SyncStatus().HasLabels)
	require.True(t, user.SyncStatus().HasMessages)
	require.NoError(t, user.SetEventID("foo"))

	// Clear the sync status.
	require.NoError(t, user.ClearSyncStatusWithoutEventID())

	// Check the user's cleared sync status.
	require.False(t, user.SyncStatus().HasLabels)
	require.False(t, user.SyncStatus().HasMessages)
	require.Empty(t, user.SyncStatus().LastMessageID)
	require.Equal(t, "foo", user.EventID())
}

func TestUser_PrimaryEmail(t *testing.T) {
	// Create a new test vault.
	s := newVault(t)

	// Create a user.
	user, err := s.AddUser("userID", "username", "username@pm.me", "authUID", "authRef", []byte("keyPass"))
	require.NoError(t, err)

	// Check that we can successfully modify a primary email
	require.Equal(t, user.PrimaryEmail(), "username@pm.me")
	require.NoError(t, user.SetPrimaryEmail("newname@pm.me"))
	require.Equal(t, user.PrimaryEmail(), "newname@pm.me")
	require.NoError(t, user.SetPrimaryEmail(""))
	require.Equal(t, user.PrimaryEmail(), "")
}

func TestUser_ForEach(t *testing.T) {
	// Create a new test vault.
	s := newVault(t)

	// Create some new users.
	user1, err := s.AddUser("userID1", "username1", "username1@pm.me", "authUID1", "authRef1", []byte("keyPass1"))
	require.NoError(t, err)
	user2, err := s.AddUser("userID2", "username2", "username2@pm.me", "authUID2", "authRef2", []byte("keyPass2"))
	require.NoError(t, err)

	// Iterate through the users.
	err = s.ForUser(runtime.NumCPU(), func(user *vault.User) error {
		switch user.UserID() {
		case "userID1":
			require.Equal(t, "username1", user.Username())
			require.Equal(t, "username1@pm.me", user.PrimaryEmail())
			require.Equal(t, "authUID1", user.AuthUID())
			require.Equal(t, "authRef1", user.AuthRef())
			require.Equal(t, "keyPass1", string(user.KeyPass()))

		case "userID2":
			require.Equal(t, "username2", user.Username())
			require.Equal(t, "username2@pm.me", user.PrimaryEmail())
			require.Equal(t, "authUID2", user.AuthUID())
			require.Equal(t, "authRef2", user.AuthRef())
			require.Equal(t, "keyPass2", string(user.KeyPass()))

		default:
			t.Fatalf("unexpected user %q", user.UserID())
		}

		return nil
	})

	require.NoError(t, err)

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

func TestUser_ShouldResync(t *testing.T) {
	// Replace the token generator with a dummy one.
	vault.RandomToken = func(_ int) ([]byte, error) {
		return []byte("token"), nil
	}

	// Create a new test vault.
	s := newVault(t)

	// There should be no users in the store.
	require.Empty(t, s.GetUserIDs())

	// Create a new user.
	user, err := s.AddUser("userID", "username", "username@pm.me", "authUID", "authRef", []byte("keyPass"))
	require.NoError(t, err)

	// The user should be listed in the store.
	require.ElementsMatch(t, []string{"userID"}, s.GetUserIDs())

	// The shouldResync field is supposed to be false for new users.
	require.False(t, user.GetShouldResync())

	// Set it to true
	if err := user.SetShouldSync(true); err != nil {
		t.Fatalf("Failed to set should-sync: %v", err)
	}

	// Check whether it matches the correct value
	require.True(t, user.GetShouldResync())
}
