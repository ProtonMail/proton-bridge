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
	"net"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/internal/safe"
	"github.com/ProtonMail/proton-bridge/v3/internal/updater"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	"github.com/ProtonMail/proton-bridge/v3/pkg/keychain"
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
	return safe.RLockRet(func() error {
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

		imapServer, err := newIMAPServer(
			bridge.vault.GetGluonDir(),
			bridge.curVersion,
			bridge.tlsConfig,
			bridge.reporter,
			bridge.logIMAPClient,
			bridge.logIMAPServer,
			bridge.imapEventCh,
			bridge.tasks,
		)
		if err != nil {
			return fmt.Errorf("failed to create new IMAP server: %w", err)
		}

		bridge.imapServer = imapServer

		for _, user := range bridge.users {
			if err := bridge.addIMAPUser(ctx, user); err != nil {
				return fmt.Errorf("failed to add users to new IMAP server: %w", err)
			}
		}

		if err := bridge.serveIMAP(); err != nil {
			return fmt.Errorf("failed to serve IMAP: %w", err)
		}

		return nil
	}, bridge.usersLock)
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
	return safe.RLockRet(func() error {
		for _, user := range bridge.users {
			user.SetShowAllMail(show)
		}

		return bridge.vault.SetShowAllMail(show)
	}, bridge.usersLock)
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

	bridge.goUpdate()

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

	bridge.goUpdate()

	return nil
}

func (bridge *Bridge) GetCurrentVersion() *semver.Version {
	return bridge.curVersion
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
	// Delete all the users.
	safe.Lock(func() {
		for _, user := range bridge.users {
			bridge.logoutUser(ctx, user, true, true)
		}
	}, bridge.usersLock)

	// Wipe the vault.
	gluonDir, err := bridge.locator.ProvideGluonPath()
	if err != nil {
		logrus.WithError(err).Error("Failed to provide gluon dir")
	} else if err := bridge.vault.Reset(gluonDir); err != nil {
		logrus.WithError(err).Error("Failed to reset vault")
	}

	// Then delete all files.
	if err := bridge.locator.Clear(); err != nil {
		logrus.WithError(err).Error("Failed to clear data paths")
	}

	// Lastly clear the keychain.
	vaultDir, err := bridge.locator.ProvideSettingsPath()
	if err != nil {
		logrus.WithError(err).Error("Failed to get vault dir")
	} else if helper, err := vault.GetHelper(vaultDir); err != nil {
		logrus.WithError(err).Error("Failed to get keychain helper")
	} else if keychain, err := keychain.NewKeychain(helper, constants.KeyChainName); err != nil {
		logrus.WithError(err).Error("Failed to get keychain")
	} else if err := keychain.Clear(); err != nil {
		logrus.WithError(err).Error("Failed to clear keychain")
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
