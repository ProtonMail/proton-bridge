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

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

const testAppName = "bridge-test"

var testConfigDir string //nolint[gochecknoglobals]

func TestMain(m *testing.M) {
	setupTestConfig()
	setupTestLogs()
	code := m.Run()
	shutdownTestConfig()
	shutdownTestLogs()
	shutdownTestPreferences()
	os.Exit(code)
}

func setupTestConfig() {
	var err error
	testConfigDir, err = ioutil.TempDir("", "config")
	if err != nil {
		panic(err)
	}
}

func shutdownTestConfig() {
	_ = os.RemoveAll(testConfigDir)
}

type mocks struct {
	t *testing.T

	ctrl          *gomock.Controller
	appDir        *MockappDirer
	appDirVersion *MockappDirer
}

func initMocks(t *testing.T) mocks {
	mockCtrl := gomock.NewController(t)
	return mocks{
		t: t,

		ctrl:          mockCtrl,
		appDir:        NewMockappDirer(mockCtrl),
		appDirVersion: NewMockappDirer(mockCtrl),
	}
}

func TestClearDataLinux(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	createTestStructureLinux(m, testConfigDir)
	cfg := newConfig(testAppName, "v1", "rev123", "c2", m.appDir, m.appDirVersion)
	require.NoError(t, cfg.ClearData())
	checkFileNames(t, testConfigDir, []string{
		"cache",
		"cache/c2",
		"cache/c2/bridge-test.lock",
		"config",
		"logs",
	})
}

func TestClearDataWindows(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	createTestStructureWindows(m, testConfigDir)
	cfg := newConfig(testAppName, "v1", "rev123", "c2", m.appDir, m.appDirVersion)
	require.NoError(t, cfg.ClearData())
	checkFileNames(t, testConfigDir, []string{
		"cache",
		"cache/c2",
		"cache/c2/bridge-test.lock",
		"config",
	})
}

// OldData touches only cache folder.
// Removes only c1 folder as nothing else is part of cache folder on Linux/Mac.
func TestClearOldDataLinux(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	createTestStructureLinux(m, testConfigDir)
	cfg := newConfig(testAppName, "v1", "rev123", "c2", m.appDir, m.appDirVersion)
	require.NoError(t, cfg.ClearOldData())
	checkFileNames(t, testConfigDir, []string{
		"cache",
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
		"logs/v1_10.log",
		"logs/v1_11.log",
		"logs/v2_12.log",
		"logs/v2_13.log",
	})
}

// OldData touches only cache folder. Removes everything except c2 folder
// and bridge log files which are part of cache folder on Windows.
func TestClearOldDataWindows(t *testing.T) {
	m := initMocks(t)
	defer m.ctrl.Finish()

	createTestStructureWindows(m, testConfigDir)
	cfg := newConfig(testAppName, "v1", "rev123", "c2", m.appDir, m.appDirVersion)
	require.NoError(t, cfg.ClearOldData())
	checkFileNames(t, testConfigDir, []string{
		"cache",
		"cache/c2",
		"cache/c2/bridge-test.lock",
		"cache/c2/events.json",
		"cache/c2/mailbox-user@pm.me.db",
		"cache/c2/prefs.json",
		"cache/c2/updates",
		"cache/c2/user_info.json",
		"cache/v1_10.log",
		"cache/v1_11.log",
		"cache/v2_12.log",
		"cache/v2_13.log",
		"config",
		"config/cert.pem",
		"config/key.pem",
	})
}

func createTestStructureLinux(m mocks, baseDir string) {
	logsDir := filepath.Join(baseDir, "logs")
	configDir := filepath.Join(baseDir, "config")
	cacheDir := filepath.Join(baseDir, "cache")
	versionedOldCacheDir := filepath.Join(baseDir, "cache", "c1")
	versionedCacheDir := filepath.Join(baseDir, "cache", "c2")
	createTestStructure(m, baseDir, logsDir, configDir, cacheDir, versionedOldCacheDir, versionedCacheDir)
}

func createTestStructureWindows(m mocks, baseDir string) {
	logsDir := filepath.Join(baseDir, "cache")
	configDir := filepath.Join(baseDir, "config")
	cacheDir := filepath.Join(baseDir, "cache")
	versionedOldCacheDir := filepath.Join(baseDir, "cache", "c1")
	versionedCacheDir := filepath.Join(baseDir, "cache", "c2")
	createTestStructure(m, baseDir, logsDir, configDir, cacheDir, versionedOldCacheDir, versionedCacheDir)
}

func createTestStructure(m mocks, baseDir, logsDir, configDir, cacheDir, versionedOldCacheDir, versionedCacheDir string) {
	m.appDir.EXPECT().UserLogs().Return(logsDir).AnyTimes()
	m.appDir.EXPECT().UserConfig().Return(configDir).AnyTimes()
	m.appDir.EXPECT().UserCache().Return(cacheDir).AnyTimes()
	m.appDirVersion.EXPECT().UserCache().Return(versionedCacheDir).AnyTimes()

	require.NoError(m.t, os.RemoveAll(baseDir))
	require.NoError(m.t, os.MkdirAll(baseDir, 0700))
	require.NoError(m.t, os.MkdirAll(logsDir, 0700))
	require.NoError(m.t, os.MkdirAll(configDir, 0700))
	require.NoError(m.t, os.MkdirAll(cacheDir, 0700))
	require.NoError(m.t, os.MkdirAll(versionedOldCacheDir, 0700))
	require.NoError(m.t, os.MkdirAll(versionedCacheDir, 0700))
	require.NoError(m.t, os.MkdirAll(filepath.Join(versionedCacheDir, "updates"), 0700))

	require.NoError(m.t, ioutil.WriteFile(filepath.Join(logsDir, "other.log"), []byte("Hello"), 0755))
	require.NoError(m.t, ioutil.WriteFile(filepath.Join(logsDir, "v1_10.log"), []byte("Hello"), 0755))
	require.NoError(m.t, ioutil.WriteFile(filepath.Join(logsDir, "v1_11.log"), []byte("Hello"), 0755))
	require.NoError(m.t, ioutil.WriteFile(filepath.Join(logsDir, "v2_12.log"), []byte("Hello"), 0755))
	require.NoError(m.t, ioutil.WriteFile(filepath.Join(logsDir, "v2_13.log"), []byte("Hello"), 0755))

	require.NoError(m.t, ioutil.WriteFile(filepath.Join(configDir, "cert.pem"), []byte("Hello"), 0755))
	require.NoError(m.t, ioutil.WriteFile(filepath.Join(configDir, "key.pem"), []byte("Hello"), 0755))

	require.NoError(m.t, ioutil.WriteFile(filepath.Join(versionedOldCacheDir, "prefs.json"), []byte("Hello"), 0755))
	require.NoError(m.t, ioutil.WriteFile(filepath.Join(versionedOldCacheDir, "events.json"), []byte("Hello"), 0755))
	require.NoError(m.t, ioutil.WriteFile(filepath.Join(versionedOldCacheDir, "user_info.json"), []byte("Hello"), 0755))
	require.NoError(m.t, ioutil.WriteFile(filepath.Join(versionedOldCacheDir, "mailbox-user@pm.me.db"), []byte("Hello"), 0755))
	require.NoError(m.t, ioutil.WriteFile(filepath.Join(versionedCacheDir, "prefs.json"), []byte("Hello"), 0755))
	require.NoError(m.t, ioutil.WriteFile(filepath.Join(versionedCacheDir, "events.json"), []byte("Hello"), 0755))
	require.NoError(m.t, ioutil.WriteFile(filepath.Join(versionedCacheDir, "user_info.json"), []byte("Hello"), 0755))
	require.NoError(m.t, ioutil.WriteFile(filepath.Join(versionedCacheDir, testAppName+".lock"), []byte("Hello"), 0755))
	require.NoError(m.t, ioutil.WriteFile(filepath.Join(versionedCacheDir, "mailbox-user@pm.me.db"), []byte("Hello"), 0755))
}

func checkFileNames(t *testing.T, dir string, expectedFileNames []string) {
	fileNames := getFileNames(t, dir)
	require.Equal(t, expectedFileNames, fileNames)
}

func getFileNames(t *testing.T, dir string) []string {
	files, err := ioutil.ReadDir(dir)
	require.NoError(t, err)

	fileNames := []string{}
	for _, file := range files {
		fileNames = append(fileNames, file.Name())
		if file.IsDir() {
			subDir := filepath.Join(dir, file.Name())
			subFileNames := getFileNames(t, subDir)
			for _, subFileName := range subFileNames {
				fileNames = append(fileNames, file.Name()+"/"+subFileName)
			}
		}
	}
	return fileNames
}
