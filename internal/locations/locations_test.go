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

package locations

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAppDirs struct {
	configDir, cacheDir string
}

func (dirs *fakeAppDirs) UserConfig() string {
	return dirs.configDir
}

func (dirs *fakeAppDirs) UserCache() string {
	return dirs.cacheDir
}

func TestClearRemovesEverythingExceptLockAndUpdateFiles(t *testing.T) {
	l := newTestLocations(t)

	assert.NoError(t, l.Clear())

	assert.FileExists(t, l.GetLockFile())
	assert.DirExists(t, l.getSettingsPath())
	assert.NoFileExists(t, filepath.Join(l.getSettingsPath(), "prefs.json"))
	assert.NoDirExists(t, l.getLogsPath())
	assert.NoDirExists(t, l.getCachePath())
	assert.DirExists(t, l.getUpdatesPath())
}

func TestClearUpdateFiles(t *testing.T) {
	l := newTestLocations(t)

	assert.NoError(t, l.ClearUpdates())

	assert.FileExists(t, l.GetLockFile())
	assert.DirExists(t, l.getSettingsPath())
	assert.FileExists(t, filepath.Join(l.getSettingsPath(), "prefs.json"))
	assert.DirExists(t, l.getLogsPath())
	assert.DirExists(t, l.getCachePath())
	assert.NoDirExists(t, l.getUpdatesPath())
}

func TestCleanLeavesStandardLocationsUntouched(t *testing.T) {
	l := newTestLocations(t)

	createFilesInDir(t, l.getLogsPath(),
		"log1.txt",
		"log2.txt",
	)

	assert.NoError(t, l.Clean())

	assert.FileExists(t, l.GetLockFile())
	assert.DirExists(t, l.getSettingsPath())
	assert.FileExists(t, filepath.Join(l.getSettingsPath(), "prefs.json"))
	assert.DirExists(t, l.getLogsPath())
	assert.FileExists(t, filepath.Join(l.getLogsPath(), "log1.txt"))
	assert.FileExists(t, filepath.Join(l.getLogsPath(), "log2.txt"))
	assert.DirExists(t, l.getCachePath())
	assert.DirExists(t, l.getUpdatesPath())
}

func TestCleanRemovesUnexpectedFilesAndFolders(t *testing.T) {
	l := newTestLocations(t)

	createFilesInDir(t, l.userCache,
		"unexpected1.txt",
		"dir1/unexpected2.txt",
		"dir1/unexpected3.txt",
		"dir2/unexpected4.txt",
		"dir3/dir4/unexpected5.txt",
	)

	require.FileExists(t, filepath.Join(l.userCache, "unexpected1.txt"))
	require.FileExists(t, filepath.Join(l.userCache, "dir1", "unexpected2.txt"))
	require.FileExists(t, filepath.Join(l.userCache, "dir1", "unexpected3.txt"))
	require.FileExists(t, filepath.Join(l.userCache, "dir2", "unexpected4.txt"))
	require.FileExists(t, filepath.Join(l.userCache, "dir3", "dir4", "unexpected5.txt"))

	assert.NoError(t, l.Clean())

	assert.FileExists(t, l.GetLockFile())
	assert.DirExists(t, l.getSettingsPath())
	assert.DirExists(t, l.getLogsPath())
	assert.DirExists(t, l.getCachePath())
	assert.DirExists(t, l.getUpdatesPath())

	assert.NoFileExists(t, filepath.Join(l.userCache, "unexpected1.txt"))
	assert.NoFileExists(t, filepath.Join(l.userCache, "dir1", "unexpected2.txt"))
	assert.NoFileExists(t, filepath.Join(l.userCache, "dir1", "unexpected3.txt"))
	assert.NoFileExists(t, filepath.Join(l.userCache, "dir2", "unexpected4.txt"))
	assert.NoFileExists(t, filepath.Join(l.userCache, "dir3", "dir4", "unexpected5.txt"))
}

func newFakeAppDirs(t *testing.T) *fakeAppDirs {
	configDir, err := ioutil.TempDir("", "test-locations-config")
	require.NoError(t, err)

	cacheDir, err := ioutil.TempDir("", "test-locations-cache")
	require.NoError(t, err)

	return &fakeAppDirs{
		configDir: configDir,
		cacheDir:  cacheDir,
	}
}

func newTestLocations(t *testing.T) *Locations {
	l := New(newFakeAppDirs(t), "configName")

	lock := l.GetLockFile()
	createFilesInDir(t, "", lock)
	require.FileExists(t, lock)

	settings, err := l.ProvideSettingsPath()
	require.NoError(t, err)
	require.DirExists(t, settings)

	createFilesInDir(t, settings, "prefs.json")
	require.FileExists(t, filepath.Join(settings, "prefs.json"))

	logs, err := l.ProvideLogsPath()
	require.NoError(t, err)
	require.DirExists(t, logs)

	cache, err := l.ProvideCachePath()
	require.NoError(t, err)
	require.DirExists(t, cache)

	updates, err := l.ProvideUpdatesPath()
	require.NoError(t, err)
	require.DirExists(t, updates)

	return l
}

func createFilesInDir(t *testing.T, dir string, files ...string) {
	for _, target := range files {
		require.NoError(t, os.MkdirAll(filepath.Dir(filepath.Join(dir, target)), 0o700))

		f, err := os.Create(filepath.Join(dir, target))
		require.NoError(t, err)
		require.NoError(t, f.Close())
	}
}
