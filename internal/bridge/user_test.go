package bridge_test

import (
	"context"
	"testing"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
	"github.com/stretchr/testify/require"
	"gitlab.protontech.ch/go/liteapi/server"
)

func TestBridge_WithoutUsers(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, dialer *bridge.TestDialer, locator bridge.Locator, storeKey []byte) {
		withBridge(t, ctx, s.GetHostURL(), dialer, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			require.Empty(t, bridge.GetUserIDs())
			require.Empty(t, getConnectedUserIDs(t, bridge))
		})

		withBridge(t, ctx, s.GetHostURL(), dialer, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			require.Empty(t, bridge.GetUserIDs())
			require.Empty(t, getConnectedUserIDs(t, bridge))
		})
	})
}

func TestBridge_Login(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, dialer *bridge.TestDialer, locator bridge.Locator, storeKey []byte) {
		withBridge(t, ctx, s.GetHostURL(), dialer, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Login the user.
			userID, err := bridge.LoginUser(ctx, username, password, nil, nil)
			require.NoError(t, err)

			// The user is now connected.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Equal(t, []string{userID}, getConnectedUserIDs(t, bridge))
		})
	})
}

func TestBridge_LoginLogoutLogin(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, dialer *bridge.TestDialer, locator bridge.Locator, storeKey []byte) {
		withBridge(t, ctx, s.GetHostURL(), dialer, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Login the user.
			userID := must(bridge.LoginUser(ctx, username, password, nil, nil))

			// The user is now connected.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Equal(t, []string{userID}, getConnectedUserIDs(t, bridge))

			// Logout the user.
			require.NoError(t, bridge.LogoutUser(ctx, userID))

			// The user is now disconnected.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Empty(t, getConnectedUserIDs(t, bridge))

			// Login the user again.
			newUserID := must(bridge.LoginUser(ctx, username, password, nil, nil))
			require.Equal(t, userID, newUserID)

			// The user is connected again.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Equal(t, []string{userID}, getConnectedUserIDs(t, bridge))
		})
	})
}

func TestBridge_LoginDeleteLogin(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, dialer *bridge.TestDialer, locator bridge.Locator, storeKey []byte) {
		withBridge(t, ctx, s.GetHostURL(), dialer, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Login the user.
			userID := must(bridge.LoginUser(ctx, username, password, nil, nil))

			// The user is now connected.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Equal(t, []string{userID}, getConnectedUserIDs(t, bridge))

			// Delete the user.
			require.NoError(t, bridge.DeleteUser(ctx, userID))

			// The user is now gone.
			require.Empty(t, bridge.GetUserIDs())
			require.Empty(t, getConnectedUserIDs(t, bridge))

			// Login the user again.
			newUserID := must(bridge.LoginUser(ctx, username, password, nil, nil))
			require.Equal(t, userID, newUserID)

			// The user is connected again.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Equal(t, []string{userID}, getConnectedUserIDs(t, bridge))
		})
	})
}

func TestBridge_LoginDeauthLogin(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, dialer *bridge.TestDialer, locator bridge.Locator, storeKey []byte) {
		withBridge(t, ctx, s.GetHostURL(), dialer, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Login the user.
			userID := must(bridge.LoginUser(ctx, username, password, nil, nil))

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
			newUserID := must(bridge.LoginUser(ctx, username, password, nil, nil))
			require.Equal(t, userID, newUserID)

			// The user is connected again.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Equal(t, []string{userID}, getConnectedUserIDs(t, bridge))
		})
	})
}

func TestBridge_LoginExpireLogin(t *testing.T) {
	const authLife = 2 * time.Second

	withEnv(t, func(ctx context.Context, s *server.Server, dialer *bridge.TestDialer, locator bridge.Locator, storeKey []byte) {
		s.SetAuthLife(authLife)

		withBridge(t, ctx, s.GetHostURL(), dialer, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Login the user. Its auth will only be valid for a short time.
			userID := must(bridge.LoginUser(ctx, username, password, nil, nil))

			// Wait until the auth expires.
			time.Sleep(authLife)

			// The user will have to refresh but the logout will still succeed.
			require.NoError(t, bridge.LogoutUser(ctx, userID))
		})
	})
}

func TestBridge_FailToLoad(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, dialer *bridge.TestDialer, locator bridge.Locator, storeKey []byte) {
		var userID string

		// Login the user.
		withBridge(t, ctx, s.GetHostURL(), dialer, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			userID = must(bridge.LoginUser(ctx, username, password, nil, nil))
		})

		// Deauth the user while bridge is stopped.
		require.NoError(t, s.RevokeUser(userID))

		// When bridge starts, the user will not be logged in.
		withBridge(t, ctx, s.GetHostURL(), dialer, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Empty(t, getConnectedUserIDs(t, bridge))
		})
	})
}

func TestBridge_LoadWithoutInternet(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, dialer *bridge.TestDialer, locator bridge.Locator, storeKey []byte) {
		var userID string

		// Login the user.
		withBridge(t, ctx, s.GetHostURL(), dialer, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			userID = must(bridge.LoginUser(ctx, username, password, nil, nil))
		})

		// Simulate loss of internet connection.
		dialer.SetCanDial(false)

		// Start bridge without internet.
		withBridge(t, ctx, s.GetHostURL(), dialer, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Initially, users are not connected.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Empty(t, getConnectedUserIDs(t, bridge))

			// Simulate internet connection.
			dialer.SetCanDial(true)

			// The user will eventually be connected.
			require.Eventually(t, func() bool {
				return len(getConnectedUserIDs(t, bridge)) == 1 && getConnectedUserIDs(t, bridge)[0] == userID
			}, 10*time.Second, time.Second)
		})
	})
}

func TestBridge_LoginRestart(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, dialer *bridge.TestDialer, locator bridge.Locator, storeKey []byte) {
		var userID string

		withBridge(t, ctx, s.GetHostURL(), dialer, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Login the user.
			userID = must(bridge.LoginUser(ctx, username, password, nil, nil))
		})

		withBridge(t, ctx, s.GetHostURL(), dialer, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// The user is still connected.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Equal(t, []string{userID}, getConnectedUserIDs(t, bridge))
		})
	})
}

func TestBridge_LoginLogoutRestart(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, dialer *bridge.TestDialer, locator bridge.Locator, storeKey []byte) {
		var userID string

		withBridge(t, ctx, s.GetHostURL(), dialer, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Login the user.
			userID = must(bridge.LoginUser(ctx, username, password, nil, nil))

			// Logout the user.
			require.NoError(t, bridge.LogoutUser(ctx, userID))
		})

		withBridge(t, ctx, s.GetHostURL(), dialer, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// The user is still disconnected.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Empty(t, getConnectedUserIDs(t, bridge))
		})
	})
}

func TestBridge_LoginDeleteRestart(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, dialer *bridge.TestDialer, locator bridge.Locator, storeKey []byte) {
		var userID string

		withBridge(t, ctx, s.GetHostURL(), dialer, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Login the user.
			userID = must(bridge.LoginUser(ctx, username, password, nil, nil))

			// Delete the user.
			require.NoError(t, bridge.DeleteUser(ctx, userID))
		})

		withBridge(t, ctx, s.GetHostURL(), dialer, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// The user is still gone.
			require.Empty(t, bridge.GetUserIDs())
			require.Empty(t, getConnectedUserIDs(t, bridge))
		})
	})
}

func TestBridge_BridgePass(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, dialer *bridge.TestDialer, locator bridge.Locator, storeKey []byte) {
		var userID, pass string

		withBridge(t, ctx, s.GetHostURL(), dialer, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Login the user.
			userID = must(bridge.LoginUser(ctx, username, password, nil, nil))

			// Retrieve the bridge pass.
			pass = must(bridge.GetUserInfo(userID)).BridgePass

			// Log the user out.
			require.NoError(t, bridge.LogoutUser(ctx, userID))

			// Log the user back in.
			must(bridge.LoginUser(ctx, username, password, nil, nil))

			// The bridge pass should be the same.
			require.Equal(t, pass, pass)
		})

		withBridge(t, ctx, s.GetHostURL(), dialer, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// The bridge should load the user.
			require.Equal(t, []string{userID}, bridge.GetUserIDs())
			require.Equal(t, []string{userID}, getConnectedUserIDs(t, bridge))

			// The bridge pass should be the same.
			require.Equal(t, pass, must(bridge.GetUserInfo(userID)).BridgePass)
		})
	})
}

func TestBridge_AddressMode(t *testing.T) {
	withEnv(t, func(ctx context.Context, s *server.Server, dialer *bridge.TestDialer, locator bridge.Locator, storeKey []byte) {
		withBridge(t, ctx, s.GetHostURL(), dialer, locator, storeKey, func(bridge *bridge.Bridge, mocks *bridge.Mocks) {
			// Login the user.
			userID, err := bridge.LoginUser(ctx, username, password, nil, nil)
			require.NoError(t, err)

			// Get the user's info.
			info, err := bridge.GetUserInfo(userID)
			require.NoError(t, err)

			// The user is in combined mode by default.
			require.Equal(t, vault.CombinedMode, info.AddressMode)

			// Put the user in split mode.
			require.NoError(t, bridge.SetAddressMode(ctx, userID, vault.SplitMode))

			// Get the user's info.
			info, err = bridge.GetUserInfo(userID)
			require.NoError(t, err)

			// The user is in split mode.
			require.Equal(t, vault.SplitMode, info.AddressMode)
		})
	})
}
