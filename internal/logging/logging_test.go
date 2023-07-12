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
	"testing"

	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestLogging_GetLogTime(t *testing.T) {
	sessionID := NewSessionID()
	fp := defaultFileProvider(os.TempDir(), sessionID, "bridge-test")
	wc, err := fp(0)
	require.NoError(t, err)
	file, ok := wc.(*os.File)
	require.True(t, ok)

	sessionIDTime := sessionID.toTime()
	require.False(t, sessionIDTime.IsZero())
	logTime := getLogTime(filepath.Base(file.Name()))
	require.False(t, logTime.IsZero())
	require.Equal(t, sessionIDTime, logTime)
}

func TestLogging_MatchLogName(t *testing.T) {
	bridgeLog := "20230602_094633102_bri_000_v3.0.99+git_5b650b1be3.log"
	crashLog := "20230602_094633102_bri_000_v3.0.99+git_5b650b1be3_crash.log"
	guiLog := "20230602_094633102_gui_000_v3.0.99+git_5b650b1be3.log"
	launcherLog := "20230602_094633102_lau_000_v3.0.99+git_5b650b1be3.log"
	require.True(t, MatchBridgeLogName(bridgeLog))
	require.False(t, MatchGUILogName(bridgeLog))
	require.False(t, MatchLauncherLogName(bridgeLog))
	require.True(t, MatchBridgeLogName(crashLog))
	require.False(t, MatchGUILogName(crashLog))
	require.False(t, MatchLauncherLogName(crashLog))
	require.False(t, MatchBridgeLogName(guiLog))
	require.True(t, MatchGUILogName(guiLog))
	require.False(t, MatchLauncherLogName(guiLog))
	require.False(t, MatchBridgeLogName(launcherLog))
	require.False(t, MatchGUILogName(launcherLog))
	require.True(t, MatchLauncherLogName(launcherLog))
}

func TestLogging_GetOrderedLogFileListForBugReport(t *testing.T) {
	dir := t.TempDir()

	filePaths, err := getOrderedLogFileListForBugReport(dir, 3)
	require.NoError(t, err)
	require.True(t, len(filePaths) == 0)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "invalid.log"), []byte("proton"), 0660))

	_ = createDummySession(t, dir, 1000, 250, 500, 3000)
	sessionID1 := createDummySession(t, dir, 1000, 250, 500, 500)
	sessionID2 := createDummySession(t, dir, 1000, 250, 500, 500)
	sessionID3 := createDummySession(t, dir, 1000, 250, 500, 4500)

	filePaths, err = getOrderedLogFileListForBugReport(dir, 3)
	fileSuffix := "_v" + constants.Version + "_" + constants.Tag + ".log"
	require.NoError(t, err)
	require.EqualValues(t, []string{
		filepath.Join(dir, string(sessionID3)+"_bri_004"+fileSuffix),
		filepath.Join(dir, string(sessionID3)+"_bri_003"+fileSuffix),
		filepath.Join(dir, string(sessionID3)+"_bri_000"+fileSuffix),
		filepath.Join(dir, string(sessionID3)+"_gui_000"+fileSuffix),
		filepath.Join(dir, string(sessionID3)+"_lau_000"+fileSuffix),
		filepath.Join(dir, string(sessionID3)+"_bri_001"+fileSuffix),
		filepath.Join(dir, string(sessionID3)+"_bri_002"+fileSuffix),
		filepath.Join(dir, string(sessionID2)+"_bri_000"+fileSuffix),
		filepath.Join(dir, string(sessionID2)+"_gui_000"+fileSuffix),
		filepath.Join(dir, string(sessionID2)+"_lau_000"+fileSuffix),
		filepath.Join(dir, string(sessionID1)+"_bri_000"+fileSuffix),
		filepath.Join(dir, string(sessionID1)+"_gui_000"+fileSuffix),
		filepath.Join(dir, string(sessionID1)+"_lau_000"+fileSuffix),
	}, filePaths)
}

func TestLogging_Close(t *testing.T) {
	d := t.TempDir()
	closer, err := Init(d, NewSessionID(), constants.AppName, 1, DefaultPruningSize, "debug")
	require.NoError(t, err)
	logrus.Debug("Test") // because we set max log file size to 1, this will force a rotation of the log file.
	require.NotNil(t, closer)
	require.NoError(t, closer.Close())
}
