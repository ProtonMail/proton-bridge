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
	"strings"

	"github.com/ProtonMail/proton-bridge/v2/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/v2/pkg/keychain"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
)

const darwin = "darwin"

func migrateRebranding(settingsObj *settings.Settings, keychainName string) (result error) {
	if err := migrateStartupBeforeRebranding(); err != nil {
		result = multierror.Append(result, err)
	}

	lastUsedVersion := settingsObj.Get(settings.LastVersionKey)

	// Skipping migration: it is first bridge start or cache was cleared.
	if lastUsedVersion == "" {
		settingsObj.SetBool(settings.RebrandingMigrationKey, true)
		return
	}

	// Skipping rest of migration: already done
	if settingsObj.GetBool(settings.RebrandingMigrationKey) {
		return
	}

	switch runtime.GOOS {
	case "windows", "linux":
		// GODT-1260 we would need admin rights to changes desktop files
		// and start menu items.
		settingsObj.SetBool(settings.RebrandingMigrationKey, true)
	case darwin:
		if shouldContinue, err := isMacBeforeRebranding(); !shouldContinue || err != nil {
			if err != nil {
				result = multierror.Append(result, err)
			}
			break
		}

		if err := migrateMacKeychainBeforeRebranding(settingsObj, keychainName); err != nil {
			result = multierror.Append(result, err)
		}

		settingsObj.SetBool(settings.RebrandingMigrationKey, true)
	}

	return result
}

// migrateMacKeychainBeforeRebranding deals with write access restriction to
// mac keychain passwords which are caused by application renaming. The old
// passwords are copied under new name in order to have write access afer
// renaming.
func migrateMacKeychainBeforeRebranding(settingsObj *settings.Settings, keychainName string) error {
	l := logrus.WithField("pkg", "app/base/migration")
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

// migrateStartupBeforeRebranding removes old startup links. The creation of new links is
// handled by bridge initialisation.
func migrateStartupBeforeRebranding() error {
	path, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	switch runtime.GOOS {
	case "windows":
		path = filepath.Join(path, `AppData\Roaming\Microsoft\Windows\Start Menu\Programs\Startup\ProtonMail Bridge.lnk`)
	case "linux":
		path = filepath.Join(path, `.config/autostart/ProtonMail Bridge.desktop`)
	case darwin:
		path = filepath.Join(path, `Library/LaunchAgents/ProtonMail Bridge.plist`)
	default:
		return errors.New("unknown GOOS")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	logrus.WithField("pkg", "app/base/migration").Warn("Migrating autostartup links")
	return os.Remove(path)
}

// startupNameForRebranding returns the name for autostart launcher based on
// type of rebranded instance i.e. update or manual.
//
// This only affects darwin when udpate re-writes the old startup and then
// manual installed it would not run proper exe. Therefore we return "old" name
// for updates and "new" name for manual which would be properly migrated.
//
// For orther (linux and windows) the link is always pointing to launcher which
// path didn't changed.
func startupNameForRebranding(origin string) string {
	if runtime.GOOS == darwin {
		if path, err := os.Executable(); err == nil && strings.Contains(path, "ProtonMail Bridge") {
			return "ProtonMail Bridge"
		}
	}

	// No need to solve for other OS. See comment above.
	return origin
}

// isBeforeRebranding decide if last used version was older than 2.2.0. If
// cannot decide it returns false with error.
func isMacBeforeRebranding() (bool, error) {
	// previous version | update | do mac migration |
	//                  | first  | false            |
	// cleared-cache    | manual | false            |
	// cleared-cache    | in-app | false            |
	// old              | in-app | false            |
	// old in-app       | in-app | false            |
	// old              | manual | true             |
	// old in-app       | manual | true             |
	// manual           | in-app | false            |

	// Skip if it was in-app update and not manual
	if path, err := os.Executable(); err != nil || strings.Contains(path, "ProtonMail Bridge") {
		return false, err
	}

	return true, nil
}
