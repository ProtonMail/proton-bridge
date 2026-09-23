// Copyright (c) 2026 Proton AG
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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package atomicfile

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteFile_CreatesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file")

	require.NoError(t, WriteFile(path, []byte("data")))

	requireContent(t, path, "data")
	requireOnlyEntry(t, path)

	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	}
}

func TestWriteFile_ReplacesExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file")

	oldStr := "old data that is longer"
	require.NoError(t, os.WriteFile(path, []byte(oldStr), 0o600))
	requireContent(t, path, "old data that is longer")

	require.NoError(t, WriteFile(path, []byte("new")))

	requireContent(t, path, "new")
	requireOnlyEntry(t, path)
}

func TestWriteFile_ReplaceFailureRemovesTempFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file")
	require.NoError(t, os.Mkdir(path, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(path, "child"), nil, 0o600))

	require.Error(t, WriteFile(path, []byte("data")))

	requireOnlyEntry(t, path)
}

func requireContent(t *testing.T, path, expected string) {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, expected, string(data))
}

func requireOnlyEntry(t *testing.T, path string) {
	t.Helper()

	entries, err := os.ReadDir(filepath.Dir(path))
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, filepath.Base(path), entries[0].Name())
}
