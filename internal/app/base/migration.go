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

package base

import (
	"os"
	"path/filepath"

	"github.com/ProtonMail/go-appdir"
	"github.com/ProtonMail/proton-bridge/internal/constants"
	"github.com/ProtonMail/proton-bridge/internal/locations"
	"github.com/sirupsen/logrus"
)

// MigrateFiles migrates files from their old (pre-refactor) locations to their new locations.
// We can remove this eventually.
//
// | entity |               old location                |            new location                |
// |--------|-------------------------------------------|----------------------------------------|
// | prefs  | ~/.cache/protonmail/<app>/c11/prefs.json  | ~/.config/protonmail/<app>/prefs.json  |
// | c11    | ~/.cache/protonmail/<app>/c11             | ~/.cache/protonmail/<app>/cache/c11    |
func MigrateFiles(configName string) error {
	appDirs := appdir.New(filepath.Join(constants.VendorName, configName))
	locations := locations.New(appDirs, configName)

	userCacheDir := appDirs.UserCache()
	newSettingsDir, err := locations.ProvideSettingsPath()
	if err != nil {
		return err
	}

	if err := moveIfExists(
		filepath.Join(userCacheDir, "c11", "prefs.json"),
		filepath.Join(newSettingsDir, "prefs.json"),
	); err != nil {
		return err
	}

	newCacheDir, err := locations.ProvideCachePath()
	if err != nil {
		return err
	}

	if err := moveIfExists(
		filepath.Join(userCacheDir, "c11"),
		filepath.Join(newCacheDir, "c11"),
	); err != nil {
		return err
	}

	return nil
}

func moveIfExists(source, destination string) error {
	if _, err := os.Stat(source); os.IsNotExist(err) {
		logrus.WithField("source", source).WithField("destination", destination).Debug("No need to migrate file")
		return nil
	}

	if _, err := os.Stat(destination); !os.IsNotExist(err) {
		logrus.WithField("source", source).WithField("destination", destination).Debug("No need to migrate file")
		return nil
	}

	return os.Rename(source, destination)
}
