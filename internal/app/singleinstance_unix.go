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

//go:build !windows
// +build !windows

package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/v2/internal/focus"
	"github.com/allan-simon/go-singleinstance"
	"golang.org/x/sys/unix"
)

// checkSingleInstance checks if another instance of the application is already running.
// It tries to create a lock file at the given path.
// If it succeeds, it returns the lock file and a nil error.
//
// For macOS and Linux when already running version is older than this instance
// it will kill old and continue with this new bridge (i.e. no error returned).
func checkSingleInstance(lockFilePath string, curVersion *semver.Version) (*os.File, error) {
	if lock, err := singleinstance.CreateLockFile(lockFilePath); err == nil {
		return lock, nil
	}

	// We couldn't create the lock file, so another instance is probably running.
	// Check if it's an older version of the app.
	lastVersion, ok := focus.TryVersion()
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

	if err := unix.Kill(pid, unix.SIGTERM); err != nil {
		return nil, err
	}

	// Need to wait some time to release file lock
	time.Sleep(time.Second)

	return singleinstance.CreateLockFile(lockFilePath)
}

func getPID(lockFilePath string) (int, error) {
	file, err := os.Open(filepath.Clean(lockFilePath))
	if err != nil {
		return 0, err
	}
	defer func() { _ = file.Close() }()

	rawPID := make([]byte, 10) // PID is probably up to 7 digits long, 10 should be enough
	n, err := file.Read(rawPID)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(strings.TrimSpace(string(rawPID[:n])))
}

/*
func runningVersionIsOlder() error {
	currentVer, err := semver.StrictNewVersion(constants.Version)
	if err != nil {
		return err
	}

	runningVer, err := semver.StrictNewVersion(settingsObj.Get(settings.LastVersionKey))
	if err != nil {
		return err
	}

	if !runningVer.LessThan(currentVer) {
		return errors.New("running version is not older")
	}

	return nil
}
*/
