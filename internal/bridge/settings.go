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

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/v3/internal/safe"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/userevents"
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
	bridge.usersLock.RLock()

	defer func() {
		logrus.Info("Restarting user event loops")
		for _, u := range bridge.users {
			u.ResumeEventLoop()
		}

		bridge.usersLock.RUnlock()
	}()

	type waiter struct {
		w  *userevents.EventPollWaiter
		id string
	}

	waiters := make([]waiter, 0, len(bridge.users))

	logrus.Info("Pausing user event loops for gluon dir change")
	for id, u := range bridge.users {
		waiters = append(waiters, waiter{w: u.PauseEventLoopWithWaiter(), id: id})
	}

	logrus.Info("Waiting on user event loop completion")
	for _, waiter := range waiters {
		if err := waiter.w.WaitPollFinished(ctx); err != nil {
			logrus.WithError(err).Errorf("Failed to wait on event loop pause for user %v", waiter.id)
			return fmt.Errorf("failed on event loop pause: %w", err)
		}
	}

	logrus.Info("Changing gluon directory")
	return bridge.serverManager.SetGluonDir(ctx, newGluonDir)
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
	if isDisabled {
		bridge.heartbeat.stop()
	} else {
		bridge.heartbeat.start()
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
	useTelemetry := !bridge.GetTelemetryDisabled()
	// Delete all the users.
	safe.Lock(func() {
		for _, user := range bridge.users {
			bridge.logoutUser(ctx, user, true, true, useTelemetry)
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
