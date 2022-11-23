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
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

func clearLogs(logDir string, maxLogs int, maxCrashes int) error {
	files, err := os.ReadDir(logDir)
	if err != nil {
		return fmt.Errorf("failed to read log directory: %w", err)
	}

	names := xslices.Map(files, func(file fs.DirEntry) string {
		return file.Name()
	})

	// Remove old logs.
	removeOldLogs(logDir, xslices.Filter(names, func(name string) bool {
		return MatchLogName(name) && !MatchStackTraceName(name)
	}), maxLogs)

	// Remove old stack traces.
	removeOldLogs(logDir, xslices.Filter(names, func(name string) bool {
		return MatchLogName(name) && MatchStackTraceName(name)
	}), maxCrashes)

	return nil
}

func removeOldLogs(dir string, names []string, max int) {
	if count := len(names); count <= max {
		return
	}

	// Sort by timestamp, oldest first.
	slices.SortFunc(names, func(a, b string) bool {
		return getLogTime(a) < getLogTime(b)
	})

	for _, path := range xslices.Map(names[:len(names)-max], func(name string) string { return filepath.Join(dir, name) }) {
		if err := os.Remove(path); err != nil {
			logrus.WithError(err).WithField("path", path).Warn("Failed to remove old log file")
		}
	}
}
