package bridge_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/emersion/go-imap/client"
	"github.com/stretchr/testify/require"
	"gitlab.protontech.ch/go/liteapi"
	"gitlab.protontech.ch/go/liteapi/server"
)

func TestBridge_Sync(t *testing.T) {
	s := server.New()
	defer s.Close()

	numMsg := 1 << 10

	withEnv(t, s, func(ctx context.Context, netCtl *liteapi.NetCtl, locator bridge.Locator, storeKey []byte) {
		userID, addrID, err := s.CreateUser("imap", "imap@pm.me", password)
		require.NoError(t, err)

		labelID, err := s.CreateLabel(userID, "folder", liteapi.LabelTypeFolder)
		require.NoError(t, err)

		literal, err := os.ReadFile(filepath.Join("testdata", "text-plain.eml"))
		require.NoError(t, err)

		for i := 0; i < numMsg; i++ {
			messageID, err := s.CreateMessage(userID, addrID, literal, liteapi.MessageFlagReceived, false, false)
			require.NoError(t, err)
			require.NoError(t, s.LabelMessage(userID, messageID, labelID))
		}

		var read uint64

		netCtl.OnRead(func(b []byte) {
			read += uint64(len(b))
		})

		// The initial user should be fully synced.
		withBridge(t, ctx, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			syncCh, done := bridge.GetEvents(events.SyncFinished{})
			defer done()

			userID, err := bridge.LoginUser(ctx, "imap", password, nil, nil)
			require.NoError(t, err)

			require.Equal(t, userID, (<-syncCh).(events.SyncFinished).UserID)
		})

		// If we then connect an IMAP client, it should see all the messages.
		withBridge(t, ctx, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			info, err := bridge.GetUserInfo(userID)
			require.NoError(t, err)
			require.True(t, info.Connected)

			client, err := client.Dial(fmt.Sprintf(":%v", bridge.GetIMAPPort()))
			require.NoError(t, err)
			require.NoError(t, client.Login("imap@pm.me", string(info.BridgePass)))
			defer client.Logout()

			status, err := client.Select(`Folders/folder`, false)
			require.NoError(t, err)
			require.Equal(t, uint32(numMsg), status.Messages)
		})

		// Now let's remove the user and simulate a network error.
		withBridge(t, ctx, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			require.NoError(t, bridge.DeleteUser(ctx, userID))
		})

		// Pretend we can only sync 2/3 of the original messages.
		netCtl.SetReadLimit(2 * read / 3)

		// Login the user; its sync should fail.
		withBridge(t, ctx, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			syncCh, done := bridge.GetEvents(events.SyncFailed{})
			defer done()

			userID, err := bridge.LoginUser(ctx, "imap", password, nil, nil)
			require.NoError(t, err)

			require.Equal(t, userID, (<-syncCh).(events.SyncFailed).UserID)

			info, err := bridge.GetUserInfo(userID)
			require.NoError(t, err)
			require.True(t, info.Connected)

			client, err := client.Dial(fmt.Sprintf(":%v", bridge.GetIMAPPort()))
			require.NoError(t, err)
			require.NoError(t, client.Login("imap@pm.me", string(info.BridgePass)))
			defer client.Logout()

			status, err := client.Select(`Folders/folder`, false)
			require.NoError(t, err)
			require.Less(t, status.Messages, uint32(numMsg))
		})

		// Remove the network limit, allowing the sync to finish.
		netCtl.SetReadLimit(0)

		// Login the user; its sync should now finish.
		// If we then connect an IMAP client, it should eventually see all the messages.
		withBridge(t, ctx, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			syncCh, done := bridge.GetEvents(events.SyncFinished{})
			defer done()

			require.Equal(t, userID, (<-syncCh).(events.SyncFinished).UserID)

			info, err := bridge.GetUserInfo(userID)
			require.NoError(t, err)
			require.True(t, info.Connected)

			client, err := client.Dial(fmt.Sprintf(":%v", bridge.GetIMAPPort()))
			require.NoError(t, err)
			require.NoError(t, client.Login("imap@pm.me", string(info.BridgePass)))
			defer client.Logout()

			status, err := client.Select(`Folders/folder`, false)
			require.NoError(t, err)
			require.Equal(t, uint32(numMsg), status.Messages)
		})
	})
}
