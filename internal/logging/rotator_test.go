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

func TestRotator(t *testing.T) {
	n := 0

	getFile := func(_ int) (io.WriteCloser, error) {
		n++
		return &WriteCloser{}, nil
	}

	r, err := NewRotator(10, getFile)
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

func TestDefaultRotator(t *testing.T) {
	fiveBytes := []byte("00000")
	tmpDir := os.TempDir()

	sessionID := NewSessionID()
	basePath := filepath.Join(tmpDir, string(sessionID))

	r, err := NewDefaultRotator(tmpDir, sessionID, "bridge", 10)
	require.NoError(t, err)
	require.Equal(t, 1, countFilesMatching(basePath+"_000_*.log"))
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
	require.Equal(t, 1, countFilesMatching(basePath+"_001_*.log"))

	for i := 0; i < 4; i++ {
		_, err = r.Write(fiveBytes)
		require.NoError(t, err)
	}

	require.NoError(t, r.wc.Close())

	// total written: 35 bytes, i.e. 4 log files
	logFileCount := countFilesMatching(basePath + "*.log")
	require.Equal(t, 4, logFileCount)
	for i := 0; i < logFileCount; i++ {
		require.Equal(t, 1, countFilesMatching(basePath+fmt.Sprintf("_%03d_*.log", i)))
	}

	cleanupLogs(t, sessionID)
}

func BenchmarkRotate(b *testing.B) {
	benchRotate(b, MaxLogSize, getTestFile(b, b.TempDir(), MaxLogSize-1))
}

func benchRotate(b *testing.B, logSize int, getFile func(index int) (io.WriteCloser, error)) {
	r, err := NewRotator(logSize, getFile)
	require.NoError(b, err)

	for n := 0; n < b.N; n++ {
		require.NoError(b, r.rotate())

		f, ok := r.wc.(*os.File)
		require.True(b, ok)
		require.NoError(b, os.Remove(f.Name()))
	}
}

func getTestFile(b *testing.B, dir string, length int) func(int) (io.WriteCloser, error) {
	return func(index int) (io.WriteCloser, error) {
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
