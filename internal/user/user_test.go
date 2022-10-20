// Copyright (c) 2022 Proton AG
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

package user_test

import (
	"context"
	"testing"
	"time"

	"github.com/ProtonMail/proton-bridge/v2/internal/certs"
	"github.com/ProtonMail/proton-bridge/v2/internal/events"
	"github.com/ProtonMail/proton-bridge/v2/internal/user"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
	"github.com/ProtonMail/proton-bridge/v2/tests"
	"github.com/stretchr/testify/require"
	"gitlab.protontech.ch/go/liteapi"
	"gitlab.protontech.ch/go/liteapi/server"
	"gitlab.protontech.ch/go/liteapi/server/backend"
)

func init() {
	user.EventPeriod = 100 * time.Millisecond
	user.EventJitter = 0
	backend.GenerateKey = tests.FastGenerateKey
	certs.GenerateCert = tests.FastGenerateCert
}

func TestUser_Data(t *testing.T) {
	withAPI(t, context.Background(), func(ctx context.Context, s *server.Server, m *liteapi.Manager) {
		withAccount(t, s, "username", "password", []string{"email@pm.me", "alias@pm.me"}, func(userID string, addrIDs []string) {
			withUser(t, ctx, s, m, "username", "password", func(user *user.User) {
				// User's ID should be correct.
				require.Equal(t, userID, user.ID())

				// User's name should be correct.
				require.Equal(t, "username", user.Name())

				// User's email should be correct.
				require.ElementsMatch(t, []string{"email@pm.me", "alias@pm.me"}, user.Emails())

				// By default, user should be in combined mode.
				require.Equal(t, vault.CombinedMode, user.GetAddressMode())

				// By default, user should have a non-empty bridge password.
				require.NotEmpty(t, user.BridgePass())
			})
		})
	})
}

func TestUser_Sync(t *testing.T) {
	withAPI(t, context.Background(), func(ctx context.Context, s *server.Server, m *liteapi.Manager) {
		withAccount(t, s, "username", "password", []string{"email@pm.me"}, func(userID string, addrIDs []string) {
			withUser(t, ctx, s, m, "username", "password", func(user *user.User) {
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

func TestUser_Deauth(t *testing.T) {
	withAPI(t, context.Background(), func(ctx context.Context, s *server.Server, m *liteapi.Manager) {
		withAccount(t, s, "username", "password", []string{"email@pm.me"}, func(userID string, addrIDs []string) {
			withUser(t, ctx, s, m, "username", "password", func(user *user.User) {
				eventCh := user.GetEventCh()

				// Revoke the user's auth token.
				require.NoError(t, s.RevokeUser(user.ID()))

				// The user should eventually be logged out.
				require.Eventually(t, func() bool { _, ok := (<-eventCh).(events.UserDeauth); return ok }, 5*time.Second, 100*time.Millisecond)
			})
		})
	})
}

func withAPI(_ *testing.T, ctx context.Context, fn func(context.Context, *server.Server, *liteapi.Manager)) { //nolint:revive
	server := server.New()
	defer server.Close()

	fn(ctx, server, liteapi.New(
		liteapi.WithHostURL(server.GetHostURL()),
		liteapi.WithTransport(liteapi.InsecureTransport()),
	))
}

func withAccount(t *testing.T, s *server.Server, username, password string, emails []string, fn func(string, []string)) {
	userID, addrID, err := s.CreateUser(username, emails[0], []byte(password))
	require.NoError(t, err)

	addrIDs := make([]string, 0, len(emails))

	addrIDs = append(addrIDs, addrID)

	for _, email := range emails[1:] {
		addrID, err := s.CreateAddress(userID, email, []byte(password))
		require.NoError(t, err)

		addrIDs = append(addrIDs, addrID)
	}

	fn(userID, addrIDs)
}

func withUser(t *testing.T, ctx context.Context, _ *server.Server, m *liteapi.Manager, username, password string, fn func(*user.User)) { //nolint:revive
	client, apiAuth, err := m.NewClientWithLogin(ctx, username, []byte(password))
	require.NoError(t, err)
	defer func() { require.NoError(t, client.Close()) }()

	apiUser, err := client.GetUser(ctx)
	require.NoError(t, err)

	salts, err := client.GetSalts(ctx)
	require.NoError(t, err)

	saltedKeyPass, err := salts.SaltForKey([]byte(password), apiUser.Keys.Primary().ID)
	require.NoError(t, err)

	vault, corrupt, err := vault.New(t.TempDir(), t.TempDir(), []byte("my secret key"))
	require.NoError(t, err)
	require.False(t, corrupt)

	vaultUser, err := vault.AddUser(apiUser.ID, username, apiAuth.UID, apiAuth.RefreshToken, saltedKeyPass)
	require.NoError(t, err)

	user, err := user.New(ctx, vaultUser, client, apiUser, true)
	require.NoError(t, err)
	defer func() { require.NoError(t, user.Close()) }()

	imapConn, err := user.NewIMAPConnectors()
	require.NoError(t, err)

	go func() {
		for _, imapConn := range imapConn {
			for update := range imapConn.GetUpdates() {
				update.Done()
			}
		}
	}()

	fn(user)
}
