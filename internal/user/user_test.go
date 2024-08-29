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

package user

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/go-proton-api/server"
	"github.com/ProtonMail/go-proton-api/server/backend"
	"github.com/ProtonMail/proton-bridge/v3/internal/certs"
	"github.com/ProtonMail/proton-bridge/v3/internal/events"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/imapservice"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/notifications"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/observability"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/smtp"
	"github.com/ProtonMail/proton-bridge/v3/internal/telemetry/mocks"
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

				// User's emails should be correct and their associated display names should be correct
				require.ElementsMatch(t, []string{"username@" + s.GetDomain(), "alias@pm.me"}, user.Emails())
				require.True(t, reflect.DeepEqual(map[string]string{
					"username@" + s.GetDomain(): "username" + " (Display Name)",
					"alias@pm.me":               "alias@pm.me (Display Name)",
				}, user.DisplayNames()))

				// By default, user should be in combined mode.
				require.Equal(t, vault.CombinedMode, user.GetAddressMode())

				// By default, user should have a non-empty bridge password.
				require.NotEmpty(t, user.BridgePass())
			})
		})
	})
}

func TestUser_AddressMode(t *testing.T) {
	withAPI(t, context.Background(), func(ctx context.Context, s *server.Server, m *proton.Manager) {
		withAccount(t, s, "username", "password", []string{}, func(string, []string) {
			withUser(t, ctx, s, m, "username", "password", func(user *User) {
				// By default, user should be in combined mode.
				require.Equal(t, vault.CombinedMode, user.GetAddressMode())

				// User should be able to switch to split mode.
				require.NoError(t, user.SetAddressMode(ctx, vault.SplitMode))
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
	require.NoError(tb, s.ChangeAddressDisplayName(userID, addrID, username+" (Display Name)"))

	addrIDs := []string{addrID}

	for _, email := range aliases {
		addrID, err := s.CreateAddress(userID, email, []byte(password))
		require.NoError(tb, err)
		require.NoError(tb, s.ChangeAddressDisplayName(userID, addrID, email+" (Display Name)"))

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

	v, corrupt, err := vault.New(tb.TempDir(), tb.TempDir(), []byte("my secret key"), nil)
	require.NoError(tb, err)
	require.NoError(tb, corrupt)

	vaultUser, err := v.AddUser(apiUser.ID, username, username+"@pm.me", apiAuth.UID, apiAuth.RefreshToken, saltedKeyPass)
	require.NoError(tb, err)

	ctl := gomock.NewController(tb)
	defer ctl.Finish()

	manager := mocks.NewMockHeartbeatManager(ctl)

	manager.EXPECT().IsTelemetryAvailable(context.Background()).AnyTimes()

	nullEventSubscription := events.NewNullSubscription()
	nullIMAPServerManager := imapservice.NewNullIMAPServerManager()
	nullSMTPServerManager := smtp.NewNullServerManager()

	user, err := New(
		ctx,
		vaultUser,
		client,
		nil,
		apiUser,
		nil,
		true,
		vault.DefaultMaxSyncMemory,
		tb.TempDir(),
		manager,
		nullIMAPServerManager,
		nullSMTPServerManager,
		nullEventSubscription,
		nil,
		observability.NewService(context.Background(), nil),
		"",
		true,
		notifications.NewStore(func() (string, error) {
			return "", nil
		}),
		func(_ string) bool {
			return false
		},
		func(_ proton.ObservabilityMetric) {},
	)
	require.NoError(tb, err)
	defer user.Close()

	fn(user)
}
