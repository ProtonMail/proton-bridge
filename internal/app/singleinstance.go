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

package app

import (
	"fmt"
	"os"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/v3/internal/focus"
	"github.com/allan-simon/go-singleinstance"
	"github.com/sirupsen/logrus"
)

// checkSingleInstance checks if another instance of the application is already running.
// It tries to create a lock file at the given path.
// If it succeeds, it returns the lock file and a nil error.
//
// For macOS and Linux when already running version is older than this instance
// it will kill old and continue with this new bridge (i.e. no error returned).
func checkSingleInstance(settingPath, lockFilePath string, curVersion *semver.Version) (*os.File, error) {
	if lock, err := singleinstance.CreateLockFile(lockFilePath); err == nil {
		logrus.WithField("path", lockFilePath).Debug("Created lock file; no other instance is running")
		return lock, nil
	}

	logrus.Warn("Failed to create lock file; another instance is running")

	// We couldn't create the lock file, so another instance is probably running.
	// Check if it's an older version of the app.
	lastVersion, ok := focus.TryVersion(settingPath)
	if !ok {
		return nil, fmt.Errorf("failed to determine version of running instance")
	}

	if !lastVersion.LessThan(curVersion) {
		return nil, fmt.Errorf("running instance is newer than this one")
	}

	// The other instance is an older version, so we should kill it.
	pid, err := getPID(lockFilePath)
	if err != nil {
		return nil, err
	}

	if err := killPID(pid); err != nil {
		return nil, err
	}

	// Need to wait some time to release file lock
	time.Sleep(time.Second)

	return singleinstance.CreateLockFile(lockFilePath)
}
