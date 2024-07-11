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

package logging

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type WriteCloser struct {
	bytes.Buffer
}

func (c *WriteCloser) Close() error {
	return nil
}

func TestLogging_Rotator(t *testing.T) {
	n := 0

	getFile := func(_ int) (io.WriteCloser, error) {
		n++
		return &WriteCloser{}, nil
	}

	r, err := NewRotator(10, getFile, nullPruner)
	require.NoError(t, err)

	_, err = r.Write([]byte("12345"))
	require.NoError(t, err)
	assert.Equal(t, 1, n)

	_, err = r.Write([]byte("12345"))
	require.NoError(t, err)
	assert.Equal(t, 1, n)

	_, err = r.Write([]byte("01234"))
	require.NoError(t, err)
	assert.Equal(t, 2, n)

	_, err = r.Write([]byte("01234"))
	require.NoError(t, err)
	assert.Equal(t, 2, n)

	_, err = r.Write([]byte("01234"))
	require.NoError(t, err)
	assert.Equal(t, 3, n)

	_, err = r.Write([]byte("01234"))
	require.NoError(t, err)
	assert.Equal(t, 3, n)

	_, err = r.Write([]byte("01234"))
	require.NoError(t, err)
	assert.Equal(t, 4, n)
}

func TestLogging_DefaultRotator(t *testing.T) {
	fiveBytes := []byte("00000")
	tmpDir := os.TempDir()

	sessionID := NewSessionID()
	basePath := filepath.Join(tmpDir, string(sessionID))

	r, err := NewDefaultRotator(tmpDir, sessionID, "bri", 10, NoPruning)
	require.NoError(t, err)
	require.Equal(t, 1, countFilesMatching(basePath+"_bri_000_*.log"))
	require.Equal(t, 1, countFilesMatching(basePath+"*.log"))

	_, err = r.Write(fiveBytes)
	require.NoError(t, err)
	require.Equal(t, 1, countFilesMatching(basePath+"*.log"))

	_, err = r.Write(fiveBytes)
	require.NoError(t, err)
	require.Equal(t, 1, countFilesMatching(basePath+"*.log"))

	_, err = r.Write(fiveBytes)
	require.NoError(t, err)
	require.Equal(t, 2, countFilesMatching(basePath+"*.log"))
	require.Equal(t, 1, countFilesMatching(basePath+"_bri_001_*.log"))

	for i := 0; i < 4; i++ {
		_, err = r.Write(fiveBytes)
		require.NoError(t, err)
	}

	require.NoError(t, r.wc.Close())

	// total written: 35 bytes, i.e. 4 log files
	logFileCount := countFilesMatching(basePath + "*.log")
	require.Equal(t, 4, logFileCount)
	for i := 0; i < logFileCount; i++ {
		require.Equal(t, 1, countFilesMatching(basePath+fmt.Sprintf("_bri_%03d_*.log", i)))
	}

	cleanupLogs(t, sessionID)
}

func TestLogging_DefaultRotatorWithPruning(t *testing.T) {
	tenBytes := []byte("0000000000")
	tmpDir := t.TempDir()

	sessionID := NewSessionID()
	basePath := filepath.Join(tmpDir, string(sessionID))

	// fill the log dir while below the pruning quota
	r, err := NewDefaultRotator(tmpDir, sessionID, "bri", 10, 40)
	require.NoError(t, err)
	for i := 0; i < 4; i++ {
		_, err = r.Write(tenBytes)
		require.NoError(t, err)
	}

	// from now on at every rotation, (i.e. every write in this case), we will prune, then create a new file.
	// we should always have 4 files, remaining after prune, plus the newly rotated file with the last written bytes.
	for i := 0; i < 10; i++ {
		_, err := r.Write(tenBytes)
		require.NoError(t, err)
		require.Equal(t, 5, countFilesMatching(basePath+"_bri_*.log"))
	}

	require.NoError(t, r.wc.Close())

	// Final check. 000, 010, 011, 012 are what's left after the last pruning, 013 never got to pass through pruning.
	checkFolderContent(t, tmpDir, []fileInfo{
		{filename: string(sessionID) + "_bri_000" + logFileSuffix, size: 10},
		{filename: string(sessionID) + "_bri_010" + logFileSuffix, size: 10},
		{filename: string(sessionID) + "_bri_011" + logFileSuffix, size: 10},
		{filename: string(sessionID) + "_bri_012" + logFileSuffix, size: 10},
		{filename: string(sessionID) + "_bri_013" + logFileSuffix, size: 10},
	}...)
}

func BenchmarkRotate(b *testing.B) {
	benchRotate(b, DefaultMaxLogFileSize, getTestFile(b, b.TempDir(), DefaultMaxLogFileSize-1))
}

func benchRotate(b *testing.B, logSize int64, getFile func(index int) (io.WriteCloser, error)) {
	r, err := NewRotator(logSize, getFile, nullPruner)
	require.NoError(b, err)

	for n := 0; n < b.N; n++ {
		require.NoError(b, r.rotate())

		f, ok := r.wc.(*os.File)
		require.True(b, ok)
		require.NoError(b, os.Remove(f.Name()))
	}
}

func getTestFile(b *testing.B, dir string, length int) func(int) (io.WriteCloser, error) {
	return func(_ int) (io.WriteCloser, error) {
		b.StopTimer()
		defer b.StartTimer()

		f, err := os.CreateTemp(dir, "log")
		if err != nil {
			return nil, err
		}

		if _, err := f.Write(make([]byte, length)); err != nil {
			return nil, err
		}

		if err := f.Sync(); err != nil {
			return nil, err
		}

		return f, nil
	}
}

func countFilesMatching(pattern string) int {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return -1
	}

	return len(files)
}

func cleanupLogs(t *testing.T, sessionID SessionID) {
	paths, err := filepath.Glob(filepath.Join(os.TempDir(), string(sessionID)+"*.log"))
	require.NoError(t, err)
	for _, path := range paths {
		require.NoError(t, os.Remove(path))
	}
}
