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

package sum

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRecursiveSum(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "verify-test")
	require.NoError(t, err)

	createFiles(t, tempDir,
		filepath.Join("a", "1"),
		filepath.Join("a", "2"),
		filepath.Join("b", "3"),
		filepath.Join("b", "4"),
		filepath.Join("b", "c", "5"),
		filepath.Join("b", "c", "6"),
	)

	sumOriginal := sum(t, tempDir)

	// Renaming files should produce a different checksum.
	require.NoError(t, os.Rename(filepath.Join(tempDir, "a", "1"), filepath.Join(tempDir, "a", "11")))
	sumRenamed := sum(t, tempDir)
	require.NotEqual(t, sumOriginal, sumRenamed)

	// Reverting to the original name should produce the same checksum again.
	require.NoError(t, os.Rename(filepath.Join(tempDir, "a", "11"), filepath.Join(tempDir, "a", "1")))
	require.Equal(t, sumOriginal, sum(t, tempDir))

	// Moving files should produce a different checksum.
	require.NoError(t, os.Rename(filepath.Join(tempDir, "a", "1"), filepath.Join(tempDir, "1")))
	sumMoved := sum(t, tempDir)
	require.NotEqual(t, sumOriginal, sumMoved)

	// Moving files back to their original location should produce the same checksum again.
	require.NoError(t, os.Rename(filepath.Join(tempDir, "1"), filepath.Join(tempDir, "a", "1")))
	require.Equal(t, sumOriginal, sum(t, tempDir))

	// Changing file data should produce a different checksum.
	originalData := modifyFile(t, filepath.Join(tempDir, "a", "1"), []byte("something"))
	require.NotEqual(t, sumOriginal, sum(t, tempDir))

	// Reverting file data should produce the original checksum.
	modifyFile(t, filepath.Join(tempDir, "a", "1"), originalData)
	require.Equal(t, sumOriginal, sum(t, tempDir))
}

func createFiles(t *testing.T, root string, paths ...string) {
	for _, path := range paths {
		makeFile(t, filepath.Join(root, path))
	}
}

func makeFile(t *testing.T, path string) {
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))

	f, err := os.Create(path)
	require.NoError(t, err)

	_, err = f.WriteString(path)
	require.NoError(t, err)

	require.NoError(t, f.Close())
}

func sum(t *testing.T, path string) []byte {
	sum, err := RecursiveSum(path, "")
	require.NoError(t, err)

	return sum
}

func modifyFile(t *testing.T, path string, data []byte) []byte {
	r, err := os.Open(path)
	require.NoError(t, err)

	b, err := ioutil.ReadAll(r)
	require.NoError(t, err)
	require.NoError(t, r.Close())

	f, err := os.Create(path)
	require.NoError(t, err)

	_, err = f.Write(data)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	return b
}
