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
	"os"
	"path/filepath"

	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
	"github.com/ProtonMail/proton-bridge/v2/internal/locations"
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
