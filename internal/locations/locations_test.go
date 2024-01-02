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

package locations

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAppDirs struct {
	configDir, dataDir, cacheDir string
}

func (dirs *fakeAppDirs) UserConfig() string {
	return dirs.configDir
}

func (dirs *fakeAppDirs) UserData() string {
	return dirs.dataDir
}

func (dirs *fakeAppDirs) UserCache() string {
	return dirs.cacheDir
}

func TestClearRemovesEverythingExceptLockAndUpdateFiles(t *testing.T) {
	l := newTestLocations(t)

	assert.NoError(t, l.Clear())

	assert.NoFileExists(t, filepath.Join(l.getSettingsPath(), "prefs.json"))
	assert.NoDirExists(t, l.getLogsPath())
	assert.DirExists(t, l.getUpdatesPath())
}

func TestClearUpdateFiles(t *testing.T) {
	l := newTestLocations(t)

	assert.NoError(t, l.ClearUpdates())

	assert.DirExists(t, l.getSettingsPath())
	assert.FileExists(t, filepath.Join(l.getSettingsPath(), "prefs.json"))
	assert.DirExists(t, l.getLogsPath())
	assert.NoDirExists(t, l.getUpdatesPath())
}

func TestRemoveOldGoIMAPCacheFolders(t *testing.T) {
	l := newTestLocations(t)

	createFilesInDir(t,
		l.getGoIMAPCachePath(),
		"foo",
		"bar",
	)

	require.FileExists(t, filepath.Join(l.getGoIMAPCachePath(), "foo"))
	require.FileExists(t, filepath.Join(l.getGoIMAPCachePath(), "bar"))

	assert.NoError(t, l.CleanGoIMAPCache())

	assert.DirExists(t, l.getSettingsPath())
	assert.DirExists(t, l.getLogsPath())
	assert.DirExists(t, l.getUpdatesPath())

	assert.NoFileExists(t, filepath.Join(l.getGoIMAPCachePath(), "foo"))
	assert.NoFileExists(t, filepath.Join(l.getGoIMAPCachePath(), "bar"))
	assert.NoDirExists(t, l.getGoIMAPCachePath())
}

func newFakeAppDirs(t *testing.T) *fakeAppDirs {
	return &fakeAppDirs{
		configDir: t.TempDir(),
		dataDir:   t.TempDir(),
		cacheDir:  t.TempDir(),
	}
}

func newTestLocations(t *testing.T) *Locations {
	l := New(newFakeAppDirs(t), "configName")

	settings, err := l.ProvideSettingsPath()
	require.NoError(t, err)
	require.DirExists(t, settings)

	createFilesInDir(t, settings, "prefs.json")
	require.FileExists(t, filepath.Join(settings, "prefs.json"))

	logs, err := l.ProvideLogsPath()
	require.NoError(t, err)
	require.DirExists(t, logs)

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
