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
	"testing"
	"time"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/go-proton-api/server"
	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/stretchr/testify/require"
)

func TestBridge_WithoutUsers(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			require.Empty(t, bridge.GetUserIDs())
			require.Empty(t, getConnectedUserIDs(t, bridge))
		})

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			require.Empty(t, bridge.GetUserIDs())
			require.Empty(t, getConnectedUserIDs(t, bridge))
		})
	})
}

func TestBridge_Login(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// Login the user.
			userID, err := bridge.LoginFull(ctx, username, password, nil, nil)
			require.NoError(t, err)

			// The user is now connected.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Equal(t, []string{userID}, getConnectedUserIDs(t, bridge))
		})
	})
}

func TestBridge_Login_DropConn(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	dropListener := proton.NewListener(l, proton.NewDropConn)
	defer func() { _ = dropListener.Close() }()

	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// Login the user.
			userID, err := bridge.LoginFull(ctx, username, password, nil, nil)
			require.NoError(t, err)

			// The user is now connected.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Equal(t, []string{userID}, getConnectedUserIDs(t, bridge))
		})

		// Whether to allow the user to be created.
		var allowUser bool

		s.AddStatusHook(func(req *http.Request) (int, bool) {
			// Drop any request to the users endpoint.
			if !allowUser && req.URL.Path == "/core/v4/users" {
				dropListener.DropAll()
			}

			// After the ping request, allow the user to be created.
			if req.URL.Path == "/tests/ping" {
				allowUser = true
			}

			return 0, false
		})

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// The user is eventually connected.
			require.Eventually(t, func() bool {
				return len(bridge.GetUserIDs()) == 1 && len(getConnectedUserIDs(t, bridge)) == 1
			}, 5*time.Second, 100*time.Millisecond)
		})
	}, server.WithListener(dropListener))
}

func TestBridge_LoginTwice(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// Login the user.
			userID, err := bridge.LoginFull(ctx, username, password, nil, nil)
			require.NoError(t, err)

			// The user is now connected.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Equal(t, []string{userID}, getConnectedUserIDs(t, bridge))

			// Additional login should fail.
			_, err = bridge.LoginFull(ctx, username, password, nil, nil)
			require.Error(t, err)
		})
	})
}

func TestBridge_LoginLogoutLogin(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// Login the user.
			userID := must(bridge.LoginFull(ctx, username, password, nil, nil))

			// The user is now connected.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Equal(t, []string{userID}, getConnectedUserIDs(t, bridge))

			// Logout the user.
			require.NoError(t, bridge.LogoutUser(ctx, userID))

			// The user is now disconnected.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Empty(t, getConnectedUserIDs(t, bridge))

			// Login the user again.
			newUserID := must(bridge.LoginFull(ctx, username, password, nil, nil))
			require.Equal(t, userID, newUserID)

			// The user is connected again.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Equal(t, []string{userID}, getConnectedUserIDs(t, bridge))
		})
	})
}

func TestBridge_LoginDeleteLogin(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// Login the user.
			userID := must(bridge.LoginFull(ctx, username, password, nil, nil))

			// The user is now connected.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Equal(t, []string{userID}, getConnectedUserIDs(t, bridge))

			// Delete the user.
			require.NoError(t, bridge.DeleteUser(ctx, userID))

			// The user is now gone.
			require.Empty(t, bridge.GetUserIDs())
			require.Empty(t, getConnectedUserIDs(t, bridge))

			// Login the user again.
			newUserID := must(bridge.LoginFull(ctx, username, password, nil, nil))
			require.Equal(t, userID, newUserID)

			// The user is connected again.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Equal(t, []string{userID}, getConnectedUserIDs(t, bridge))
		})
	})
}

func TestBridge_LoginDeauthLogin(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// Login the user.
			userID := must(bridge.LoginFull(ctx, username, password, nil, nil))

			// Get a channel to receive the deauth event.
			eventCh, done := bridge.GetEvents(events.UserDeauth{})
			defer done()

			// Deauth the user.
			require.NoError(t, s.RevokeUser(userID))

			// The user is eventually disconnected.
			require.Eventually(t, func() bool {
				return len(getConnectedUserIDs(t, bridge)) == 0
			}, 10*time.Second, time.Second)

			// We should get a deauth event.
			require.IsType(t, events.UserDeauth{}, <-eventCh)

			// Login the user after the disconnection.
			newUserID := must(bridge.LoginFull(ctx, username, password, nil, nil))
			require.Equal(t, userID, newUserID)

			// The user is connected again.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Equal(t, []string{userID}, getConnectedUserIDs(t, bridge))
		})
	})
}

func TestBridge_LoginDeauthRestartLogin(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		var userID string

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// Login the user.
			userID = must(bridge.LoginFull(ctx, username, password, nil, nil))

			// Get a channel to receive the deauth event.
			eventCh, done := bridge.GetEvents(events.UserDeauth{})
			defer done()

			// Deauth the user.
			require.NoError(t, s.RevokeUser(userID))

			// The user is eventually disconnected.
			require.Eventually(t, func() bool {
				return len(getConnectedUserIDs(t, bridge)) == 0
			}, 10*time.Second, time.Second)

			// We should get a deauth event.
			require.IsType(t, events.UserDeauth{}, <-eventCh)
		})

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// The user should be disconnected at startup.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Empty(t, getConnectedUserIDs(t, bridge))

			// Login the user after the disconnection.
			newUserID := must(bridge.LoginFull(ctx, username, password, nil, nil))
			require.Equal(t, userID, newUserID)

			// The user is connected again.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Equal(t, []string{userID}, getConnectedUserIDs(t, bridge))
		})
	})
}

func TestBridge_LoginExpireLogin(t *testing.T) {
	const authLife = 2 * time.Second

	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		s.SetAuthLife(authLife)

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// Login the user. Its auth will only be valid for a short time.
			userID := must(bridge.LoginFull(ctx, username, password, nil, nil))

			// Wait until the auth expires.
			time.Sleep(authLife)

			// The user will have to refresh but the logout will still succeed.
			require.NoError(t, bridge.LogoutUser(ctx, userID))
		})
	})
}

func TestBridge_FailToLoad(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		var userID string

		// Login the user.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			userID = must(bridge.LoginFull(ctx, username, password, nil, nil))
		})

		// Deauth the user while bridge is stopped.
		require.NoError(t, s.RevokeUser(userID))

		// When bridge starts, the user will not be logged in.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Empty(t, getConnectedUserIDs(t, bridge))
		})
	})
}

func TestBridge_LoadWithoutInternet(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		var userID string

		// Login the user.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			userID = must(bridge.LoginFull(ctx, username, password, nil, nil))
		})

		// Simulate loss of internet connection.
		netCtl.Disable()

		// Start bridge without internet.
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// Initially, users are not connected.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Empty(t, getConnectedUserIDs(t, bridge))

			time.Sleep(5 * time.Second)

			// Simulate internet connection.
			netCtl.Enable()

			// The user will eventually be connected.
			require.Eventually(t, func() bool {
				return len(getConnectedUserIDs(t, bridge)) == 1 && getConnectedUserIDs(t, bridge)[0] == userID
			}, 10*time.Second, time.Second)
		})
	})
}

func TestBridge_LoginRestart(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		var userID string

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			userID = must(bridge.LoginFull(ctx, username, password, nil, nil))
		})

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Equal(t, []string{userID}, getConnectedUserIDs(t, bridge))
		})
	})
}

func TestBridge_LoginLogoutRestart(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		var userID string

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// Login the user.
			userID = must(bridge.LoginFull(ctx, username, password, nil, nil))

			// Logout the user.
			require.NoError(t, bridge.LogoutUser(ctx, userID))
		})

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// The user is still disconnected.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Empty(t, getConnectedUserIDs(t, bridge))
		})
	})
}

func TestBridge_LoginDeleteRestart(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		var userID string

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// Login the user.
			userID = must(bridge.LoginFull(ctx, username, password, nil, nil))

			// Delete the user.
			require.NoError(t, bridge.DeleteUser(ctx, userID))
		})

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// The user is still gone.
			require.Empty(t, bridge.GetUserIDs())
			require.Empty(t, getConnectedUserIDs(t, bridge))
		})
	})
}

func TestBridge_FailLoginRecover(t *testing.T) {
	for i := uint64(1); i < 10; i++ {
		t.Run(fmt.Sprintf("read %v%% of the data", 100*i/10), func(t *testing.T) {
			withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
				var userID string

				// Log the user in, wait for it to sync, then log it out.
				// (We don't want to count message sync data in the test.)
				withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
					syncCh, done := chToType[events.Event, events.SyncFinished](bridge.GetEvents(events.SyncFinished{}))
					defer done()

					userID = must(bridge.LoginFull(ctx, username, password, nil, nil))
					require.Equal(t, userID, (<-syncCh).UserID)
					require.NoError(t, bridge.LogoutUser(ctx, userID))
				})

				var total uint64

				// Now that the user is synced, we can measure exactly how much data is needed during login.
				withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
					total = countBytesRead(netCtl, func() {
						must(bridge.LoginFull(ctx, username, password, nil, nil))
					})

					require.NoError(t, bridge.LogoutUser(ctx, userID))
				})

				// Now simulate failing to login.
				withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
					// Simulate a partial read.
					netCtl.SetReadLimit(i * total / 10)

					// We should fail to log the user in because we can't fully read its data.
					require.Error(t, getErr(bridge.LoginFull(ctx, username, password, nil, nil)))

					// The user should still be there (but disconnected).
					require.Equal(t, []string{userID}, bridge.GetUserIDs())
					require.Empty(t, getConnectedUserIDs(t, bridge))
				})

				// Simulate the network recovering.
				netCtl.SetReadLimit(0)

				// We should now be able to log the user in.
				withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
					require.NoError(t, getErr(bridge.LoginFull(ctx, username, password, nil, nil)))

					// The user should be there, now connected.
					require.Equal(t, []string{userID}, bridge.GetUserIDs())
					require.Equal(t, []string{userID}, getConnectedUserIDs(t, bridge))
				})
			})
		})
	}
}

func TestBridge_FailLoadRecover(t *testing.T) {
	for i := uint64(1); i < 10; i++ {
		t.Run(fmt.Sprintf("read %v%% of the data", 100*i/10), func(t *testing.T) {
			withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
				var userID string

				// Log the user in and wait for it to sync.
				// (We don't want to count message sync data in the test.)
				withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
					syncCh, done := chToType[events.Event, events.SyncFinished](bridge.GetEvents(events.SyncFinished{}))
					defer done()

					userID = must(bridge.LoginFull(ctx, username, password, nil, nil))
					require.Equal(t, userID, (<-syncCh).UserID)
				})

				// See how much data it takes to load the user at startup.
				total := countBytesRead(netCtl, func() {
					withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(_ *bridge.Bridge, _ *bridge.Mocks) {
						// ...
					})
				})

				// Simulate a partial read.
				netCtl.SetReadLimit(i * total / 10)

				// We should fail to load the user; it should be listed but disconnected.
				withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
					require.Equal(t, []string{userID}, bridge.GetUserIDs())
					require.Empty(t, getConnectedUserIDs(t, bridge))
				})

				// Simulate the network recovering.
				netCtl.SetReadLimit(0)

				// We should now be able to load the user.
				withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
					require.Equal(t, []string{userID}, bridge.GetUserIDs())
					require.Equal(t, []string{userID}, getConnectedUserIDs(t, bridge))
				})
			})
		})
	}
}

func TestBridge_BridgePass(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		var userID string

		var pass []byte

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// Login the user.
			userID = must(bridge.LoginFull(ctx, username, password, nil, nil))

			// Retrieve the bridge pass.
			pass = must(bridge.GetUserInfo(userID)).BridgePass

			// Log the user out.
			require.NoError(t, bridge.LogoutUser(ctx, userID))

			// Log the user back in.
			must(bridge.LoginFull(ctx, username, password, nil, nil))

			// The bridge pass should be the same.
			require.Equal(t, pass, pass)
		})

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// The bridge should load the user.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Equal(t, []string{userID}, getConnectedUserIDs(t, bridge))

			// The bridge pass should be the same.
			require.Equal(t, pass, must(bridge.GetUserInfo(userID)).BridgePass)
		})
	})
}

func TestBridge_AddressMode(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// Login the user.
			userID, err := bridge.LoginFull(ctx, username, password, nil, nil)
			require.NoError(t, err)

			// Get the user's info.
			info, err := bridge.GetUserInfo(userID)
			require.NoError(t, err)

			// The user is in combined mode by default.
			require.Equal(t, vault.CombinedMode, info.AddressMode)

			// Repeatedly switch between address modes.
			for i := 1; i <= 10; i++ {
				var target vault.AddressMode

				if i%2 == 0 {
					target = vault.CombinedMode
				} else {
					target = vault.SplitMode
				}

				// Put the user in the target mode.
				require.NoError(t, bridge.SetAddressMode(ctx, userID, target))

				// Get the user's info.
				info, err = bridge.GetUserInfo(userID)
				require.NoError(t, err)

				// The user is in the target mode.
				require.Equal(t, target, info.AddressMode)
			}
		})
	})
}

func TestBridge_LoginLogoutRepeated(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			for i := 0; i < 10; i++ {
				// Log the user in.
				userID := must(bridge.LoginFull(ctx, username, password, nil, nil))

				// Log the user out.
				require.NoError(t, bridge.LogoutUser(ctx, userID))
			}
		})
	})
}

func TestBridge_LogoutOffline(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		var userID string

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// Login the user.
			userID = must(bridge.LoginFull(ctx, username, password, nil, nil))

			// The user is now connected.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Equal(t, []string{userID}, getConnectedUserIDs(t, bridge))

			// Go offline.
			netCtl.Disable()

			// We can still log the user out.
			require.NoError(t, bridge.LogoutUser(ctx, userID))

			// The user is now disconnected.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Empty(t, getConnectedUserIDs(t, bridge))
		})

		// Go back online.
		netCtl.Enable()

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// The user is still disconnected.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Empty(t, getConnectedUserIDs(t, bridge))
		})
	})
}

func TestBridge_DeleteDisconnected(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// Login the user.
			userID, err := bridge.LoginFull(ctx, username, password, nil, nil)
			require.NoError(t, err)

			// The user is now connected.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Equal(t, []string{userID}, getConnectedUserIDs(t, bridge))

			// Logout the user.
			require.NoError(t, bridge.LogoutUser(ctx, userID))

			// The user is now disconnected.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Empty(t, getConnectedUserIDs(t, bridge))

			// Delete the user.
			require.NoError(t, bridge.DeleteUser(ctx, userID))

			// The user is now deleted.
			require.Empty(t, bridge.GetUserIDs())
			require.Empty(t, getConnectedUserIDs(t, bridge))
		})
	})
}

func TestBridge_DeleteOffline(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// Login the user.
			userID, err := bridge.LoginFull(ctx, username, password, nil, nil)
			require.NoError(t, err)

			// The user is now connected.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Equal(t, []string{userID}, getConnectedUserIDs(t, bridge))

			// Go offline.
			netCtl.Disable()

			// We can still log the user out.
			require.NoError(t, bridge.DeleteUser(ctx, userID))

			// The user is now gone.
			require.Empty(t, bridge.GetUserIDs())
			require.Empty(t, getConnectedUserIDs(t, bridge))
		})
	})
}

func TestBridge_UserInfo_Alias(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, vaultKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// Create a new user.
			userID, _, err := s.CreateUser("primary", []byte("password"))
			require.NoError(t, err)

			// Give the new user an alias.
			require.NoError(t, getErr(s.CreateAddress(userID, "alias@pm.me", []byte("password"))))

			// Login the user.
			require.NoError(t, getErr(bridge.LoginFull(ctx, "primary", []byte("password"), nil, nil)))

			// Get user info.
			info, err := bridge.GetUserInfo(userID)
			require.NoError(t, err)

			// The user should have two addresses, the primary should be first.
			require.Equal(t, []string{"primary@" + s.GetDomain(), "alias@pm.me"}, info.Addresses)
		})
	})
}

func TestBridge_User_Refresh(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, vaultKey []byte) {
		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, vaultKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			// Get a channel of sync started events.
			syncStartCh, done := chToType[events.Event, events.SyncStarted](bridge.GetEvents(events.SyncStarted{}))
			defer done()

			// Get a channel of sync finished events.
			syncFinishCh, done := chToType[events.Event, events.SyncFinished](bridge.GetEvents(events.SyncFinished{}))
			defer done()

			// Log the user in.
			userID := must(bridge.LoginFull(ctx, username, password, nil, nil))

			// The sync should start and finish.
			require.Equal(t, userID, (<-syncStartCh).UserID)
			require.Equal(t, userID, (<-syncFinishCh).UserID)

			// Trigger a refresh.
			require.NoError(t, s.RefreshUser(userID, proton.RefreshAll))

			// The sync should start and finish again.
			require.Equal(t, userID, (<-syncStartCh).UserID)
			require.Equal(t, userID, (<-syncFinishCh).UserID)
		})
	})
}

func TestBridge_User_GetAddresses(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, netCtl *proton.NetCtl, locator bridge.Locator, storeKey []byte) {
		// Create a user.
		userID, _, err := s.CreateUser("user", password)
		require.NoError(t, err)
		addrID2, err := s.CreateAddress(userID, "user@external.com", []byte("password"))
		require.NoError(t, err)
		require.NoError(t, s.ChangeAddressType(userID, addrID2, proton.AddressTypeExternal))

		withBridge(ctx, t, s.GetHostURL(), netCtl, locator, storeKey, func(bridge *bridge.Bridge, _ *bridge.Mocks) {
			userLoginAndSync(ctx, t, bridge, "user", password)
			info, err := bridge.GetUserInfo(userID)
			require.NoError(t, err)
			require.Equal(t, 1, len(info.Addresses))
			require.Equal(t, info.Addresses[0], "user@proton.local")
		})
	})
}

// getErr returns the error that was passed to it.
func getErr[T any](_ T, err error) error {
	return err
}
