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

package logging

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/sirupsen/logrus"
)

func clearLogs(logDir string, maxLogs int, maxCrashes int) error {
	files, err := ioutil.ReadDir(logDir)
	if err != nil {
		return err
	}

	var logsWithPrefix []string
	var crashesWithPrefix []string

	for _, file := range files {
		if MatchLogName(file.Name()) {
			if MatchStackTraceName(file.Name()) {
				crashesWithPrefix = append(crashesWithPrefix, file.Name())
			} else {
				logsWithPrefix = append(logsWithPrefix, file.Name())
			}
		} else {
			// Older versions of Bridge stored logs in subfolders for each version.
			// That also has to be cleared and the functionality can be removed after some time.
			if file.IsDir() {
				if err := clearLogs(filepath.Join(logDir, file.Name()), maxLogs, maxCrashes); err != nil {
					return err
				}
			} else {
				removeLog(logDir, file.Name())
			}
		}
	}

	removeOldLogs(logDir, logsWithPrefix, maxLogs)
	removeOldLogs(logDir, crashesWithPrefix, maxCrashes)

	return nil
}

func removeOldLogs(logDir string, filenames []string, maxLogs int) {
	count := len(filenames)
	if count <= maxLogs {
		return
	}

	sort.Strings(filenames) // Sorted by timestamp: oldest first.
	for _, filename := range filenames[:count-maxLogs] {
		removeLog(logDir, filename)
	}
}

func removeLog(logDir, filename string) {
	// We need to be sure to delete only log files.
	// Directory with logs can also contain other files.
	if !MatchLogName(filename) {
		return
	}
	if err := os.Remove(filepath.Join(logDir, filename)); err != nil {
		logrus.WithError(err).Error("Failed to remove", filepath.Join(logDir, filename))
	}
}
