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
	"os"
	"testing"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/go-proton-api/server"
	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/stretchr/testify/require"
)

func TestBridge_Settings_GluonDir(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// Create a user.
			_, err := bridge.LoginFull(context.Background(), username, password, nil, nil)
			require.NoError(t, err)

			// Create a new location for the Gluon data.
			newGluonDir := t.TempDir()

			// Move the gluon dir; it should also move the user's data.
			require.NoError(t, bridge.SetGluonDir(context.Background(), newGluonDir))

			// Check that the new directory is not empty.
			entries, err := os.ReadDir(newGluonDir)
			require.NoError(t, err)

			// There should be at least one entry.
			require.NotEmpty(t, entries)
		})
	})
}

func TestBridge_Settings_GluonDirWithOnGoingEvents(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		userID, addrID, err := s.CreateUser("imap", password)
		require.NoError(t, err)

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			syncCh, done := chToType[events.Event, events.SyncFinished](bridge.GetEvents(events.SyncFinished{}))
			defer done()

			_, err := bridge.LoginFull(context.Background(), "imap", password, nil, nil)
			require.NoError(t, err)

			<-syncCh
		})

		labelID, err := s.CreateLabel(userID, "folder", "", proton.LabelTypeFolder)
		require.NoError(t, err)

		withClient(ctx, t, s, "imap", password, func(ctx context.Context, c *proton.Client) {
			createNumMessages(ctx, t, c, addrID, labelID, 200)
		})

		withBridgeWaitForServers(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// Create a new location for the Gluon data.
			newGluonDir := t.TempDir()

			// Move the gluon dir; it should also move the user's data.
			require.NoError(t, bridge.SetGluonDir(context.Background(), newGluonDir))

			// Check that the new directory is not empty.
			entries, err := os.ReadDir(newGluonDir)
			require.NoError(t, err)

			// There should be at least one entry.
			require.NotEmpty(t, entries)
		})
	})
}

func TestBridge_Settings_IMAPPort(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			curPort := bridge.GetIMAPPort()

			// Set the port to 1144.
			require.NoError(t, bridge.SetIMAPPort(ctx, 1144))

			// Get the new setting.
			require.Equal(t, 1144, bridge.GetIMAPPort())

			// Assert that it has changed.
			require.NotEqual(t, curPort, bridge.GetIMAPPort())
		})
	})
}

func TestBridge_Settings_IMAPSSL(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// By default, IMAP SSL is disabled.
			require.False(t, bridge.GetIMAPSSL())

			// Enable IMAP SSL.
			require.NoError(t, bridge.SetIMAPSSL(ctx, true))

			// Get the new setting.
			require.True(t, bridge.GetIMAPSSL())
		})
	})
}

func TestBridge_Settings_SMTPPort(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			curPort := bridge.GetSMTPPort()

			// Set the port to 1024.
			require.NoError(t, bridge.SetSMTPPort(ctx, 1024))

			// Get the new setting.
			require.Equal(t, 1024, bridge.GetSMTPPort())

			// Assert that it has changed.
			require.NotEqual(t, curPort, bridge.GetSMTPPort())
		})
	})
}

func TestBridge_Settings_SMTPSSL(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// By default, SMTP SSL is disabled.
			require.False(t, bridge.GetSMTPSSL())

			// Enable SMTP SSL.
			require.NoError(t, bridge.SetSMTPSSL(ctx, true))

			// Get the new setting.
			require.True(t, bridge.GetSMTPSSL())
		})
	})
}

func TestBridge_Settings_Proxy(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// By default, proxy is allowed.
			require.False(t, bridge.GetProxyAllowed())

			// Disallow proxy.
			mocks.ProxyCtl.EXPECT().AllowProxy()
			require.NoError(t, bridge.SetProxyAllowed(true))

			// Get the new setting.
			require.True(t, bridge.GetProxyAllowed())
		})
	})
}

func TestBridge_Settings_Autostart(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// By default, autostart is enabled.
			require.True(t, bridge.GetAutostart())

			// Disable autostart.
			mocks.Autostarter.EXPECT().IsEnabled().Return(true)
			mocks.Autostarter.EXPECT().Disable().Return(nil)
			require.NoError(t, bridge.SetAutostart(false))

			// Get the new setting.
			require.False(t, bridge.GetAutostart())

			// Re Enable autostart.
			mocks.Autostarter.EXPECT().IsEnabled().Return(false)
			mocks.Autostarter.EXPECT().Enable().Return(nil)
			require.NoError(t, bridge.SetAutostart(true))

			// Get the new setting.
			require.True(t, bridge.GetAutostart())
		})
	})
}

func TestBridge_Settings_FirstStart(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// By default, first start is true.
			require.True(t, bridge.GetFirstStart())

			// the setting of the first start value is managed by bridge itself, so the setter is not exported.
		})
	})
}
