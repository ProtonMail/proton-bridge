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

package logging

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestClearLogs tests that cearLogs removes only bridge old log files keeping last three of them.
func TestClearLogs(t *testing.T) {
	dir, err := ioutil.TempDir("", "clear-logs-test")
	require.NoError(t, err)

	require.NoError(t, ioutil.WriteFile(filepath.Join(dir, "other.log"), []byte("Hello"), 0o755))
	require.NoError(t, ioutil.WriteFile(filepath.Join(dir, "v1_10.log"), []byte("Hello"), 0o755))
	require.NoError(t, ioutil.WriteFile(filepath.Join(dir, "v1_11.log"), []byte("Hello"), 0o755))
	require.NoError(t, ioutil.WriteFile(filepath.Join(dir, "v2_12.log"), []byte("Hello"), 0o755))
	require.NoError(t, ioutil.WriteFile(filepath.Join(dir, "v2_13.log"), []byte("Hello"), 0o755))

	require.NoError(t, clearLogs(dir, 3, 0))
	checkFileNames(t, dir, []string{
		"other.log",
		"v1_11.log",
		"v2_12.log",
		"v2_13.log",
	})
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
