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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package app

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/allan-simon/go-singleinstance"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func migrateOldVersions() (allErrors error) {
	cacheDir, cacheError := os.UserCacheDir()
	if cacheError != nil {
		allErrors = multierror.Append(allErrors, errors.Wrap(cacheError, "cannot get os cache"))
		return // not need to continue for now (with more migrations might be still ok to continue)
	}

	if err := killV2AppAndRemoveV2LockFiles(filepath.Join(cacheDir, "protonmail", "bridge", "bridge.lock")); err != nil {
		allErrors = multierror.Append(allErrors, errors.Wrap(err, "cannot migrate lockfiles"))
	}

	return
}

func killV2AppAndRemoveV2LockFiles(lockFilePathV2 string) error {
	l := logrus.WithField("path", lockFilePathV2)

	if _, err := os.Stat(lockFilePathV2); os.IsNotExist(err) {
		l.Debug("no v2 lockfile")
		return nil
	}

	lock, err := singleinstance.CreateLockFile(lockFilePathV2)

	if err == nil {
		l.Debug("no other v2 instance is running")

		if errClose := lock.Close(); errClose != nil {
			l.WithError(errClose).Error("Cannot close lock file")
		}

		return os.Remove(lockFilePathV2)
	}

	// The other instance is an older version, so we should kill it.
	pid, err := getPID(lockFilePathV2)
	if err != nil {
		return errors.Wrap(err, "cannot get v2 pid")
	}

	if err := killPID(pid); err != nil {
		return errors.Wrapf(err, "cannot kill v2 app (PID %d)", pid)
	}

	// Need to wait some time to release file lock
	time.Sleep(time.Second)

	return nil
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

func killPID(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	return p.Kill()
}
