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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

// Package locations implements a type that provides cross-platform access to
// standard filesystem locations, including config, cache and log directories.
package locations

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/ProtonMail/proton-bridge/v3/pkg/files"
	"github.com/sirupsen/logrus"
)

// Locations provides cross-platform access to standard locations.
// On linux:
// - settings: ~/.config/protonmail/<app>
// - gluon     ~/.local/share/protonmail/<app>/gluon
// - logs:     ~/.local/share/protonmail/<app>/logs
// - updates:  ~/.local/share/protonmail/<app>/updates
// - locks:    ~/.cache/protonmail/<app>/*.lock
// Other OSes are similar.
type Locations struct {
	// userConfig is the path to the user config directory, for storing persistent config data.
	userConfig string

	// userData is the path to the user data directory, for storing persistent data.
	userData string

	// userCache is the path to the user cache directory, for storing non-essential data.
	userCache string

	configName    string
	configGuiName string
}

// New returns a new locations object.
func New(provider Provider, configName string) *Locations {
	return &Locations{
		userConfig: provider.UserConfig(),
		userData:   provider.UserData(),
		userCache:  provider.UserCache(),

		configName:    configName,
		configGuiName: configName + "-gui",
	}
}

// GetLockFile returns the path to the bridge lock file (e.g. ~/.cache/<company>/<app>/<app>.lock).
func (l *Locations) GetLockFile() string {
	return filepath.Join(l.userCache, l.configName+".lock")
}

// GetGuiLockFile returns the path to the GUI lock file (e.g. ~/.cache/<company>/<app>/<app>.lock).
func (l *Locations) GetGuiLockFile() string {
	return filepath.Join(l.userCache, l.configGuiName+".lock")
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
	return "https://github.com/ProtonMail/proton-bridge/blob/master/COPYING_NOTES.md#dependencies"
}

// ProvideSettingsPath returns a location for user settings (e.g. ~/.config/<company>/<app>).
// It creates it if it doesn't already exist.
func (l *Locations) ProvideSettingsPath() (string, error) {
	if err := os.MkdirAll(l.getSettingsPath(), 0o700); err != nil {
		return "", err
	}

	return l.getSettingsPath(), nil
}

// ProvideGluonPath returns a location for gluon data.
// It creates it if it doesn't already exist.
func (l *Locations) ProvideGluonPath() (string, error) {
	if err := os.MkdirAll(l.getGluonPath(), 0o700); err != nil {
		return "", err
	}

	return l.getGluonPath(), nil
}

// ProvideLogsPath returns a location for user logs (e.g. ~/.local/share/<company>/<app>/logs).
// It creates it if it doesn't already exist.
func (l *Locations) ProvideLogsPath() (string, error) {
	if err := os.MkdirAll(l.getLogsPath(), 0o700); err != nil {
		return "", err
	}

	return l.getLogsPath(), nil
}

// ProvideGUICertPath returns a location for TLS certs used for the connection between bridge and the GUI.
// It creates it if it doesn't already exist.
func (l *Locations) ProvideGUICertPath() (string, error) {
	if err := os.MkdirAll(l.getGUICertPath(), 0o700); err != nil {
		return "", err
	}

	return l.getGUICertPath(), nil
}

// ProvideUpdatesPath returns a location for update files (e.g. ~/.local/share/<company>/<app>/updates).
// It creates it if it doesn't already exist.
func (l *Locations) ProvideUpdatesPath() (string, error) {
	if err := os.MkdirAll(l.getUpdatesPath(), 0o700); err != nil {
		return "", err
	}

	return l.getUpdatesPath(), nil
}

func (l *Locations) getGluonPath() string {
	return filepath.Join(l.userData, "gluon")
}

func (l *Locations) getGUICertPath() string {
	return l.userConfig
}

func (l *Locations) getSettingsPath() string {
	return l.userConfig
}

func (l *Locations) getLogsPath() string {
	return filepath.Join(l.userData, "logs")
}

func (l *Locations) getGoIMAPCachePath() string {
	return filepath.Join(l.userConfig, "cache")
}

func (l *Locations) getUpdatesPath() string {
	return filepath.Join(l.userData, "updates")
}

// Clear removes everything except the lock and update files.
func (l *Locations) Clear() error {
	return files.Remove(
		l.userConfig,
		l.userData,
		l.userCache,
	).Except(
		l.GetGuiLockFile(),
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
	return files.Remove(
		l.userCache,
		l.userData,
	).Except(
		l.GetGuiLockFile(),
		l.getLogsPath(),
		l.getUpdatesPath(),
		l.getGluonPath(),
	).Do()
}

// CleanGoIMAPCache removes all cache data from the go-imap implementation.
func (l *Locations) CleanGoIMAPCache() error {
	return files.Remove(l.getGoIMAPCachePath()).Do()
}
