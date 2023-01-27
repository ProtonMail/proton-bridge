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

// TestClearLogs tests that cearLogs removes only bridge old log files keeping last three of them.
func TestClearLogs(t *testing.T) {
	dir := t.TempDir()

	// Create some old log files.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "other.log"), []byte("Hello"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "v2.4.7_debe87f2f5_0000000001.log"), []byte("Hello"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "v2.4.8_debe87f2f5_0000000002.log"), []byte("Hello"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "v2.4.9_debe87f2f5_0000000003.log"), []byte("Hello"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "v2.5.0_debe87f2f5_0000000004.log"), []byte("Hello"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "v2.5.1_debe87f2f5_0000000005.log"), []byte("Hello"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "v2.5.2_debe87f2f5_0000000006.log"), []byte("Hello"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "v2.5.3_debe87f2f5_0000000007.log"), []byte("Hello"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "v2.5.4_debe87f2f5_0000000008.log"), []byte("Hello"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "v2.5.5_debe87f2f5_0000000009.log"), []byte("Hello"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "v2.5.6_debe87f2f5_0000000010.log"), []byte("Hello"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "v2.5.7_debe87f2f5_0000000011.log"), []byte("Hello"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "v2.5.8_debe87f2f5_0000000012.log"), []byte("Hello"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "v2.5.12_debe87f2f5_0000000013.log"), []byte("Hello"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "v2.5.9_debe87f2f5_0000000014.log"), []byte("Hello"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "v2.5.10_debe87f2f5_0000000015.log"), []byte("Hello"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "v2.5.11_debe87f2f5_0000000016.log"), []byte("Hello"), 0o755))

	// Clear the logs.
	require.NoError(t, clearLogs(dir, 3, 0))

	// We should only clear matching files, and keep the 3 most recent ones.
	checkFileNames(t, dir, []string{
		"other.log",
		"v2.5.9_debe87f2f5_0000000014.log",
		"v2.5.10_debe87f2f5_0000000015.log",
		"v2.5.11_debe87f2f5_0000000016.log",
	})
}

func checkFileNames(t *testing.T, dir string, expectedFileNames []string) {
	require.ElementsMatch(t, expectedFileNames, getFileNames(t, dir))
}

func getFileNames(t *testing.T, dir string) []string {
	files, err := os.ReadDir(dir)
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
