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

package app

import (
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/ProtonMail/proton-bridge/v2/internal/cookies"
	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
	"github.com/stretchr/testify/require"
)

func TestMigratePrefsToVault(t *testing.T) {
	// Create a new vault.
	vault, corrupt, err := vault.New(t.TempDir(), t.TempDir(), []byte("my secret key"))
	require.NoError(t, err)
	require.False(t, corrupt)

	// load the old prefs file.
	b, err := os.ReadFile(filepath.Join("testdata", "prefs.json"))
	require.NoError(t, err)

	// Migrate the old prefs file to the new vault.
	require.NoError(t, migratePrefsToVault(vault, b))

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
	require.True(t, vault.GetFirstStartGUI())
	require.Equal(t, "blablabla", vault.GetColorScheme())
	require.Equal(t, "2.3.0+git", vault.GetLastVersion().String())
	require.True(t, vault.GetAutostart())

	// Check that the other app settings have been migrated.
	require.Equal(t, 16, vault.SyncWorkers())
	require.Equal(t, 16, vault.SyncBuffer())
	require.Equal(t, 16, vault.SyncAttPool())
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
