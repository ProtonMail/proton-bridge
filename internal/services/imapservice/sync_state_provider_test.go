// Copyright (c) 2026 Proton AG
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

package imapservice

import (
	"context"
	"testing"

	"github.com/ProtonMail/proton-bridge/v3/internal/sentry"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/syncservice"
	"github.com/ProtonMail/proton-bridge/v3/pkg/utils"
	"github.com/bradenaw/juniper/xmaps"
	"github.com/stretchr/testify/require"
)

func TestMigrateSyncSettings_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := GetSyncConfigPath(tmpDir, "test")

	expected, err := generateTestState(testFile)
	require.NoError(t, err)

	migrated, err := MigrateVaultSettings(tmpDir, "test", true, true, nil)
	require.NoError(t, err)
	require.False(t, migrated)

	state, err := NewSyncState(testFile, sentry.NullSentryReporter{})
	require.NoError(t, err)
	status, err := state.GetSyncStatus(context.Background())
	require.NoError(t, err)
	require.Equal(t, expected, status)
}

func TestMigrateSyncSettings_DoesNotExist(t *testing.T) {
	tmpDir := t.TempDir()

	failedIDs := []string{"foo", "bar"}
	migrated, err := MigrateVaultSettings(tmpDir, "test", true, true, failedIDs)
	require.NoError(t, err)
	require.True(t, migrated)

	state, err := NewSyncState(GetSyncConfigPath(tmpDir, "test"), sentry.NullSentryReporter{})
	require.NoError(t, err)
	status, err := state.GetSyncStatus(context.Background())
	require.NoError(t, err)
	require.Zero(t, status.NumSyncedMessages)
	require.Zero(t, status.TotalMessageCount)
	require.Empty(t, status.LastSyncedMessageID)
	require.ElementsMatch(t, failedIDs, utils.Keys(status.FailedMessages))
	require.True(t, status.HasLabels)
	require.True(t, status.HasMessageCount)
	require.True(t, status.HasMessages)
}

func generateTestState(path string) (syncservice.Status, error) {
	status := syncservice.DefaultStatus()

	status.HasMessages = true
	status.HasLabels = false
	status.FailedMessages = xmaps.SetFromSlice([]string{"foo", "bar"})
	status.TotalMessageCount = 1204
	status.NumSyncedMessages = 100
	status.HasMessages = true

	return status, storeImpl(&status, path)
}

func TestSyncState_StartSyncEventID(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := GetSyncConfigPath(tmpDir, "test")

	state, err := NewSyncState(testFile, sentry.NullSentryReporter{})
	require.NoError(t, err)

	ctx := context.Background()
	const eventID = "EVENT_START"

	status, err := state.GetSyncStatus(ctx)
	require.NoError(t, err)
	require.Empty(t, status.StartSyncEventID)

	require.NoError(t, state.SetStartSyncEventID(ctx, eventID))

	status, err = state.GetSyncStatus(ctx)
	require.NoError(t, err)
	require.Equal(t, eventID, status.StartSyncEventID)

	reloaded, err := NewSyncState(testFile, sentry.NullSentryReporter{})
	require.NoError(t, err)

	status, err = reloaded.GetSyncStatus(ctx)
	require.NoError(t, err)
	require.Equal(t, eventID, status.StartSyncEventID)
}

func TestSyncState_ClearSyncStatusClearsStartSyncEventID(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := GetSyncConfigPath(tmpDir, "test")

	state, err := NewSyncState(testFile, sentry.NullSentryReporter{})
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, state.SetHasLabels(ctx, true))
	require.NoError(t, state.SetStartSyncEventID(ctx, "EVENT_START"))

	require.NoError(t, state.ClearSyncStatus(ctx))

	status, err := state.GetSyncStatus(ctx)
	require.NoError(t, err)
	require.Empty(t, status.StartSyncEventID)
	require.False(t, status.HasLabels)
}
