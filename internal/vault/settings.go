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

package vault

import (
	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/v2/internal/updater"
)

// GetIMAPPort sets the port that the IMAP server should listen on.
func (vault *Vault) GetIMAPPort() int {
	return vault.get().Settings.IMAPPort
}

// SetIMAPPort sets the port that the IMAP server should listen on.
func (vault *Vault) SetIMAPPort(port int) error {
	return vault.mod(func(data *Data) {
		data.Settings.IMAPPort = port
	})
}

// GetSMTPPort sets the port that the SMTP server should listen on.
func (vault *Vault) GetSMTPPort() int {
	return vault.get().Settings.SMTPPort
}

// SetSMTPPort sets the port that the SMTP server should listen on.
func (vault *Vault) SetSMTPPort(port int) error {
	return vault.mod(func(data *Data) {
		data.Settings.SMTPPort = port
	})
}

// GetIMAPSSL sets whether the IMAP server should use SSL.
func (vault *Vault) GetIMAPSSL() bool {
	return vault.get().Settings.IMAPSSL
}

// SetIMAPSSL sets whether the IMAP server should use SSL.
func (vault *Vault) SetIMAPSSL(ssl bool) error {
	return vault.mod(func(data *Data) {
		data.Settings.IMAPSSL = ssl
	})
}

// GetSMTPSSL sets whether the SMTP server should use SSL.
func (vault *Vault) GetSMTPSSL() bool {
	return vault.get().Settings.SMTPSSL
}

// SetSMTPSSL sets whether the SMTP server should use SSL.
func (vault *Vault) SetSMTPSSL(ssl bool) error {
	return vault.mod(func(data *Data) {
		data.Settings.SMTPSSL = ssl
	})
}

// GetGluonDir sets the directory where the gluon should store its data.
func (vault *Vault) GetGluonDir() string {
	return vault.get().Settings.GluonDir
}

// SetGluonDir sets the directory where the gluon should store its data.
func (vault *Vault) SetGluonDir(dir string) error {
	return vault.mod(func(data *Data) {
		data.Settings.GluonDir = dir
	})
}

// GetUpdateChannel sets the update channel.
func (vault *Vault) GetUpdateChannel() updater.Channel {
	return vault.get().Settings.UpdateChannel
}

// SetUpdateChannel sets the update channel.
func (vault *Vault) SetUpdateChannel(channel updater.Channel) error {
	return vault.mod(func(data *Data) {
		data.Settings.UpdateChannel = channel
	})
}

// GetUpdateRollout sets the update rollout.
func (vault *Vault) GetUpdateRollout() float64 {
	return vault.get().Settings.UpdateRollout
}

// SetUpdateRollout sets the update rollout.
func (vault *Vault) SetUpdateRollout(rollout float64) error {
	return vault.mod(func(data *Data) {
		data.Settings.UpdateRollout = rollout
	})
}

// GetColorScheme sets the color scheme to be used by the bridge GUI.
func (vault *Vault) GetColorScheme() string {
	return vault.get().Settings.ColorScheme
}

// SetColorScheme sets the color scheme to be used by the bridge GUI.
func (vault *Vault) SetColorScheme(colorScheme string) error {
	return vault.mod(func(data *Data) {
		data.Settings.ColorScheme = colorScheme
	})
}

// GetProxyAllowed sets whether the bridge is allowed to use alternative routing.
func (vault *Vault) GetProxyAllowed() bool {
	return vault.get().Settings.ProxyAllowed
}

// SetProxyAllowed sets whether the bridge is allowed to use alternative routing.
func (vault *Vault) SetProxyAllowed(allowed bool) error {
	return vault.mod(func(data *Data) {
		data.Settings.ProxyAllowed = allowed
	})
}

// GetShowAllMail sets whether the bridge should show the All Mail folder.
func (vault *Vault) GetShowAllMail() bool {
	return vault.get().Settings.ShowAllMail
}

// SetShowAllMail sets whether the bridge should show the All Mail folder.
func (vault *Vault) SetShowAllMail(showAllMail bool) error {
	return vault.mod(func(data *Data) {
		data.Settings.ShowAllMail = showAllMail
	})
}

// GetAutostart sets whether the bridge should autostart.
func (vault *Vault) GetAutostart() bool {
	return vault.get().Settings.Autostart
}

// SetAutostart sets whether the bridge should autostart.
func (vault *Vault) SetAutostart(autostart bool) error {
	return vault.mod(func(data *Data) {
		data.Settings.Autostart = autostart
	})
}

// GetAutoUpdate sets whether the bridge should automatically update.
func (vault *Vault) GetAutoUpdate() bool {
	return vault.get().Settings.AutoUpdate
}

// SetAutoUpdate sets whether the bridge should automatically update.
func (vault *Vault) SetAutoUpdate(autoUpdate bool) error {
	return vault.mod(func(data *Data) {
		data.Settings.AutoUpdate = autoUpdate
	})
}

// GetLastVersion returns the last version of the bridge that was run.
func (vault *Vault) GetLastVersion() *semver.Version {
	return semver.MustParse(vault.get().Settings.LastVersion)
}

// SetLastVersion sets the last version of the bridge that was run.
func (vault *Vault) SetLastVersion(version *semver.Version) error {
	return vault.mod(func(data *Data) {
		data.Settings.LastVersion = version.String()
	})
}

// GetFirstStart sets whether this is the first time the bridge has been started.
func (vault *Vault) GetFirstStart() bool {
	return vault.get().Settings.FirstStart
}

// SetFirstStart sets whether this is the first time the bridge has been started.
func (vault *Vault) SetFirstStart(firstStart bool) error {
	return vault.mod(func(data *Data) {
		data.Settings.FirstStart = firstStart
	})
}

// GetFirstStartGUI sets whether this is the first time the bridge GUI has been started.
func (vault *Vault) GetFirstStartGUI() bool {
	return vault.get().Settings.FirstStartGUI
}

// SetFirstStartGUI sets whether this is the first time the bridge GUI has been started.
func (vault *Vault) SetFirstStartGUI(firstStartGUI bool) error {
	return vault.mod(func(data *Data) {
		data.Settings.FirstStartGUI = firstStartGUI
	})
}

// SyncWorkers returns the number of workers to use for syncing.
func (vault *Vault) SyncWorkers() int {
	return vault.get().Settings.SyncWorkers
}

// SyncBuffer returns the number of buffer workers to use for syncing.
func (vault *Vault) SyncBuffer() int {
	return vault.get().Settings.SyncBuffer
}
