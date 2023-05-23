// Copyright (c) 2023 Proton AG
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

//go:build !windows

package bridge_test

import (
	"context"
	"syscall"
	"testing"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/go-proton-api/server"
	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/stretchr/testify/require"
)

// Disabled due to flakyness.
func _TestBridge_SyncExistsWithErrorWhenTooManyFilesAreOpen(t *testing.T) { //nolint:unused
	var rlimitCurrent syscall.Rlimit

	require.NoError(t, syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlimitCurrent))

	// Restore RLimit for Process at the end of this test
	defer func() {
		require.NoError(t, syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rlimitCurrent))
	}()

	rlimit := syscall.Rlimit{
		Max: 100,
		Cur: 100,
	}

	require.NoError(t, syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rlimit))

	numMsg := 1 << 8

	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		userID, addrID, err := s.CreateUser("imap", password)
		require.NoError(t, err)

		labelID, err := s.CreateLabel(userID, "folder", "", proton.LabelTypeFolder)
		require.NoError(t, err)

		withClient(ctx, t, s, "imap", password, func(ctx context.Context, c *proton.Client) {
			createNumMessages(ctx, t, c, addrID, labelID, numMsg)
		})

		// The initial user should be fully synced.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			syncCh, done := bridge.GetEvents(events.SyncFailed{})
			defer done()

			userID, err := bridge.LoginFull(ctx, "imap", password, nil, nil)
			require.NoError(t, err)

			evt := <-syncCh
			switch e := evt.(type) {
			case events.SyncFailed:
				require.Equal(t, userID, e.UserID)
			default:
				require.Fail(t, "Expected events.SyncFailed{}")
			}
		})
	}, server.WithTLS(false))
}
