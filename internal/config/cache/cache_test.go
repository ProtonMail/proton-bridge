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

package cache

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveOldVersions(t *testing.T) {
	dir, err := ioutil.TempDir("", "test-cache")
	require.NoError(t, err)

	cache, err := New(dir, "c4")
	require.NoError(t, err)

	createFilesInDir(t, dir,
		"unexpected1.txt",
		"c1/unexpected1.txt",
		"c2/unexpected2.txt",
		"c3/unexpected3.txt",
		"something.txt",
	)

	require.DirExists(t, filepath.Join(dir, "c4"))
	require.FileExists(t, filepath.Join(dir, "unexpected1.txt"))
	require.FileExists(t, filepath.Join(dir, "c1", "unexpected1.txt"))
	require.FileExists(t, filepath.Join(dir, "c2", "unexpected2.txt"))
	require.FileExists(t, filepath.Join(dir, "c3", "unexpected3.txt"))
	require.FileExists(t, filepath.Join(dir, "something.txt"))

	assert.NoError(t, cache.RemoveOldVersions())

	assert.DirExists(t, filepath.Join(dir, "c4"))
	assert.NoFileExists(t, filepath.Join(dir, "unexpected1.txt"))
	assert.NoFileExists(t, filepath.Join(dir, "c1", "unexpected1.txt"))
	assert.NoFileExists(t, filepath.Join(dir, "c2", "unexpected2.txt"))
	assert.NoFileExists(t, filepath.Join(dir, "c3", "unexpected3.txt"))
	assert.NoFileExists(t, filepath.Join(dir, "something.txt"))
}

func createFilesInDir(t *testing.T, dir string, files ...string) {
	for _, target := range files {
		require.NoError(t, os.MkdirAll(filepath.Dir(filepath.Join(dir, target)), 0o700))

		f, err := os.Create(filepath.Join(dir, target))
		require.NoError(t, err)
		require.NoError(t, f.Close())
	}
}
