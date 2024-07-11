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
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/go-proton-api/server"
	"github.com/ProtonMail/go-proton-api/server/backend"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v3/internal/certs"
	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/internal/cookies"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/focus"
	"github.com/ProtonMail/proton-bridge/v3/internal/locations"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/imapsmtpserver"
	"github.com/ProtonMail/proton-bridge/v3/internal/updater"
	"github.com/ProtonMail/proton-bridge/v3/internal/user"
	"github.com/ProtonMail/proton-bridge/v3/internal/useragent"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/ProtonMail/proton-bridge/v3/pkg/keychain"
	"github.com/ProtonMail/proton-bridge/v3/tests"
	"github.com/bradenaw/juniper/xslices"
	imapid "github.com/emersion/go-imap-id"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
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
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
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
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// Get a stream of TLS issue events.
			raiseCh, done := bridge.GetEvents(events.Raise{})
			defer done()

			settingsFolder, err := locator.ProvideSettingsPath()
			require.NoError(t, err)

			// Simulate a focus event.
			focus.TryRaise(settingsFolder)

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

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
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

func TestBridge_UserAgent_Persistence(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, vaultKey []byte) {
		otherPassword := []byte("bar")
		otherUser := "foo"
		_, _, err := s.CreateUser(otherUser, otherPassword)
		require.NoError(t, err)

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(b *bridge.Bridge, _ *bridge.Mocks) {
			currentUserAgent := b.GetCurrentUserAgent()
			require.Contains(t, currentUserAgent, useragent.DefaultUserAgent)

			require.NoError(t, getErr(b.LoginFull(ctx, otherUser, otherPassword, nil, nil)))

			imapClient, err := eventuallyDial(fmt.Sprintf("%v:%v", constants.Host, b.GetIMAPPort()))
			require.NoError(t, err)
			defer func() { _ = imapClient.Logout() }()

			idClient := imapid.NewClient(imapClient)

			// Set IMAP ID before Login to have the value capture in the Login API Call.
			_, err = idClient.ID(imapid.ID{
				imapid.FieldName:    "MyFancyClient",
				imapid.FieldVersion: "0.1.2",
			})

			require.NoError(t, err)

			// Login the user.
			_, err = b.LoginFull(context.Background(), username, password, nil, nil)
			require.NoError(t, err)

			// Assert that the user agent then contains the platform.
			require.Contains(t, b.GetCurrentUserAgent(), "MyFancyClient/0.1.2")
		})

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			currentUserAgent := bridge.GetCurrentUserAgent()
			require.Contains(t, currentUserAgent, "MyFancyClient/0.1.2")
		})
	})
}

func TestBridge_UserAgentFromUnknownClient(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, vaultKey []byte) {
		otherPassword := []byte("bar")
		otherUser := "foo"
		_, _, err := s.CreateUser(otherUser, otherPassword)
		require.NoError(t, err)

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(b *bridge.Bridge, _ *bridge.Mocks) {
			currentUserAgent := b.GetCurrentUserAgent()
			require.Contains(t, currentUserAgent, useragent.DefaultUserAgent)

			userID, err := b.LoginFull(context.Background(), username, password, nil, nil)
			require.NoError(t, err)

			imapClient, err := eventuallyDial(fmt.Sprintf("%v:%v", constants.Host, b.GetIMAPPort()))
			require.NoError(t, err)
			defer func() { _ = imapClient.Logout() }()

			info, err := b.GetUserInfo(userID)
			require.NoError(t, err)
			require.True(t, info.State == bridge.Connected)

			require.NoError(t, imapClient.Login(info.Addresses[0], string(info.BridgePass)))

			currentUserAgent = b.GetCurrentUserAgent()
			require.Contains(t, currentUserAgent, "UnknownClient/0.0.1")
		})
	})
}

func TestBridge_UserAgentFromSMTPClient(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, vaultKey []byte) {
		otherPassword := []byte("bar")
		otherUser := "foo"
		_, _, err := s.CreateUser(otherUser, otherPassword)
		require.NoError(t, err)

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(b *bridge.Bridge, _ *bridge.Mocks) {
			currentUserAgent := b.GetCurrentUserAgent()
			require.Contains(t, currentUserAgent, useragent.DefaultUserAgent)

			userID, err := b.LoginFull(context.Background(), username, password, nil, nil)
			require.NoError(t, err)

			client, err := smtp.Dial(net.JoinHostPort(constants.Host, fmt.Sprint(b.GetSMTPPort())))
			require.NoError(t, err)
			defer client.Close() //nolint:errcheck

			info, err := b.GetUserInfo(userID)
			require.NoError(t, err)
			require.True(t, info.State == bridge.Connected)

			// Upgrade to TLS.
			require.NoError(t, client.StartTLS(&tls.Config{InsecureSkipVerify: true}))
			require.NoError(t, client.Auth(sasl.NewLoginClient(
				info.Addresses[0],
				string(info.BridgePass)),
			))

			require.Eventually(t, func() bool {
				currentUserAgent = b.GetCurrentUserAgent()

				return strings.Contains(currentUserAgent, "UnknownClient/0.0.1")
			}, time.Minute, 5*time.Second)
		})
	})
}

func TestBridge_UserAgentFromIMAPID(t *testing.T) {
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

		otherPassword := []byte("bar")
		otherUser := "foo"
		_, _, err := s.CreateUser(otherUser, otherPassword)
		require.NoError(t, err)

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(b *bridge.Bridge, _ *bridge.Mocks) {
			require.NoError(t, getErr(b.LoginFull(ctx, otherUser, otherPassword, nil, nil)))

			imapClient, err := eventuallyDial(fmt.Sprintf("%v:%v", constants.Host, b.GetIMAPPort()))
			require.NoError(t, err)
			defer func() { _ = imapClient.Logout() }()

			idClient := imapid.NewClient(imapClient)

			// Set IMAP ID before Login to have the value capture in the Login API Call.
			_, err = idClient.ID(imapid.ID{
				imapid.FieldName:    "MyFancyClient",
				imapid.FieldVersion: "0.1.2",
			})

			require.NoError(t, err)

			// Login the user.
			userID, err := b.LoginFull(context.Background(), username, password, nil, nil)
			require.NoError(t, err)

			info, err := b.GetUserInfo(userID)
			require.NoError(t, err)
			require.True(t, info.State == bridge.Connected)

			require.NoError(t, imapClient.Login(info.Addresses[0], string(info.BridgePass)))

			lock.Lock()
			defer lock.Unlock()

			userAgent := calls[len(calls)-1].RequestHeader.Get("User-Agent")

			// Assert that the user agent was sent to the API.
			require.Contains(t, userAgent, b.GetCurrentUserAgent())
			require.Contains(t, userAgent, "MyFancyClient/0.1.2")
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
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			_, err := bridge.LoginFull(context.Background(), username, password, nil, nil)
			require.NoError(t, err)
		})

		// Start bridge again and check that it uses the same session ID.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(_ *bridge.Bridge, _ *bridge.Mocks) {
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
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
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
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			newUserID, err := bridge.LoginFull(context.Background(), username, password, nil, nil)
			require.NoError(t, err)

			userID = newUserID
		})

		// Start bridge with the correct vault key -- it should load the users correctly.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			require.ElementsMatch(t, []string{userID}, bridge.GetUserIDs())
		})

		// Start bridge with a bad vault key, the vault will be wiped and bridge will show no users.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, []byte("bad"), func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			require.Empty(t, bridge.GetUserIDs())
		})

		// Start bridge with a nil vault key, the vault will be wiped and bridge will show no users.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, nil, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			require.Empty(t, bridge.GetUserIDs())
		})
	})
}

func TestBridge_MissingGluonStore(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, vaultKey []byte) {
		var gluonDir string

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			_, err := bridge.LoginFull(context.Background(), username, password, nil, nil)
			require.NoError(t, err)

			// Move the gluon dir.
			require.NoError(t, bridge.SetGluonDir(ctx, t.TempDir()))

			// Get the gluon dir.
			gluonDir = bridge.GetGluonCacheDir()
		})

		// The user removes the gluon dir while bridge is not running.
		require.NoError(t, os.RemoveAll(gluonDir))

		// Bridge starts but can't find the gluon store dir; there should be no error.
		withBridgeWaitForServers(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(_ *bridge.Bridge, _ *bridge.Mocks) {
			// ...
		})
	})
}

func TestBridge_MissingGluonDatabase(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, vaultKey []byte) {
		var gluonDir string

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			_, err := bridge.LoginFull(context.Background(), username, password, nil, nil)
			require.NoError(t, err)

			// Get the gluon dir.
			gluonDir, err = bridge.GetGluonDataDir()
			require.NoError(t, err)
		})

		// The user removes the gluon dir while bridge is not running.
		require.NoError(t, os.RemoveAll(gluonDir))

		// Bridge starts but can't find the gluon database dir; there should be no error.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(_ *bridge.Bridge, _ *bridge.Mocks) {
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

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// Watch for sync finished event.
			syncCh, done := chToType[events.Event, events.SyncFinished](bridge.GetEvents(events.SyncFinished{}))
			defer done()

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

func TestBridge_InitGluonDirectory(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, vaultKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(b *bridge.Bridge, _ *bridge.Mocks) {
			configDir, err := b.GetGluonDataDir()
			require.NoError(t, err)

			_, err = os.ReadDir(imapsmtpserver.ApplyGluonCachePathSuffix(b.GetGluonCacheDir()))
			require.False(t, os.IsNotExist(err))

			_, err = os.ReadDir(imapsmtpserver.ApplyGluonConfigPathSuffix(configDir))
			require.False(t, os.IsNotExist(err))
		})
	})
}

func TestBridge_LoginFailed(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, vaultKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			failCh, done := chToType[events.Event, events.IMAPLoginFailed](bridge.GetEvents(events.IMAPLoginFailed{}))
			defer done()

			_, err := bridge.LoginFull(ctx, username, password, nil, nil)
			require.NoError(t, err)

			imapClient, err := eventuallyDial(net.JoinHostPort(constants.Host, fmt.Sprint(bridge.GetIMAPPort())))
			require.NoError(t, err)

			require.Error(t, imapClient.Login("badUser", "badPass"))
			require.Equal(t, "badUser", (<-failCh).Username)
		})
	})
}

func TestBridge_ChangeCacheDirectory(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, vaultKey []byte) {
		userID, addrID, err := s.CreateUser("imap", password)
		require.NoError(t, err)

		labelID, err := s.CreateLabel(userID, "folder", "", proton.LabelTypeFolder)
		require.NoError(t, err)

		withClient(ctx, t, s, "imap", password, func(ctx context.Context, c *proton.Client) {
			createNumMessages(ctx, t, c, addrID, labelID, 10)
		})

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(b *bridge.Bridge, _ *bridge.Mocks) {
			newCacheDir := t.TempDir()
			currentCacheDir := b.GetGluonCacheDir()
			configDir, err := b.GetGluonDataDir()
			require.NoError(t, err)

			// Login the user.
			syncCh, done := chToType[events.Event, events.SyncFinished](b.GetEvents(events.SyncFinished{}))
			defer done()
			userID, err := b.LoginFull(ctx, "imap", password, nil, nil)
			require.NoError(t, err)
			require.Equal(t, userID, (<-syncCh).UserID)

			// The user is now connected.
			require.Equal(t, []string{userID}, b.GetUserIDs())
			require.Equal(t, []string{userID}, getConnectedUserIDs(t, b))

			// Change directory
			err = b.SetGluonDir(ctx, newCacheDir)
			require.NoError(t, err)

			// Old store should no more exists.
			_, err = os.ReadDir(imapsmtpserver.ApplyGluonCachePathSuffix(currentCacheDir))
			require.True(t, os.IsNotExist(err))
			// Database should not have changed.
			_, err = os.ReadDir(imapsmtpserver.ApplyGluonConfigPathSuffix(configDir))
			require.False(t, os.IsNotExist(err))

			// New path should have Gluon sub-folder.
			require.Equal(t, filepath.Join(newCacheDir, "gluon"), b.GetGluonCacheDir())
			// And store should be inside it.
			_, err = os.ReadDir(imapsmtpserver.ApplyGluonCachePathSuffix(b.GetGluonCacheDir()))
			require.False(t, os.IsNotExist(err))

			// We should be able to fetch.
			info, err := b.GetUserInfo(userID)
			require.NoError(t, err)
			require.True(t, info.State == bridge.Connected)

			client, err := eventuallyDial(fmt.Sprintf("%v:%v", constants.Host, b.GetIMAPPort()))
			require.NoError(t, err)
			require.NoError(t, client.Login(info.Addresses[0], string(info.BridgePass)))
			defer func() { _ = client.Logout() }()

			status, err := client.Select(`Folders/folder`, false)
			require.NoError(t, err)
			require.Equal(t, uint32(10), status.Messages)
		})
	})
}

func TestBridge_ChangeAddressOrder(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, vaultKey []byte) {
		// Create a user.
		userID, addrID, err := s.CreateUser("imap", password)
		require.NoError(t, err)

		// Create a second address for the user.
		aliasID, err := s.CreateAddress(userID, "alias@"+s.GetDomain(), password)
		require.NoError(t, err)

		// Create 10 messages for the user.
		withClient(ctx, t, s, "imap", password, func(ctx context.Context, c *proton.Client) {
			createNumMessages(ctx, t, c, addrID, proton.InboxLabel, 10)
		})

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(b *bridge.Bridge, _ *bridge.Mocks) {
			// Log the user in with its first address.
			syncCh, done := chToType[events.Event, events.SyncFinished](b.GetEvents(events.SyncFinished{}))
			defer done()
			userID, err := b.LoginFull(ctx, "imap", password, nil, nil)
			require.NoError(t, err)
			require.Equal(t, userID, (<-syncCh).UserID)

			// We should see 10 messages in the inbox.
			info, err := b.GetUserInfo(userID)
			require.NoError(t, err)
			require.True(t, info.State == bridge.Connected)

			client, err := eventuallyDial(fmt.Sprintf("%v:%v", constants.Host, b.GetIMAPPort()))
			require.NoError(t, err)
			require.NoError(t, client.Login(info.Addresses[0], string(info.BridgePass)))
			defer func() { _ = client.Logout() }()

			status, err := client.Select(`Inbox`, false)
			require.NoError(t, err)
			require.Equal(t, uint32(10), status.Messages)
		})

		// Make the second address the primary one.
		withClient(ctx, t, s, "imap", password, func(ctx context.Context, c *proton.Client) {
			require.NoError(t, c.OrderAddresses(ctx, proton.OrderAddressesReq{AddressIDs: []string{aliasID, addrID}}))
		})

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(b *bridge.Bridge, _ *bridge.Mocks) {
			// We should still see 10 messages in the inbox.
			info, err := b.GetUserInfo(userID)
			require.NoError(t, err)
			require.True(t, info.State == bridge.Connected)

			client, err := eventuallyDial(fmt.Sprintf("%v:%v", constants.Host, b.GetIMAPPort()))
			require.NoError(t, err)
			require.NoError(t, client.Login(info.Addresses[0], string(info.BridgePass)))
			defer func() { _ = client.Logout() }()

			require.Eventually(t, func() bool {
				status, err := client.Select(`Inbox`, false)
				require.NoError(t, err)
				return status.Messages == 10
			}, 5*time.Second, 100*time.Millisecond)
		})
	})
}

// withEnv creates the full test environment and runs the tests.
func withEnv(t *testing.T, tests func(context.Context, *server.Server, *proton.NetCtl, bridge.Locator, []byte), opts ...server.Option) {
	opt := goleak.IgnoreCurrent()
	defer goleak.VerifyNone(t, opt)

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

// withMocks creates the mock objects used in the tests.
func withMocks(t *testing.T, tests func(*bridge.Mocks)) {
	mocks := bridge.NewMocks(t, v2_3_0, v2_3_0)
	defer mocks.Close()

	tests(mocks)
}

// Needs to be global to survive bridge shutdown/startup in unit tests as they happen to fast.
var testUIDValidityGenerator = imap.DefaultEpochUIDValidityGenerator()

// withBridge creates a new bridge which points to the given API URL and uses the given keychain, and closes it when done.
func withBridgeNoMocks(
	ctx context.Context,
	t *testing.T,
	mocks *bridge.Mocks,
	apiURL string,
	netCtl *proton.NetCtl,
	locator bridge.Locator,
	vaultKey []byte,
	tests func(*bridge.Bridge),
	waitOnServers bool,
) {
	// Bridge will disable the proxy by default at startup.
	mocks.ProxyCtl.EXPECT().DisallowProxy()

	// Get the path to the vault.
	vaultDir, err := locator.ProvideSettingsPath()
	require.NoError(t, err)

	// Create the vault.
	vault, _, err := vault.New(vaultDir, t.TempDir(), vaultKey, async.NoopPanicHandler{})
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
		keychain.NewTestKeychainsList(),

		// The API stuff.
		apiURL,
		cookieJar,
		useragent.New(),
		mocks.TLSReporter,
		netCtl.NewRoundTripper(&tls.Config{InsecureSkipVerify: true}),
		mocks.ProxyCtl,
		mocks.CrashHandler,
		mocks.Reporter,
		testUIDValidityGenerator,
		mocks.Heartbeat,

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
	require.NoError(t, bridge.SetIMAPPort(ctx, 0))
	require.NoError(t, bridge.SetSMTPPort(ctx, 0))

	if waitOnServers {
		// Wait for bridge to start the IMAP server.
		waitForEvent(t, eventCh, events.IMAPServerReady{})
		// Wait for bridge to start the SMTP server.
		waitForEvent(t, eventCh, events.SMTPServerReady{})
	}

	// Close the bridge when done.
	defer bridge.Close(ctx)

	// Use the bridge.
	tests(bridge)
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
	withMocks(t, func(mocks *bridge.Mocks) {
		withBridgeNoMocks(ctx, t, mocks, apiURL, netCtl, locator, vaultKey, func(bridge *bridge.Bridge) {
			tests(bridge, mocks)
		}, false)
	})
}

// withBridgeWaitForServers is the same as withBridge, but it will wait until IMAP & SMTP servers are ready.
func withBridgeWaitForServers(
	ctx context.Context,
	t *testing.T,
	apiURL string,
	netCtl *proton.NetCtl,
	locator bridge.Locator,
	vaultKey []byte,
	tests func(*bridge.Bridge, *bridge.Mocks),
) {
	withMocks(t, func(mocks *bridge.Mocks) {
		withBridgeNoMocks(ctx, t, mocks, apiURL, netCtl, locator, vaultKey, func(bridge *bridge.Bridge) {
			tests(bridge, mocks)
		}, true)
	})
}

func waitForEvent[T any](t *testing.T, eventCh <-chan events.Event, _ T) {
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

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer close(outCh)

		for in := range inCh {
			out, ok := any(in).(Out)
			if !ok {
				panic(fmt.Sprintf("unexpected type %T", in))
			}

			select {
			case <-ctx.Done():
				return

			case outCh <- out:
			}
		}
	}()

	return outCh, func() {
		cancel()
		done()
	}
}

type eventWaiter struct {
	evtCh  <-chan events.Event
	cancel func()
}

func (e *eventWaiter) Done() {
	e.cancel()
}

func (e *eventWaiter) Wait() {
	<-e.evtCh
}

func waitForSMTPServerReady(b *bridge.Bridge) *eventWaiter {
	evtCh, cancel := b.GetEvents(events.SMTPServerReady{})
	return &eventWaiter{
		evtCh:  evtCh,
		cancel: cancel,
	}
}

func waitForSMTPServerStopped(b *bridge.Bridge) *eventWaiter {
	evtCh, cancel := b.GetEvents(events.SMTPServerStopped{})
	return &eventWaiter{
		evtCh:  evtCh,
		cancel: cancel,
	}
}

func waitForIMAPServerReady(b *bridge.Bridge) *eventWaiter {
	evtCh, cancel := b.GetEvents(events.IMAPServerReady{})
	return &eventWaiter{
		evtCh:  evtCh,
		cancel: cancel,
	}
}

func waitForIMAPServerStopped(b *bridge.Bridge) *eventWaiter {
	evtCh, cancel := b.GetEvents(events.IMAPServerStopped{})
	return &eventWaiter{
		evtCh:  evtCh,
		cancel: cancel,
	}
}

func TestBridge_GetUpdatedCachePath(t *testing.T) {
	type TestData struct {
		gluonDBPath    string
		gluonCachePath string
		shouldChange   bool
	}

	dataArr := []TestData{
		{
			gluonDBPath:    "/Users/test/",
			gluonCachePath: "/Users/test/gluon",
			shouldChange:   false,
		}, {
			gluonDBPath:    "/Users/test/",
			gluonCachePath: "/Users/tester/gluon",
			shouldChange:   true,
		}, {
			gluonDBPath:    "/Users/testing/",
			gluonCachePath: "/Users/test/gluon",
			shouldChange:   true,
		},
		{
			gluonDBPath:    "/Users/testing/",
			gluonCachePath: "/Users/test/gluon",
			shouldChange:   true,
		},
		{
			gluonDBPath:    "/Users/testing/",
			gluonCachePath: "/Volumes/test/gluon",
			shouldChange:   false,
		},
		{
			gluonDBPath:    "/Volumes/test/",
			gluonCachePath: "/Users/test/gluon",
			shouldChange:   false,
		},
		{
			gluonDBPath:    "/XXX/test/",
			gluonCachePath: "/Users/test/gluon",
			shouldChange:   false,
		},
		{
			gluonDBPath:    "/XXX/test/",
			gluonCachePath: "/YYY/test/gluon",
			shouldChange:   false,
		},
	}

	for _, el := range dataArr {
		newCachePath := bridge.GetUpdatedCachePath(el.gluonDBPath, el.gluonCachePath)
		require.Equal(t, el.shouldChange, newCachePath != "" && newCachePath != el.gluonCachePath)
	}
}
