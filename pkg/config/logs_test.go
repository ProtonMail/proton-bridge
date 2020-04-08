// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type testLogConfig struct{ logDir, logPrefix string }

func (c *testLogConfig) GetLogDir() string    { return c.logDir }
func (c *testLogConfig) GetLogPrefix() string { return c.logPrefix }

var testLogDir string //nolint[gochecknoglobals]

func setupTestLogs() {
	var err error
	testLogDir, err = ioutil.TempDir("", "log")
	if err != nil {
		panic(err)
	}
}

func shutdownTestLogs() {
	_ = os.RemoveAll(testLogDir)
}

func TestLogNameLength(t *testing.T) {
	cfg := New("bridge-test", "longVersion123", "longRevision1234567890", "c2")
	name := getLogFilename(cfg.GetLogPrefix())
	if len(name) > 128 {
		t.Fatal("Name of the log is too long - limit for encrypted linux is 128 characters")
	}
}

// Info and higher levels writes to the file.
func TestSetupLogInfo(t *testing.T) {
	dir := beforeEachCreateTestDir(t, "setupInfo")

	SetupLog(&testLogConfig{dir, "v"}, "info")
	require.Equal(t, "info", logrus.GetLevel().String())

	logrus.Info("test message")
	files := checkLogFiles(t, dir, 1)
	checkLogContains(t, dir, files[0].Name(), "test message")
}

// Debug levels writes to stdout.
func TestSetupLogDebug(t *testing.T) {
	dir := beforeEachCreateTestDir(t, "setupDebug")

	SetupLog(&testLogConfig{dir, "v"}, "debug")
	require.Equal(t, "debug", logrus.GetLevel().String())

	logrus.Info("test message")
	checkLogFiles(t, dir, 0)
}

func TestReopenLogFile(t *testing.T) {
	dir := beforeEachCreateTestDir(t, "reopenLogFile")

	setLogFile(dir, "v1")

	done := make(chan interface{})

	log.Info("first message")

	go func() {
		<-done // Wait for closing file and opening new one.
		log.Info("second message")
		done <- nil
	}()

	closeLogFile()
	setLogFile(dir, "v2")

	done <- nil
	<-done // Wait for second log message.

	files := checkLogFiles(t, dir, 2)
	checkLogContains(t, dir, files[0].Name(), "first message")
	checkLogContains(t, dir, files[1].Name(), "second message")
}

func TestCheckLogFileSizeSmall(t *testing.T) {
	dir := beforeEachCreateTestDir(t, "logFileSizeSmall")

	setLogFile(dir, "v1")
	originalFileName := logFile.Name()

	_, _ = logFile.WriteString("small file")
	checkLogFileSize(dir, "v2")

	require.Equal(t, originalFileName, logFile.Name())
}

func TestCheckLogFileSizeBig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	dir := beforeEachCreateTestDir(t, "logFileSizeBig")

	setLogFile(dir, "v1")
	originalFileName := logFile.Name()

	// The limit for big file is 10*1024*1024 - keep the string 10 letters long.
	for i := 0; i < 1024*1024; i++ {
		_, _ = logFile.WriteString("big file!\n")
	}
	checkLogFileSize(dir, "v2")

	require.NotEqual(t, originalFileName, logFile.Name())
}

// ClearLogs removes only bridge old log files keeping last three of them.
func TestClearLogsLinux(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	dir := beforeEachCreateTestDir(t, "clearLogs")

	createTestStructureLinux(m, dir)
	require.NoError(t, clearLogs(dir))
	checkFileNames(t, dir, []string{
		"cache",
		"cache/c1",
		"cache/c1/events.json",
		"cache/c1/mailbox-user@pm.me.db",
		"cache/c1/prefs.json",
		"cache/c1/user_info.json",
		"cache/c2",
		"cache/c2/bridge-test.lock",
		"cache/c2/events.json",
		"cache/c2/mailbox-user@pm.me.db",
		"cache/c2/prefs.json",
		"cache/c2/updates",
		"cache/c2/user_info.json",
		"config",
		"config/cert.pem",
		"config/key.pem",
		"logs",
		"logs/other.log",
		"logs/v1_11.log",
		"logs/v2_12.log",
		"logs/v2_13.log",
	})
}

// ClearLogs removes only bridge old log files even when log folder
// is shared with other files on Windows.
func TestClearLogsWindows(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	dir := beforeEachCreateTestDir(t, "clearLogs")

	createTestStructureWindows(m, dir)
	require.NoError(t, clearLogs(dir))
	checkFileNames(t, dir, []string{
		"cache",
		"cache/c1",
		"cache/c1/events.json",
		"cache/c1/mailbox-user@pm.me.db",
		"cache/c1/prefs.json",
		"cache/c1/user_info.json",
		"cache/c2",
		"cache/c2/bridge-test.lock",
		"cache/c2/events.json",
		"cache/c2/mailbox-user@pm.me.db",
		"cache/c2/prefs.json",
		"cache/c2/updates",
		"cache/c2/user_info.json",
		"cache/other.log",
		"cache/v1_11.log",
		"cache/v2_12.log",
		"cache/v2_13.log",
		"config",
		"config/cert.pem",
		"config/key.pem",
	})
}

func beforeEachCreateTestDir(t *testing.T, dir string) string {
	// Make sure opened file (from the previous test) is cleared.
	closeLogFile()

	dir = filepath.Join(testLogDir, dir)
	require.NoError(t, os.MkdirAll(dir, 0700))
	return dir
}

func checkLogFiles(t *testing.T, dir string, expectedCount int) []os.FileInfo {
	files, err := ioutil.ReadDir(dir)
	require.NoError(t, err)
	require.Equal(t, expectedCount, len(files))
	return files
}

func checkLogContains(t *testing.T, dir, fileName, expectedSubstr string) {
	data, err := ioutil.ReadFile(filepath.Join(dir, fileName)) //nolint[gosec]
	require.NoError(t, err)
	require.Contains(t, string(data), expectedSubstr)
}
