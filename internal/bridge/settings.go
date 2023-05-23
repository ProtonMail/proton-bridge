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

package bridge

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/v3/internal/safe"
	"github.com/ProtonMail/proton-bridge/v3/internal/updater"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/sirupsen/logrus"
)

func (bridge *Bridge) GetKeychainApp() (string, error) {
	vaultDir, err := bridge.locator.ProvideSettingsPath()
	if err != nil {
		return "", err
	}

	return vault.GetHelper(vaultDir)
}

func (bridge *Bridge) SetKeychainApp(helper string) error {
	vaultDir, err := bridge.locator.ProvideSettingsPath()
	if err != nil {
		return err
	}

	bridge.heartbeat.SetKeyChainPref(helper)

	return vault.SetHelper(vaultDir, helper)
}

func (bridge *Bridge) GetIMAPPort() int {
	return bridge.vault.GetIMAPPort()
}

func (bridge *Bridge) SetIMAPPort(ctx context.Context, newPort int) error {
	if newPort == bridge.vault.GetIMAPPort() {
		return nil
	}

	if err := bridge.vault.SetIMAPPort(newPort); err != nil {
		return err
	}

	bridge.heartbeat.SetIMAPPort(newPort)

	return bridge.restartIMAP(ctx)
}

func (bridge *Bridge) GetIMAPSSL() bool {
	return bridge.vault.GetIMAPSSL()
}

func (bridge *Bridge) SetIMAPSSL(ctx context.Context, newSSL bool) error {
	if newSSL == bridge.vault.GetIMAPSSL() {
		return nil
	}

	if err := bridge.vault.SetIMAPSSL(newSSL); err != nil {
		return err
	}

	bridge.heartbeat.SetIMAPConnectionMode(newSSL)

	return bridge.restartIMAP(ctx)
}

func (bridge *Bridge) GetSMTPPort() int {
	return bridge.vault.GetSMTPPort()
}

func (bridge *Bridge) SetSMTPPort(ctx context.Context, newPort int) error {
	if newPort == bridge.vault.GetSMTPPort() {
		return nil
	}

	if err := bridge.vault.SetSMTPPort(newPort); err != nil {
		return err
	}

	bridge.heartbeat.SetSMTPPort(newPort)

	return bridge.restartSMTP(ctx)
}

func (bridge *Bridge) GetSMTPSSL() bool {
	return bridge.vault.GetSMTPSSL()
}

func (bridge *Bridge) SetSMTPSSL(ctx context.Context, newSSL bool) error {
	if newSSL == bridge.vault.GetSMTPSSL() {
		return nil
	}

	if err := bridge.vault.SetSMTPSSL(newSSL); err != nil {
		return err
	}

	bridge.heartbeat.SetSMTPConnectionMode(newSSL)

	return bridge.restartSMTP(ctx)
}

func (bridge *Bridge) GetGluonCacheDir() string {
	return bridge.vault.GetGluonCacheDir()
}

func (bridge *Bridge) GetGluonDataDir() (string, error) {
	return bridge.locator.ProvideGluonDataPath()
}

func (bridge *Bridge) SetGluonDir(ctx context.Context, newGluonDir string) error {
	return bridge.serverManager.SetGluonDir(ctx, newGluonDir)
}

func (bridge *Bridge) moveGluonCacheDir(oldGluonDir, newGluonDir string) error {
	logrus.Infof("gluon cache moving from %s to %s", oldGluonDir, newGluonDir)
	oldCacheDir := ApplyGluonCachePathSuffix(oldGluonDir)
	if err := copyDir(oldCacheDir, ApplyGluonCachePathSuffix(newGluonDir)); err != nil {
		return fmt.Errorf("failed to copy gluon dir: %w", err)
	}

	if err := bridge.vault.SetGluonDir(newGluonDir); err != nil {
		return fmt.Errorf("failed to set new gluon cache dir: %w", err)
	}

	if err := os.RemoveAll(oldCacheDir); err != nil {
		logrus.WithError(err).Error("failed to remove old gluon cache dir")
	}
	return nil
}

func (bridge *Bridge) GetProxyAllowed() bool {
	return bridge.vault.GetProxyAllowed()
}

func (bridge *Bridge) SetProxyAllowed(allowed bool) error {
	if allowed {
		bridge.proxyCtl.AllowProxy()
	} else {
		bridge.proxyCtl.DisallowProxy()
	}

	bridge.heartbeat.SetDoh(allowed)

	return bridge.vault.SetProxyAllowed(allowed)
}

func (bridge *Bridge) GetShowAllMail() bool {
	return bridge.vault.GetShowAllMail()
}

func (bridge *Bridge) SetShowAllMail(show bool) error {
	return safe.RLockRet(func() error {
		for _, user := range bridge.users {
			user.SetShowAllMail(show)
		}

		bridge.heartbeat.SetShowAllMail(show)

		return bridge.vault.SetShowAllMail(show)
	}, bridge.usersLock)
}

func (bridge *Bridge) GetAutostart() bool {
	return bridge.vault.GetAutostart()
}

func (bridge *Bridge) SetAutostart(autostart bool) error {
	if autostart != bridge.vault.GetAutostart() {
		if err := bridge.vault.SetAutostart(autostart); err != nil {
			return err
		}

		bridge.heartbeat.SetAutoStart(autostart)
	}

	var err error
	if autostart {
		// do nothing if already enabled
		if bridge.autostarter.IsEnabled() {
			return nil
		}
		err = bridge.autostarter.Enable()
	} else {
		// do nothing if already disabled
		if !bridge.autostarter.IsEnabled() {
			return nil
		}
		err = bridge.autostarter.Disable()
	}

	return err
}

func (bridge *Bridge) GetUpdateRollout() float64 {
	return bridge.vault.GetUpdateRollout()
}

func (bridge *Bridge) GetAutoUpdate() bool {
	return bridge.vault.GetAutoUpdate()
}

func (bridge *Bridge) SetAutoUpdate(autoUpdate bool) error {
	if bridge.vault.GetAutoUpdate() == autoUpdate {
		return nil
	}

	if err := bridge.vault.SetAutoUpdate(autoUpdate); err != nil {
		return err
	}

	bridge.heartbeat.SetAutoUpdate(autoUpdate)

	bridge.goUpdate()

	return nil
}

func (bridge *Bridge) GetTelemetryDisabled() bool {
	return bridge.vault.GetTelemetryDisabled()
}

func (bridge *Bridge) SetTelemetryDisabled(isDisabled bool) error {
	if err := bridge.vault.SetTelemetryDisabled(isDisabled); err != nil {
		return err
	}
	// If telemetry is re-enabled locally, try to send the heartbeat.
	if !isDisabled {
		defer bridge.goHeartbeat()
	}
	return nil
}

func (bridge *Bridge) GetUpdateChannel() updater.Channel {
	return bridge.vault.GetUpdateChannel()
}

func (bridge *Bridge) SetUpdateChannel(channel updater.Channel) error {
	if bridge.vault.GetUpdateChannel() == channel {
		return nil
	}

	if err := bridge.vault.SetUpdateChannel(channel); err != nil {
		return err
	}

	bridge.heartbeat.SetBeta(channel)

	bridge.goUpdate()

	return nil
}

func (bridge *Bridge) GetCurrentVersion() *semver.Version {
	return bridge.curVersion
}

func (bridge *Bridge) GetLastVersion() *semver.Version {
	return bridge.lastVersion
}

func (bridge *Bridge) GetFirstStart() bool {
	return bridge.firstStart
}

func (bridge *Bridge) GetColorScheme() string {
	return bridge.vault.GetColorScheme()
}

func (bridge *Bridge) SetColorScheme(colorScheme string) error {
	return bridge.vault.SetColorScheme(colorScheme)
}

// FactoryReset deletes all users, wipes the vault, and deletes all files.
// Note: it does not clear the keychain. The only entry in the keychain is the vault password,
// which we need at next startup to decrypt the vault.
func (bridge *Bridge) FactoryReset(ctx context.Context) {
	// Delete all the users.
	safe.Lock(func() {
		for _, user := range bridge.users {
			bridge.logoutUser(ctx, user, true, true)
		}
	}, bridge.usersLock)

	// Wipe the vault.
	gluonCacheDir, err := bridge.locator.ProvideGluonCachePath()
	if err != nil {
		logrus.WithError(err).Error("Failed to provide gluon dir")
	} else if err := bridge.vault.Reset(gluonCacheDir); err != nil {
		logrus.WithError(err).Error("Failed to reset vault")
	}

	// Lastly, delete all files except the vault.
	if err := bridge.locator.Clear(bridge.vault.Path()); err != nil {
		logrus.WithError(err).Error("Failed to clear data paths")
	}
}

func getPort(addr net.Addr) int {
	switch addr := addr.(type) {
	case *net.TCPAddr:
		return addr.Port

	case *net.UDPAddr:
		return addr.Port

	default:
		return 0
	}
}
