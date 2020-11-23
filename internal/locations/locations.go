// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

// Package locations implements a type that provides cross-platform access to
// standard filesystem locations, including config, cache and log directories.
package locations

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/ProtonMail/proton-bridge/pkg/files"
	"github.com/sirupsen/logrus"
)

// Locations provides cross-platform access to standard locations.
// On linux:
// - settings: ~/.config/protonmail/<app>
// - logs: ~/.cache/protonmail/<app>/logs
// - cache: ~/.cache/protonmail/<app>/cache
// - updates: ~/.cache/protonmail/<app>/updates
// - lockfile: ~/.cache/protonmail/<app>/<app>.lock
type Locations struct {
	userConfig, userCache string
	configName            string
}

type appDirsProvider interface {
	UserConfig() string
	UserCache() string
}

func New(appDirs appDirsProvider, configName string) *Locations {
	return &Locations{
		userConfig: appDirs.UserConfig(),
		userCache:  appDirs.UserCache(),
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
		appName := l.configName
		if l.configName == "importExport" {
			appName = "import-export"
		}
		// Most Linux distributions.
		path := "/usr/share/doc/protonmail/" + appName + "/LICENSE"
		if _, err := os.Stat(path); err == nil {
			return path
		}
		// Arch distributions.
		return "/usr/share/licenses/protonmail-" + appName + "/LICENSE"
	case "darwin": //nolint[goconst]
		path := filepath.Join(filepath.Dir(os.Args[0]), "..", "Resources", "LICENSE")
		if _, err := os.Stat(path); err == nil {
			return path
		}

		appName := "ProtonMail Bridge.app"
		if l.configName == "importExport" {
			appName = "ProtonMail Import-Export.app"
		}
		return "/Applications/" + appName + "/Contents/Resources/LICENSE"
	case "windows":
		path := filepath.Join(filepath.Dir(os.Args[0]), "LICENSE.txt")
		if _, err := os.Stat(path); err == nil {
			return path
		}
		// This should not happen, Windows should be handled by relative
		// location to the binary above. This is just fallback which may
		// or may not work, depends where user installed the app and how
		// user started the app.
		return filepath.FromSlash("C:/Program Files/Proton Technologies AG/ProtonMail Bridge/LICENSE.txt")
	}
	return ""
}

// ProvideSettingsPath returns a location for user settings (e.g. ~/.config/<company>/<app>).
// It creates it if it doesn't already exist.
func (l *Locations) ProvideSettingsPath() (string, error) {
	if err := os.MkdirAll(l.getSettingsPath(), 0700); err != nil {
		return "", err
	}

	return l.getSettingsPath(), nil
}

// ProvideLogsPath returns a location for user logs (e.g. ~/.cache/<company>/<app>/logs).
// It creates it if it doesn't already exist.
func (l *Locations) ProvideLogsPath() (string, error) {
	if err := os.MkdirAll(l.getLogsPath(), 0700); err != nil {
		return "", err
	}

	return l.getLogsPath(), nil
}

// ProvideCachePath returns a location for user cache dirs (e.g. ~/.cache/<company>/<app>/cache).
// It creates it if it doesn't already exist.
func (l *Locations) ProvideCachePath() (string, error) {
	if err := os.MkdirAll(l.getCachePath(), 0700); err != nil {
		return "", err
	}

	return l.getCachePath(), nil
}

// ProvideUpdatesPath returns a location for update files (e.g. ~/.cache/<company>/<app>/updates).
// It creates it if it doesn't already exist.
func (l *Locations) ProvideUpdatesPath() (string, error) {
	if err := os.MkdirAll(l.getUpdatesPath(), 0700); err != nil {
		return "", err
	}

	return l.getUpdatesPath(), nil
}

func (l *Locations) getSettingsPath() string {
	return l.userConfig
}

func (l *Locations) getLogsPath() string {
	return filepath.Join(l.userCache, "logs")
}

func (l *Locations) getCachePath() string {
	return filepath.Join(l.userCache, "cache")
}

func (l *Locations) getUpdatesPath() string {
	return filepath.Join(l.userCache, "updates")
}

// Clear removes everything except the lock file.
func (l *Locations) Clear() error {
	return files.Remove(
		l.getSettingsPath(),
		l.getLogsPath(),
		l.getCachePath(),
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
