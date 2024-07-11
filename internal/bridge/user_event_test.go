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
	"net"
	"net/http"
	"net/mail"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gluon/rfc822"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/go-proton-api/server"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/user"
	"github.com/bradenaw/juniper/stream"
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

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			userLoginAndSync(ctx, t, bridge, "user", password)
		})

		// Remove a message
		withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
			require.NoError(t, c.DeleteMessage(ctx, messageIDs[0]))
		})

		require.NoError(t, s.RefreshUser(userID, proton.RefreshMail))

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			syncCh, closeCh := chToType[events.Event, events.SyncFinished](bridge.GetEvents(events.SyncFinished{}))

			require.Equal(t, userID, (<-syncCh).UserID)
			closeCh()

			userContinueEventProcess(ctx, t, s, bridge)
		})

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
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
		logoutCh, closeCh := chToType[events.Event, events.UserLoggedOut](bridge.GetEvents(events.UserLoggedOut{}))

		// User feedback is logout
		require.NoError(t, bridge.SendBadEventUserFeedback(ctx, badUserID, false))

		require.Equal(t, badUserID, (<-logoutCh).UserID)
		closeCh()

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

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			userLoginAndSync(ctx, t, bridge, "user", password)

			var messageIDs []string

			// Create 10 more messages for the user, generating events.
			withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
				messageIDs = createNumMessages(ctx, t, c, addrID, proton.InboxLabel, 10)
			})

			// If bridge attempts to sync the new messages, it should get a BadRequest error.
			s.AddStatusHook(func(req *http.Request) (int, bool) {
				if strings.Contains(req.URL.Path, "/mail/v4/messages/"+messageIDs[2]) {
					return http.StatusUnprocessableEntity, true
				}

				return 0, false
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

func TestBridge_User_AddressEventUpdatedForAddressThatDoesNotExist_NoBadEvent(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		// Create a user.
		userID, _, err := s.CreateUser("user", password)
		require.NoError(t, err)

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			userLoginAndSync(ctx, t, bridge, "user", password)
		})

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			_, err := s.CreateAddressAsUpdate(userID, "another@pm.me", password)
			require.NoError(t, err)
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

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
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

func TestBridge_User_DropConn_NoBadEvent(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	dropListener := proton.NewListener(l, proton.NewDropConn)
	defer func() { _ = dropListener.Close() }()

	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		// Create a user.
		_, addrID, err := s.CreateUser("user", password)
		require.NoError(t, err)

		// Create 10 messages for the user.
		withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
			createNumMessages(ctx, t, c, addrID, proton.InboxLabel, 10)
		})

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			var count int32
			// The first 10 times bridge attempts to sync any of the messages, drop the connection.
			s.AddStatusHook(func(req *http.Request) (int, bool) {
				if strings.Contains(req.URL.Path, "/mail/v4/messages") {
					if atomic.AddInt32(&count, 1) < 10 {
						dropListener.DropAll()
					}
				}

				return 0, false
			})
			userLoginAndSync(ctx, t, bridge, "user", password)

			mocks.Reporter.EXPECT().ReportMessageWithContext(gomock.Any(), gomock.Any()).AnyTimes()

			// Create 10 more messages for the user, generating events.
			withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
				createNumMessages(ctx, t, c, addrID, proton.InboxLabel, 10)
			})

			info, err := bridge.QueryUserInfo("user")
			require.NoError(t, err)

			cli, err := eventuallyDial(fmt.Sprintf("%v:%v", constants.Host, bridge.GetIMAPPort()))
			require.NoError(t, err)
			require.NoError(t, cli.Login(info.Addresses[0], string(info.BridgePass)))
			defer func() { _ = cli.Logout() }()

			// The IMAP client will eventually see 20 messages.
			require.Eventually(t, func() bool {
				status, err := cli.Status("INBOX", []imap.StatusItem{imap.StatusMessages})
				return err == nil && status.Messages == 20
			}, 10*time.Second, 100*time.Millisecond)
		})
	}, server.WithListener(dropListener))
}

func TestBridge_User_UpdateDraft(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		// Create a bridge user.
		_, _, err := s.CreateUser("user", password)
		require.NoError(t, err)

		// Initially sync the user.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			userLoginAndSync(ctx, t, bridge, "user", password)
		})

		withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
			user, err := c.GetUser(ctx)
			require.NoError(t, err)

			addrs, err := c.GetAddresses(ctx)
			require.NoError(t, err)

			salts, err := c.GetSalts(ctx)
			require.NoError(t, err)

			keyPass, err := salts.SaltForKey(password, user.Keys.Primary().ID)
			require.NoError(t, err)

			_, addrKRs, err := proton.Unlock(user, addrs, keyPass, async.NoopPanicHandler{})
			require.NoError(t, err)

			// Create a draft (generating a "create draft message" event).
			draft, err := c.CreateDraft(ctx, addrKRs[addrs[0].ID], proton.CreateDraftReq{
				Message: proton.DraftTemplate{
					Subject:  "subject",
					Sender:   &mail.Address{Name: "sender", Address: addrs[0].Email},
					Body:     "body",
					MIMEType: rfc822.TextPlain,
				},
			})
			require.NoError(t, err)
			require.Empty(t, draft.ReplyTos)

			// Process those events
			withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
				userContinueEventProcess(ctx, t, s, bridge)
			})

			// Update the draft (generating an "update draft message" event).
			draft2, err := c.UpdateDraft(ctx, draft.ID, addrKRs[addrs[0].ID], proton.UpdateDraftReq{
				Message: proton.DraftTemplate{
					Subject:  "subject 2",
					Sender:   &mail.Address{Name: "sender", Address: addrs[0].Email},
					Body:     "body 2",
					MIMEType: rfc822.TextPlain,
				},
			})
			require.NoError(t, err)
			require.Empty(t, draft2.ReplyTos)
		})
	})
}

func TestBridge_User_UpdateDraftAndCreateOtherMessage(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		// Create a bridge user.
		_, _, err := s.CreateUser("user", password)
		require.NoError(t, err)

		// Initially sync the user.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			userLoginAndSync(ctx, t, bridge, "user", password)
		})

		withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
			user, err := c.GetUser(ctx)
			require.NoError(t, err)

			addrs, err := c.GetAddresses(ctx)
			require.NoError(t, err)

			salts, err := c.GetSalts(ctx)
			require.NoError(t, err)

			keyPass, err := salts.SaltForKey(password, user.Keys.Primary().ID)
			require.NoError(t, err)

			_, addrKRs, err := proton.Unlock(user, addrs, keyPass, async.NoopPanicHandler{})
			require.NoError(t, err)

			// Create a draft (generating a "create draft message" event).
			draft, err := c.CreateDraft(ctx, addrKRs[addrs[0].ID], proton.CreateDraftReq{
				Message: proton.DraftTemplate{
					Subject:  "subject",
					Sender:   &mail.Address{Name: "sender", Address: addrs[0].Email},
					Body:     "body",
					MIMEType: rfc822.TextPlain,
				},
			})
			require.NoError(t, err)

			// Process those events
			withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
				userContinueEventProcess(ctx, t, s, bridge)
			})

			// Update the draft (generating an "update draft message" event).
			require.NoError(t, getErr(c.UpdateDraft(ctx, draft.ID, addrKRs[addrs[0].ID], proton.UpdateDraftReq{
				Message: proton.DraftTemplate{
					Subject:  "subject 2",
					Sender:   &mail.Address{Name: "sender", Address: addrs[0].Email},
					Body:     "body 2",
					MIMEType: rfc822.TextPlain,
				},
			})))

			// Import a message (generating a "create message" event).
			str, err := c.ImportMessages(ctx, addrKRs[addrs[0].ID], 1, 1, proton.ImportReq{
				Metadata: proton.ImportMetadata{
					AddressID: addrs[0].ID,
					Flags:     proton.MessageFlagReceived,
				},
				Message: []byte("From: someone@example.com\r\nTo: blabla@example.com\r\n\r\nhello"),
			})
			require.NoError(t, err)

			res, err := stream.Collect(ctx, str)
			require.NoError(t, err)

			// Process those events.
			withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
				userContinueEventProcess(ctx, t, s, bridge)
			})

			// Update the imported message (generating an "update message" event).
			require.NoError(t, c.MarkMessagesUnread(ctx, res[0].MessageID))

			// Process those events.
			withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
				userContinueEventProcess(ctx, t, s, bridge)
			})
		})
	})
}

func TestBridge_User_SendDraftRemoveDraftFlag(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		// Create a bridge user.
		_, _, err := s.CreateUser("user", password)
		require.NoError(t, err)

		// Initially sync the user.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			userLoginAndSync(ctx, t, bridge, "user", password)
		})

		withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
			user, err := c.GetUser(ctx)
			require.NoError(t, err)

			addrs, err := c.GetAddresses(ctx)
			require.NoError(t, err)

			salts, err := c.GetSalts(ctx)
			require.NoError(t, err)

			keyPass, err := salts.SaltForKey(password, user.Keys.Primary().ID)
			require.NoError(t, err)

			_, addrKRs, err := proton.Unlock(user, addrs, keyPass, async.NoopPanicHandler{})
			require.NoError(t, err)

			// Create a draft (generating a "create draft message" event).
			draft, err := c.CreateDraft(ctx, addrKRs[addrs[0].ID], proton.CreateDraftReq{
				Message: proton.DraftTemplate{
					Subject:  "subject",
					ToList:   []*mail.Address{{Address: addrs[0].Email}},
					Sender:   &mail.Address{Name: "sender", Address: addrs[0].Email},
					Body:     "body",
					MIMEType: rfc822.TextPlain,
				},
			})
			require.NoError(t, err)

			// Process those events
			withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
				userContinueEventProcess(ctx, t, s, bridge)

				info, err := bridge.QueryUserInfo("user")
				require.NoError(t, err)

				cli, err := eventuallyDial(fmt.Sprintf("%v:%v", constants.Host, bridge.GetIMAPPort()))
				require.NoError(t, err)
				require.NoError(t, cli.Login(info.Addresses[0], string(info.BridgePass)))
				defer func() { _ = cli.Logout() }()

				messages, err := clientFetch(cli, "Drafts")
				require.NoError(t, err)
				require.Len(t, messages, 1)
				require.Contains(t, messages[0].Flags, imap.DraftFlag)
			})

			// Send the draft (generating an "update message" event).
			{
				pubKeys, recType, err := c.GetPublicKeys(ctx, addrs[0].Email)
				require.NoError(t, err)
				require.Equal(t, recType, proton.RecipientTypeInternal)

				var req proton.SendDraftReq

				require.NoError(t, req.AddTextPackage(addrKRs[addrs[0].ID], "body", rfc822.TextPlain, map[string]proton.SendPreferences{
					addrs[0].Email: {
						Encrypt:          true,
						PubKey:           must(crypto.NewKeyRing(must(crypto.NewKeyFromArmored(pubKeys[0].PublicKey)))),
						SignatureType:    proton.DetachedSignature,
						EncryptionScheme: proton.InternalScheme,
						MIMEType:         rfc822.TextPlain,
					},
				}, nil))

				require.NoError(t, getErr(c.SendDraft(ctx, draft.ID, req)))
			}

			// Process those events; the draft will move to the sent folder and lose the draft flag.
			withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
				userContinueEventProcess(ctx, t, s, bridge)

				info, err := bridge.QueryUserInfo("user")
				require.NoError(t, err)

				cli, err := eventuallyDial(fmt.Sprintf("%v:%v", constants.Host, bridge.GetIMAPPort()))
				require.NoError(t, err)
				require.NoError(t, cli.Login(info.Addresses[0], string(info.BridgePass)))
				defer func() { _ = cli.Logout() }()

				messages, err := clientFetch(cli, "Sent")
				require.NoError(t, err)
				require.Len(t, messages, 1)
				require.NotContains(t, messages[0].Flags, imap.DraftFlag)
			})
		})
	})
}

func TestBridge_User_DisableEnableAddress(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		// Create a user.
		userID, _, err := s.CreateUser("user", password)
		require.NoError(t, err)

		// Create an additional address for the user.
		aliasID, err := s.CreateAddress(userID, "alias@"+s.GetDomain(), password)
		require.NoError(t, err)

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			require.NoError(t, getErr(bridge.LoginFull(ctx, "user", password, nil, nil)))

			// Initially we should list the address.
			info, err := bridge.QueryUserInfo("user")
			require.NoError(t, err)
			require.Contains(t, info.Addresses, "alias@"+s.GetDomain())
		})

		// Disable the address.
		withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
			require.NoError(t, c.DisableAddress(ctx, aliasID))
		})

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// Eventually we shouldn't list the address.
			require.Eventually(t, func() bool {
				info, err := bridge.QueryUserInfo("user")
				require.NoError(t, err)

				return xslices.Index(info.Addresses, "alias@"+s.GetDomain()) < 0
			}, 5*time.Second, 100*time.Millisecond)
		})

		// Enable the address.
		withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
			require.NoError(t, c.EnableAddress(ctx, aliasID))
		})

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// Eventually we should list the address.
			require.Eventually(t, func() bool {
				info, err := bridge.QueryUserInfo("user")
				require.NoError(t, err)

				return xslices.Index(info.Addresses, "alias@"+s.GetDomain()) >= 0
			}, 5*time.Second, 100*time.Millisecond)
		})
	})
}

func TestBridge_User_CreateDisabledAddress(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		// Create a user.
		userID, _, err := s.CreateUser("user", password)
		require.NoError(t, err)

		// Create an additional address for the user.
		aliasID, err := s.CreateAddress(userID, "alias@"+s.GetDomain(), password)
		require.NoError(t, err)

		// Immediately disable the address.
		withClient(ctx, t, s, "user", password, func(ctx context.Context, c *proton.Client) {
			require.NoError(t, c.DisableAddress(ctx, aliasID))
		})

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			require.NoError(t, getErr(bridge.LoginFull(ctx, "user", password, nil, nil)))

			// Initially we shouldn't list the address.
			info, err := bridge.QueryUserInfo("user")
			require.NoError(t, err)
			require.NotContains(t, info.Addresses, "alias@"+s.GetDomain())
		})
	})
}

func TestBridge_User_HandleParentLabelRename(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			require.NoError(t, getErr(bridge.LoginFull(ctx, username, password, nil, nil)))

			info, err := bridge.QueryUserInfo(username)
			require.NoError(t, err)

			cli, err := eventuallyDial(fmt.Sprintf("%v:%v", constants.Host, bridge.GetIMAPPort()))
			require.NoError(t, err)
			require.NoError(t, cli.Login(info.Addresses[0], string(info.BridgePass)))
			defer func() { _ = cli.Logout() }()

			withClient(ctx, t, s, username, password, func(ctx context.Context, c *proton.Client) {
				parentName := uuid.NewString()
				childName := uuid.NewString()

				// Create a folder.
				parentLabel, err := c.CreateLabel(ctx, proton.CreateLabelReq{
					Name:  parentName,
					Type:  proton.LabelTypeFolder,
					Color: "#f66",
				})
				require.NoError(t, err)

				// Wait for the parent folder to be created.
				require.Eventually(t, func() bool {
					return xslices.IndexFunc(clientList(cli), func(mailbox *imap.MailboxInfo) bool {
						return mailbox.Name == fmt.Sprintf("Folders/%v", parentName)
					}) >= 0
				}, 100*user.EventPeriod, user.EventPeriod)

				// Create a subfolder.
				childLabel, err := c.CreateLabel(ctx, proton.CreateLabelReq{
					Name:     childName,
					Type:     proton.LabelTypeFolder,
					Color:    "#f66",
					ParentID: parentLabel.ID,
				})
				require.NoError(t, err)
				require.Equal(t, parentLabel.ID, childLabel.ParentID)

				// Wait for the parent folder to be created.
				require.Eventually(t, func() bool {
					return xslices.IndexFunc(clientList(cli), func(mailbox *imap.MailboxInfo) bool {
						return mailbox.Name == fmt.Sprintf("Folders/%v/%v", parentName, childName)
					}) >= 0
				}, 100*user.EventPeriod, user.EventPeriod)

				newParentName := uuid.NewString()

				// Rename the parent folder.
				require.NoError(t, getErr(c.UpdateLabel(ctx, parentLabel.ID, proton.UpdateLabelReq{
					Color: "#f66",
					Name:  newParentName,
				})))

				// Wait for the parent folder to be renamed.
				require.Eventually(t, func() bool {
					return xslices.IndexFunc(clientList(cli), func(mailbox *imap.MailboxInfo) bool {
						return mailbox.Name == fmt.Sprintf("Folders/%v", newParentName)
					}) >= 0
				}, 100*user.EventPeriod, user.EventPeriod)

				// Wait for the child folder to be renamed.
				require.Eventually(t, func() bool {
					return xslices.IndexFunc(clientList(cli), func(mailbox *imap.MailboxInfo) bool {
						return mailbox.Name == fmt.Sprintf("Folders/%v/%v", newParentName, childName)
					}) >= 0
				}, 100*user.EventPeriod, user.EventPeriod)
			})
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

	cli, err := eventuallyDial(fmt.Sprintf("%v:%v", constants.Host, bridge.GetIMAPPort()))
	require.NoError(t, err)
	require.NoError(t, cli.Login(info.Addresses[0], string(info.BridgePass)))
	defer func() { _ = cli.Logout() }()

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
		return xslices.IndexFunc(clientList(cli), func(mailbox *imap.MailboxInfo) bool {
			return mailbox.Name == "Labels/"+randomLabel
		}) >= 0
	}, 100*user.EventPeriod, user.EventPeriod)
}

func eventuallyDial(addr string) (cli *client.Client, err error) {
	var sleep = 1 * time.Second
	for i := 0; i < 5; i++ {
		cli, err := client.Dial(addr)
		if err == nil {
			return cli, nil
		}
		time.Sleep(sleep)
		sleep *= 2
	}
	return nil, fmt.Errorf("after 5 attempts, last error: %s", err)
}
