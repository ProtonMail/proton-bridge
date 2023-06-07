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
