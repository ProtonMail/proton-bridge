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
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/go-proton-api/server"
	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/user"
	"github.com/bradenaw/juniper/xslices"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestBridge_User_RefreshEvent(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		// Create a user.
		userID, addrID, err := s.CreateUser("user", password)
		require.NoError(t, err)

		labelID, err := s.CreateLabel(userID, "folder", "", proton.LabelTypeFolder)
		require.NoError(t, err)

		var messageIDs []string

		// Create 10 messages for the user.
		withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
			messageIDs = createNumMessages(ctx, t, c, addrID, labelID, 10)
		})

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			userLoginAndSync(ctx, t, bridge, "user", password)
		})

		// Remove a message
		withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
			require.NoError(t, c.DeleteMessage(ctx, messageIDs[0]))
		})

		require.NoError(t, s.RefreshUser(userID, proton.RefreshMail))

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			syncCh, closeCh := chToType[events.Event, events.SyncFinished](bridge.GetEvents(events.SyncFinished{}))

			mocks.Reporter.EXPECT().ReportMessageWithContext(gomock.Any(), gomock.Any()).MinTimes(1)

			require.Equal(t, userID, (<-syncCh).UserID)
			closeCh()

			userContinueEventProcess(ctx, t, s, bridge)
		})

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
				createNumMessages(ctx, t, c, addrID, labelID, 10)
			})

			userContinueEventProcess(ctx, t, s, bridge)
		})
	})
}

func TestBridge_User_BadMessage_BadEvent(t *testing.T) {
	t.Run("Resync", test_badMessage_badEvent(func(t *testing.T, ctx context.Context, bridge *bridge.Bridge, badUserID string) {
		// User feedback is resync
		require.NoError(t, bridge.SendBadEventUserFeedback(ctx, badUserID, true))

		// Wait for sync to finish
		syncCh, closeCh := chToType[events.Event, events.SyncFinished](bridge.GetEvents(events.SyncFinished{}))
		require.Equal(t, badUserID, (<-syncCh).UserID)
		closeCh()
	}))

	t.Run("LogoutAndLogin", test_badMessage_badEvent(func(t *testing.T, ctx context.Context, bridge *bridge.Bridge, badUserID string) {
		// User feedback is logout
		require.NoError(t, bridge.SendBadEventUserFeedback(ctx, badUserID, false))

		// The user will eventually be logged out due to the bad request errors.
		require.Eventually(t, func() bool {
			return len(bridge.GetUserIDs()) == 1 && len(getConnectedUserIDs(t, bridge)) == 0
		}, 100*user.EventPeriod, user.EventPeriod)

		// Login again
		_, err := bridge.LoginFull(ctx, "user", password, nil, nil)
		require.NoError(t, err)
	}))
}

func test_badMessage_badEvent(userFeedback func(t *testing.T, ctx context.Context, bridge *bridge.Bridge, badUserID string)) func(t *testing.T) {
	return func(t *testing.T) {
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

			withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
				userLoginAndSync(ctx, t, bridge, "user", password)

				var messageIDs []string

				// Create 10 more messages for the user, generating events.
				withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
					messageIDs = createNumMessages(ctx, t, c, addrID, labelID, 10)
				})

				// If bridge attempts to sync the new messages, it should get a BadRequest error.
				doBadRequest := true
				s.AddStatusHook(func(req *http.Request) (int, bool) {
					if !doBadRequest {
						return 0, false
					}

					if xslices.Index(xslices.Map(messageIDs[0:5], func(messageID string) string {
						return "/mail/v4/messages/" + messageID
					}), req.URL.Path) < 0 {
						return 0, false
					}

					return http.StatusBadRequest, true
				})

				badUserID := userReceivesBadError(t, bridge, mocks)

				// Remove messages, make response OK again
				withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
					require.NoError(t, c.DeleteMessage(ctx, messageIDs[0:5]...))
				})
				doBadRequest = false

				userFeedback(t, ctx, bridge, badUserID)

				userContinueEventProcess(ctx, t, s, bridge)
			})
		})
	}
}

func TestBridge_User_BadMessage_NoBadEvent(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		// Create a user.
		_, addrID, err := s.CreateUser("user", password)
		require.NoError(t, err)

		// Create 10 messages for the user.
		withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
			createNumMessages(ctx, t, c, addrID, proton.InboxLabel, 10)
		})

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			userLoginAndSync(ctx, t, bridge, "user", password)

			var messageIDs []string

			// If bridge attempts to sync the new messages, it should get a BadRequest error.
			s.AddStatusHook(func(req *http.Request) (int, bool) {
				if len(messageIDs) < 3 {
					return 0, false
				}

				if strings.Contains(req.URL.Path, "/mail/v4/messages/"+messageIDs[2]) {
					return http.StatusUnprocessableEntity, true
				}

				return 0, false
			})

			// Create 10 more messages for the user, generating events.
			withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
				messageIDs = createNumMessages(ctx, t, c, addrID, proton.InboxLabel, 10)
			})

			// Remove messages
			withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
				require.NoError(t, c.DeleteMessage(ctx, messageIDs...))
			})

			userContinueEventProcess(ctx, t, s, bridge)
		})
	})
}

func TestBridge_User_SameMessageLabelCreated_NoBadEvent(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		// Create a user.
		userID, addrID, err := s.CreateUser("user", password)
		require.NoError(t, err)

		var messageIDs []string

		// Create 10 messages for the user.
		withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
			messageIDs = createNumMessages(ctx, t, c, addrID, proton.InboxLabel, 10)
		})

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			userLoginAndSync(ctx, t, bridge, "user", password)

			labelID, err := s.CreateLabel(userID, "folder", "", proton.LabelTypeFolder)
			require.NoError(t, err)

			// Add NOOP events
			require.NoError(t, s.AddLabelCreatedEvent(userID, labelID))
			require.NoError(t, s.AddMessageCreatedEvent(userID, messageIDs[9]))

			userContinueEventProcess(ctx, t, s, bridge)
		})
	})
}

func TestBridge_User_MessageLabelDeleted_NoBadEvent(t *testing.T) {
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

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			userLoginAndSync(ctx, t, bridge, "user", password)

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

			userContinueEventProcess(ctx, t, s, bridge)
		})
	})
}

func TestBridge_User_AddressEvents_NoBadEvent(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		// Create a user.
		userID, addrID, err := s.CreateUser("user", password)
		require.NoError(t, err)

		// Create 10 messages for the user.
		withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
			createNumMessages(ctx, t, c, addrID, proton.InboxLabel, 10)
		})

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			userLoginAndSync(ctx, t, bridge, "user", password)

			addrID, err = s.CreateAddress(userID, "other@pm.me", password)
			require.NoError(t, err)
			userContinueEventProcess(ctx, t, s, bridge)

			require.NoError(t, s.AddAddressCreatedEvent(userID, addrID))
			userContinueEventProcess(ctx, t, s, bridge)
		})

		otherID, err := s.CreateAddress(userID, "another@pm.me", password)
		require.NoError(t, err)
		require.NoError(t, s.RemoveAddress(userID, otherID))

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			userContinueEventProcess(ctx, t, s, bridge)

			require.NoError(t, s.CreateAddressKey(userID, addrID, password))
			userContinueEventProcess(ctx, t, s, bridge)

			require.NoError(t, s.RemoveAddress(userID, addrID))
			userContinueEventProcess(ctx, t, s, bridge)
		})
	})
}

func TestBridge_User_Network_NoBadEvents(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		retVal := int32(0)

		setResponseAndWait := func(status int32) {
			atomic.StoreInt32(&retVal, status)
			time.Sleep(user.EventPeriod)
		}

		s.AddStatusHook(func(req *http.Request) (int, bool) {
			status := atomic.LoadInt32(&retVal)
			if strings.Contains(req.URL.Path, "/core/v4/events/") {
				return int(status), status != 0
			}

			return 0, false
		})

		// Create a user.
		_, addrID, err := s.CreateUser("user", password)
		require.NoError(t, err)

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			userLoginAndSync(ctx, t, bridge, "user", password)

			// Create 10 more messages for the user, generating events.
			withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
				createNumMessages(ctx, t, c, addrID, proton.InboxLabel, 10)

				setResponseAndWait(http.StatusInternalServerError)
				setResponseAndWait(http.StatusServiceUnavailable)
				setResponseAndWait(http.StatusPaymentRequired)
				setResponseAndWait(http.StatusForbidden)
				setResponseAndWait(http.StatusBadRequest)
				setResponseAndWait(http.StatusUnprocessableEntity)
				setResponseAndWait(http.StatusTooManyRequests)
				time.Sleep(10 * time.Second) // needs minimum of 10 seconds to retry
			})

			setResponseAndWait(0)
			time.Sleep(10 * time.Second) // needs up to 20 seconds to retry
			userContinueEventProcess(ctx, t, s, bridge)
		})
	})
}

// userLoginAndSync logs in user and waits until user is fully synced.
func userLoginAndSync(
	ctx context.Context,
	t *testing.T,
	bridge *bridge.Bridge,
	username string, password []byte, //nolint:unparam
) {
	syncCh, done := chToType[events.Event, events.SyncFinished](bridge.GetEvents(events.SyncFinished{}))
	defer done()

	userID, err := bridge.LoginFull(ctx, username, password, nil, nil)
	require.NoError(t, err)

	require.Equal(t, userID, (<-syncCh).UserID)
}

func userReceivesBadError(
	t *testing.T,
	bridge *bridge.Bridge,
	mocks *bridge.Mocks,
) (userID string) {
	badEventCh, closeCh := bridge.GetEvents(events.UserBadEvent{})

	// The user will continue to process events and will receive bad request errors.
	mocks.Reporter.EXPECT().ReportMessageWithContext(gomock.Any(), gomock.Any()).MinTimes(1)

	badEvent, ok := (<-badEventCh).(events.UserBadEvent)
	require.True(t, ok)

	closeCh()

	return badEvent.UserID
}

func userContinueEventProcess(
	ctx context.Context,
	t *testing.T,
	s *server.Server,
	bridge *bridge.Bridge,
) {
	info, err := bridge.QueryUserInfo("user")
	require.NoError(t, err)

	client, err := client.Dial(fmt.Sprintf("%v:%v", constants.Host, bridge.GetIMAPPort()))
	require.NoError(t, err)
	require.NoError(t, client.Login(info.Addresses[0], string(info.BridgePass)))
	defer func() { _ = client.Logout() }()

	randomLabel := uuid.NewString()

	// Create a new label.
	withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
		require.NoError(t, getErr(c.CreateLabel(ctx, proton.CreateLabelReq{
			Name:  randomLabel,
			Color: "#f66",
			Type:  proton.LabelTypeLabel,
		})))
	})

	// Wait for the label to be created.
	require.Eventually(t, func() bool {
		return xslices.IndexFunc(clientList(client), func(mailbox *imap.MailboxInfo) bool {
			return mailbox.Name == "Labels/"+randomLabel
		}) >= 0
	}, 100*user.EventPeriod, user.EventPeriod)
}
