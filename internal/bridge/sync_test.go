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

package bridge_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
	"github.com/ProtonMail/proton-bridge/v3/internal/services/imapservice"
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

		// Now let's remove the user and stop the network at 2/3 of the data.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			require.NoError(t, bridge.DeleteUser(ctx, userID))
		})

		// Pretend we can only sync 2/3 of the original messages.
		netCtl.SetReadLimit(2 * total / 3)

		// Login the user; its sync should fail.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(b *bridge.Bridge, _ *bridge.Mocks) {
			syncCh, done := chToType[events.Event, events.SyncFinished](b.GetEvents(events.SyncFinished{}))
			defer done()

			{
				syncFailedCh, syncFailedDone := chToType[events.Event, events.SyncFailed](b.GetEvents(events.SyncFailed{}))
				defer syncFailedDone()

				userID, err := b.LoginFull(ctx, "imap", password, nil, nil)
				require.NoError(t, err)

				require.Equal(t, userID, (<-syncFailedCh).UserID)

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
				require.Equal(t, userID, (<-syncCh).UserID)

				info, err := b.GetUserInfo(userID)
				require.NoError(t, err)
				require.True(t, info.State == bridge.Connected)

				client, err := eventuallyDial(fmt.Sprintf("%v:%v", constants.Host, b.GetIMAPPort()))
				require.NoError(t, err)
				require.NoError(t, client.Login(info.Addresses[0], string(info.BridgePass)))
				defer func() { _ = client.Logout() }()

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

func TestBridge_CanProcessEventsDuringSync(t *testing.T) {
	numMsg := 1 << 8

	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		userID, addrID, err := s.CreateUser("imap", password)
		require.NoError(t, err)

		labelID, err := s.CreateLabel(userID, "folder", "", proton.LabelTypeFolder)
		require.NoError(t, err)

		withClient(ctx, t, s, "imap", password, func(ctx context.Context, c *proton.Client) {
			createNumMessages(ctx, t, c, addrID, labelID, numMsg)
		})

		// Simulate 429 to prevent sync from progressing.
		s.AddStatusHook(func(request *http.Request) (int, bool) {
			if strings.Contains(request.URL.Path, "/mail/v4/messages/") {
				return http.StatusTooManyRequests, true
			}

			return 0, false
		})

		// The initial user should be fully synced.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			syncStartedCh, syncStartedDone := chToType[events.Event, events.SyncStarted](bridge.GetEvents(events.SyncStarted{}))
			defer syncStartedDone()

			addressCreatedCh, addressCreatedDone := chToType[events.Event, events.UserAddressCreated](bridge.GetEvents(events.UserAddressCreated{}))
			defer addressCreatedDone()

			userID, err := bridge.LoginFull(ctx, "imap", password, nil, nil)
			require.NoError(t, err)

			require.Equal(t, userID, (<-syncStartedCh).UserID)

			// Create a new address
			newAddress := "foo@proton.ch"
			addrID, err := s.CreateAddress(userID, newAddress, password)
			require.NoError(t, err)

			event := <-addressCreatedCh
			require.Equal(t, userID, event.UserID)
			require.Equal(t, newAddress, event.Email)
			require.Equal(t, addrID, event.AddressID)
		})
	}, server.WithTLS(false))
}

func TestBridge_RefreshDuringSyncRestartSync(t *testing.T) {
	numMsg := 1 << 8

	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		userID, addrID, err := s.CreateUser("imap", password)
		require.NoError(t, err)

		labelID, err := s.CreateLabel(userID, "folder", "", proton.LabelTypeFolder)
		require.NoError(t, err)

		withClient(ctx, t, s, "imap", password, func(ctx context.Context, c *proton.Client) {
			createNumMessages(ctx, t, c, addrID, labelID, numMsg)
		})

		var refreshPerformed atomic.Bool
		refreshPerformed.Store(false)

		// Simulate 429 to prevent sync from progressing.
		s.AddStatusHook(func(request *http.Request) (int, bool) {
			if strings.Contains(request.URL.Path, "/mail/v4/messages/") {
				if !refreshPerformed.Load() {
					return http.StatusTooManyRequests, true
				}
			}

			return 0, false
		})

		// The initial user should be fully synced.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			syncCh, done := chToType[events.Event, events.SyncFinished](bridge.GetEvents(events.SyncFinished{}))
			defer done()

			userID, err := bridge.LoginFull(ctx, "imap", password, nil, nil)
			require.NoError(t, err)

			syncStartedCh, syncStartedDone := chToType[events.Event, events.SyncStarted](bridge.GetEvents(events.SyncStarted{}))
			defer syncStartedDone()

			require.Equal(t, userID, (<-syncStartedCh).UserID)

			require.NoError(t, err, s.RefreshUser(userID, proton.RefreshMail))
			require.Equal(t, userID, (<-syncStartedCh).UserID)
			refreshPerformed.Store(true)

			require.Equal(t, userID, (<-syncCh).UserID)
		})
	}, server.WithTLS(false))
}

func TestBridge_EventReplayAfterSyncHasFinished(t *testing.T) {
	numMsg := 1 << 8

	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		userID, addrID, err := s.CreateUser("imap", password)
		require.NoError(t, err)

		labelID, err := s.CreateLabel(userID, "folder", "", proton.LabelTypeFolder)
		require.NoError(t, err)

		withClient(ctx, t, s, "imap", password, func(ctx context.Context, c *proton.Client) {
			createNumMessages(ctx, t, c, addrID, labelID, numMsg)
		})

		addrID1, err := s.CreateAddress(userID, "foo@proton.ch", password)
		require.NoError(t, err)

		var allowSyncToProgress atomic.Bool
		allowSyncToProgress.Store(false)

		// Simulate 429 to prevent sync from progressing.
		s.AddStatusHook(func(request *http.Request) (int, bool) {
			if request.Method == "GET" && strings.Contains(request.URL.Path, "/mail/v4/messages/") {
				if !allowSyncToProgress.Load() {
					return http.StatusTooManyRequests, true
				}
			}

			return 0, false
		})

		// The initial user should be fully synced.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			syncCh, done := chToType[events.Event, events.SyncFinished](bridge.GetEvents(events.SyncFinished{}))
			defer done()

			syncStartedCh, syncStartedDone := chToType[events.Event, events.SyncStarted](bridge.GetEvents(events.SyncStarted{}))
			defer syncStartedDone()

			addressCreatedCh, addressCreatedDone := chToType[events.Event, events.UserAddressCreated](bridge.GetEvents(events.UserAddressCreated{}))
			defer addressCreatedDone()

			userID, err := bridge.LoginFull(ctx, "imap", password, nil, nil)
			require.NoError(t, err)

			require.Equal(t, userID, (<-syncStartedCh).UserID)

			// create 20 more messages and move them to inbox
			withClient(ctx, t, s, "imap", password, func(ctx context.Context, c *proton.Client) {
				createNumMessages(ctx, t, c, addrID, proton.InboxLabel, 20)
			})

			// User AddrID2 event as a check point to see when the new address was created.
			addrID2, err := s.CreateAddress(userID, "bar@proton.ch", password)
			require.NoError(t, err)

			allowSyncToProgress.Store(true)
			require.Equal(t, userID, (<-syncCh).UserID)

			// At most two events can be published, one for the first address, then for the second.
			// if the second event is not `addrID2` then something went wrong.
			event := <-addressCreatedCh
			if event.AddressID == addrID1 {
				event = <-addressCreatedCh
			}

			require.Equal(t, addrID2, event.AddressID)

			info, err := bridge.GetUserInfo(userID)
			require.NoError(t, err)

			client, err := eventuallyDial(fmt.Sprintf("%v:%v", constants.Host, bridge.GetIMAPPort()))
			require.NoError(t, err)
			require.NoError(t, client.Login(info.Addresses[0], string(info.BridgePass)))
			defer func() { _ = client.Logout() }()

			// Finally check if the 20 messages are in INBOX.
			status, err := client.Status("INBOX", []imap.StatusItem{imap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(20), status.Messages)

			// Finally check if the numMsg are in the folder.
			status, err = client.Status("Folders/folder", []imap.StatusItem{imap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(numMsg), status.Messages)
		})
	}, server.WithTLS(false))
}

func TestBridge_MessageCreateDuringSync(t *testing.T) {
	numMsg := 1 << 8

	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		userID, addrID, err := s.CreateUser("imap", password)
		require.NoError(t, err)

		labelID, err := s.CreateLabel(userID, "folder", "", proton.LabelTypeFolder)
		require.NoError(t, err)

		withClient(ctx, t, s, "imap", password, func(ctx context.Context, c *proton.Client) {
			createNumMessages(ctx, t, c, addrID, labelID, numMsg)
		})

		var allowSyncToProgress atomic.Bool
		allowSyncToProgress.Store(false)

		// Simulate 429 to prevent sync from progressing.
		s.AddStatusHook(func(request *http.Request) (int, bool) {
			if request.Method == "GET" && strings.Contains(request.URL.Path, "/mail/v4/messages/") {
				if !allowSyncToProgress.Load() {
					return http.StatusTooManyRequests, true
				}
			}

			return 0, false
		})

		// The initial user should be fully synced.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			syncStartedCh, syncStartedDone := chToType[events.Event, events.SyncStarted](bridge.GetEvents(events.SyncStarted{}))
			defer syncStartedDone()

			addressCreatedCh, addressCreatedDone := chToType[events.Event, events.UserAddressCreated](bridge.GetEvents(events.UserAddressCreated{}))
			defer addressCreatedDone()

			userID, err := bridge.LoginFull(ctx, "imap", password, nil, nil)
			require.NoError(t, err)

			require.Equal(t, userID, (<-syncStartedCh).UserID)

			// create 20 more messages and move them to inbox
			withClient(ctx, t, s, "imap", password, func(ctx context.Context, c *proton.Client) {
				createNumMessages(ctx, t, c, addrID, proton.InboxLabel, 20)
			})

			// User AddrID2 event as a check point to see when the new address was created.
			addrID, err := s.CreateAddress(userID, "bar@proton.ch", password)
			require.NoError(t, err)

			// At most two events can be published, one for the first address, then for the second.
			// if the second event is not `addrID` then something went wrong.
			event := <-addressCreatedCh
			require.Equal(t, addrID, event.AddressID)
			allowSyncToProgress.Store(true)

			info, err := bridge.GetUserInfo(userID)
			require.NoError(t, err)

			client, err := eventuallyDial(fmt.Sprintf("%v:%v", constants.Host, bridge.GetIMAPPort()))
			require.NoError(t, err)
			require.NoError(t, client.Login(info.Addresses[0], string(info.BridgePass)))
			defer func() { _ = client.Logout() }()

			require.Eventually(t, func() bool {
				// Finally check if the 20 messages are in INBOX.
				status, err := client.Status("INBOX", []imap.StatusItem{imap.StatusMessages})
				require.NoError(t, err)

				return uint32(20) == status.Messages
			}, 10*time.Second, time.Second)
		})
	}, server.WithTLS(false))
}

func TestBridge_CorruptedVaultClearsPreviousIMAPSyncState(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, vaultKey []byte) {
		userID, addrID, err := s.CreateUser("imap", password)
		require.NoError(t, err)

		labelID, err := s.CreateLabel(userID, "folder", "", proton.LabelTypeFolder)
		require.NoError(t, err)

		withClient(ctx, t, s, "imap", password, func(ctx context.Context, c *proton.Client) {
			createNumMessages(ctx, t, c, addrID, labelID, 100)
		})

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			syncCh, done := chToType[events.Event, events.SyncFinished](bridge.GetEvents(events.SyncFinished{}))
			defer done()

			var err error

			userID, err = bridge.LoginFull(context.Background(), "imap", password, nil, nil)
			require.NoError(t, err)

			// Wait for sync to finish
			require.Equal(t, userID, (<-syncCh).UserID)
		})

		settingsPath, err := locator.ProvideSettingsPath()
		require.NoError(t, err)

		syncConfigPath, err := locator.ProvideIMAPSyncConfigPath()
		require.NoError(t, err)

		syncStatePath := imapservice.GetSyncConfigPath(syncConfigPath, userID)
		// Check sync state is complete
		{
			state, err := imapservice.NewSyncState(syncStatePath)
			require.NoError(t, err)
			syncStatus, err := state.GetSyncStatus(context.Background())
			require.NoError(t, err)
			require.True(t, syncStatus.IsComplete())
		}

		// corrupt the vault
		require.NoError(t, os.WriteFile(filepath.Join(settingsPath, "vault.enc"), []byte("Trash!"), 0o600))

		// Bridge starts but can't find the gluon database dir; there should be no error.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			_, err := bridge.LoginFull(context.Background(), "imap", password, nil, nil)
			require.NoError(t, err)
		})

		// Check sync state is reset.
		{
			state, err := imapservice.NewSyncState(syncStatePath)
			require.NoError(t, err)
			syncStatus, err := state.GetSyncStatus(context.Background())
			require.NoError(t, err)
			require.False(t, syncStatus.IsComplete())
		}
	})
}

func TestBridge_AddressOrderChangeDuringSyncInCombinedModeDoesNotTriggerBadEventOnNewMessage(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		// Create a user.
		userID, addrID, err := s.CreateUser("user", password)
		require.NoError(t, err)

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			userInfoChanged, done := chToType[events.Event, events.UserChanged](bridge.GetEvents(events.UserChanged{}))
			defer done()

			withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
				createNumMessages(ctx, t, c, addrID, proton.InboxLabel, 300)
			})

			_, err := bridge.LoginFull(ctx, "user", password, nil, nil)
			require.NoError(t, err)

			info, err := bridge.GetUserInfo(userID)
			require.NoError(t, err)
			require.Equal(t, 1, len(info.Addresses))
			require.Equal(t, info.Addresses[0], "user@proton.local")

			addrID2, err := s.CreateAddress(userID, "foo@"+s.GetDomain(), password)
			require.NoError(t, err)

			require.NoError(t, s.SetAddressOrder(userID, []string{addrID2, addrID}))

			withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
				createNumMessages(ctx, t, c, addrID2, proton.InboxLabel, 1)
			})

			// Since we can't intercept events at this time, we sleep for a bit to make sure the
			// new message does not get combined into the event below. This ensures the newly created
			// goes through the full code flow which triggered the original bad event.
			time.Sleep(time.Second)
			require.NoError(t, s.SetAddressOrder(userID, []string{addrID, addrID2}))

			for i := 0; i < 2; i++ {
				select {
				case <-ctx.Done():
					return
				case e := <-userInfoChanged:
					require.Equal(t, userID, e.UserID)
				}
			}
		})
	})
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
	return createMessagesWithFlags(ctx, t, c, addrID, labelID, 0, messages...)
}

func createMessagesWithFlags(ctx context.Context, t *testing.T, c *proton.Client, addrID, labelID string, flags proton.MessageFlag, messages ...[]byte) []string {
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

	var msgFlags proton.MessageFlag
	if flags == 0 {
		msgFlags = proton.MessageFlagReceived
	} else {
		msgFlags = flags
	}

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
					Flags:     msgFlags,
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
