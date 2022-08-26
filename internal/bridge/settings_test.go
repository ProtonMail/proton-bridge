package bridge_test

import (
	"context"
	"os"
	"testing"

	"github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/stretchr/testify/require"
	"gitlab.protontech.ch/go/liteapi/server"
)

func TestBridge_Settings_GluonDir(t *testing.T) {
	withEnv(t, func(s *server.Server, locator bridge.Locator, storeKey []byte) {
		withBridge(t, s.GetHostURL(), locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Create a user.
			_, err := bridge.LoginUser(context.Background(), username, password, nil, nil)
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

func TestBridge_Settings_IMAPPort(t *testing.T) {
	withEnv(t, func(s *server.Server, locator bridge.Locator, storeKey []byte) {
		withBridge(t, s.GetHostURL(), locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// By default, the port is 1143.
			require.Equal(t, 1143, bridge.GetIMAPPort())

			// Set the port to 1144.
			require.NoError(t, bridge.SetIMAPPort(1144))

			// Get the new setting.
			require.Equal(t, 1144, bridge.GetIMAPPort())
		})
	})
}

func TestBridge_Settings_IMAPSSL(t *testing.T) {
	withEnv(t, func(s *server.Server, locator bridge.Locator, storeKey []byte) {
		withBridge(t, s.GetHostURL(), locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// By default, IMAP SSL is disabled.
			require.False(t, bridge.GetIMAPSSL())

			// Enable IMAP SSL.
			require.NoError(t, bridge.SetIMAPSSL(true))

			// Get the new setting.
			require.True(t, bridge.GetIMAPSSL())
		})
	})
}

func TestBridge_Settings_SMTPPort(t *testing.T) {
	withEnv(t, func(s *server.Server, locator bridge.Locator, storeKey []byte) {
		withBridge(t, s.GetHostURL(), locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// By default, the port is 1025.
			require.Equal(t, 1025, bridge.GetSMTPPort())

			// Set the port to 1024.
			require.NoError(t, bridge.SetSMTPPort(1024))

			// Get the new setting.
			require.Equal(t, 1024, bridge.GetSMTPPort())
		})
	})
}

func TestBridge_Settings_SMTPSSL(t *testing.T) {
	withEnv(t, func(s *server.Server, locator bridge.Locator, storeKey []byte) {
		withBridge(t, s.GetHostURL(), locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// By default, SMTP SSL is disabled.
			require.False(t, bridge.GetSMTPSSL())

			// Enable SMTP SSL.
			require.NoError(t, bridge.SetSMTPSSL(true))

			// Get the new setting.
			require.True(t, bridge.GetSMTPSSL())
		})
	})
}

func TestBridge_Settings_Proxy(t *testing.T) {
	withEnv(t, func(s *server.Server, locator bridge.Locator, storeKey []byte) {
		withBridge(t, s.GetHostURL(), locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// By default, proxy is allowed.
			require.True(t, bridge.GetProxyAllowed())

			// Disallow proxy.
			mocks.ProxyDialer.EXPECT().DisallowProxy()
			require.NoError(t, bridge.SetProxyAllowed(false))

			// Get the new setting.
			require.False(t, bridge.GetProxyAllowed())
		})
	})
}

func TestBridge_Settings_Autostart(t *testing.T) {
	withEnv(t, func(s *server.Server, locator bridge.Locator, storeKey []byte) {
		withBridge(t, s.GetHostURL(), locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// By default, autostart is disabled.
			require.False(t, bridge.GetAutostart())

			// Enable autostart.
			mocks.Autostarter.EXPECT().Enable().Return(nil)
			require.NoError(t, bridge.SetAutostart(true))

			// Get the new setting.
			require.True(t, bridge.GetAutostart())
		})
	})
}

func TestBridge_Settings_FirstStart(t *testing.T) {
	withEnv(t, func(s *server.Server, locator bridge.Locator, storeKey []byte) {
		withBridge(t, s.GetHostURL(), locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// By default, first start is true.
			require.True(t, bridge.GetFirstStart())

			// Set first start to false.
			require.NoError(t, bridge.SetFirstStart(false))

			// Get the new setting.
			require.False(t, bridge.GetFirstStart())
		})
	})
}

func TestBridge_Settings_FirstStartGUI(t *testing.T) {
	withEnv(t, func(s *server.Server, locator bridge.Locator, storeKey []byte) {
		withBridge(t, s.GetHostURL(), locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// By default, first start is true.
			require.True(t, bridge.GetFirstStartGUI())

			// Set first start to false.
			require.NoError(t, bridge.SetFirstStartGUI(false))

			// Get the new setting.
			require.False(t, bridge.GetFirstStartGUI())
		})
	})
}
