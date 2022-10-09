package bridge_test

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v2/internal/certs"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/focus"
	"github.com/ProtonMail/proton-bridge/v2/internal/locations"
	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
	"github.com/ProtonMail/proton-bridge/v2/internal/user"
	"github.com/ProtonMail/proton-bridge/v2/internal/useragent"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
	"github.com/ProtonMail/proton-bridge/v2/tests"
	"github.com/bradenaw/juniper/xslices"
	"github.com/stretchr/testify/require"
	"gitlab.protontech.ch/go/liteapi"
	"gitlab.protontech.ch/go/liteapi/server"
	"gitlab.protontech.ch/go/liteapi/server/backend"
)

var (
	username = "username"
	password = []byte("password")

	v2_3_0 = semver.MustParse("2.3.0")
	v2_4_0 = semver.MustParse("2.4.0")
)

func init() {
	user.EventPeriod = 100 * time.Millisecond
	user.EventJitter = 0
	backend.GenerateKey = tests.FastGenerateKey
	certs.GenerateCert = tests.FastGenerateCert
}

func TestBridge_ConnStatus(t *testing.T) {
	withTLSEnv(t, func(ctx context.Context, s *server.Server, netCtl *liteapi.NetCtl, locator bridge.Locator, vaultKey []byte) {
		withBridge(t, ctx, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Get a stream of connection status events.
			eventCh, done := bridge.GetEvents(events.ConnStatusUp{}, events.ConnStatusDown{})
			defer done()

			// Simulate network disconnect.
			netCtl.Disable()

			// Trigger some operation that will fail due to the network disconnect.
			_, err := bridge.LoginUser(context.Background(), username, password, nil, nil)
			require.Error(t, err)

			// Wait for the event.
			require.Equal(t, events.ConnStatusDown{}, <-eventCh)

			// Simulate network reconnect.
			netCtl.Enable()

			// Trigger some operation that will succeed due to the network reconnect.
			userID, err := bridge.LoginUser(context.Background(), username, password, nil, nil)
			require.NoError(t, err)
			require.NotEmpty(t, userID)

			// Wait for the event.
			require.Equal(t, events.ConnStatusUp{}, <-eventCh)
		})
	})
}

func TestBridge_TLSIssue(t *testing.T) {
	withTLSEnv(t, func(ctx context.Context, s *server.Server, netCtl *liteapi.NetCtl, locator bridge.Locator, vaultKey []byte) {
		withBridge(t, ctx, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
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
	withTLSEnv(t, func(ctx context.Context, s *server.Server, netCtl *liteapi.NetCtl, locator bridge.Locator, vaultKey []byte) {
		withBridge(t, ctx, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
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
	withTLSEnv(t, func(ctx context.Context, s *server.Server, netCtl *liteapi.NetCtl, locator bridge.Locator, vaultKey []byte) {
		var calls []server.Call

		s.AddCallWatcher(func(call server.Call) {
			calls = append(calls, call)
		})

		withBridge(t, ctx, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Set the platform to something other than the default.
			bridge.SetCurrentPlatform("platform")

			// Assert that the user agent then contains the platform.
			require.Contains(t, bridge.GetCurrentUserAgent(), "platform")

			// Login the user.
			_, err := bridge.LoginUser(context.Background(), username, password, nil, nil)
			require.NoError(t, err)

			// Assert that the user agent was sent to the API.
			require.Contains(t, calls[len(calls)-1].Header.Get("User-Agent"), bridge.GetCurrentUserAgent())
		})
	})
}

func TestBridge_Cookies(t *testing.T) {
	withTLSEnv(t, func(ctx context.Context, s *server.Server, netCtl *liteapi.NetCtl, locator bridge.Locator, vaultKey []byte) {
		var calls []server.Call

		s.AddCallWatcher(func(call server.Call) {
			calls = append(calls, call)
		})

		var sessionID string

		// Start bridge and add a user so that API assigns us a session ID via cookie.
		withBridge(t, ctx, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			_, err := bridge.LoginUser(context.Background(), username, password, nil, nil)
			require.NoError(t, err)

			cookie, err := (&http.Request{Header: calls[len(calls)-1].Header}).Cookie("Session-Id")
			require.NoError(t, err)

			sessionID = cookie.Value
		})

		// Start bridge again and check that it uses the same session ID.
		withBridge(t, ctx, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			cookie, err := (&http.Request{Header: calls[len(calls)-1].Header}).Cookie("Session-Id")
			require.NoError(t, err)

			require.Equal(t, sessionID, cookie.Value)
		})
	})
}

func TestBridge_CheckUpdate(t *testing.T) {
	withTLSEnv(t, func(ctx context.Context, s *server.Server, netCtl *liteapi.NetCtl, locator bridge.Locator, vaultKey []byte) {
		withBridge(t, ctx, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Disable autoupdate for this test.
			require.NoError(t, bridge.SetAutoUpdate(false))

			// Get a stream of update not available events.
			noUpdateCh, done := bridge.GetEvents(events.UpdateNotAvailable{})
			defer done()

			// We are currently on the latest version.
			bridge.CheckForUpdates()

			// we should receive an event indicating that no update is available.
			require.Equal(t, events.UpdateNotAvailable{}, <-noUpdateCh)

			// Simulate a new version being available.
			mocks.Updater.SetLatestVersion(v2_4_0, v2_3_0)

			// Get a stream of update available events.
			updateCh, done := bridge.GetEvents(events.UpdateAvailable{})
			defer done()

			// Check for updates.
			bridge.CheckForUpdates()

			// We should receive an event indicating that an update is available.
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
	withTLSEnv(t, func(ctx context.Context, s *server.Server, netCtl *liteapi.NetCtl, locator bridge.Locator, vaultKey []byte) {
		withBridge(t, ctx, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Enable autoupdate for this test.
			require.NoError(t, bridge.SetAutoUpdate(true))

			// Get a stream of update events.
			updateCh, done := bridge.GetEvents(events.UpdateInstalled{})
			defer done()

			// Simulate a new version being available.
			mocks.Updater.SetLatestVersion(v2_4_0, v2_3_0)

			// Check for updates.
			bridge.CheckForUpdates()

			// We should receive an event indicating that the update was installed.
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
	withTLSEnv(t, func(ctx context.Context, s *server.Server, netCtl *liteapi.NetCtl, locator bridge.Locator, vaultKey []byte) {
		withBridge(t, ctx, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Disable autoupdate for this test.
			require.NoError(t, bridge.SetAutoUpdate(false))

			// Get a stream of update available events.
			updateCh, done := bridge.GetEvents(events.UpdateAvailable{})
			defer done()

			// Simulate a new version being available, but it's too new for us.
			mocks.Updater.SetLatestVersion(v2_4_0, v2_4_0)

			// Check for updates.
			bridge.CheckForUpdates()

			// We should receive an event indicating an update is available, but we can't install it.
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
	withTLSEnv(t, func(ctx context.Context, s *server.Server, netCtl *liteapi.NetCtl, locator bridge.Locator, vaultKey []byte) {
		withBridge(t, ctx, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
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
	withTLSEnv(t, func(ctx context.Context, s *server.Server, netCtl *liteapi.NetCtl, locator bridge.Locator, vaultKey []byte) {
		var userID string

		// Login a user.
		withBridge(t, ctx, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			newUserID, err := bridge.LoginUser(context.Background(), username, password, nil, nil)
			require.NoError(t, err)

			userID = newUserID
		})

		// Start bridge with the correct vault key -- it should load the users correctly.
		withBridge(t, ctx, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			require.ElementsMatch(t, []string{userID}, bridge.GetUserIDs())
		})

		// Start bridge with a bad vault key, the vault will be wiped and bridge will show no users.
		withBridge(t, ctx, s.GetHostURL(), netCtl, locator, []byte("bad"), func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			require.Empty(t, bridge.GetUserIDs())
		})

		// Start bridge with a nil vault key, the vault will be wiped and bridge will show no users.
		withBridge(t, ctx, s.GetHostURL(), netCtl, locator, nil, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			require.Empty(t, bridge.GetUserIDs())
		})
	})
}

func TestBridge_MissingGluonDir(t *testing.T) {
	withTLSEnv(t, func(ctx context.Context, s *server.Server, netCtl *liteapi.NetCtl, locator bridge.Locator, vaultKey []byte) {
		var gluonDir string

		withBridge(t, ctx, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			_, err := bridge.LoginUser(context.Background(), username, password, nil, nil)
			require.NoError(t, err)

			// Move the gluon dir.
			bridge.SetGluonDir(ctx, t.TempDir())

			// Get the gluon dir.
			gluonDir = bridge.GetGluonDir()
		})

		// The user removes the gluon dir while bridge is not running.
		require.NoError(t, os.RemoveAll(gluonDir))

		// Bridge starts but can't find the gluon dir; there should be no error.
		withBridge(t, ctx, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// ...
		})
	})
}

// withTLSEnv creates the full test environment and runs the tests.
func withTLSEnv(t *testing.T, tests func(context.Context, *server.Server, *liteapi.NetCtl, bridge.Locator, []byte)) {
	// Create test API.
	server := server.NewTLS()
	defer server.Close()

	// Add test user.
	_, _, err := server.CreateUser(username, username+"@pm.me", password)
	require.NoError(t, err)

	// Generate a random vault key.
	vaultKey, err := crypto.RandomToken(32)
	require.NoError(t, err)

	// Create a context used for the test.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a net controller so we can simulate network connectivity issues.
	netCtl := liteapi.NewNetCtl()

	// Create a locations object to provide temporary locations for bridge data during the test.
	locations := locations.New(bridge.NewTestLocationsProvider(t.TempDir()), "config-name")

	// Run the tests.
	tests(ctx, server, netCtl, locations, vaultKey)
}

// withEnv creates the full test environment and runs the tests.
func withEnv(t *testing.T, server *server.Server, tests func(context.Context, *liteapi.NetCtl, bridge.Locator, []byte)) {
	// Add test user.
	_, _, err := server.CreateUser(username, username+"@pm.me", password)
	require.NoError(t, err)

	// Generate a random vault key.
	vaultKey, err := crypto.RandomToken(32)
	require.NoError(t, err)

	// Create a context used for the test.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a net controller so we can simulate network connectivity issues.
	netCtl := liteapi.NewNetCtl()

	// Create a locations object to provide temporary locations for bridge data during the test.
	locations := locations.New(bridge.NewTestLocationsProvider(t.TempDir()), "config-name")

	// Run the tests.
	tests(ctx, netCtl, locations, vaultKey)
}

// withBridge creates a new bridge which points to the given API URL and uses the given keychain, and closes it when done.
func withBridge(
	t *testing.T,
	ctx context.Context,
	apiURL string,
	netCtl *liteapi.NetCtl,
	locator bridge.Locator,
	vaultKey []byte,
	tests func(*bridge.Bridge, *bridge.Mocks),
) {
	// Create the mock objects used in the tests.
	mocks := bridge.NewMocks(t, v2_3_0, v2_3_0)
	defer mocks.Close()

	// Bridge will enable the proxy by default at startup.
	mocks.ProxyCtl.EXPECT().AllowProxy()

	// Get the path to the vault.
	vaultDir, err := locator.ProvideSettingsPath()
	require.NoError(t, err)

	// Create the vault.
	vault, _, err := vault.New(vaultDir, t.TempDir(), vaultKey)
	require.NoError(t, err)

	// Let the IMAP and SMTP servers choose random available ports for this test.
	require.NoError(t, vault.SetIMAPPort(0))
	require.NoError(t, vault.SetSMTPPort(0))

	// Create a new bridge.
	bridge, err := bridge.New(
		apiURL,
		locator,
		vault,
		useragent.New(),
		mocks.TLSReporter,
		liteapi.NewDialer(netCtl, &tls.Config{InsecureSkipVerify: true}).GetRoundTripper(),
		mocks.ProxyCtl,
		mocks.Autostarter,
		mocks.Updater,
		v2_3_0,
	)
	require.NoError(t, err)

	// Close the bridge when done.
	defer bridge.Close(ctx)

	// Use the bridge.
	tests(bridge, mocks)
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
