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
	"github.com/stretchr/testify/require"
)

func TestService_beforeStartSyncing_SetsBookmarkOnce(t *testing.T) {
	state, err := NewSyncState(GetSyncConfigPath(t.TempDir(), "test-user"), sentry.NullSentryReporter{})
	require.NoError(t, err)

	service := &Service{syncStateProvider: state}

	require.NoError(t, service.beforeStartSyncing(context.Background(), "EVENT_A"))

	status, err := state.GetSyncStatus(context.Background())
	require.NoError(t, err)
	require.Equal(t, "EVENT_A", status.StartSyncEventID)

	require.NoError(t, service.beforeStartSyncing(context.Background(), "EVENT_B"))

	status, err = state.GetSyncStatus(context.Background())
	require.NoError(t, err)
	require.Equal(t, "EVENT_A", status.StartSyncEventID)
}

func TestService_beforeStartSyncing_SkipsWhenBookmarkExists(t *testing.T) {
	state, err := NewSyncState(GetSyncConfigPath(t.TempDir(), "test-user"), sentry.NullSentryReporter{})
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, state.SetStartSyncEventID(ctx, "EVENT_EXISTING"))

	service := &Service{syncStateProvider: state}

	require.NoError(t, service.beforeStartSyncing(ctx, "EVENT_NEW"))

	status, err := state.GetSyncStatus(ctx)
	require.NoError(t, err)
	require.Equal(t, "EVENT_EXISTING", status.StartSyncEventID)
}

func TestService_clearSyncStatusResetsStartSyncBookmark(t *testing.T) {
	tmpDir := t.TempDir()
	state, err := NewSyncState(GetSyncConfigPath(tmpDir, "test-user"), sentry.NullSentryReporter{})
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, state.SetStartSyncEventID(ctx, "EVENT_OLD"))
	require.NoError(t, state.SetHasLabels(ctx, true))

	service := &Service{
		syncStateProvider: state,
		labels:            newRWLabels(),
	}

	require.NoError(t, service.syncStateProvider.ClearSyncStatus(ctx))

	status, err := state.GetSyncStatus(ctx)
	require.NoError(t, err)
	require.Empty(t, status.StartSyncEventID)
	require.False(t, status.HasLabels)

	require.NoError(t, service.beforeStartSyncing(ctx, "EVENT_REFRESH"))

	status, err = state.GetSyncStatus(ctx)
	require.NoError(t, err)
	require.Equal(t, "EVENT_REFRESH", status.StartSyncEventID)
}
