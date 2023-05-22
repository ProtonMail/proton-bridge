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

package bridge_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/go-proton-api/server"
	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/bradenaw/juniper/iterator"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestBridge_Refresh(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		userID, _, err := s.CreateUser("imap", password)
		require.NoError(t, err)

		names := iterator.Collect(iterator.Map(iterator.Counter(10), func(i int) string {
			return fmt.Sprintf("folder%v", i)
		}))

		for _, name := range names {
			must(s.CreateLabel(userID, name, "", proton.LabelTypeFolder))
		}

		// The initial user should be fully synced.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(b *bridge.Bridge, _ *bridge.Mocks) {
			syncCh, done := chToType[events.Event, events.SyncFinished](b.GetEvents(events.SyncFinished{}))
			defer done()

			userID, err := b.LoginFull(ctx, "imap", password, nil, nil)
			require.NoError(t, err)

			require.Equal(t, userID, (<-syncCh).UserID)
		})

		var uidValidities = make(map[string]uint32, len(names))
		// If we then connect an IMAP client, it should see all the labels with UID validity of 1.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(b *bridge.Bridge, mocks *bridge.Mocks) {
			mocks.Reporter.EXPECT().ReportMessageWithContext(gomock.Any(), gomock.Any()).AnyTimes()

			info, err := b.GetUserInfo(userID)
			require.NoError(t, err)
			require.True(t, info.State == bridge.Connected)

			client, err := eventuallyDial(fmt.Sprintf("%v:%v", constants.Host, b.GetIMAPPort()))
			require.NoError(t, err)
			require.NoError(t, client.Login(info.Addresses[0], string(info.BridgePass)))
			defer func() { _ = client.Logout() }()

			for _, name := range names {
				status, err := client.Select("Folders/"+name, false)
				require.NoError(t, err)
				uidValidities[name] = status.UidValidity
			}
		})

		// Refresh the user; this will force a resync.
		require.NoError(t, s.RefreshUser(userID, proton.RefreshAll))

		// If we then connect an IMAP client, it should see all the labels with UID validity of 1.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(b *bridge.Bridge, mocks *bridge.Mocks) {
			mocks.Reporter.EXPECT().ReportMessageWithContext(gomock.Any(), gomock.Any()).AnyTimes()

			syncCh, done := chToType[events.Event, events.SyncFinished](b.GetEvents(events.SyncFinished{}))
			defer done()

			require.Equal(t, userID, (<-syncCh).UserID)
		})

		// After resync, the IMAP client should see all the labels with UID validity of 2.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(b *bridge.Bridge, mocks *bridge.Mocks) {
			mocks.Reporter.EXPECT().ReportMessageWithContext(gomock.Any(), gomock.Any()).AnyTimes()

			info, err := b.GetUserInfo(userID)
			require.NoError(t, err)
			require.True(t, info.State == bridge.Connected)

			client, err := eventuallyDial(fmt.Sprintf("%v:%v", constants.Host, b.GetIMAPPort()))
			require.NoError(t, err)
			require.NoError(t, client.Login(info.Addresses[0], string(info.BridgePass)))
			defer func() { _ = client.Logout() }()

			for _, name := range names {
				status, err := client.Select("Folders/"+name, false)
				require.NoError(t, err)
				require.Greater(t, status.UidValidity, uidValidities[name])
			}
		})
	})
}
