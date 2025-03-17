// Copyright (c) 2025 Proton AG
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

package app

import (
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v3/internal/bridge"
	"github.com/ProtonMail/proton-bridge/v3/internal/cookies"
	"github.com/ProtonMail/proton-bridge/v3/internal/legacy/credentials"
	"github.com/ProtonMail/proton-bridge/v3/internal/locations"
	"github.com/ProtonMail/proton-bridge/v3/internal/updater"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/ProtonMail/proton-bridge/v3/pkg/algo"
	"github.com/ProtonMail/proton-bridge/v3/pkg/keychain"
	"github.com/stretchr/testify/require"
)

func TestMigratePrefsToVaultWithKeys(t *testing.T) {
	// Create a new vault.
	vault, corrupt, err := vault.New(t.TempDir(), t.TempDir(), []byte("my secret key"), async.NoopPanicHandler{})
	require.NoError(t, err)
	require.NoError(t, corrupt)

	// load the old prefs file.
	configDir := filepath.Join("testdata", "with_keys")

	// Migrate the old prefs file to the new vault.
	require.NoError(t, migrateOldSettingsWithDir(configDir, vault))

	// Check Json Settings
	validateJSONPrefs(t, vault)

	cert, key := vault.GetBridgeTLSCert()
	// Check the keys were found and collected.
	require.Equal(t, "-----BEGIN CERTIFICATE-----", string(cert))
	require.Equal(t, "-----BEGIN RSA PRIVATE KEY-----", string(key))
}

func TestMigratePrefsToVaultWithoutKeys(t *testing.T) {
	// Create a new vault.
	vault, corrupt, err := vault.New(t.TempDir(), t.TempDir(), []byte("my secret key"), async.NoopPanicHandler{})
	require.NoError(t, err)
	require.NoError(t, corrupt)

	// load the old prefs file.
	configDir := filepath.Join("testdata", "without_keys")

	// Migrate the old prefs file to the new vault.
	require.NoError(t, migrateOldSettingsWithDir(configDir, vault))

	// Migrate the old prefs file to the new vault.
	require.NoError(t, migrateOldSettingsWithDir(configDir, vault))

	// Check Json Settings
	validateJSONPrefs(t, vault)

	// Check the keys were found and collected.
	cert, key := vault.GetBridgeTLSCert()
	require.NotEqual(t, []byte("-----BEGIN CERTIFICATE-----"), cert)
	require.NotEqual(t, []byte("-----BEGIN RSA PRIVATE KEY-----"), key)
}

func TestKeychainMigration(t *testing.T) {
	// Migration tested only for linux.
	if runtime.GOOS != "linux" {
		return
	}

	tmpDir := t.TempDir()

	// Prepare for keychain migration test
	{
		require.NoError(t, os.Setenv("XDG_CONFIG_HOME", tmpDir))
		oldCacheDir := filepath.Join(tmpDir, "protonmail", "bridge")
		require.NoError(t, os.MkdirAll(oldCacheDir, 0o700))

		oldPrefs, err := os.ReadFile(filepath.Join("testdata", "without_keys", "protonmail", "bridge", "prefs.json"))
		require.NoError(t, err)

		require.NoError(t, os.WriteFile(
			filepath.Join(oldCacheDir, "prefs.json"),
			oldPrefs, 0o600,
		))
	}

	locations := locations.New(bridge.NewTestLocationsProvider(tmpDir), "config-name")
	settingsFolder, err := locations.ProvideSettingsPath()
	require.NoError(t, err)

	// Check that there is nothing yet
	keychainName, err := vault.GetHelper(settingsFolder)
	require.NoError(t, err)
	require.Equal(t, "", keychainName)

	// Check migration
	require.NoError(t, migrateKeychainHelper(locations))
	keychainName, err = vault.GetHelper(settingsFolder)
	require.NoError(t, err)
	require.Equal(t, "secret-service", keychainName)

	// Change the migrated value
	require.NoError(t, vault.SetHelper(settingsFolder, "different"))

	// Calling migration again will not overwrite existing prefs
	require.NoError(t, migrateKeychainHelper(locations))
	keychainName, err = vault.GetHelper(settingsFolder)
	require.NoError(t, err)
	require.Equal(t, "different", keychainName)
}

func TestUserMigration(t *testing.T) {
	kcl := keychain.NewTestKeychainsList()

	kc, _, err := keychain.NewKeychain("mock", "bridge", kcl.GetHelpers(), kcl.GetDefaultHelper())
	require.NoError(t, err)

	require.NoError(t, kc.Put("brokenID", "broken"))
	require.NoError(t, kc.Put(
		"emptyID",
		(&credentials.Credentials{}).Marshal(),
	))

	wantUID := "uidtoken"
	wantRefresh := "refreshtoken"

	wantCredentials := credentials.Credentials{
		UserID:                "validID",
		Name:                  "user@pm.me",
		Emails:                "user@pm.me;alias@pm.me",
		APIToken:              wantUID + ":" + wantRefresh,
		MailboxPassword:       []byte("secret"),
		BridgePassword:        "bElu2Q1Vusy28J3Wf56cIg",
		Version:               "v2.3.X",
		Timestamp:             100,
		IsCombinedAddressMode: true,
	}
	require.NoError(t, kc.Put(
		wantCredentials.UserID,
		wantCredentials.Marshal(),
	))

	tmpDir := t.TempDir()
	locations := locations.New(bridge.NewTestLocationsProvider(tmpDir), "config-name")
	settingsFolder, err := locations.ProvideSettingsPath()
	require.NoError(t, err)
	require.NoError(t, vault.SetHelper(settingsFolder, "mock"))

	token, err := crypto.RandomToken(32)
	require.NoError(t, err)

	v, corrupt, err := vault.New(settingsFolder, settingsFolder, token, async.NoopPanicHandler{})
	require.NoError(t, err)
	require.NoError(t, corrupt)

	require.NoError(t, migrateOldAccounts(locations, kcl, v))
	require.Equal(t, []string{wantCredentials.UserID}, v.GetUserIDs())

	require.NoError(t, v.GetUser(wantCredentials.UserID, func(u *vault.User) {
		require.Equal(t, wantCredentials.UserID, u.UserID())
		require.Equal(t, wantUID, u.AuthUID())
		require.Equal(t, wantRefresh, u.AuthRef())
		require.Equal(t, wantCredentials.MailboxPassword, u.KeyPass())
		require.Equal(t,
			[]byte(wantCredentials.BridgePassword),
			algo.B64RawEncode(u.BridgePass()),
		)
		require.Equal(t, vault.CombinedMode, u.AddressMode())
	}))
}

func validateJSONPrefs(t *testing.T, vault *vault.Vault) {
	// Check that the IMAP and SMTP prefs are migrated.
	require.Equal(t, 2143, vault.GetIMAPPort())
	require.Equal(t, 2025, vault.GetSMTPPort())
	require.True(t, vault.GetSMTPSSL())

	// Check that the update channel is migrated.
	require.True(t, vault.GetAutoUpdate())
	require.Equal(t, updater.EarlyChannel, vault.GetUpdateChannel())
	require.Equal(t, 0.4849529004202015, vault.GetUpdateRollout())

	// Check that the app settings have been migrated.
	require.False(t, vault.GetFirstStart())
	require.Equal(t, "blablabla", vault.GetColorScheme())
	require.Equal(t, "2.3.0+git", vault.GetLastVersion().String())
	require.True(t, vault.GetAutostart())

	// Check that the other app settings have been migrated.
	require.False(t, vault.GetProxyAllowed())
	require.False(t, vault.GetShowAllMail())

	// Check that the cookies have been migrated.
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	cookies, err := cookies.NewCookieJar(jar, vault)
	require.NoError(t, err)

	url, err := url.Parse("https://api.protonmail.ch")
	require.NoError(t, err)

	// There should be a cookie for the API.
	require.NotEmpty(t, cookies.Cookies(url))
}
