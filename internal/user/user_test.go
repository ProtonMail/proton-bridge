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

package user

import (
	"context"
	"testing"
	"time"

	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/go-proton-api/server"
	"github.com/ProtonMail/go-proton-api/server/backend"
	"github.com/ProtonMail/proton-bridge/v3/internal/bridge/mocks"
	"github.com/ProtonMail/proton-bridge/v3/internal/certs"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/ProtonMail/proton-bridge/v3/tests"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func init() {
	EventPeriod = 100 * time.Millisecond
	EventJitter = 0
	backend.GenerateKey = backend.FastGenerateKey
	certs.GenerateCert = tests.FastGenerateCert
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreCurrent())
}

func TestUser_Info(t *testing.T) {
	withAPI(t, context.Background(), func(ctx context.Context, s *server.Server, m *proton.Manager) {
		withAccount(t, s, "username", "password", []string{"alias@pm.me"}, func(userID string, _ []string) {
			withUser(t, ctx, s, m, "username", "password", func(user *User) {
				// User's ID should be correct.
				require.Equal(t, userID, user.ID())

				// User's name should be correct.
				require.Equal(t, "username", user.Name())

				// User's email should be correct.
				require.ElementsMatch(t, []string{"username@" + s.GetDomain(), "alias@pm.me"}, user.Emails())

				// By default, user should be in combined mode.
				require.Equal(t, vault.CombinedMode, user.GetAddressMode())

				// By default, user should have a non-empty bridge password.
				require.NotEmpty(t, user.BridgePass())
			})
		})
	})
}

func TestUser_Sync(t *testing.T) {
	withAPI(t, context.Background(), func(ctx context.Context, s *server.Server, m *proton.Manager) {
		withAccount(t, s, "username", "password", []string{}, func(string, []string) {
			withUser(t, ctx, s, m, "username", "password", func(user *User) {
				// User starts a sync at startup.
				require.IsType(t, events.SyncStarted{}, <-user.GetEventCh())

				// User sends sync progress.
				require.IsType(t, events.SyncProgress{}, <-user.GetEventCh())

				// User finishes a sync at startup.
				require.IsType(t, events.SyncFinished{}, <-user.GetEventCh())
			})
		})
	})
}

func TestUser_AddressMode(t *testing.T) {
	withAPI(t, context.Background(), func(ctx context.Context, s *server.Server, m *proton.Manager) {
		withAccount(t, s, "username", "password", []string{}, func(string, []string) {
			withUser(t, ctx, s, m, "username", "password", func(user *User) {
				// User finishes syncing at startup.
				require.IsType(t, events.SyncStarted{}, <-user.GetEventCh())
				require.IsType(t, events.SyncProgress{}, <-user.GetEventCh())
				require.IsType(t, events.SyncFinished{}, <-user.GetEventCh())

				// By default, user should be in combined mode.
				require.Equal(t, vault.CombinedMode, user.GetAddressMode())

				// User should be able to switch to split mode.
				require.NoError(t, user.SetAddressMode(ctx, vault.SplitMode))

				// Create a new set of IMAP connectors (after switching to split mode).
				imapConn, err := user.NewIMAPConnectors()
				require.NoError(t, err)

				// Process updates from the new set of IMAP connectors.
				for _, imapConn := range imapConn {
					go func(imapConn connector.Connector) {
						for update := range imapConn.GetUpdates() {
							update.Done()
						}
					}(imapConn)
				}

				// User finishes syncing after switching to split mode.
				require.IsType(t, events.SyncStarted{}, <-user.GetEventCh())
				require.IsType(t, events.SyncProgress{}, <-user.GetEventCh())
				require.IsType(t, events.SyncFinished{}, <-user.GetEventCh())
			})
		})
	})
}

func TestUser_Deauth(t *testing.T) {
	withAPI(t, context.Background(), func(ctx context.Context, s *server.Server, m *proton.Manager) {
		withAccount(t, s, "username", "password", []string{}, func(string, []string) {
			withUser(t, ctx, s, m, "username", "password", func(user *User) {
				require.IsType(t, events.SyncStarted{}, <-user.GetEventCh())
				require.IsType(t, events.SyncProgress{}, <-user.GetEventCh())
				require.IsType(t, events.SyncFinished{}, <-user.GetEventCh())

				// Revoke the user's auth token.
				require.NoError(t, s.RevokeUser(user.ID()))

				// The user should eventually be logged out.
				require.Eventually(t, func() bool { _, ok := (<-user.GetEventCh()).(events.UserDeauth); return ok }, 500*time.Second, 100*time.Millisecond)
			})
		})
	})
}

func TestUser_Refresh(t *testing.T) {
	ctl := gomock.NewController(t)
	mockReporter := mocks.NewMockReporter(ctl)

	withAPI(t, context.Background(), func(ctx context.Context, s *server.Server, m *proton.Manager) {
		withAccount(t, s, "username", "password", []string{}, func(string, []string) {
			withUser(t, ctx, s, m, "username", "password", func(user *User) {
				require.IsType(t, events.SyncStarted{}, <-user.GetEventCh())
				require.IsType(t, events.SyncProgress{}, <-user.GetEventCh())
				require.IsType(t, events.SyncFinished{}, <-user.GetEventCh())

				user.reporter = mockReporter

				mockReporter.EXPECT().ReportMessageWithContext(
					gomock.Eq("Warning: refresh occurred"),
					mocks.NewRefreshContextMatcher(proton.RefreshAll),
				).Return(nil)

				// Send refresh event
				require.NoError(t, s.RefreshUser(user.ID(), proton.RefreshAll))

				// The user should eventually be re-synced.
				require.Eventually(t, func() bool { _, ok := (<-user.GetEventCh()).(events.UserRefreshed); return ok }, 5*time.Second, 100*time.Millisecond)
			})
		})
	})
}

func withAPI(_ testing.TB, ctx context.Context, fn func(context.Context, *server.Server, *proton.Manager)) { //nolint:revive
	server := server.New()
	defer server.Close()

	fn(ctx, server, proton.New(
		proton.WithHostURL(server.GetHostURL()),
		proton.WithTransport(proton.InsecureTransport()),
	))
}

func withAccount(tb testing.TB, s *server.Server, username, password string, aliases []string, fn func(string, []string)) { //nolint:unparam
	userID, addrID, err := s.CreateUser(username, []byte(password))
	require.NoError(tb, err)

	addrIDs := []string{addrID}

	for _, email := range aliases {
		addrID, err := s.CreateAddress(userID, email, []byte(password))
		require.NoError(tb, err)

		addrIDs = append(addrIDs, addrID)
	}

	fn(userID, addrIDs)
}

func withUser(tb testing.TB, ctx context.Context, _ *server.Server, m *proton.Manager, username, password string, fn func(*User)) { //nolint:unparam,revive
	client, apiAuth, err := m.NewClientWithLogin(ctx, username, []byte(password))
	require.NoError(tb, err)

	apiUser, err := client.GetUser(ctx)
	require.NoError(tb, err)

	salts, err := client.GetSalts(ctx)
	require.NoError(tb, err)

	saltedKeyPass, err := salts.SaltForKey([]byte(password), apiUser.Keys.Primary().ID)
	require.NoError(tb, err)

	vault, corrupt, err := vault.New(tb.TempDir(), tb.TempDir(), []byte("my secret key"))
	require.NoError(tb, err)
	require.False(tb, corrupt)

	vaultUser, err := vault.AddUser(apiUser.ID, username, apiAuth.UID, apiAuth.RefreshToken, saltedKeyPass)
	require.NoError(tb, err)

	user, err := New(ctx, vaultUser, client, nil, apiUser, nil, vault.SyncWorkers(), true)
	require.NoError(tb, err)
	defer user.Close()

	imapConn, err := user.NewIMAPConnectors()
	require.NoError(tb, err)

	for _, imapConn := range imapConn {
		go func(imapConn connector.Connector) {
			for update := range imapConn.GetUpdates() {
				update.Done()
			}
		}(imapConn)
	}

	fn(user)
}
