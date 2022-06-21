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
	"bytes"
	"io"
	"io/ioutil"
	"os"
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

	getFile := func() (io.WriteCloser, error) {
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

func BenchmarkRotateRAMFile(b *testing.B) {
	dir, err := ioutil.TempDir("", "rotate-benchmark")
	require.NoError(b, err)
	defer os.RemoveAll(dir) //nolint:errcheck

	benchRotate(b, MaxLogSize, getTestFile(b, dir, MaxLogSize-1))
}

func BenchmarkRotateDiskFile(b *testing.B) {
	cache, err := os.UserCacheDir()
	require.NoError(b, err)

	dir, err := ioutil.TempDir(cache, "rotate-benchmark")
	require.NoError(b, err)
	defer os.RemoveAll(dir) //nolint:errcheck

	benchRotate(b, MaxLogSize, getTestFile(b, dir, MaxLogSize-1))
}

func benchRotate(b *testing.B, logSize int, getFile func() (io.WriteCloser, error)) {
	r, err := NewRotator(logSize, getFile)
	require.NoError(b, err)

	for n := 0; n < b.N; n++ {
		require.NoError(b, r.rotate())

		f, ok := r.wc.(*os.File)
		require.True(b, ok)
		require.NoError(b, os.Remove(f.Name()))
	}
}

func getTestFile(b *testing.B, dir string, length int) func() (io.WriteCloser, error) {
	return func() (io.WriteCloser, error) {
		b.StopTimer()
		defer b.StartTimer()

		f, err := ioutil.TempFile(dir, "log")
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
