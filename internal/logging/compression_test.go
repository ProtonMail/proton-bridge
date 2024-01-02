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
	"archive/zip"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/bradenaw/juniper/xslices"
	"github.com/stretchr/testify/require"
)

func TestLogging_LogCompression(t *testing.T) {
	dir := t.TempDir()

	files := []fileInfo{
		{filepath.Join(dir, "1.log"), 100000},
		{filepath.Join(dir, "2.log"), 200000},
		{filepath.Join(dir, "3.log"), 300000},
	}

	// Files will have a content and size (relative to the zip format overhead) that ensure a compression ratio of roughly 2:1.
	createRandomFiles(t, files)
	paths := xslices.Map(files, func(fileInfo fileInfo) string { return fileInfo.filename })

	// Case 1: no input file.
	_, _, err := zipFilesWithMaxSize([]string{}, 10)
	require.ErrorIs(t, err, errNoInputFile)

	// Case 2: limit to low, no file can be included.
	_, _, err = zipFilesWithMaxSize(paths, 100)
	require.ErrorIs(t, err, errCannotFitAnyFile)

	// case 3: 1 file fits.
	buffer, fileCount, err := zipFilesWithMaxSize(paths, 100000)
	require.NoError(t, err)
	require.Equal(t, 1, fileCount)
	checkZipFileContent(t, buffer, paths[0:1])

	// case 4: 2 files fit.
	buffer, fileCount, err = zipFilesWithMaxSize(paths, 200000)
	require.NoError(t, err)
	require.Equal(t, 2, fileCount)
	checkZipFileContent(t, buffer, paths[0:2])

	// case 5: 3 files fit.
	buffer, fileCount, err = zipFilesWithMaxSize(paths, 500000)
	require.NoError(t, err)
	require.Equal(t, 3, fileCount)
	checkZipFileContent(t, buffer, paths)
}

func createRandomFiles(t *testing.T, files []fileInfo) {
	// The file is crafted to have a compression ratio of roughly 2:1 by filling the first half with random data, and the second with zeroes.
	for _, file := range files {
		randomData := make([]byte, file.size)
		_, err := rand.Read(randomData[:file.size/2])
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(file.filename, randomData, 0660))
	}
}

func checkZipFileContent(t *testing.T, buffer *bytes.Buffer, expectedFilePaths []string) {
	dir := t.TempDir()
	count := unzipFile(t, buffer, dir)
	require.Equal(t, len(expectedFilePaths), count)
	for _, file := range expectedFilePaths {
		checkFilesAreIdentical(t, file, filepath.Join(dir, filepath.Base(file)))
	}
}

func unzipFile(t *testing.T, buffer *bytes.Buffer, dir string) int {
	reader, err := zip.NewReader(bytes.NewReader(buffer.Bytes()), int64(len(buffer.Bytes())))
	require.NoError(t, err)

	for _, f := range reader.File {
		info := f.FileInfo()
		require.False(t, info.IsDir())
		require.Equal(t, filepath.Base(info.Name()), info.Name()) // no sub-folder
		extractFileFromZip(t, f, filepath.Join(dir, f.Name))
	}

	return len(reader.File)
}

func extractFileFromZip(t *testing.T, zip *zip.File, path string) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zip.Mode())
	require.NoError(t, err)
	defer func() { _ = file.Close() }()

	reader, err := zip.Open()
	require.NoError(t, err)
	defer func() { _ = reader.Close() }()

	_, err = io.Copy(file, reader)
	require.NoError(t, err)
}

func checkFilesAreIdentical(t *testing.T, path1, path2 string) {
	require.EqualValues(t, sha256Sum(t, path1), sha256Sum(t, path2))
}

func sha256Sum(t *testing.T, path string) []byte {
	f, err := os.Open(path)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	hash := sha256.New()
	_, err = io.Copy(hash, f)
	require.NoError(t, err)

	return hash.Sum(nil)
}
