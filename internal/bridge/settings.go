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

package bridge

import (
	"context"
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
	"github.com/ProtonMail/proton-bridge/v2/internal/user"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
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

	return vault.SetHelper(vaultDir, helper)
}

func (bridge *Bridge) GetIMAPPort() int {
	return bridge.vault.GetIMAPPort()
}

func (bridge *Bridge) SetIMAPPort(newPort int) error {
	if newPort == bridge.vault.GetIMAPPort() {
		return nil
	}

	if err := bridge.vault.SetIMAPPort(newPort); err != nil {
		return err
	}

	return bridge.restartIMAP()
}

func (bridge *Bridge) GetIMAPSSL() bool {
	return bridge.vault.GetIMAPSSL()
}

func (bridge *Bridge) SetIMAPSSL(newSSL bool) error {
	if newSSL == bridge.vault.GetIMAPSSL() {
		return nil
	}

	if err := bridge.vault.SetIMAPSSL(newSSL); err != nil {
		return err
	}

	return bridge.restartIMAP()
}

func (bridge *Bridge) GetSMTPPort() int {
	return bridge.vault.GetSMTPPort()
}

func (bridge *Bridge) SetSMTPPort(newPort int) error {
	if newPort == bridge.vault.GetSMTPPort() {
		return nil
	}

	if err := bridge.vault.SetSMTPPort(newPort); err != nil {
		return err
	}

	return bridge.restartSMTP()
}

func (bridge *Bridge) GetSMTPSSL() bool {
	return bridge.vault.GetSMTPSSL()
}

func (bridge *Bridge) SetSMTPSSL(newSSL bool) error {
	if newSSL == bridge.vault.GetSMTPSSL() {
		return nil
	}

	if err := bridge.vault.SetSMTPSSL(newSSL); err != nil {
		return err
	}

	return bridge.restartSMTP()
}

func (bridge *Bridge) GetGluonDir() string {
	return bridge.vault.GetGluonDir()
}

func (bridge *Bridge) SetGluonDir(ctx context.Context, newGluonDir string) error {
	if newGluonDir == bridge.GetGluonDir() {
		return fmt.Errorf("new gluon dir is the same as the old one")
	}

	if err := bridge.closeIMAP(context.Background()); err != nil {
		return fmt.Errorf("failed to close IMAP: %w", err)
	}

	if err := moveDir(bridge.GetGluonDir(), newGluonDir); err != nil {
		return fmt.Errorf("failed to move gluon dir: %w", err)
	}

	if err := bridge.vault.SetGluonDir(newGluonDir); err != nil {
		return fmt.Errorf("failed to set new gluon dir: %w", err)
	}

	imapServer, err := newIMAPServer(bridge.vault.GetGluonDir(), bridge.curVersion, bridge.tlsConfig, bridge.logIMAPClient, bridge.logIMAPServer)
	if err != nil {
		return fmt.Errorf("failed to create new IMAP server: %w", err)
	}

	bridge.imapServer = imapServer

	if err := bridge.users.IterValuesErr(func(user *user.User) error {
		return bridge.addIMAPUser(ctx, user)
	}); err != nil {
		return fmt.Errorf("failed to add users to new IMAP server: %w", err)
	}

	if err := bridge.serveIMAP(); err != nil {
		return fmt.Errorf("failed to serve IMAP: %w", err)
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

	return bridge.vault.SetProxyAllowed(allowed)
}

func (bridge *Bridge) GetShowAllMail() bool {
	return bridge.vault.GetShowAllMail()
}

func (bridge *Bridge) SetShowAllMail(show bool) error {
	panic("TODO")
}

func (bridge *Bridge) GetAutostart() bool {
	return bridge.vault.GetAutostart()
}

func (bridge *Bridge) SetAutostart(autostart bool) error {
	if err := bridge.vault.SetAutostart(autostart); err != nil {
		return err
	}

	var err error

	if autostart {
		err = bridge.autostarter.Enable()
	} else {
		err = bridge.autostarter.Disable()
	}

	return err
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

	bridge.updateCheckCh <- struct{}{}

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

	bridge.updateCheckCh <- struct{}{}

	return nil
}

func (bridge *Bridge) GetLastVersion() *semver.Version {
	return bridge.vault.GetLastVersion()
}

func (bridge *Bridge) GetFirstStart() bool {
	return bridge.vault.GetFirstStart()
}

func (bridge *Bridge) SetFirstStart(firstStart bool) error {
	return bridge.vault.SetFirstStart(firstStart)
}

func (bridge *Bridge) GetFirstStartGUI() bool {
	return bridge.vault.GetFirstStartGUI()
}

func (bridge *Bridge) SetFirstStartGUI(firstStart bool) error {
	return bridge.vault.SetFirstStartGUI(firstStart)
}

func (bridge *Bridge) GetColorScheme() string {
	return bridge.vault.GetColorScheme()
}

func (bridge *Bridge) SetColorScheme(colorScheme string) error {
	return bridge.vault.SetColorScheme(colorScheme)
}

func (bridge *Bridge) FactoryReset(ctx context.Context) {
	// First delete all users.
	for _, userID := range bridge.GetUserIDs() {
		if bridge.users.Has(userID) {
			if err := bridge.DeleteUser(ctx, userID); err != nil {
				logrus.WithError(err).Errorf("Failed to delete user %s", userID)
			}
		}
	}

	// Then delete all files.
	if err := bridge.locator.Clear(); err != nil {
		logrus.WithError(err).Error("Failed to clear data paths")
	}
}
