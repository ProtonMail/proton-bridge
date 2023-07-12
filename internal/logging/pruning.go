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

package logging

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bradenaw/juniper/xslices"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

const (
	DefaultPruningSize = 1024 * 1024 * 200
	NoPruning          = -1
)

type Pruner func() (failureCount int, err error)

type logFileInfo struct {
	filename string
	size     int64
}

type sessionInfo struct {
	dir          string
	sessionID    SessionID
	launcherLogs []logFileInfo
	guiLogs      []logFileInfo
	bridgeLogs   []logFileInfo
}

func defaultPruner(logsDir string, currentSessionID SessionID, pruningSize int64) func() (failureCount int, err error) {
	return func() (int, error) {
		return pruneLogs(logsDir, currentSessionID, pruningSize)
	}
}

func nullPruner() (failureCount int, err error) {
	return 0, nil
}

// DefaultPruner gets rid of the older log files according to the following policy:
//   - We will limit the total size of the log files to roughly pruningSize.
//     The current session is included in this quota, in order not to grow indefinitely on setups where bridge can run uninterrupted for months.
//   - If the current session's log files total size is above the pruning size, we delete all other sessions log. For the current we keep
//     launcher and gui log (they're limited to a few kb at most by nature), and we have n bridge log files,
//     We keep the first and list log file (startup information contains relevant information, notably the SentryID), and start deleting the other
//     starting with the oldest until the total size drops below the pruning size.
//   - Otherwise: If the total size of log files for all sessions exceeds pruningSize, sessions gets deleted starting with the oldest, until the size
//     drops below the pruning size. Sessions are treated atomically. Current session is left untouched in that case.
func pruneLogs(logDir string, currentSessionID SessionID, pruningSize int64) (failureCount int, err error) {
	sessionInfoList, err := buildSessionInfoList(logDir)
	if err != nil {
		return 0, err
	}

	// we want total size to include the current session.
	totalSize := xslices.Reduce(maps.Values(sessionInfoList), int64(0), func(sum int64, info *sessionInfo) int64 { return sum + info.size() })
	if totalSize <= pruningSize {
		return 0, nil
	}

	currentSessionInfo, ok := sessionInfoList[currentSessionID]
	if ok {
		delete(sessionInfoList, currentSessionID)

		if currentSessionInfo.size() > pruningSize {
			// current session is already too big. We delete all other sessions and prune the current session.
			for _, session := range sessionInfoList {
				failureCount += session.deleteFiles()
			}

			failureCount += currentSessionInfo.pruneAsCurrentSession(pruningSize)
			return failureCount, nil
		}
	}

	// current session size if below max size, so we erase older session starting with the eldest until we go below maxFileSize
	sortedSessions := maps.Values(sessionInfoList)
	slices.SortFunc(sortedSessions, func(lhs, rhs *sessionInfo) bool { return lhs.sessionID < rhs.sessionID })
	for _, sessionInfo := range sortedSessions {
		totalSize -= sessionInfo.size()
		failureCount += sessionInfo.deleteFiles()
		if totalSize <= pruningSize {
			return failureCount, nil
		}
	}

	return failureCount, nil
}

func newSessionInfo(dir string, sessionID SessionID) (*sessionInfo, error) {
	paths, err := filepath.Glob(filepath.Join(dir, string(sessionID)+"_*.log"))
	if err != nil {
		return nil, err
	}

	rx := regexp.MustCompile(`^\Q` + string(sessionID) + `\E_([^_]*)_\d+_.*\.log$`)

	result := sessionInfo{sessionID: sessionID, dir: dir}
	for _, path := range paths {
		filename := filepath.Base(path)
		match := rx.FindStringSubmatch(filename)
		if len(match) != 2 {
			continue
		}

		stats, err := os.Stat(path)
		if err != nil {
			continue
		}

		fileInfo := logFileInfo{
			filename: filename,
			size:     stats.Size(),
		}

		switch AppName(match[1]) {
		case LauncherShortAppName:
			result.launcherLogs = append(result.launcherLogs, fileInfo)
		case GUIShortAppName:
			result.guiLogs = append(result.guiLogs, fileInfo)
		case BridgeShortAppName:
			result.bridgeLogs = append(result.bridgeLogs, fileInfo)
		}
	}

	lessFunc := func(lhs, rhs logFileInfo) bool { return strings.Compare(lhs.filename, rhs.filename) < 0 }
	slices.SortFunc(result.launcherLogs, lessFunc)
	slices.SortFunc(result.guiLogs, lessFunc)
	slices.SortFunc(result.bridgeLogs, lessFunc)

	return &result, nil
}

func (s *sessionInfo) size() int64 {
	summer := func(accum int64, logInfo logFileInfo) int64 { return accum + logInfo.size }
	size := xslices.Reduce(s.launcherLogs, 0, summer)
	size = xslices.Reduce(s.guiLogs, size, summer)
	size = xslices.Reduce(s.bridgeLogs, size, summer)
	return size
}

func (s *sessionInfo) deleteFiles() (failureCount int) {
	var allLogs []logFileInfo
	allLogs = append(allLogs, s.launcherLogs...)
	allLogs = append(allLogs, s.guiLogs...)
	allLogs = append(allLogs, s.bridgeLogs...)

	for _, log := range allLogs {
		if err := os.Remove(filepath.Join(s.dir, log.filename)); err != nil {
			failureCount++
		}
	}

	return failureCount
}

func (s *sessionInfo) pruneAsCurrentSession(pruningSize int64) (failureCount int) {
	// when pruning the current session, we keep the launcher and GUI logs, the first and last bridge log file
	// and we delete intermediate bridge logs until the size constraint is satisfied (or there nothing left to delete).
	if len(s.bridgeLogs) < 3 {
		return 0
	}

	size := s.size()
	if size <= pruningSize {
		return 0
	}

	for _, fileInfo := range s.bridgeLogs[1 : len(s.bridgeLogs)-1] {
		if err := os.Remove(filepath.Join(s.dir, fileInfo.filename)); err != nil {
			failureCount++
		}
		size -= fileInfo.size
		if size <= pruningSize {
			return failureCount
		}
	}

	return failureCount
}

func buildSessionInfoList(dir string) (map[SessionID]*sessionInfo, error) {
	result := make(map[SessionID]*sessionInfo)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		rx := regexp.MustCompile(`^(\d{8}_\d{9})_.*\.log$`)
		match := rx.FindStringSubmatch(entry.Name())
		if match == nil || len(match) < 2 {
			continue
		}

		sessionID := SessionID(match[1])
		if _, ok := result[sessionID]; !ok {
			sessionInfo, err := newSessionInfo(dir, sessionID)
			if err != nil {
				return nil, err
			}
			result[sessionID] = sessionInfo
		}
	}

	return result, nil
}
