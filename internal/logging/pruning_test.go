// Copyright (c) 2024 Proton AG
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
	"testing"
	"time"

	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/stretchr/testify/require"
)

type fileInfo struct {
	filename string
	size     int64
}

var logFileSuffix = "_v" + constants.Version + "_" + constants.Tag + ".log"

func TestLogging_Pruning(t *testing.T) {
	dir := t.TempDir()
	const maxLogSize = 100
	sessionID1 := createDummySession(t, dir, maxLogSize, 50, 50, 250)
	sessionID2 := createDummySession(t, dir, maxLogSize, 100, 100, 350)
	sessionID3 := createDummySession(t, dir, maxLogSize, 150, 100, 350)

	// Expected files per session
	session1Files := []fileInfo{
		{filename: string(sessionID1) + "_lau_000" + logFileSuffix, size: 50},
		{filename: string(sessionID1) + "_gui_000" + logFileSuffix, size: 50},
		{filename: string(sessionID1) + "_bri_000" + logFileSuffix, size: 100},
		{filename: string(sessionID1) + "_bri_001" + logFileSuffix, size: 100},
		{filename: string(sessionID1) + "_bri_002" + logFileSuffix, size: 50},
	}

	session2Files := []fileInfo{
		{filename: string(sessionID2) + "_lau_000" + logFileSuffix, size: 100},
		{filename: string(sessionID2) + "_gui_000" + logFileSuffix, size: 100},
		{filename: string(sessionID2) + "_bri_000" + logFileSuffix, size: 100},
		{filename: string(sessionID2) + "_bri_001" + logFileSuffix, size: 100},
		{filename: string(sessionID2) + "_bri_002" + logFileSuffix, size: 100},
		{filename: string(sessionID2) + "_bri_003" + logFileSuffix, size: 50},
	}

	session3Files := []fileInfo{
		{filename: string(sessionID3) + "_lau_000" + logFileSuffix, size: 100},
		{filename: string(sessionID3) + "_lau_001" + logFileSuffix, size: 50},
		{filename: string(sessionID3) + "_gui_000" + logFileSuffix, size: 100},
		{filename: string(sessionID3) + "_bri_000" + logFileSuffix, size: 100},
		{filename: string(sessionID3) + "_bri_001" + logFileSuffix, size: 100},
		{filename: string(sessionID3) + "_bri_002" + logFileSuffix, size: 100},
		{filename: string(sessionID3) + "_bri_003" + logFileSuffix, size: 50},
	}

	allSessions := session1Files
	allSessions = append(allSessions, append(session2Files, session3Files...)...)
	checkFolderContent(t, dir, allSessions...)

	failureCount, err := pruneLogs(dir, sessionID3, 2000) // nothing to prune
	require.Equal(t, failureCount, 0)
	require.NoError(t, err)
	checkFolderContent(t, dir, allSessions...)

	failureCount, err = pruneLogs(dir, sessionID3, 1200) // session 1 is pruned
	require.Equal(t, failureCount, 0)
	require.NoError(t, err)

	checkFolderContent(t, dir, append(session2Files, session3Files...)...)
	failureCount, err = pruneLogs(dir, sessionID3, 1000) // session 2 is pruned
	require.Equal(t, failureCount, 0)
	require.NoError(t, err)

	checkFolderContent(t, dir, session3Files...)
}

func TestLogging_PruningBigCurrentSession(t *testing.T) {
	dir := t.TempDir()
	const maxLogFileSize = 1000
	sessionID1 := createDummySession(t, dir, maxLogFileSize, 500, 500, 2500)
	sessionID2 := createDummySession(t, dir, maxLogFileSize, 1000, 1000, 3500)
	sessionID3 := createDummySession(t, dir, maxLogFileSize, 500, 500, 10500)

	// Expected files per session
	session1Files := []fileInfo{
		{filename: string(sessionID1) + "_lau_000" + logFileSuffix, size: 500},
		{filename: string(sessionID1) + "_gui_000" + logFileSuffix, size: 500},
		{filename: string(sessionID1) + "_bri_000" + logFileSuffix, size: 1000},
		{filename: string(sessionID1) + "_bri_001" + logFileSuffix, size: 1000},
		{filename: string(sessionID1) + "_bri_002" + logFileSuffix, size: 500},
	}

	session2Files := []fileInfo{
		{filename: string(sessionID2) + "_lau_000" + logFileSuffix, size: 1000},
		{filename: string(sessionID2) + "_gui_000" + logFileSuffix, size: 1000},
		{filename: string(sessionID2) + "_bri_000" + logFileSuffix, size: 1000},
		{filename: string(sessionID2) + "_bri_001" + logFileSuffix, size: 1000},
		{filename: string(sessionID2) + "_bri_002" + logFileSuffix, size: 1000},
		{filename: string(sessionID2) + "_bri_003" + logFileSuffix, size: 500},
	}

	session3Files := []fileInfo{
		{filename: string(sessionID3) + "_lau_000" + logFileSuffix, size: 500},
		{filename: string(sessionID3) + "_gui_000" + logFileSuffix, size: 500},
		{filename: string(sessionID3) + "_bri_000" + logFileSuffix, size: 1000},
		{filename: string(sessionID3) + "_bri_001" + logFileSuffix, size: 1000},
		{filename: string(sessionID3) + "_bri_002" + logFileSuffix, size: 1000},
		{filename: string(sessionID3) + "_bri_003" + logFileSuffix, size: 1000},
		{filename: string(sessionID3) + "_bri_004" + logFileSuffix, size: 1000},
		{filename: string(sessionID3) + "_bri_005" + logFileSuffix, size: 1000},
		{filename: string(sessionID3) + "_bri_006" + logFileSuffix, size: 1000},
		{filename: string(sessionID3) + "_bri_007" + logFileSuffix, size: 1000},
		{filename: string(sessionID3) + "_bri_008" + logFileSuffix, size: 1000},
		{filename: string(sessionID3) + "_bri_009" + logFileSuffix, size: 1000},
		{filename: string(sessionID3) + "_bri_010" + logFileSuffix, size: 500},
	}

	allSessions := session1Files
	allSessions = append(allSessions, append(session2Files, session3Files...)...)
	checkFolderContent(t, dir, allSessions...)

	// current session is bigger than maxFileSize. We keep launcher and gui logs, the first and last bridge log
	// and only the last bridge log that keep the total file size under the limit.
	failureCount, err := pruneLogs(dir, sessionID3, 8000)
	require.Equal(t, failureCount, 0)
	require.NoError(t, err)
	checkFolderContent(t, dir, []fileInfo{
		{filename: string(sessionID3) + "_lau_000" + logFileSuffix, size: 500},
		{filename: string(sessionID3) + "_gui_000" + logFileSuffix, size: 500},
		{filename: string(sessionID3) + "_bri_000" + logFileSuffix, size: 1000},
		{filename: string(sessionID3) + "_bri_005" + logFileSuffix, size: 1000},
		{filename: string(sessionID3) + "_bri_006" + logFileSuffix, size: 1000},
		{filename: string(sessionID3) + "_bri_007" + logFileSuffix, size: 1000},
		{filename: string(sessionID3) + "_bri_008" + logFileSuffix, size: 1000},
		{filename: string(sessionID3) + "_bri_009" + logFileSuffix, size: 1000},
		{filename: string(sessionID3) + "_bri_010" + logFileSuffix, size: 500},
	}...)

	failureCount, err = pruneLogs(dir, sessionID3, 5000)
	require.Equal(t, failureCount, 0)
	require.NoError(t, err)
	checkFolderContent(t, dir, []fileInfo{
		{filename: string(sessionID3) + "_lau_000" + logFileSuffix, size: 500},
		{filename: string(sessionID3) + "_gui_000" + logFileSuffix, size: 500},
		{filename: string(sessionID3) + "_bri_000" + logFileSuffix, size: 1000},
		{filename: string(sessionID3) + "_bri_008" + logFileSuffix, size: 1000},
		{filename: string(sessionID3) + "_bri_009" + logFileSuffix, size: 1000},
		{filename: string(sessionID3) + "_bri_010" + logFileSuffix, size: 500},
	}...)

	// whatever maxFileSize is, we will always keep the following files
	minimalFiles := []fileInfo{
		{filename: string(sessionID3) + "_lau_000" + logFileSuffix, size: 500},
		{filename: string(sessionID3) + "_gui_000" + logFileSuffix, size: 500},
		{filename: string(sessionID3) + "_bri_000" + logFileSuffix, size: 1000},
		{filename: string(sessionID3) + "_bri_010" + logFileSuffix, size: 500},
	}
	failureCount, err = pruneLogs(dir, sessionID3, 2000)
	require.Equal(t, failureCount, 0)
	require.NoError(t, err)
	checkFolderContent(t, dir, minimalFiles...)

	failureCount, err = pruneLogs(dir, sessionID3, 0)
	require.Equal(t, failureCount, 0)
	require.NoError(t, err)
	checkFolderContent(t, dir, minimalFiles...)
}

func createDummySession(t *testing.T, dir string, maxLogFileSize int64, launcherLogSize, guiLogSize, bridgeLogSize int64) SessionID {
	time.Sleep(2 * time.Millisecond) // ensure our sessionID is unused.
	sessionID := NewSessionID()
	if launcherLogSize > 0 {
		createDummyRotatedLogFile(t, dir, sessionID, LauncherShortAppName, launcherLogSize, maxLogFileSize)
	}

	if guiLogSize > 0 {
		createDummyRotatedLogFile(t, dir, sessionID, GUIShortAppName, guiLogSize, maxLogFileSize)
	}

	if bridgeLogSize > 0 {
		createDummyRotatedLogFile(t, dir, sessionID, BridgeShortAppName, bridgeLogSize, maxLogFileSize)
	}

	return sessionID
}

func createDummyRotatedLogFile(t *testing.T, dir string, sessionID SessionID, appName AppName, totalSize, maxLogFileSize int64) {
	rotator, err := NewDefaultRotator(dir, sessionID, appName, maxLogFileSize, NoPruning)
	require.NoError(t, err)
	for i := int64(0); i < totalSize/maxLogFileSize; i++ {
		count, err := rotator.Write(make([]byte, maxLogFileSize))
		require.NoError(t, err)
		require.Equal(t, int64(count), maxLogFileSize)
	}

	remainder := totalSize % maxLogFileSize
	if remainder > 0 {
		count, err := rotator.Write(make([]byte, remainder))
		require.NoError(t, err)
		require.Equal(t, int64(count), remainder)
	}

	require.NoError(t, rotator.wc.Close())
}

func checkFolderContent(t *testing.T, dir string, fileInfos ...fileInfo) {
	for _, fi := range fileInfos {
		checkFileExistsWithSize(t, dir, fi)
	}

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Equal(t, len(fileInfos), len(entries))
}

func checkFileExistsWithSize(t *testing.T, dir string, info fileInfo) {
	stat, err := os.Stat(filepath.Join(dir, info.filename))
	require.NoError(t, err)
	require.Equal(t, stat.Size(), info.size)
}
