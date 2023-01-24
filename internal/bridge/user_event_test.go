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
	"net/http"
	"testing"
	"time"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/go-proton-api/server"
	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/bradenaw/juniper/xslices"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestBridge_User_BadEvents(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		// Create a user.
		userID, addrID, err := s.CreateUser("user", password)
		require.NoError(t, err)

		labelID, err := s.CreateLabel(userID, "folder", "", proton.LabelTypeFolder)
		require.NoError(t, err)

		// Create 10 messages for the user.
		withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
			createNumMessages(ctx, t, c, addrID, labelID, 10)
		})

		// The initial user should be fully synced.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			syncCh, done := chToType[events.Event, events.SyncFinished](bridge.GetEvents(events.SyncFinished{}))
			defer done()

			userID, err := bridge.LoginFull(ctx, "user", password, nil, nil)
			require.NoError(t, err)

			require.Equal(t, userID, (<-syncCh).UserID)
		})

		var messageIDs []string

		// Create 10 more messages for the user, generating events.
		withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
			messageIDs = createNumMessages(ctx, t, c, addrID, labelID, 10)
		})

		// If bridge attempts to sync the new messages, it should get a BadRequest error.
		s.AddStatusHook(func(req *http.Request) (int, bool) {
			if xslices.Index(xslices.Map(messageIDs, func(messageID string) string {
				return "/mail/v4/messages/" + messageID
			}), req.URL.Path) < 0 {
				return 0, false
			}

			return http.StatusBadRequest, true
		})

		// The user will continue to process events and will receive bad request errors.
		withMocks(t, func(mocks *bridge.Mocks) {
			mocks.Reporter.EXPECT().ReportMessageWithContext(gomock.Any(), gomock.Any()).MinTimes(1)

			// The user will eventually be logged out due to the bad request errors.
			withBridgeNoMocks(ctx, t, mocks, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge) {
				require.Eventually(t, func() bool {
					return len(bridge.GetUserIDs()) == 1 && len(getConnectedUserIDs(t, bridge)) == 0
				}, 10*time.Second, 100*time.Millisecond)
			})
		})
	})
}

func TestBridge_User_NoBadEvent_SameMessageLabelCreated(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		// Create a user.
		userID, addrID, err := s.CreateUser("user", password)
		require.NoError(t, err)

		var messageIDs []string

		// Create 10 messages for the user.
		withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
			messageIDs = createNumMessages(ctx, t, c, addrID, proton.InboxLabel, 10)
		})

		// The initial user should be fully synced.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			syncCh, done := chToType[events.Event, events.SyncFinished](bridge.GetEvents(events.SyncFinished{}))
			defer done()

			userID, err := bridge.LoginFull(ctx, "user", password, nil, nil)
			require.NoError(t, err)

			require.Equal(t, userID, (<-syncCh).UserID)
		})

		labelID, err := s.CreateLabel(userID, "folder", "", proton.LabelTypeFolder)
		require.NoError(t, err)

		// Add NOOP events
		require.NoError(t, s.AddLabelCreatedEvent(userID, labelID))
		require.NoError(t, s.AddMessageCreatedEvent(userID, messageIDs[9]))

		userContinuEventProcess(ctx, t, s, netCtl, locator, storeKey)
	})
}

func TestBridge_User_NoBadEvents_MessageLabelDeleted(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		// Create a user.
		userID, addrID, err := s.CreateUser("user", password)
		require.NoError(t, err)

		labelID, err := s.CreateLabel(userID, "folder", "", proton.LabelTypeFolder)
		require.NoError(t, err)

		// Create 10 messages for the user.
		withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
			createNumMessages(ctx, t, c, addrID, labelID, 10)
		})

		// The initial user should be fully synced.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			syncCh, done := chToType[events.Event, events.SyncFinished](bridge.GetEvents(events.SyncFinished{}))
			defer done()

			userID, err := bridge.LoginFull(ctx, "user", password, nil, nil)
			require.NoError(t, err)

			require.Equal(t, userID, (<-syncCh).UserID)
		})

		// Create and delete 10 more messages for the user, generating delete events.
		withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
			messageIDs := createNumMessages(ctx, t, c, addrID, labelID, 10)
			require.NoError(t, c.DeleteMessage(ctx, messageIDs...))
		})

		// Create and delete 10 labels for the user, generating delete events.
		withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
			for i := 0; i < 10; i++ {
				label, err := c.CreateLabel(ctx, proton.CreateLabelReq{
					Name:  uuid.NewString(),
					Color: "#f66",
					Type:  proton.LabelTypeLabel,
				})
				require.NoError(t, err)

				require.NoError(t, c.DeleteLabel(ctx, label.ID))
			}
		})

		userContinuEventProcess(ctx, t, s, netCtl, locator, storeKey)
	})
}

// userContinuEventProcess checks that user will continue to process events and will not receive any bad request errors.
func userContinuEventProcess(
	ctx context.Context,
	t *testing.T,
	s *server.Server,
	netCtl *proton.NetCtl,
	locator bridge.Locator,
	storeKey []byte,
) {
	withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
		info, err := bridge.QueryUserInfo("user")
		require.NoError(t, err)

		client, err := client.Dial(fmt.Sprintf("%v:%v", constants.Host, bridge.GetIMAPPort()))
		require.NoError(t, err)
		require.NoError(t, client.Login(info.Addresses[0], string(info.BridgePass)))
		defer func() { _ = client.Logout() }()

		// Create a new label.
		withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
			require.NoError(t, getErr(c.CreateLabel(ctx, proton.CreateLabelReq{
				Name:  "blabla",
				Color: "#f66",
				Type:  proton.LabelTypeLabel,
			})))
		})

		// Wait for the label to be created.
		require.Eventually(t, func() bool {
			return xslices.IndexFunc(clientList(client), func(mailbox *imap.MailboxInfo) bool {
				return mailbox.Name == "Labels/blabla"
			}) >= 0
		}, 10*time.Second, 100*time.Millisecond)
	})
}
