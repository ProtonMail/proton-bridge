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

package vault_test

import (
	"math"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/proton-bridge/v3/internal/updater"
	"github.com/ProtonMail/proton-bridge/v3/internal/useragent"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/stretchr/testify/require"
)

func TestVault_Settings_IMAP(t *testing.T) {
	// Create a new test vault.
	s := newVault(t)

	// Check the default IMAP port and SSL setting.
	require.Equal(t, 1143, s.GetIMAPPort())
	require.Equal(t, false, s.GetIMAPSSL())

	// Modify the IMAP port and SSL setting.
	require.NoError(t, s.SetIMAPPort(1234))
	require.NoError(t, s.SetIMAPSSL(true))

	// Check the new IMAP port and SSL setting.
	require.Equal(t, 1234, s.GetIMAPPort())
	require.Equal(t, true, s.GetIMAPSSL())
}

func TestVault_Settings_SMTP(t *testing.T) {
	// Create a new test vault.
	s := newVault(t)

	// Check the default SMTP port and SSL setting.
	require.Equal(t, 1025, s.GetSMTPPort())
	require.Equal(t, false, s.GetSMTPSSL())

	// Modify the SMTP port and SSL setting.
	require.NoError(t, s.SetSMTPPort(1234))
	require.NoError(t, s.SetSMTPSSL(true))

	// Check the new SMTP port and SSL setting.
	require.Equal(t, 1234, s.GetSMTPPort())
	require.Equal(t, true, s.GetSMTPSSL())
}

func TestVault_Settings_GluonDir(t *testing.T) {
	// create a new test vault.
	s, corrupt, err := vault.New(t.TempDir(), "/path/to/gluon", []byte("my secret key"), async.NoopPanicHandler{})
	require.NoError(t, err)
	require.False(t, corrupt)

	// Check the default gluon dir.
	require.Equal(t, "/path/to/gluon", s.GetGluonCacheDir())

	// Modify the gluon dir.
	require.NoError(t, s.SetGluonDir("/tmp/gluon"))

	// Check the new gluon dir.
	require.Equal(t, "/tmp/gluon", s.GetGluonCacheDir())
}

func TestVault_Settings_UpdateChannel(t *testing.T) {
	// create a new test vault.
	s := newVault(t)

	// Check the default update channel.
	require.Equal(t, updater.StableChannel, s.GetUpdateChannel())

	// Modify the update channel.
	require.NoError(t, s.SetUpdateChannel(updater.EarlyChannel))

	// Check the new update channel.
	require.Equal(t, updater.EarlyChannel, s.GetUpdateChannel())
}

func TestVault_Settings_UpdateRollout(t *testing.T) {
	// create a new test vault.
	s := newVault(t)

	// Check the default update rollout.
	require.GreaterOrEqual(t, s.GetUpdateRollout(), float64(0))
	require.LessOrEqual(t, s.GetUpdateRollout(), float64(1))

	// Modify the update rollout.
	require.NoError(t, s.SetUpdateRollout(0.5))

	// Check the new update rollout.
	require.Equal(t, float64(0.5), s.GetUpdateRollout())

	// Since GODT-2319 0.6046602879796196 is not allowed as a rollout value (RNG was not seeded)
	require.NoError(t, s.SetUpdateRollout(vault.ForbiddenRollout))
	require.GreaterOrEqual(t, math.Abs(s.GetUpdateRollout()-vault.ForbiddenRollout), 0.00000001)
}

func TestVault_Settings_ColorScheme(t *testing.T) {
	// create a new test vault.
	s := newVault(t)

	// Check the default color scheme.
	require.Equal(t, "", s.GetColorScheme())

	// Modify the color scheme.
	require.NoError(t, s.SetColorScheme("dark"))

	// Check the new color scheme.
	require.Equal(t, "dark", s.GetColorScheme())
}

func TestVault_Settings_ProxyAllowed(t *testing.T) {
	// create a new test vault.
	s := newVault(t)

	// Check the default proxy allowed setting.
	require.Equal(t, false, s.GetProxyAllowed())

	// Modify the proxy allowed setting.
	require.NoError(t, s.SetProxyAllowed(true))

	// Check the new proxy allowed setting.
	require.Equal(t, true, s.GetProxyAllowed())
}

func TestVault_Settings_ShowAllMail(t *testing.T) {
	// create a new test vault.
	s := newVault(t)

	// Check the default show all mail setting.
	require.Equal(t, true, s.GetShowAllMail())

	// Modify the show all mail setting.
	require.NoError(t, s.SetShowAllMail(false))

	// Check the new show all mail setting.
	require.Equal(t, false, s.GetShowAllMail())
}

func TestVault_Settings_TelemetryDisabled(t *testing.T) {
	// create a new test vault.
	s := newVault(t)

	// Check the default show all mail setting.
	require.Equal(t, false, s.GetTelemetryDisabled())

	// Modify the show all mail setting.
	require.NoError(t, s.SetTelemetryDisabled(true))

	// Check the new show all mail setting.
	require.Equal(t, true, s.GetTelemetryDisabled())
}

func TestVault_Settings_Autostart(t *testing.T) {
	// create a new test vault.
	s := newVault(t)

	// Check the default autostart setting.
	require.Equal(t, true, s.GetAutostart())

	// Modify the autostart setting.
	require.NoError(t, s.SetAutostart(false))

	// Check the new autostart setting.
	require.Equal(t, false, s.GetAutostart())
}

func TestVault_Settings_AutoUpdate(t *testing.T) {
	// create a new test vault.
	s := newVault(t)

	// Check the default auto update setting.
	require.Equal(t, true, s.GetAutoUpdate())

	// Modify the auto update setting.
	require.NoError(t, s.SetAutoUpdate(false))

	// Check the new auto update setting.
	require.Equal(t, false, s.GetAutoUpdate())
}

func TestVault_Settings_LastVersion(t *testing.T) {
	// create a new test vault.
	s := newVault(t)

	// Check the default first start value.
	require.True(t, semver.MustParse("0.0.0").Equal(s.GetLastVersion()))

	// Modify the first start value.
	require.NoError(t, s.SetLastVersion(semver.MustParse("1.2.3")))

	// Check the new first start value.
	require.True(t, semver.MustParse("1.2.3").Equal(s.GetLastVersion()))
}

func TestVault_Settings_FirstStart(t *testing.T) {
	// create a new test vault.
	s := newVault(t)

	// Check the default first start value.
	require.Equal(t, true, s.GetFirstStart())

	// Modify the first start value.
	require.NoError(t, s.SetFirstStart(false))

	// Check the new first start value.
	require.Equal(t, false, s.GetFirstStart())
}

func TestVault_Settings_MaxSyncMemory(t *testing.T) {
	// create a new test vault.
	s := newVault(t)

	// Check the default first start value.
	require.Equal(t, vault.DefaultMaxSyncMemory, s.GetMaxSyncMemory())
}

func TestVault_Settings_LastUserAgent(t *testing.T) {
	// create a new test vault.
	s := newVault(t)

	// Check the default first start value.
	require.Equal(t, useragent.DefaultUserAgent, s.GetLastUserAgent())
}

func Test_Settings_PasswordArchive(t *testing.T) {
	// Create a new test vault.
	s := newVault(t)

	// The store should have no users.
	require.Empty(t, s.GetUserIDs())

	// Create a new user.
	user, err := s.AddUser("userID1", "username1", "username1@pm.me", "authUID1", "authRef1", []byte("keyPass1"))
	require.NoError(t, err)
	bridgePass := user.BridgePass()

	// Remove the user.
	require.NoError(t, user.Close())
	require.NoError(t, s.DeleteUser("userID1"))

	// Add a different user. Another password is generated.
	user, err = s.AddUser("userID2", "username2", "username2@pm.me", "authUID2", "authRef2", []byte("keyPass2"))
	require.NoError(t, err)
	require.NotEqual(t, user.BridgePass(), bridgePass)

	// Add the first user again. The password is restored.
	user, err = s.AddUser("userID1", "username1", "username1@pm.me", "authUID1", "authRef1", []byte("keyPass1"))
	require.NoError(t, err)
	require.Equal(t, user.BridgePass(), bridgePass)
}
