package bridge_test

import (
	"context"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/focus"
	"github.com/ProtonMail/proton-bridge/v2/internal/locations"
	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
	"github.com/ProtonMail/proton-bridge/v2/internal/useragent"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
	"github.com/bradenaw/juniper/xslices"
	"github.com/stretchr/testify/require"
	"gitlab.protontech.ch/go/liteapi"
	"gitlab.protontech.ch/go/liteapi/server"
)

const (
	username = "username"
	password = "password"
)

var (
	v2_3_0 = semver.MustParse("2.3.0")
	v2_4_0 = semver.MustParse("2.4.0")
)

func TestBridge_ConnStatus(t *testing.T) {
	withEnv(t, func(s *server.Server, locator bridge.Locator, vaultKey []byte) {
		withBridge(t, s.GetHostURL(), locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Get a stream of connection status events.
			eventCh, done := bridge.GetEvents(events.ConnStatus{})
			defer done()

			// Simulate network disconnect.
			mocks.TLSDialer.SetCanDial(false)

			// Trigger some operation that will fail due to the network disconnect.
			_, err := bridge.LoginUser(context.Background(), username, password, nil, nil)
			require.Error(t, err)

			// Wait for the event.
			require.Equal(t, events.ConnStatus{Status: liteapi.StatusDown}, <-eventCh)

			// Simulate network reconnect.
			mocks.TLSDialer.SetCanDial(true)

			// Trigger some operation that will succeed due to the network reconnect.
			userID, err := bridge.LoginUser(context.Background(), username, password, nil, nil)
			require.NoError(t, err)
			require.NotEmpty(t, userID)

			// Wait for the event.
			require.Equal(t, events.ConnStatus{Status: liteapi.StatusUp}, <-eventCh)
		})
	})
}

func TestBridge_TLSIssue(t *testing.T) {
	withEnv(t, func(s *server.Server, locator bridge.Locator, vaultKey []byte) {
		withBridge(t, s.GetHostURL(), locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Get a stream of TLS issue events.
			tlsEventCh, done := bridge.GetEvents(events.TLSIssue{})
			defer done()

			// Simulate a TLS issue.
			go func() {
				mocks.TLSIssueCh <- struct{}{}
			}()

			// Wait for the event.
			require.IsType(t, events.TLSIssue{}, <-tlsEventCh)
		})
	})
}

func TestBridge_Focus(t *testing.T) {
	withEnv(t, func(s *server.Server, locator bridge.Locator, vaultKey []byte) {
		withBridge(t, s.GetHostURL(), locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Get a stream of TLS issue events.
			raiseCh, done := bridge.GetEvents(events.Raise{})
			defer done()

			// Simulate a focus event.
			focus.TryRaise()

			// Wait for the event.
			require.IsType(t, events.Raise{}, <-raiseCh)
		})
	})
}

func TestBridge_UserAgent(t *testing.T) {
	withEnv(t, func(s *server.Server, locator bridge.Locator, vaultKey []byte) {
		var calls []server.Call

		s.AddCallWatcher(func(call server.Call) {
			calls = append(calls, call)
		})

		withBridge(t, s.GetHostURL(), locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Set the platform to something other than the default.
			bridge.SetCurrentPlatform("platform")

			// Assert that the user agent then contains the platform.
			require.Contains(t, bridge.GetCurrentUserAgent(), "platform")

			// Login the user.
			_, err := bridge.LoginUser(context.Background(), username, password, nil, nil)
			require.NoError(t, err)

			// Assert that the user agent was sent to the API.
			require.Contains(t, calls[len(calls)-1].Request.Header.Get("User-Agent"), bridge.GetCurrentUserAgent())
		})
	})
}

func TestBridge_Cookies(t *testing.T) {
	withEnv(t, func(s *server.Server, locator bridge.Locator, vaultKey []byte) {
		var calls []server.Call

		s.AddCallWatcher(func(call server.Call) {
			calls = append(calls, call)
		})

		var sessionID string

		// Start bridge and add a user so that API assigns us a session ID via cookie.
		withBridge(t, s.GetHostURL(), locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			_, err := bridge.LoginUser(context.Background(), username, password, nil, nil)
			require.NoError(t, err)

			cookie, err := calls[len(calls)-1].Request.Cookie("Session-Id")
			require.NoError(t, err)

			sessionID = cookie.Value
		})

		// Start bridge again and check that it uses the same session ID.
		withBridge(t, s.GetHostURL(), locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			cookie, err := calls[len(calls)-1].Request.Cookie("Session-Id")
			require.NoError(t, err)

			require.Equal(t, sessionID, cookie.Value)
		})
	})
}

func TestBridge_CheckUpdate(t *testing.T) {
	withEnv(t, func(s *server.Server, locator bridge.Locator, vaultKey []byte) {
		withBridge(t, s.GetHostURL(), locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Disable autoupdate for this test.
			require.NoError(t, bridge.SetAutoUpdate(false))

			// Get a stream of update events.
			updateCh, done := bridge.GetEvents(events.UpdateNotAvailable{}, events.UpdateAvailable{})
			defer done()

			// We are currently on the latest version.
			bridge.CheckForUpdates()
			require.Equal(t, events.UpdateNotAvailable{}, <-updateCh)

			// Simulate a new version being available.
			mocks.Updater.SetLatestVersion(v2_4_0, v2_3_0)

			// Check for updates.
			bridge.CheckForUpdates()
			require.Equal(t, events.UpdateAvailable{
				Version: updater.VersionInfo{
					Version:           v2_4_0,
					MinAuto:           v2_3_0,
					RolloutProportion: 1.0,
				},
				CanInstall: true,
			}, <-updateCh)
		})
	})
}

func TestBridge_AutoUpdate(t *testing.T) {
	withEnv(t, func(s *server.Server, locator bridge.Locator, vaultKey []byte) {
		withBridge(t, s.GetHostURL(), locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Enable autoupdate for this test.
			require.NoError(t, bridge.SetAutoUpdate(true))

			// Get a stream of update events.
			updateCh, done := bridge.GetEvents(events.UpdateNotAvailable{}, events.UpdateInstalled{})
			defer done()

			// Simulate a new version being available.
			mocks.Updater.SetLatestVersion(v2_4_0, v2_3_0)

			// Check for updates.
			bridge.CheckForUpdates()
			require.Equal(t, events.UpdateInstalled{
				Version: updater.VersionInfo{
					Version:           v2_4_0,
					MinAuto:           v2_3_0,
					RolloutProportion: 1.0,
				},
			}, <-updateCh)
		})
	})
}

func TestBridge_ManualUpdate(t *testing.T) {
	withEnv(t, func(s *server.Server, locator bridge.Locator, vaultKey []byte) {
		withBridge(t, s.GetHostURL(), locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Disable autoupdate for this test.
			require.NoError(t, bridge.SetAutoUpdate(false))

			// Get a stream of update events.
			updateCh, done := bridge.GetEvents(events.UpdateNotAvailable{}, events.UpdateAvailable{})
			defer done()

			// Simulate a new version being available, but it's too new for us.
			mocks.Updater.SetLatestVersion(v2_4_0, v2_4_0)

			// Check for updates.
			bridge.CheckForUpdates()
			require.Equal(t, events.UpdateAvailable{
				Version: updater.VersionInfo{
					Version:           v2_4_0,
					MinAuto:           v2_4_0,
					RolloutProportion: 1.0,
				},
				CanInstall: false,
			}, <-updateCh)
		})
	})
}

func TestBridge_ForceUpdate(t *testing.T) {
	withEnv(t, func(s *server.Server, locator bridge.Locator, vaultKey []byte) {
		withBridge(t, s.GetHostURL(), locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Get a stream of update events.
			updateCh, done := bridge.GetEvents(events.UpdateForced{})
			defer done()

			// Set the minimum accepted app version to something newer than the current version.
			s.SetMinAppVersion(v2_4_0)

			// Try to login the user. It will fail because the bridge is too old.
			_, err := bridge.LoginUser(context.Background(), username, password, nil, nil)
			require.Error(t, err)

			// We should get an update required event.
			require.Equal(t, events.UpdateForced{}, <-updateCh)
		})
	})
}

func TestBridge_BadVaultKey(t *testing.T) {
	withEnv(t, func(s *server.Server, locator bridge.Locator, vaultKey []byte) {
		var userID string

		// Login a user.
		withBridge(t, s.GetHostURL(), locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			newUserID, err := bridge.LoginUser(context.Background(), username, password, nil, nil)
			require.NoError(t, err)

			userID = newUserID
		})

		// Start bridge with the correct vault key -- it should load the users correctly.
		withBridge(t, s.GetHostURL(), locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			require.ElementsMatch(t, []string{userID}, bridge.GetUserIDs())
		})

		// Start bridge with a bad vault key, the vault will be wiped and bridge will show no users.
		withBridge(t, s.GetHostURL(), locator, []byte("bad"), func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			require.Empty(t, bridge.GetUserIDs())
		})

		// Start bridge with a nil vault key, the vault will be wiped and bridge will show no users.
		withBridge(t, s.GetHostURL(), locator, nil, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			require.Empty(t, bridge.GetUserIDs())
		})
	})
}

// withEnv creates the full test environment and runs the tests.
func withEnv(t *testing.T, tests func(server *server.Server, locator bridge.Locator, vaultKey []byte)) {
	// Create test API.
	server := server.NewTLS()
	defer server.Close()

	// Add test user.
	_, _, err := server.AddUser(username, password, username+"@pm.me")
	require.NoError(t, err)

	// Generate a random vault key.
	vaultKey, err := crypto.RandomToken(32)
	require.NoError(t, err)

	// Run the tests.
	tests(server, locations.New(bridge.NewTestLocationsProvider(t), "config-name"), vaultKey)
}

// withBridge creates a new bridge which points to the given API URL and uses the given keychain, and closes it when done.
func withBridge(t *testing.T, apiURL string, locator bridge.Locator, vaultKey []byte, tests func(bridge *bridge.Bridge, mocks *bridge.Mocks)) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create the mock objects used in the tests.
	mocks := bridge.NewMocks(t, v2_3_0, v2_3_0)

	// Bridge will enable the proxy by default at startup.
	mocks.ProxyDialer.EXPECT().AllowProxy()

	// Get the path to the vault.
	vaultDir, err := locator.ProvideSettingsPath()
	require.NoError(t, err)

	// Create the vault.
	vault, _, err := vault.New(vaultDir, t.TempDir(), vaultKey)
	require.NoError(t, err)

	// Create a new bridge.
	bridge, err := bridge.New(
		apiURL,
		locator,
		vault,
		useragent.New(),
		mocks.TLSReporter,
		mocks.ProxyDialer,
		mocks.Autostarter,
		mocks.Updater,
		v2_3_0,
	)
	require.NoError(t, err)

	// Use the bridge.
	tests(bridge, mocks)

	// Close the bridge.
	require.NoError(t, bridge.Close(ctx))
}

// must is a helper function that panics on error.
func must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}

	return val
}

func getConnectedUserIDs(t *testing.T, bridge *bridge.Bridge) []string {
	t.Helper()

	return xslices.Filter(bridge.GetUserIDs(), func(userID string) bool {
		info, err := bridge.GetUserInfo(userID)
		require.NoError(t, err)

		return info.Connected
	})
}
