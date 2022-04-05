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

package base

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/internal/constants"
	"github.com/ProtonMail/proton-bridge/internal/locations"
	"github.com/ProtonMail/proton-bridge/pkg/keychain"
	"github.com/sirupsen/logrus"
)

// migrateFiles migrates files from their old (pre-refactor) locations to their new locations.
// We can remove this eventually.
//
// | entity    |               old location                |            new location                |
// |-----------|-------------------------------------------|----------------------------------------|
// | prefs     | ~/.cache/protonmail/<app>/c11/prefs.json  | ~/.config/protonmail/<app>/prefs.json  |
// | c11 1.5.x | ~/.cache/protonmail/<app>/c11             | ~/.cache/protonmail/<app>/cache/c11    |
// | c11 1.6.x | ~/.cache/protonmail/<app>/cache/c11       | ~/.config/protonmail/<app>/cache/c11   |
// | updates   | ~/.cache/protonmail/<app>/updates         | ~/.config/protonmail/<app>/updates     |.
func migrateFiles(configName string) error {
	locationsProvider, err := locations.NewDefaultProvider(filepath.Join(constants.VendorName, configName))
	if err != nil {
		return err
	}

	locations := locations.New(locationsProvider, configName)
	userCacheDir := locationsProvider.UserCache()

	if err := migratePrefsFrom15x(locations, userCacheDir); err != nil {
		return err
	}
	if err := migrateCacheFromBoth15xAnd16x(locations, userCacheDir); err != nil {
		return err
	}
	if err := migrateUpdatesFrom16x(configName, locations); err != nil { //nolint:revive It is more clear to structure this way
		return err
	}
	return nil
}

func migratePrefsFrom15x(locations *locations.Locations, userCacheDir string) error {
	newSettingsDir, err := locations.ProvideSettingsPath()
	if err != nil {
		return err
	}

	return moveIfExists(
		filepath.Join(userCacheDir, "c11", "prefs.json"),
		filepath.Join(newSettingsDir, "prefs.json"),
	)
}

func migrateCacheFromBoth15xAnd16x(locations *locations.Locations, userCacheDir string) error {
	olderCacheDir := userCacheDir
	newerCacheDir := locations.GetOldCachePath()
	latestCacheDir, err := locations.ProvideCachePath()
	if err != nil {
		return err
	}

	// Migration for versions before 1.6.x.
	if err := moveIfExists(
		filepath.Join(olderCacheDir, "c11"),
		filepath.Join(latestCacheDir, "c11"),
	); err != nil {
		return err
	}

	// Migration for versions 1.6.x.
	return moveIfExists(
		filepath.Join(newerCacheDir, "c11"),
		filepath.Join(latestCacheDir, "c11"),
	)
}

func migrateUpdatesFrom16x(configName string, locations *locations.Locations) error {
	// In order to properly update Bridge 1.6.X and higher we need to
	// change the launcher first. Since this is not part of automatic
	// updates the migration must wait until manual update. Until that
	// we need to keep old path.
	if configName == "bridge" {
		return nil
	}

	oldUpdatesPath := locations.GetOldUpdatesPath()
	// Do not use ProvideUpdatesPath, that creates dir right away.
	newUpdatesPath := locations.GetUpdatesPath()

	return moveIfExists(oldUpdatesPath, newUpdatesPath)
}

// migrateMacKeychainBefore220 deals with write access restriction to mac
// keychain passwords which are caused by application renaming. The old
// passwords are copied under new name in order to have write access afer
// renaming.
func migrateMacKeychainBefore220(settingsObj *settings.Settings, keychainName string) error {
	l := logrus.WithField("pkg", "app/base/migration")
	if runtime.GOOS != "darwin" {
		return nil
	}

	if shouldContinue, err := isBefore220(settingsObj); !shouldContinue || err != nil {
		return err
	}

	l.Warn("Migrating mac keychain")

	helperConstructor, ok := keychain.Helpers["macos-keychain"]
	if !ok {
		return errors.New("cannot find macos-keychain helper")
	}

	oldKC, err := helperConstructor("ProtonMailBridgeService")
	if err != nil {
		l.WithError(err).Error("Keychain constructor failed")
		return err
	}

	idByURL, err := oldKC.List()
	if err != nil {
		l.WithError(err).Error("List old keychain failed")
		return err
	}

	newKC, err := keychain.NewKeychain(settingsObj, keychainName)
	if err != nil {
		return err
	}

	for url, id := range idByURL {
		li := l.WithField("id", id).WithField("url", url)
		userID, secret, err := oldKC.Get(url)
		if err != nil {
			li.WithField("userID", userID).
				WithField("err", err).
				Error("Faild to get old item")
			continue
		}

		if _, _, err := newKC.Get(userID); err == nil {
			li.Warn("Skipping migration, item already exists.")
			continue
		}

		if err := newKC.Put(userID, secret); err != nil {
			li.WithError(err).Error("Failed to migrate user")
		}

		li.Info("Item migrated")
	}

	return nil
}

// migrateStartup220 removes old startup links. The creation of new links is
// handled by bridge initialisation.
func migrateStartup220(settingsObj *settings.Settings) error {
	if shouldContinue, err := isBefore220(settingsObj); !shouldContinue || err != nil {
		return err
	}

	logrus.WithField("pkg", "app/base/migration").Warn("Migrating autostartup links")

	path, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	switch runtime.GOOS {
	case "windows":
		path = filepath.Join(path, `AppData\Roaming\Microsoft\Windows\Start Menu\Programs\Startup\ProtonMail Bridge.lnk`)
	case "linux":
		path = filepath.Join(path, `.config/autostart/ProtonMail Bridge.desktop`)
	case "darwin":
		path = filepath.Join(path, `Library/LaunchAgents/ProtonMail Bridge.plist`)
	default:
		return errors.New("unknown GOOS")
	}

	return os.Remove(path)
}

// isBefore220 decide if last used version was older than 2.2.0. If cannot decide it returns false with error.
func isBefore220(settingsObj *settings.Settings) (bool, error) {
	lastUsedVersion := settingsObj.Get(settings.LastVersionKey)

	// Skipping migration: it is first bridge start or cache was cleared.
	if lastUsedVersion == "" {
		return false, nil
	}

	v220 := semver.MustParse("2.2.0")
	lastVer, err := semver.NewVersion(lastUsedVersion)

	// Skipping migration: Should not happen but cannot decide what to do.
	if err != nil {
		return false, err
	}

	// Skipping migration: 2.2.0>= was already used hence old stuff was already migrated.
	if !lastVer.LessThan(v220) {
		return false, nil
	}

	return true, nil
}

func moveIfExists(source, destination string) error {
	l := logrus.WithField("source", source).WithField("destination", destination)

	if _, err := os.Stat(source); os.IsNotExist(err) {
		l.Info("No need to migrate file, source doesn't exist")
		return nil
	}

	if _, err := os.Stat(destination); !os.IsNotExist(err) {
		// Once migrated, files should not stay in source anymore. Therefore
		// if some files are still in source location but target already exist,
		// it's suspicious. Could happen by installing new version, then the
		// old one because of some reason, and then the new one again.
		// Good to see as warning because it could be a reason why Bridge is
		// behaving weirdly, like wrong configuration, or db re-sync and so on.
		l.Warn("No need to migrate file, target already exists")
		return nil
	}

	l.Info("Migrating files")
	return os.Rename(source, destination)
}
