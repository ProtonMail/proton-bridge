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

package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ProtonMail/go-appdir"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
)

var (
	log = logrus.WithField("pkg", "config") //nolint[gochecknoglobals]
)

type appDirProvider interface {
	UserConfig() string
	UserCache() string
	UserLogs() string
}

type Config struct {
	appName        string
	version        string
	revision       string
	cacheVersion   string
	appDirs        appDirProvider
	appDirsVersion appDirProvider
}

// New returns fully initialized config struct.
// `appName` should be in camelCase format for folder or file names. It's also used in API
// as `AppVersion` which is converted to CamelCase.
// `version` is the version of the app (e.g. v1.2.3).
// `cacheVersion` is the version of the cache files (setting a different number will remove the old ones).
func New(appName, version, revision, cacheVersion string) *Config {
	appDirs := appdir.New(filepath.Join("protonmail", appName))
	appDirsVersion := appdir.New(filepath.Join("protonmail", appName, cacheVersion))
	return newConfig(appName, version, revision, cacheVersion, appDirs, appDirsVersion)
}

func newConfig(appName, version, revision, cacheVersion string, appDirs, appDirsVersion appDirProvider) *Config {
	return &Config{
		appName:        appName,
		version:        version,
		revision:       revision,
		cacheVersion:   cacheVersion,
		appDirs:        appDirs,
		appDirsVersion: appDirsVersion,
	}
}

// CreateDirs creates all folders that are necessary for bridge to properly function.
func (c *Config) CreateDirs() error {
	// Log files.
	if err := os.MkdirAll(c.appDirs.UserLogs(), 0700); err != nil {
		return err
	}
	// TLS files.
	if err := os.MkdirAll(c.appDirs.UserConfig(), 0750); err != nil {
		return err
	}
	// Lock, events, preferences, user_info, db files.
	if err := os.MkdirAll(c.appDirsVersion.UserCache(), 0750); err != nil {
		return err
	}
	return nil
}

// ClearData removes all files except the lock file.
// The lock file will be removed when the Bridge stops.
func (c *Config) ClearData() error {
	dirs := []string{
		c.appDirs.UserLogs(),
		c.appDirs.UserConfig(),
		c.appDirs.UserCache(),
	}
	shouldRemove := func(filePath string) bool {
		return filePath != c.GetLockPath()
	}
	return c.removeAllExcept(dirs, shouldRemove)
}

// ClearOldData removes all old files, such as old log files or old versions of cache and so on.
func (c *Config) ClearOldData() error {
	// `appDirs` is parent for `appDirsVersion`.
	// `dir` then contains all subfolders and only `cacheVersion` should stay.
	// But on Windows all files (dirs) are in the same one - we cannot remove log, lock or tls files.
	dir := c.appDirs.UserCache()

	return c.removeExcept(dir, func(filePath string) bool {
		fileName := filepath.Base(filePath)
		return (fileName != c.cacheVersion &&
			!logFileRgx.MatchString(fileName) &&
			filePath != c.GetLogDir() &&
			filePath != c.GetTLSCertPath() &&
			filePath != c.GetTLSKeyPath() &&
			filePath != c.GetEventsPath() &&
			filePath != c.GetIMAPCachePath() &&
			filePath != c.GetLockPath() &&
			filePath != c.GetPreferencesPath())
	})
}

func (c *Config) removeAllExcept(dirs []string, shouldRemove func(string) bool) error {
	var result *multierror.Error
	for _, dir := range dirs {
		if err := c.removeExcept(dir, shouldRemove); err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result.ErrorOrNil()
}

func (c *Config) removeExcept(dir string, shouldRemove func(string) bool) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	var result *multierror.Error
	for _, file := range files {
		filePath := filepath.Join(dir, file.Name())
		if !shouldRemove(filePath) {
			continue
		}

		if !file.IsDir() {
			if err := os.RemoveAll(filePath); err != nil {
				result = multierror.Append(result, err)
			}
			continue
		}

		subDir := filepath.Join(dir, file.Name())
		if err := c.removeExcept(subDir, shouldRemove); err != nil {
			result = multierror.Append(result, err)
		} else {
			// Remove dir itself only if it's empty.
			subFiles, err := ioutil.ReadDir(subDir)
			if err != nil {
				result = multierror.Append(result, err)
			} else if len(subFiles) == 0 {
				if err := os.RemoveAll(subDir); err != nil {
					result = multierror.Append(result, err)
				}
			}
		}
	}
	return result.ErrorOrNil()
}

// IsDevMode should be used for development conditions such us whether to send sentry reports.
func (c *Config) IsDevMode() bool {
	return os.Getenv("PROTONMAIL_ENV") == "dev"
}

// GetVersion returns the version.
func (c *Config) GetVersion() string {
	return c.version
}

// GetLogDir returns folder for log files.
func (c *Config) GetLogDir() string {
	return c.appDirs.UserLogs()
}

// GetLogPrefix returns prefix for log files. Bridge uses format vVERSION.
func (c *Config) GetLogPrefix() string {
	return "v" + c.version + "_" + c.revision
}

// GetTLSCertPath returns path to certificate; used for TLS servers (IMAP, SMTP and API).
func (c *Config) GetTLSCertPath() string {
	return filepath.Join(c.appDirs.UserConfig(), "cert.pem")
}

// GetTLSKeyPath returns path to private key; used for TLS servers (IMAP, SMTP and API).
func (c *Config) GetTLSKeyPath() string {
	return filepath.Join(c.appDirs.UserConfig(), "key.pem")
}

// GetDBDir returns folder for db files.
func (c *Config) GetDBDir() string {
	return c.appDirsVersion.UserCache()
}

// GetEventsPath returns path to events file containing the last processed event IDs.
func (c *Config) GetEventsPath() string {
	return filepath.Join(c.appDirsVersion.UserCache(), "events.json")
}

// GetIMAPCachePath returns path to file with IMAP status.
func (c *Config) GetIMAPCachePath() string {
	return filepath.Join(c.appDirsVersion.UserCache(), "user_info.json")
}

// GetLockPath returns path to lock file to check if bridge is already running.
func (c *Config) GetLockPath() string {
	return filepath.Join(c.appDirsVersion.UserCache(), c.appName+".lock")
}

// GetUpdateDir returns folder for update files; such as new binary.
func (c *Config) GetUpdateDir() string {
	return filepath.Join(c.appDirsVersion.UserCache(), "updates")
}

// GetPreferencesPath returns path to preference file.
func (c *Config) GetPreferencesPath() string {
	return filepath.Join(c.appDirsVersion.UserCache(), "prefs.json")
}

// GetTransferDir returns folder for import-export rules files.
func (c *Config) GetTransferDir() string {
	return c.appDirsVersion.UserCache()
}

// GetDefaultAPIPort returns default Bridge local API port.
func (c *Config) GetDefaultAPIPort() int {
	return 1042
}

// GetDefaultIMAPPort returns default Bridge IMAP port.
func (c *Config) GetDefaultIMAPPort() int {
	return 1143
}

// GetDefaultSMTPPort returns default Bridge SMTP port.
func (c *Config) GetDefaultSMTPPort() int {
	return 1025
}
