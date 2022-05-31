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

package files

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemove(t *testing.T) {
	dir := newTestDir(t,
		"subdir1",
		"subdir2/subdir3",
	)
	defer delTestDir(t, dir)

	createTestFiles(t, dir,
		"subdir1/file1",
		"subdir1/file2",
		"subdir2/file3",
		"subdir2/file4",
		"subdir2/subdir3/file5",
		"subdir2/subdir3/file6",
	)

	require.NoError(t, Remove(
		filepath.Join(dir, "subdir1"),
		filepath.Join(dir, "subdir2", "file3"),
		filepath.Join(dir, "subdir2", "subdir3", "file5"),
	).Do())

	assert.NoFileExists(t, filepath.Join(dir, "subdir1", "file1"))
	assert.NoFileExists(t, filepath.Join(dir, "subdir1", "file2"))
	assert.NoFileExists(t, filepath.Join(dir, "subdir2", "file3"))
	assert.FileExists(t, filepath.Join(dir, "subdir2", "file4"))
	assert.NoFileExists(t, filepath.Join(dir, "subdir2", "subdir3", "file5"))
	assert.FileExists(t, filepath.Join(dir, "subdir2", "subdir3", "file6"))
}

func TestRemoveWithExceptions(t *testing.T) {
	dir := newTestDir(t,
		"subdir1",
		"subdir2/subdir3",
		"subdir4",
	)
	defer delTestDir(t, dir)

	createTestFiles(t, dir,
		"subdir1/file1",
		"subdir1/file2",
		"subdir2/file3",
		"subdir2/file4",
		"subdir2/subdir3/file5",
		"subdir2/subdir3/file6",
		"subdir4/file7",
		"subdir4/file8",
	)

	require.NoError(t, Remove(dir).Except(
		filepath.Join(dir, "subdir2", "file4"),
		filepath.Join(dir, "subdir2", "subdir3", "file6"),
		filepath.Join(dir, "subdir4"),
	).Do())

	assert.NoFileExists(t, filepath.Join(dir, "subdir1", "file1"))
	assert.NoFileExists(t, filepath.Join(dir, "subdir1", "file2"))
	assert.NoFileExists(t, filepath.Join(dir, "subdir2", "file3"))
	assert.FileExists(t, filepath.Join(dir, "subdir2", "file4"))
	assert.NoFileExists(t, filepath.Join(dir, "subdir2", "subdir3", "file5"))
	assert.FileExists(t, filepath.Join(dir, "subdir2", "subdir3", "file6"))
	assert.FileExists(t, filepath.Join(dir, "subdir4", "file7"))
	assert.FileExists(t, filepath.Join(dir, "subdir4", "file8"))
}

func newTestDir(t *testing.T, subdirs ...string) string {
	dir, err := ioutil.TempDir("", "test-files-dir")
	require.NoError(t, err)

	for _, target := range subdirs {
		require.NoError(t, os.MkdirAll(filepath.Join(dir, target), 0o700))
	}

	return dir
}

func createTestFiles(t *testing.T, dir string, files ...string) {
	for _, target := range files {
		f, err := os.Create(filepath.Join(dir, target))
		require.NoError(t, err)
		require.NoError(t, f.Close())
	}
}

func delTestDir(t *testing.T, dir string) {
	require.NoError(t, os.RemoveAll(dir))
}
