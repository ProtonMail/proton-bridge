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
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gluon/rfc822"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/go-proton-api/server"
	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/bradenaw/juniper/iterator"
	"github.com/bradenaw/juniper/stream"
	"github.com/bradenaw/juniper/xslices"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestBridge_Sync(t *testing.T) {
	numMsg := 1 << 8

	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		userID, addrID, err := s.CreateUser("imap", password)
		require.NoError(t, err)

		labelID, err := s.CreateLabel(userID, "folder", "", proton.LabelTypeFolder)
		require.NoError(t, err)

		withClient(ctx, t, s, "imap", password, func(ctx context.Context, c *proton.Client) {
			createNumMessages(ctx, t, c, addrID, labelID, numMsg)
		})

		var total uint64

		// The initial user should be fully synced.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
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
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(b *bridge.Bridge, _ *bridge.Mocks) {
			info, err := b.GetUserInfo(userID)
			require.NoError(t, err)
			require.True(t, info.State == bridge.Connected)

			client, err := eventuallyDial(fmt.Sprintf("%v:%v", constants.Host, b.GetIMAPPort()))
			require.NoError(t, err)
			require.NoError(t, client.Login(info.Addresses[0], string(info.BridgePass)))
			defer func() { _ = client.Logout() }()

			status, err := client.Select(`Folders/folder`, false)
			require.NoError(t, err)
			require.Equal(t, uint32(numMsg), status.Messages)
		})

		// Now let's remove the user and simulate a network error.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			require.NoError(t, bridge.DeleteUser(ctx, userID))
		})

		// Pretend we can only sync 2/3 of the original messages.
		netCtl.SetReadLimit(2 * total / 3)

		// Login the user; its sync should fail.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(b *bridge.Bridge, _ *bridge.Mocks) {
			{
				syncCh, done := chToType[events.Event, events.SyncFailed](b.GetEvents(events.SyncFailed{}))
				defer done()

				userID, err := b.LoginFull(ctx, "imap", password, nil, nil)
				require.NoError(t, err)

				require.Equal(t, userID, (<-syncCh).UserID)

				info, err := b.GetUserInfo(userID)
				require.NoError(t, err)
				require.True(t, info.State == bridge.Connected)
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

				client, err := eventuallyDial(fmt.Sprintf("%v:%v", constants.Host, b.GetIMAPPort()))
				require.NoError(t, err)
				require.NoError(t, client.Login(info.Addresses[0], string(info.BridgePass)))
				defer func() { _ = client.Logout() }()

				status, err := client.Select(`Folders/folder`, false)
				require.NoError(t, err)
				require.Equal(t, uint32(numMsg), status.Messages)
			}
		})
	}, server.WithTLS(false))
}

// GODT-2215: This test no longer works since it's now possible to import messages into Gluon with bad ContentType header.
func _TestBridge_Sync_BadMessage(t *testing.T) { //nolint:unused,deadcode
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		userID, addrID, err := s.CreateUser("imap", password)
		require.NoError(t, err)

		labelID, err := s.CreateLabel(userID, "folder", "", proton.LabelTypeFolder)
		require.NoError(t, err)

		var messageIDs []string

		withClient(ctx, t, s, "imap", password, func(ctx context.Context, c *proton.Client) {
			messageIDs = createMessages(ctx, t, c, addrID, labelID,
				[]byte("To: someone@pm.me\r\nSubject: Good message\r\n\r\nHello!"),
				[]byte("To: someone@pm.me\r\nSubject: Bad message\r\nContentType: this is not a valid content type\r\n\r\nHello!"),
			)
		})

		// The initial user should be fully synced and should skip the bad message.
		// We should report the bad message to sentry.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			mocks.Reporter.EXPECT().ReportMessageWithContext("Failed to build message (sync)", gomock.Any())

			syncCh, done := chToType[events.Event, events.SyncFinished](bridge.GetEvents(events.SyncFinished{}))
			defer done()

			userID, err := bridge.LoginFull(ctx, "imap", password, nil, nil)
			require.NoError(t, err)

			require.Equal(t, userID, (<-syncCh).UserID)
		})

		// If we then connect an IMAP client, it should see the good message but not the bad one.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(b *bridge.Bridge, _ *bridge.Mocks) {
			info, err := b.GetUserInfo(userID)
			require.NoError(t, err)
			require.True(t, info.State == bridge.Connected)

			client, err := eventuallyDial(fmt.Sprintf("%v:%v", constants.Host, b.GetIMAPPort()))
			require.NoError(t, err)
			require.NoError(t, client.Login(info.Addresses[0], string(info.BridgePass)))
			defer func() { _ = client.Logout() }()

			status, err := client.Select(`Folders/folder`, false)
			require.NoError(t, err)
			require.Equal(t, uint32(1), status.Messages)

			messages, err := clientFetch(client, `Folders/folder`)
			require.NoError(t, err)
			require.Len(t, messages, 1)

			// The bad message should have been skipped.
			literal, err := io.ReadAll(messages[0].GetBody(must(imap.ParseBodySectionName("BODY[]"))))
			require.NoError(t, err)

			header, err := rfc822.Parse(literal).ParseHeader()
			require.NoError(t, err)
			require.Equal(t, "Good message", header.Get("Subject"))
			require.Equal(t, messageIDs[0], header.Get("X-Pm-Internal-Id"))
		})
	})
}

func TestBridge_SyncWithOngoingEvents(t *testing.T) {
	numMsg := 1 << 8
	messageSplitIndex := numMsg * 2 / 3
	renmainingMessageCount := numMsg - messageSplitIndex

	messages := make([]string, 0, numMsg)

	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		userID, addrID, err := s.CreateUser("imap", password)
		require.NoError(t, err)

		labelID, err := s.CreateLabel(userID, "folder", "", proton.LabelTypeFolder)
		require.NoError(t, err)

		withClient(ctx, t, s, "imap", password, func(ctx context.Context, c *proton.Client) {
			importResults := createNumMessages(ctx, t, c, addrID, labelID, numMsg)
			for _, v := range importResults {
				if len(v) != 0 {
					messages = append(messages, v)
				}
			}
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

		// Now let's remove the user and stop the network at 2/3 of the data.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			require.NoError(t, bridge.DeleteUser(ctx, userID))
		})

		// Pretend we can only sync 2/3 of the original messages.
		netCtl.SetReadLimit(2 * total / 3)

		// Login the user; its sync should fail.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(b *bridge.Bridge, mocks *bridge.Mocks) {
			{
				syncCh, done := chToType[events.Event, events.SyncFailed](b.GetEvents(events.SyncFailed{}))
				defer done()

				userID, err := b.LoginFull(ctx, "imap", password, nil, nil)
				require.NoError(t, err)

				require.Equal(t, userID, (<-syncCh).UserID)

				info, err := b.GetUserInfo(userID)
				require.NoError(t, err)
				require.True(t, info.State == bridge.Connected)
			}

			// Create a new mailbox and move that last 1/3 of the messages into it to simulate user
			// actions during sync.
			{
				newLabelID, err := s.CreateLabel(userID, "folder2", "", proton.LabelTypeFolder)
				require.NoError(t, err)

				messages := messages[messageSplitIndex:]

				withClient(ctx, t, s, "imap", password, func(ctx context.Context, c *proton.Client) {
					require.NoError(t, c.UnlabelMessages(ctx, messages, labelID))
					require.NoError(t, c.LabelMessages(ctx, messages, newLabelID))
				})
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

				client, err := eventuallyDial(fmt.Sprintf("%v:%v", constants.Host, b.GetIMAPPort()))
				require.NoError(t, err)
				require.NoError(t, client.Login(info.Addresses[0], string(info.BridgePass)))
				defer func() { _ = client.Logout() }()

				status, err := client.Select(`Folders/folder`, false)
				require.NoError(t, err)
				// Original folder should have more than 0 messages and less than the total.
				require.Greater(t, status.Messages, uint32(0))
				require.Less(t, status.Messages, uint32(numMsg))

				// Check that the new messages arrive in the right location.
				require.Eventually(t, func() bool {
					status, err := client.Select(`Folders/folder2`, true)
					if err != nil {
						return false
					}
					if status.Messages != uint32(renmainingMessageCount) {
						return false
					}

					return true
				}, 10*time.Second, 500*time.Millisecond)
			}
		})
	}, server.WithTLS(false))
}

func withClient(ctx context.Context, t *testing.T, s *server.Server, username string, password []byte, fn func(context.Context, *proton.Client)) { //nolint:unparam
	m := proton.New(
		proton.WithHostURL(s.GetHostURL()),
		proton.WithTransport(proton.InsecureTransport()),
	)

	c, _, err := m.NewClientWithLogin(ctx, username, password)
	require.NoError(t, err)
	defer c.Close()

	fn(ctx, c)
}

func clientFetch(client *client.Client, mailbox string, extraItems ...imap.FetchItem) ([]*imap.Message, error) {
	status, err := client.Select(mailbox, false)
	if err != nil {
		return nil, err
	}

	if status.Messages == 0 {
		return nil, nil
	}

	resCh := make(chan *imap.Message)

	fetchItems := []imap.FetchItem{imap.FetchFlags, imap.FetchEnvelope, imap.FetchUid, imap.FetchBodyStructure, "BODY.PEEK[]"}
	fetchItems = append(fetchItems, extraItems...)

	go func() {
		if err := client.Fetch(
			&imap.SeqSet{Set: []imap.Seq{{Start: 1, Stop: status.Messages}}},
			fetchItems,
			resCh,
		); err != nil {
			panic(err)
		}
	}()

	return iterator.Collect(iterator.Chan(resCh)), nil
}

func clientStore(client *client.Client, from, to int, isUID bool, item imap.StoreItem, flags ...string) error {
	var storeFunc func(seqset *imap.SeqSet, item imap.StoreItem, value interface{}, ch chan *imap.Message) error

	if isUID {
		storeFunc = client.UidStore
	} else {
		storeFunc = client.Store
	}

	return storeFunc(
		&imap.SeqSet{Set: []imap.Seq{{Start: uint32(from), Stop: uint32(to)}}},
		item,
		xslices.Map(flags, func(flag string) interface{} { return flag }),
		nil,
	)
}

func clientList(client *client.Client) []*imap.MailboxInfo {
	resCh := make(chan *imap.MailboxInfo)

	go func() {
		if err := client.List("", "*", resCh); err != nil {
			panic(err)
		}
	}()

	return iterator.Collect(iterator.Chan(resCh))
}

func createNumMessages(ctx context.Context, t *testing.T, c *proton.Client, addrID, labelID string, count int) []string {
	literal, err := os.ReadFile(filepath.Join("testdata", "text-plain.eml"))
	require.NoError(t, err)

	return createMessages(ctx, t, c, addrID, labelID, xslices.Repeat(literal, count)...)
}

func createMessages(ctx context.Context, t *testing.T, c *proton.Client, addrID, labelID string, messages ...[]byte) []string {
	user, err := c.GetUser(ctx)
	require.NoError(t, err)

	addr, err := c.GetAddresses(ctx)
	require.NoError(t, err)

	salt, err := c.GetSalts(ctx)
	require.NoError(t, err)

	keyPass, err := salt.SaltForKey(password, user.Keys.Primary().ID)
	require.NoError(t, err)

	_, addrKRs, err := proton.Unlock(user, addr, keyPass, async.NoopPanicHandler{})
	require.NoError(t, err)

	_, ok := addrKRs[addrID]
	require.True(t, ok)

	str, err := c.ImportMessages(
		ctx,
		addrKRs[addrID],
		runtime.NumCPU(),
		runtime.NumCPU(),
		xslices.Map(messages, func(message []byte) proton.ImportReq {
			return proton.ImportReq{
				Metadata: proton.ImportMetadata{
					AddressID: addrID,
					LabelIDs:  []string{labelID},
					Flags:     proton.MessageFlagReceived,
				},
				Message: message,
			}
		})...,
	)
	require.NoError(t, err)

	res, err := stream.Collect(ctx, str)
	require.NoError(t, err)

	return xslices.Map(res, func(res proton.ImportRes) string {
		return res.MessageID
	})
}

func countBytesRead(ctl *proton.NetCtl, fn func()) uint64 {
	var read uint64

	ctl.OnRead(func(b []byte) {
		atomic.AddUint64(&read, uint64(len(b)))
	})

	fn()

	return read
}
