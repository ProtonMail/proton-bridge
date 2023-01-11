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
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/go-proton-api/server"
	"github.com/ProtonMail/go-proton-api/server/backend"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v3/internal/certs"
	"github.com/ProtonMail/proton-bridge/v3/internal/cookies"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/focus"
	"github.com/ProtonMail/proton-bridge/v3/internal/locations"
	"github.com/ProtonMail/proton-bridge/v3/internal/updater"
	"github.com/ProtonMail/proton-bridge/v3/internal/user"
	"github.com/ProtonMail/proton-bridge/v3/internal/useragent"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/ProtonMail/proton-bridge/v3/tests"
	"github.com/bradenaw/juniper/xslices"
	"github.com/stretchr/testify/require"
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
	backend.GenerateKey = backend.FastGenerateKey
	certs.GenerateCert = tests.FastGenerateCert
}

func TestBridge_ConnStatus(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, vaultKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Get a stream of connection status events.
			eventCh, done := bridge.GetEvents(events.ConnStatusUp{}, events.ConnStatusDown{})
			defer done()

			// Simulate network disconnect.
			netCtl.Disable()

			// Trigger some operation that will fail due to the network disconnect.
			_, err := bridge.LoginFull(context.Background(), username, password, nil, nil)
			require.Error(t, err)

			// Wait for the event.
			require.Equal(t, events.ConnStatusDown{}, <-eventCh)

			// Simulate network reconnect.
			netCtl.Enable()

			// Trigger some operation that will succeed due to the network reconnect.
			userID, err := bridge.LoginFull(context.Background(), username, password, nil, nil)
			require.NoError(t, err)
			require.NotEmpty(t, userID)

			// Wait for the event.
			require.Equal(t, events.ConnStatusUp{}, <-eventCh)
		})
	})
}

func TestBridge_TLSIssue(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, vaultKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
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
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, vaultKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
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
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, vaultKey []byte) {
		var (
			calls []server.Call
			lock  sync.Mutex
		)

		s.AddCallWatcher(func(call server.Call) {
			lock.Lock()
			defer lock.Unlock()

			calls = append(calls, call)
		})

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Set the platform to something other than the default.
			bridge.SetCurrentPlatform("platform")

			// Assert that the user agent then contains the platform.
			require.Contains(t, bridge.GetCurrentUserAgent(), "platform")

			// Login the user.
			_, err := bridge.LoginFull(context.Background(), username, password, nil, nil)
			require.NoError(t, err)

			lock.Lock()
			defer lock.Unlock()

			// Assert that the user agent was sent to the API.
			require.Contains(t, calls[len(calls)-1].RequestHeader.Get("User-Agent"), bridge.GetCurrentUserAgent())
		})
	})
}

func TestBridge_Cookies(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, vaultKey []byte) {
		var (
			sessionIDs     []string
			sessionIDsLock sync.RWMutex
		)

		// Save any session IDs we use.
		s.AddCallWatcher(func(call server.Call) {
			cookie, err := (&http.Request{Header: call.RequestHeader}).Cookie("Session-Id")
			if err != nil {
				return
			}

			sessionIDsLock.Lock()
			defer sessionIDsLock.Unlock()

			sessionIDs = append(sessionIDs, cookie.Value)
		})

		// Start bridge and add a user so that API assigns us a session ID via cookie.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			_, err := bridge.LoginFull(context.Background(), username, password, nil, nil)
			require.NoError(t, err)
		})

		// Start bridge again and check that it uses the same session ID.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// ...
		})

		// We should have used just one session ID.
		sessionIDsLock.Lock()
		defer sessionIDsLock.Unlock()

		require.Len(t, xslices.Unique(sessionIDs), 1)
	})
}

func TestBridge_CheckUpdate(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, vaultKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
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
				Silent:     false,
				Compatible: true,
			}, <-updateCh)
		})
	})
}

func TestBridge_AutoUpdate(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, vaultKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Enable autoupdate for this test.
			require.NoError(t, bridge.SetAutoUpdate(true))

			// Get a stream of update events.
			updateCh, done := bridge.GetEvents(events.UpdateInstalled{})
			defer done()

			// Simulate a new version being available.
			mocks.Updater.SetLatestVersion(v2_4_0, v2_3_0)

			// Check for updates.
			bridge.CheckForUpdates()

			// We should receive an event indicating that the update was silently installed.
			require.Equal(t, events.UpdateInstalled{
				Version: updater.VersionInfo{
					Version:           v2_4_0,
					MinAuto:           v2_3_0,
					RolloutProportion: 1.0,
				},
				Silent: true,
			}, <-updateCh)
		})
	})
}

func TestBridge_ManualUpdate(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, vaultKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
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
				Silent:     false,
				Compatible: false,
			}, <-updateCh)
		})
	})
}

func TestBridge_ForceUpdate(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, vaultKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Get a stream of update events.
			updateCh, done := bridge.GetEvents(events.UpdateForced{})
			defer done()

			// Set the minimum accepted app version to something newer than the current version.
			s.SetMinAppVersion(v2_4_0)

			// Try to login the user. It will fail because the bridge is too old.
			_, err := bridge.LoginFull(context.Background(), username, password, nil, nil)
			require.Error(t, err)

			// We should get an update required event.
			require.Equal(t, events.UpdateForced{}, <-updateCh)
		})
	})
}

func TestBridge_BadVaultKey(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, vaultKey []byte) {
		var userID string

		// Login a user.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			newUserID, err := bridge.LoginFull(context.Background(), username, password, nil, nil)
			require.NoError(t, err)

			userID = newUserID
		})

		// Start bridge with the correct vault key -- it should load the users correctly.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			require.ElementsMatch(t, []string{userID}, bridge.GetUserIDs())
		})

		// Start bridge with a bad vault key, the vault will be wiped and bridge will show no users.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, []byte("bad"), func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			require.Empty(t, bridge.GetUserIDs())
		})

		// Start bridge with a nil vault key, the vault will be wiped and bridge will show no users.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, nil, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			require.Empty(t, bridge.GetUserIDs())
		})
	})
}

func TestBridge_MissingGluonDir(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, vaultKey []byte) {
		var gluonDir string

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			_, err := bridge.LoginFull(context.Background(), username, password, nil, nil)
			require.NoError(t, err)

			// Move the gluon dir.
			require.NoError(t, bridge.SetGluonDir(ctx, t.TempDir()))

			// Get the gluon dir.
			gluonDir = bridge.GetGluonDir()
		})

		// The user removes the gluon dir while bridge is not running.
		require.NoError(t, os.RemoveAll(gluonDir))

		// Bridge starts but can't find the gluon dir; there should be no error.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// ...
		})
	})
}

func TestBridge_AddressWithoutKeys(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, vaultKey []byte) {
		m := proton.New(
			proton.WithHostURL(s.GetHostURL()),
			proton.WithTransport(proton.InsecureTransport()),
		)
		defer m.Close()

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Create a user which will have an address without keys.
			userID, _, err := s.CreateUser("nokeys", []byte("password"))
			require.NoError(t, err)

			// Create an additional address for the user; it will not have keys.
			aliasAddrID, err := s.CreateAddress(userID, "alias@pm.me", []byte("password"))
			require.NoError(t, err)

			// Create an API client so we can remove the address keys.
			c, _, err := m.NewClientWithLogin(ctx, "nokeys", []byte("password"))
			require.NoError(t, err)
			defer c.Close()

			// Get the alias address.
			aliasAddr, err := c.GetAddress(ctx, aliasAddrID)
			require.NoError(t, err)

			// Remove the address keys.
			require.NoError(t, s.RemoveAddressKey(userID, aliasAddrID, aliasAddr.Keys[0].ID))

			// Watch for sync finished event.
			syncCh, done := chToType[events.Event, events.SyncFinished](bridge.GetEvents(events.SyncFinished{}))
			defer done()

			// We should be able to log the user in.
			require.NoError(t, getErr(bridge.LoginFull(context.Background(), "nokeys", []byte("password"), nil, nil)))
			require.NoError(t, err)

			// The sync should eventually finish for the user without keys.
			require.Equal(t, userID, (<-syncCh).UserID)
		})
	})
}

func TestBridge_FactoryReset(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, vaultKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// The settings should be their default values.
			require.True(t, bridge.GetAutoUpdate())
			require.Equal(t, updater.StableChannel, bridge.GetUpdateChannel())

			// Login the user.
			userID, err := bridge.LoginFull(ctx, username, password, nil, nil)
			require.NoError(t, err)

			// Change some settings.
			require.NoError(t, bridge.SetAutoUpdate(false))
			require.NoError(t, bridge.SetUpdateChannel(updater.EarlyChannel))

			// The user is now connected.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Equal(t, []string{userID}, getConnectedUserIDs(t, bridge))

			// The settings should be changed.
			require.False(t, bridge.GetAutoUpdate())
			require.Equal(t, updater.EarlyChannel, bridge.GetUpdateChannel())

			// Perform a factory reset.
			bridge.FactoryReset(ctx)

			// The user is gone.
			require.Equal(t, []string{}, bridge.GetUserIDs())
			require.Equal(t, []string{}, getConnectedUserIDs(t, bridge))
		})

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// The settings should be reset.
			require.True(t, bridge.GetAutoUpdate())
			require.Equal(t, updater.StableChannel, bridge.GetUpdateChannel())
		})
	})
}

func TestBridge_ChangeCacheDirectoryFailsBetweenDifferentVolumes(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Test only necessary on windows")
	}
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, vaultKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Change directory
			err := bridge.SetGluonDir(ctx, "XX:\\")
			require.Error(t, err)
		})
	})
}

func TestBridge_ChangeCacheDirectory(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, vaultKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			newCacheDir := t.TempDir()
			currentCacheDir := bridge.GetGluonDir()

			// Login the user.
			userID, err := bridge.LoginFull(ctx, username, password, nil, nil)
			require.NoError(t, err)

			// The user is now connected.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Equal(t, []string{userID}, getConnectedUserIDs(t, bridge))

			// Change directory
			err = bridge.SetGluonDir(ctx, newCacheDir)
			require.NoError(t, err)

			_, err = os.ReadDir(currentCacheDir)
			require.True(t, os.IsNotExist(err))

			require.Equal(t, newCacheDir, bridge.GetGluonDir())
		})
	})
}

// withEnv creates the full test environment and runs the tests.
func withEnv(t *testing.T, tests func(context.Context, *server.Server, *proton.NetCtl, bridge.Locator, []byte), opts ...server.Option) {
	server := server.New(opts...)
	defer server.Close()

	// Add test user.
	_, _, err := server.CreateUser(username, password)
	require.NoError(t, err)

	// Generate a random vault key.
	vaultKey, err := crypto.RandomToken(32)
	require.NoError(t, err)

	// Create a context used for the test.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a net controller so we can simulate network connectivity issues.
	netCtl := proton.NewNetCtl()

	// Create a locations object to provide temporary locations for bridge data during the test.
	locations := locations.New(bridge.NewTestLocationsProvider(t.TempDir()), "config-name")

	// Run the tests.
	tests(ctx, server, netCtl, locations, vaultKey)
}

// withBridge creates a new bridge which points to the given API URL and uses the given keychain, and closes it when done.
func withBridge(
	ctx context.Context,
	t *testing.T,
	apiURL string,
	netCtl *proton.NetCtl,
	locator bridge.Locator,
	vaultKey []byte,
	tests func(*bridge.Bridge, *bridge.Mocks),
) {
	// Create the mock objects used in the tests.
	mocks := bridge.NewMocks(t, v2_3_0, v2_3_0)
	defer mocks.Close()

	// Bridge will disable the proxy by default at startup.
	mocks.ProxyCtl.EXPECT().DisallowProxy()

	// Get the path to the vault.
	vaultDir, err := locator.ProvideSettingsPath()
	require.NoError(t, err)

	// Create the vault.
	vault, _, err := vault.New(vaultDir, t.TempDir(), vaultKey)
	require.NoError(t, err)

	// Create a new cookie jar.
	cookieJar, err := cookies.NewCookieJar(bridge.NewTestCookieJar(), vault)
	require.NoError(t, err)
	defer func() { require.NoError(t, cookieJar.PersistCookies()) }()

	// Create a new bridge.
	bridge, eventCh, err := bridge.New(
		// The app stuff.
		locator,
		vault,
		mocks.Autostarter,
		mocks.Updater,
		v2_3_0,

		// The API stuff.
		apiURL,
		cookieJar,
		useragent.New(),
		mocks.TLSReporter,
		netCtl.NewRoundTripper(&tls.Config{InsecureSkipVerify: true}),
		mocks.ProxyCtl,
		mocks.CrashHandler,
		mocks.Reporter,

		// The logging stuff.
		os.Getenv("BRIDGE_LOG_IMAP_CLIENT") == "1",
		os.Getenv("BRIDGE_LOG_IMAP_SERVER") == "1",
		os.Getenv("BRIDGE_LOG_SMTP") == "1",
	)
	require.NoError(t, err)
	require.Empty(t, bridge.GetErrors())

	// Wait for bridge to finish loading users.
	waitForEvent(t, eventCh, events.AllUsersLoaded{})

	// Set random IMAP and SMTP ports for the tests.
	require.NoError(t, bridge.SetIMAPPort(0))
	require.NoError(t, bridge.SetSMTPPort(0))

	// Close the bridge when done.
	defer bridge.Close(ctx)

	// Use the bridge.
	tests(bridge, mocks)
}

func waitForEvent[T any](t *testing.T, eventCh <-chan events.Event, wantEvent T) {
	t.Helper()

	for event := range eventCh {
		switch event.(type) { // nolint:gocritic
		case T:
			return
		}
	}
}

// must is a helper function that panics on error.
func must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}

	return val
}

func getConnectedUserIDs(t *testing.T, b *bridge.Bridge) []string {
	t.Helper()

	return xslices.Filter(b.GetUserIDs(), func(userID string) bool {
		info, err := b.GetUserInfo(userID)
		require.NoError(t, err)

		return info.State == bridge.Connected
	})
}

func chToType[In, Out any](inCh <-chan In, done func()) (<-chan Out, func()) {
	outCh := make(chan Out)

	go func() {
		defer close(outCh)

		for in := range inCh {
			out, ok := any(in).(Out)
			if !ok {
				panic(fmt.Sprintf("unexpected type %T", in))
			}

			outCh <- out
		}
	}()

	return outCh, done
}
