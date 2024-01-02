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

package sum

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRecursiveSum(t *testing.T) {
	dir := t.TempDir()

	createFiles(t, dir,
		filepath.Join("a", "1"),
		filepath.Join("a", "2"),
		filepath.Join("b", "3"),
		filepath.Join("b", "4"),
		filepath.Join("b", "c", "5"),
		filepath.Join("b", "c", "6"),
	)

	sumOriginal := sum(t, dir)

	// Renaming files should produce a different checksum.
	require.NoError(t, os.Rename(filepath.Join(dir, "a", "1"), filepath.Join(dir, "a", "11")))
	sumRenamed := sum(t, dir)
	require.NotEqual(t, sumOriginal, sumRenamed)

	// Reverting to the original name should produce the same checksum again.
	require.NoError(t, os.Rename(filepath.Join(dir, "a", "11"), filepath.Join(dir, "a", "1")))
	require.Equal(t, sumOriginal, sum(t, dir))

	// Moving files should produce a different checksum.
	require.NoError(t, os.Rename(filepath.Join(dir, "a", "1"), filepath.Join(dir, "1")))
	sumMoved := sum(t, dir)
	require.NotEqual(t, sumOriginal, sumMoved)

	// Moving files back to their original location should produce the same checksum again.
	require.NoError(t, os.Rename(filepath.Join(dir, "1"), filepath.Join(dir, "a", "1")))
	require.Equal(t, sumOriginal, sum(t, dir))

	// Changing file data should produce a different checksum.
	originalData := modifyFile(t, filepath.Join(dir, "a", "1"), []byte("something"))
	require.NotEqual(t, sumOriginal, sum(t, dir))

	// Reverting file data should produce the original checksum.
	modifyFile(t, filepath.Join(dir, "a", "1"), originalData)
	require.Equal(t, sumOriginal, sum(t, dir))
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

	b, err := io.ReadAll(r)
	require.NoError(t, err)
	require.NoError(t, r.Close())

	f, err := os.Create(path)
	require.NoError(t, err)

	_, err = f.Write(data)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	return b
}
