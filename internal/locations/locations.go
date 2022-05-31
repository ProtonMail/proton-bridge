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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

// Package locations implements a type that provides cross-platform access to
// standard filesystem locations, including config, cache and log directories.
package locations

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/ProtonMail/proton-bridge/v2/pkg/files"
	"github.com/sirupsen/logrus"
)

// Locations provides cross-platform access to standard locations.
// On linux:
// - settings: ~/.config/protonmail/<app>
// - logs: ~/.cache/protonmail/<app>/logs
// - cache: ~/.config/protonmail/<app>/cache
// - updates: ~/.config/protonmail/<app>/updates
// - lockfile: ~/.cache/protonmail/<app>/<app>.lock .
type Locations struct {
	userConfig, userCache string
	configName            string
}

// New returns a new locations object.
func New(provider Provider, configName string) *Locations {
	return &Locations{
		userConfig: provider.UserConfig(),
		userCache:  provider.UserCache(),
		configName: configName,
	}
}

// GetLockFile returns the path to the lock file (e.g. ~/.cache/<company>/<app>/<app>.lock).
func (l *Locations) GetLockFile() string {
	return filepath.Join(l.userCache, l.configName+".lock")
}

// GetLicenseFilePath returns path to liense file.
func (l *Locations) GetLicenseFilePath() string {
	path := l.getLicenseFilePath()
	logrus.WithField("path", path).Info("License file path")
	return path
}

func (l *Locations) getLicenseFilePath() string {
	// User can install app to different location, or user can run it
	// directly from the package without installation, or it could be
	// automatically updated (app started from differenet location).
	// For all those cases, first let's check LICENSE next to the binary.
	path := filepath.Join(filepath.Dir(os.Args[0]), "LICENSE")
	if _, err := os.Stat(path); err == nil {
		return path
	}

	switch runtime.GOOS {
	case "linux":
		// Most Linux distributions.
		path := "/usr/share/doc/protonmail/" + l.configName + "/LICENSE"
		if _, err := os.Stat(path); err == nil {
			return path
		}
		// Arch distributions.
		return "/usr/share/licenses/protonmail-" + l.configName + "/LICENSE"
	case "darwin": //nolint:goconst
		path := filepath.Join(filepath.Dir(os.Args[0]), "..", "Resources", "LICENSE")
		if _, err := os.Stat(path); err == nil {
			return path
		}

		// This should not happen, macOS should be handled by relative
		// location to the binary above. This is just fallback which may
		// or may not work, depends where user installed the app and how
		// user started the app.
		return "/Applications/Proton Mail Bridge.app/Contents/Resources/LICENSE"
	case "windows":
		path := filepath.Join(filepath.Dir(os.Args[0]), "LICENSE.txt")
		if _, err := os.Stat(path); err == nil {
			return path
		}
		// This should not happen, Windows should be handled by relative
		// location to the binary above. This is just fallback which may
		// or may not work, depends where user installed the app and how
		// user started the app.
		return filepath.FromSlash("C:/Program Files/Proton/Proton Mail Bridge/LICENSE.txt")
	}
	return ""
}

// GetDependencyLicensesLink returns link to page listing dependencies.
func (l *Locations) GetDependencyLicensesLink() string {
	return "https://github.com/ProtonMail/proton-bridge/v2/blob/master/COPYING_NOTES.md#dependencies"
}

// ProvideSettingsPath returns a location for user settings (e.g. ~/.config/<company>/<app>).
// It creates it if it doesn't already exist.
func (l *Locations) ProvideSettingsPath() (string, error) {
	if err := os.MkdirAll(l.getSettingsPath(), 0o700); err != nil {
		return "", err
	}

	return l.getSettingsPath(), nil
}

// ProvideLogsPath returns a location for user logs (e.g. ~/.cache/<company>/<app>/logs).
// It creates it if it doesn't already exist.
func (l *Locations) ProvideLogsPath() (string, error) {
	if err := os.MkdirAll(l.getLogsPath(), 0o700); err != nil {
		return "", err
	}

	return l.getLogsPath(), nil
}

// ProvideCachePath returns a location for user cache dirs (e.g. ~/.config/<company>/<app>/cache).
// It creates it if it doesn't already exist.
func (l *Locations) ProvideCachePath() (string, error) {
	if err := os.MkdirAll(l.getCachePath(), 0o700); err != nil {
		return "", err
	}

	return l.getCachePath(), nil
}

// GetOldCachePath returns a former location for user cache dirs used for migration scripts only.
func (l *Locations) GetOldCachePath() string {
	return filepath.Join(l.userCache, "cache")
}

// ProvideUpdatesPath returns a location for update files (e.g. ~/.cache/<company>/<app>/updates).
// It creates it if it doesn't already exist.
func (l *Locations) ProvideUpdatesPath() (string, error) {
	if err := os.MkdirAll(l.getUpdatesPath(), 0o700); err != nil {
		return "", err
	}

	return l.getUpdatesPath(), nil
}

// GetUpdatesPath returns a new location for update files used for migration scripts only.
func (l *Locations) GetUpdatesPath() string {
	return l.getUpdatesPath()
}

// GetOldUpdatesPath returns a former location for update files used for migration scripts only.
func (l *Locations) GetOldUpdatesPath() string {
	return filepath.Join(l.userCache, "updates")
}

func (l *Locations) getSettingsPath() string {
	return l.userConfig
}

func (l *Locations) getLogsPath() string {
	return filepath.Join(l.userCache, "logs")
}

func (l *Locations) getCachePath() string {
	// Bridge cache is not a typical cache which can be deleted with only
	// downside that the app has to download everything again.
	// Cache for bridge is database with IMAP UIDs and UIDVALIDITY, and also
	// other IMAP setup. Deleting such data leads to either re-sync of client,
	// or mix of headers and bodies. Both is caused because of need of re-sync
	// between Bridge and API which will happen in different order than before.
	// In the first case, UIDVALIDITY is also changed and causes the better
	// outcome to "just" re-sync everything; in the later, UIDVALIDITY stays
	// the same, causing the client to not re-sync but UIDs in the client does
	// not match UIDs in Bridge.
	// Because users might use tools to regularly clear caches, Bridge cache
	// cannot be located in a standard cache folder.
	return filepath.Join(l.userConfig, "cache")
}

func (l *Locations) getUpdatesPath() string {
	// In order to properly update Bridge 1.6.X and higher we need to
	// change the launcher first. Since this is not part of automatic
	// updates the migration must wait until manual update. Until that
	// we need to keep old path.
	if l.configName == "bridge" {
		return l.GetOldUpdatesPath()
	}

	// Users might use tools to regularly clear caches, which would mean always
	// removing updates, therefore Bridge updates have to be somewhere else.
	return filepath.Join(l.userConfig, "updates")
}

// Clear removes everything except the lock and update files.
func (l *Locations) Clear() error {
	return files.Remove(
		l.userConfig,
		l.userCache,
	).Except(
		l.GetLockFile(),
		l.getUpdatesPath(),
	).Do()
}

// ClearUpdates removes update files.
func (l *Locations) ClearUpdates() error {
	return files.Remove(
		l.getUpdatesPath(),
	).Do()
}

// Clean removes any unexpected files from the app cache folder
// while leaving files in the standard locations untouched.
func (l *Locations) Clean() error {
	return files.Remove(l.userCache).Except(
		l.GetLockFile(),
		l.getLogsPath(),
		l.getCachePath(),
		l.getUpdatesPath(),
	).Do()
}
