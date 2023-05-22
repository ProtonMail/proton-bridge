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
	"github.com/stretchr/testify/require"
)

func TestServerManager_NoLoadedUsersNoServers(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			_, err := eventuallyDial(fmt.Sprintf("%v:%v", constants.Host, bridge.GetIMAPPort()))
			require.Error(t, err)
		})
	})
}

func TestServerManager_ServersStartAfterFirstConnectedUser(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			imapWaiter := waitForIMAPServerReady(bridge)
			defer imapWaiter.Done()

			smtpWaiter := waitForSMTPServerReady(bridge)
			defer smtpWaiter.Done()

			_, err := bridge.LoginFull(ctx, username, password, nil, nil)
			require.NoError(t, err)

			imapWaiter.Wait()
			smtpWaiter.Wait()
		})
	})
}

func TestServerManager_ServersStopsAfterUserLogsOut(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			imapWaiter := waitForIMAPServerReady(bridge)
			defer imapWaiter.Done()

			smtpWaiter := waitForSMTPServerReady(bridge)
			defer smtpWaiter.Done()

			userID, err := bridge.LoginFull(ctx, username, password, nil, nil)
			require.NoError(t, err)

			imapWaiter.Wait()
			smtpWaiter.Wait()

			imapWaiterStopped := waitForIMAPServerStopped(bridge)
			defer imapWaiterStopped.Done()

			require.NoError(t, bridge.LogoutUser(ctx, userID))

			imapWaiterStopped.Wait()
		})
	})
}

func TestServerManager_ServersDoNotStopWhenThereIsStillOneActiveUser(t *testing.T) {
	otherPassword := []byte("bar")
	otherUser := "foo"
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		_, _, err := s.CreateUser(otherUser, otherPassword)
		require.NoError(t, err)

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			imapWaiter := waitForIMAPServerReady(bridge)
			defer imapWaiter.Done()

			smtpWaiter := waitForSMTPServerReady(bridge)
			defer smtpWaiter.Done()

			_, err := bridge.LoginFull(ctx, username, password, nil, nil)
			require.NoError(t, err)

			userIDOther, err := bridge.LoginFull(ctx, otherUser, otherPassword, nil, nil)
			require.NoError(t, err)

			imapWaiter.Wait()
			smtpWaiter.Wait()

			evtCh, cancel := bridge.GetEvents(events.UserDeauth{})
			defer cancel()

			require.NoError(t, s.RevokeUser(userIDOther))

			waitForEvent(t, evtCh, events.UserDeauth{})

			imapClient, err := eventuallyDial(fmt.Sprintf("%v:%v", constants.Host, bridge.GetIMAPPort()))
			require.NoError(t, err)
			require.NoError(t, imapClient.Logout())
		})
	})
}

func TestServerManager_ServersStartIfAtLeastOneUserIsLoggedIn(t *testing.T) {
	otherPassword := []byte("bar")
	otherUser := "foo"
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		userIDOther, _, err := s.CreateUser(otherUser, otherPassword)
		require.NoError(t, err)

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			_, err := bridge.LoginFull(ctx, username, password, nil, nil)
			require.NoError(t, err)

			_, err = bridge.LoginFull(ctx, otherUser, otherPassword, nil, nil)
			require.NoError(t, err)
		})

		require.NoError(t, s.RevokeUser(userIDOther))

		withBridgeWaitForServers(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			imapClient, err := eventuallyDial(fmt.Sprintf("%v:%v", constants.Host, bridge.GetIMAPPort()))
			require.NoError(t, err)
			require.NoError(t, imapClient.Logout())
		})
	})
}

func TestServerManager_NetworkLossStopsServers(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			imapWaiter := waitForIMAPServerReady(bridge)
			defer imapWaiter.Done()

			smtpWaiter := waitForSMTPServerReady(bridge)
			defer smtpWaiter.Done()

			imapWaiterStop := waitForIMAPServerStopped(bridge)
			defer imapWaiterStop.Done()

			smtpWaiterStop := waitForSMTPServerStopped(bridge)
			defer smtpWaiterStop.Done()

			_, err := bridge.LoginFull(ctx, username, password, nil, nil)
			require.NoError(t, err)

			imapWaiter.Wait()
			smtpWaiter.Wait()

			netCtl.Disable()

			imapWaiterStop.Wait()
			smtpWaiterStop.Wait()

			netCtl.Enable()

			imapWaiter.Wait()
			smtpWaiter.Wait()
		})
	})
}
