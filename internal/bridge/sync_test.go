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

package bridge_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"

	"github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/bradenaw/juniper/iterator"
	"github.com/bradenaw/juniper/stream"
	"github.com/emersion/go-imap/client"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"gitlab.protontech.ch/go/liteapi"
	"gitlab.protontech.ch/go/liteapi/server"
)

func TestBridge_Sync(t *testing.T) {
	numMsg := 1 << 8

	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *liteapi.NetCtl, locator bridge.Locator, storeKey []byte) {
		userID, addrID, err := s.CreateUser("imap", "imap@pm.me", password)
		require.NoError(t, err)

		labelID, err := s.CreateLabel(userID, "folder", "", liteapi.LabelTypeFolder)
		require.NoError(t, err)

		withClient(ctx, t, s, "imap", password, func(ctx context.Context, c *liteapi.Client) {
			createMessages(ctx, t, c, addrID, labelID, numMsg)
		})

		var total uint64

		// The initial user should be fully synced.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			syncCh, done := chToType[events.Event, events.SyncFinished](bridge.GetEvents(events.SyncFinished{}))
			defer done()

			// Count how many bytes it takes to fully sync the user.
			total = countBytesRead(netCtl, func() {
				userID, err := bridge.LoginFull(ctx, "imap", password, nil, nil)
				require.NoError(t, err)

				require.Equal(t, userID, (<-syncCh).UserID)
			})
		})

		// If we then connect an IMAP client, it should see all the messages.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(b *bridge.Bridge, mocks *bridge.Mocks) {
			mocks.Reporter.EXPECT().ReportMessageWithContext(gomock.Any(), gomock.Any()).AnyTimes()

			info, err := b.GetUserInfo(userID)
			require.NoError(t, err)
			require.True(t, info.State == bridge.Connected)

			client, err := client.Dial(fmt.Sprintf("%v:%v", constants.Host, b.GetIMAPPort()))
			require.NoError(t, err)
			require.NoError(t, client.Login("imap@pm.me", string(info.BridgePass)))
			defer func() { _ = client.Logout() }()

			status, err := client.Select(`Folders/folder`, false)
			require.NoError(t, err)
			require.Equal(t, uint32(numMsg), status.Messages)
		})

		// Now let's remove the user and simulate a network error.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			require.NoError(t, bridge.DeleteUser(ctx, userID))
		})

		// Pretend we can only sync 2/3 of the original messages.
		netCtl.SetReadLimit(2 * total / 3)

		// Login the user; its sync should fail.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(b *bridge.Bridge, mocks *bridge.Mocks) {
			mocks.Reporter.EXPECT().ReportMessageWithContext(gomock.Any(), gomock.Any()).AnyTimes()

			{
				syncCh, done := chToType[events.Event, events.SyncFailed](b.GetEvents(events.SyncFailed{}))
				defer done()

				userID, err := b.LoginFull(ctx, "imap", password, nil, nil)
				require.NoError(t, err)

				require.Equal(t, userID, (<-syncCh).UserID)

				info, err := b.GetUserInfo(userID)
				require.NoError(t, err)
				require.True(t, info.State == bridge.Connected)

				client, err := client.Dial(fmt.Sprintf("%v:%v", constants.Host, b.GetIMAPPort()))
				require.NoError(t, err)
				require.NoError(t, client.Login("imap@pm.me", string(info.BridgePass)))
				defer func() { _ = client.Logout() }()

				status, err := client.Select(`Folders/folder`, false)
				require.NoError(t, err)
				require.Less(t, status.Messages, uint32(numMsg))
			}

			// Remove the network limit, allowing the sync to finish.
			netCtl.SetReadLimit(0)

			{
				syncCh, done := chToType[events.Event, events.SyncFinished](b.GetEvents(events.SyncFinished{}))
				defer done()

				require.Equal(t, userID, (<-syncCh).UserID)

				info, err := b.GetUserInfo(userID)
				require.NoError(t, err)
				require.True(t, info.State == bridge.Connected)

				client, err := client.Dial(fmt.Sprintf("%v:%v", constants.Host, b.GetIMAPPort()))
				require.NoError(t, err)
				require.NoError(t, client.Login("imap@pm.me", string(info.BridgePass)))
				defer func() { _ = client.Logout() }()

				status, err := client.Select(`Folders/folder`, false)
				require.NoError(t, err)
				require.Equal(t, uint32(numMsg), status.Messages)
			}
		})
	}, server.WithTLS(false))
}

func withClient(ctx context.Context, t *testing.T, s *server.Server, username string, password []byte, fn func(context.Context, *liteapi.Client)) {
	m := liteapi.New(
		liteapi.WithHostURL(s.GetHostURL()),
		liteapi.WithTransport(liteapi.InsecureTransport()),
	)

	c, _, err := m.NewClientWithLogin(ctx, username, password)
	require.NoError(t, err)
	defer c.Close()

	fn(ctx, c)
}

func createMessages(ctx context.Context, t *testing.T, c *liteapi.Client, addrID, labelID string, count int) {
	literal, err := os.ReadFile(filepath.Join("testdata", "text-plain.eml"))
	require.NoError(t, err)

	user, err := c.GetUser(ctx)
	require.NoError(t, err)

	addr, err := c.GetAddresses(ctx)
	require.NoError(t, err)

	salt, err := c.GetSalts(ctx)
	require.NoError(t, err)

	keyPass, err := salt.SaltForKey(password, user.Keys.Primary().ID)
	require.NoError(t, err)

	_, addrKRs, err := liteapi.Unlock(user, addr, keyPass)
	require.NoError(t, err)

	require.NoError(t, getErr(stream.Collect(ctx, c.ImportMessages(
		ctx,
		addrKRs[addrID],
		runtime.NumCPU(),
		runtime.NumCPU(),
		iterator.Collect(iterator.Map(iterator.Counter(count), func(i int) liteapi.ImportReq {
			return liteapi.ImportReq{
				Metadata: liteapi.ImportMetadata{
					AddressID: addrID,
					LabelIDs:  []string{labelID},
					Flags:     liteapi.MessageFlagReceived,
				},
				Message: literal,
			}
		}))...,
	))))
}

func countBytesRead(ctl *liteapi.NetCtl, fn func()) uint64 {
	var read uint64

	ctl.OnRead(func(b []byte) {
		atomic.AddUint64(&read, uint64(len(b)))
	})

	fn()

	return read
}
