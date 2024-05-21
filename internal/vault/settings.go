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

package vault

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/internal/updater"
	"github.com/ProtonMail/proton-bridge/v3/internal/useragent"
	"github.com/sirupsen/logrus"
)

const (
	ForbiddenRollout = 0.6046602879796196
)

// GetIMAPPort sets the port that the IMAP server should listen on.
func (vault *Vault) GetIMAPPort() int {
	return vault.getSafe().Settings.IMAPPort
}

// SetIMAPPort sets the port that the IMAP server should listen on.
func (vault *Vault) SetIMAPPort(port int) error {
	return vault.modSafe(func(data *Data) {
		data.Settings.IMAPPort = port
	})
}

// GetSMTPPort sets the port that the SMTP server should listen on.
func (vault *Vault) GetSMTPPort() int {
	return vault.getSafe().Settings.SMTPPort
}

// SetSMTPPort sets the port that the SMTP server should listen on.
func (vault *Vault) SetSMTPPort(port int) error {
	return vault.modSafe(func(data *Data) {
		data.Settings.SMTPPort = port
	})
}

// GetIMAPSSL sets whether the IMAP server should use SSL.
func (vault *Vault) GetIMAPSSL() bool {
	return vault.getSafe().Settings.IMAPSSL
}

// SetIMAPSSL sets whether the IMAP server should use SSL.
func (vault *Vault) SetIMAPSSL(ssl bool) error {
	return vault.modSafe(func(data *Data) {
		data.Settings.IMAPSSL = ssl
	})
}

// GetSMTPSSL sets whether the SMTP server should use SSL.
func (vault *Vault) GetSMTPSSL() bool {
	return vault.getSafe().Settings.SMTPSSL
}

// SetSMTPSSL sets whether the SMTP server should use SSL.
func (vault *Vault) SetSMTPSSL(ssl bool) error {
	return vault.modSafe(func(data *Data) {
		data.Settings.SMTPSSL = ssl
	})
}

// GetGluonCacheDir sets the directory where the gluon should store its data.
func (vault *Vault) GetGluonCacheDir() string {
	return vault.getSafe().Settings.GluonDir
}

// SetGluonDir sets the directory where the gluon should store its data.
func (vault *Vault) SetGluonDir(dir string) error {
	return vault.modSafe(func(data *Data) {
		data.Settings.GluonDir = dir
	})
}

// GetUpdateChannel sets the update channel.
func (vault *Vault) GetUpdateChannel() updater.Channel {
	return vault.getSafe().Settings.UpdateChannel
}

// SetUpdateChannel sets the update channel.
func (vault *Vault) SetUpdateChannel(channel updater.Channel) error {
	return vault.modSafe(func(data *Data) {
		data.Settings.UpdateChannel = channel
	})
}

// GetUpdateRollout sets the update rollout.
func (vault *Vault) GetUpdateRollout() float64 {
	// The rollout value 0.6046602879796196 is forbidden. The RNG was not seeded when it was picked (GODT-2319).
	rollout := vault.getSafe().Settings.UpdateRollout
	if math.Abs(rollout-ForbiddenRollout) >= 0.00000001 {
		return rollout
	}

	rollout = rand.Float64() //nolint:gosec
	if err := vault.SetUpdateRollout(rollout); err != nil {
		logrus.WithError(err).Warning("Failed writing updateRollout value in vault")
	}
	return rollout
}

// SetUpdateRollout sets the update rollout.
func (vault *Vault) SetUpdateRollout(rollout float64) error {
	return vault.modSafe(func(data *Data) {
		data.Settings.UpdateRollout = rollout
	})
}

// GetColorScheme sets the color scheme to be used by the bridge GUI.
func (vault *Vault) GetColorScheme() string {
	return vault.getSafe().Settings.ColorScheme
}

// SetColorScheme sets the color scheme to be used by the bridge GUI.
func (vault *Vault) SetColorScheme(colorScheme string) error {
	return vault.modSafe(func(data *Data) {
		data.Settings.ColorScheme = colorScheme
	})
}

// GetProxyAllowed sets whether the bridge is allowed to use alternative routing.
func (vault *Vault) GetProxyAllowed() bool {
	return vault.getSafe().Settings.ProxyAllowed
}

// SetProxyAllowed sets whether the bridge is allowed to use alternative routing.
func (vault *Vault) SetProxyAllowed(allowed bool) error {
	return vault.modSafe(func(data *Data) {
		data.Settings.ProxyAllowed = allowed
	})
}

// GetShowAllMail sets whether the bridge should show the All Mail folder.
func (vault *Vault) GetShowAllMail() bool {
	return vault.getSafe().Settings.ShowAllMail
}

// SetShowAllMail sets whether the bridge should show the All Mail folder.
func (vault *Vault) SetShowAllMail(showAllMail bool) error {
	return vault.modSafe(func(data *Data) {
		data.Settings.ShowAllMail = showAllMail
	})
}

// GetAutostart sets whether the bridge should autostart.
func (vault *Vault) GetAutostart() bool {
	return vault.getSafe().Settings.Autostart
}

// SetAutostart sets whether the bridge should autostart.
func (vault *Vault) SetAutostart(autostart bool) error {
	return vault.modSafe(func(data *Data) {
		data.Settings.Autostart = autostart
	})
}

// GetAutoUpdate sets whether the bridge should automatically update.
func (vault *Vault) GetAutoUpdate() bool {
	return vault.getSafe().Settings.AutoUpdate
}

// SetAutoUpdate sets whether the bridge should automatically update.
func (vault *Vault) SetAutoUpdate(autoUpdate bool) error {
	return vault.modSafe(func(data *Data) {
		data.Settings.AutoUpdate = autoUpdate
	})
}

// GetTelemetryDisabled checks whether telemetry is disabled.
func (vault *Vault) GetTelemetryDisabled() bool {
	return vault.getSafe().Settings.TelemetryDisabled
}

// SetTelemetryDisabled sets whether telemetry is disabled.
func (vault *Vault) SetTelemetryDisabled(telemetryDisabled bool) error {
	return vault.modSafe(func(data *Data) {
		data.Settings.TelemetryDisabled = telemetryDisabled
	})
}

// GetLastVersion returns the last version of the bridge that was run.
func (vault *Vault) GetLastVersion() *semver.Version {
	lastVersion := vault.getSafe().Settings.LastVersion
	version, err := semver.NewVersion(lastVersion)
	if err != nil {
		logrus.WithError(err).Error(fmt.Sprintf("Error encountered when trying to get last version from vault: %s", lastVersion))
		version, _ = semver.NewVersion(constants.Version)
	}

	return version
}

// SetLastVersion sets the last version of the bridge that was run.
func (vault *Vault) SetLastVersion(version *semver.Version) error {
	return vault.modSafe(func(data *Data) {
		data.Settings.LastVersion = version.String()
	})
}

// GetFirstStart returns whether this is the first time the bridge has been started.
func (vault *Vault) GetFirstStart() bool {
	return vault.getSafe().Settings.FirstStart
}

// SetFirstStart sets whether this is the first time the bridge has been started.
func (vault *Vault) SetFirstStart(firstStart bool) error {
	return vault.modSafe(func(data *Data) {
		data.Settings.FirstStart = firstStart
	})
}

// GetMaxSyncMemory returns the maximum amount of memory the sync process should use.
func (vault *Vault) GetMaxSyncMemory() uint64 {
	v := vault.getSafe().Settings.MaxSyncMemory
	// can be zero if never written to vault before.
	if v == 0 {
		return DefaultMaxSyncMemory
	}

	return v
}

// SetMaxSyncMemory sets the maximum amount of memory the sync process should use.
func (vault *Vault) SetMaxSyncMemory(maxMemory uint64) error {
	return vault.modSafe(func(data *Data) {
		data.Settings.MaxSyncMemory = maxMemory
	})
}

// GetLastUserAgent returns the last user agent recorded by bridge.
func (vault *Vault) GetLastUserAgent() string {
	v := vault.getSafe().Settings.LastUserAgent

	// Handle case where there may be no value.
	if len(v) == 0 {
		v = useragent.DefaultUserAgent
	}

	return v
}

// SetLastUserAgent store the last user agent recorded by bridge.
func (vault *Vault) SetLastUserAgent(userAgent string) error {
	return vault.modSafe(func(data *Data) {
		data.Settings.LastUserAgent = userAgent
	})
}

// GetLastHeartbeatSent returns the last time heartbeat was sent.
func (vault *Vault) GetLastHeartbeatSent() time.Time {
	return vault.getSafe().Settings.LastHeartbeatSent
}

// SetLastHeartbeatSent store the last time heartbeat was sent.
func (vault *Vault) SetLastHeartbeatSent(timestamp time.Time) error {
	return vault.modSafe(func(data *Data) {
		data.Settings.LastHeartbeatSent = timestamp
	})
}
