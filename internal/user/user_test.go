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
	user.DefaultEventPeriod = 100 * time.Millisecond
	user.DefaultEventJitter = 0
	backend.GenerateKey = tests.FastGenerateKey
	certs.GenerateCert = tests.FastGenerateCert
}

func TestUser_Data(t *testing.T) {
	withAPI(t, context.Background(), "username", "password", []string{"email@pm.me", "alias@pm.me"}, func(ctx context.Context, s *server.Server, userID string, addrIDs []string) {
		withUser(t, ctx, s.GetHostURL(), "username", "password", func(user *user.User) {
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
}

func TestUser_Sync(t *testing.T) {
	withAPI(t, context.Background(), "username", "password", []string{"email@pm.me"}, func(ctx context.Context, s *server.Server, userID string, addrIDs []string) {
		withUser(t, ctx, s.GetHostURL(), "username", "password", func(user *user.User) {
			// Get the user's IMAP connectors.
			imapConn, err := user.NewIMAPConnectors()
			require.NoError(t, err)

			// Pretend to be gluon applying all the updates.
			go func() {
				for _, imapConn := range imapConn {
					for update := range imapConn.GetUpdates() {
						update.Done()
					}
				}
			}()

			// Trigger a user sync.
			errCh := user.DoSync(ctx)

			// User starts a sync at startup.
			require.IsType(t, events.SyncStarted{}, <-user.GetEventCh())

			// User finishes a sync at startup.
			require.IsType(t, events.SyncFinished{}, <-user.GetEventCh())

			// The sync completes without error.
			require.NoError(t, <-errCh)
		})
	})
}

func TestUser_Deauth(t *testing.T) {
	withAPI(t, context.Background(), "username", "password", []string{"email@pm.me"}, func(ctx context.Context, s *server.Server, userID string, addrIDs []string) {
		withUser(t, ctx, s.GetHostURL(), "username", "password", func(user *user.User) {
			eventCh := user.GetEventCh()

			// Revoke the user's auth token.
			require.NoError(t, s.RevokeUser(userID))

			// The user should eventually be logged out.
			require.Eventually(t, func() bool { _, ok := (<-eventCh).(events.UserDeauth); return ok }, 5*time.Second, 100*time.Millisecond)
		})
	})
}

func withAPI(t *testing.T, ctx context.Context, username, password string, emails []string, fn func(context.Context, *server.Server, string, []string)) {
	server := server.New()
	defer server.Close()

	var addrIDs []string

	userID, addrID, err := server.CreateUser(username, password, emails[0])
	require.NoError(t, err)

	addrIDs = append(addrIDs, addrID)

	for _, email := range emails[1:] {
		addrID, err := server.CreateAddress(userID, email, password)
		require.NoError(t, err)

		addrIDs = append(addrIDs, addrID)
	}

	fn(ctx, server, userID, addrIDs)
}

func withUser(t *testing.T, ctx context.Context, apiURL, username, password string, fn func(*user.User)) {
	c, apiAuth, err := liteapi.New(liteapi.WithHostURL(apiURL)).NewClientWithLogin(ctx, username, []byte(password))
	require.NoError(t, err)
	defer func() { require.NoError(t, c.Close()) }()

	apiUser, apiAddrs, userKR, addrKRs, passphrase, err := c.Unlock(ctx, []byte(password))
	require.NoError(t, err)

	vault, corrupt, err := vault.New(t.TempDir(), t.TempDir(), []byte("my secret key"))
	require.NoError(t, err)
	require.False(t, corrupt)

	vaultUser, err := vault.AddUser(apiUser.ID, username, apiAuth.UID, apiAuth.RefreshToken, passphrase)
	require.NoError(t, err)

	user, err := user.New(ctx, vaultUser, c, apiUser, apiAddrs, userKR, addrKRs)
	require.NoError(t, err)
	defer func() { require.NoError(t, user.Close(ctx)) }()

	fn(user)
}
