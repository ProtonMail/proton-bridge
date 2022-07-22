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

package base

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/ProtonMail/proton-bridge/v2/internal/config/settings"
	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
	"github.com/allan-simon/go-singleinstance"
	"golang.org/x/sys/unix"
)

// checkSingleInstance returns error if a bridge instance is already running
// This instance should be stop and window of running window should be brought
// to focus.
//
// For macOS and Linux when already running version is older than this instance
// it will kill old and continue with this new bridge (i.e. no error returned).
func checkSingleInstance(lockFilePath string, settingsObj *settings.Settings) (*os.File, error) {
	if lock, err := singleinstance.CreateLockFile(lockFilePath); err == nil {
		// Bridge is not runnig, continue normally
		return lock, nil
	}

	if err := runningVersionIsOlder(settingsObj); err != nil {
		return nil, err
	}

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

func runningVersionIsOlder(settingsObj *settings.Settings) error {
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
