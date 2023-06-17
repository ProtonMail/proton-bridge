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
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
)

var (
	errNoInputFile      = errors.New("no file was provided to put in the archive")
	errCannotFitAnyFile = errors.New("no file can fit in the archive")
)

// zipFilesWithMaxSize compress the maximum number of files from the given list that can fit a ZIP archive file whose size does not exceed
// maxSize. Input files are taken in order and the function returns as soon as the next file cannot fit, even if another file further in the list
// may fit. The function return the number of files that were included in the archive. The files included are filePath[:fileCount].
func zipFilesWithMaxSize(filePaths []string, maxSize int64) (buffer *bytes.Buffer, fileCount int, err error) {
	if len(filePaths) == 0 {
		return nil, 0, errNoInputFile
	}
	buffer, err = createZipFromFile(filePaths[0])
	if err != nil {
		return nil, 0, err
	}

	if int64(buffer.Len()) > maxSize {
		return nil, 0, errCannotFitAnyFile
	}

	fileCount = 1
	var previousBuffer *bytes.Buffer

	for _, filePath := range filePaths[1:] {
		previousBuffer = cloneBuffer(buffer)

		zipReader, err := zip.NewReader(bytes.NewReader(buffer.Bytes()), int64(len(buffer.Bytes())))
		if err != nil {
			return nil, 0, err
		}

		buffer, err = addFileToArchive(zipReader, filePath)
		if err != nil {
			return nil, 0, err
		}

		if int64(buffer.Len()) > maxSize {
			return previousBuffer, fileCount, nil
		}

		fileCount++
	}

	return buffer, fileCount, nil
}

// cloneBuffer clones a buffer.
func cloneBuffer(buffer *bytes.Buffer) *bytes.Buffer {
	return bytes.NewBuffer(bytes.Clone(buffer.Bytes()))
}

// createZip creates a zip archive containing a single file.
func createZipFromFile(filePath string) (*bytes.Buffer, error) {
	file, err := os.Open(filePath) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	return createZip(file, filepath.Base(filePath))
}

// createZip creates a zip file containing a file names filename with content read from reader.
func createZip(reader io.Reader, filename string) (*bytes.Buffer, error) {
	b := bytes.NewBuffer(make([]byte, 0))
	zipWriter := zip.NewWriter(b)

	f, err := zipWriter.Create(filename)
	if err != nil {
		return nil, err
	}

	if _, err = io.Copy(f, reader); err != nil {
		return nil, err
	}

	if err = zipWriter.Close(); err != nil {
		return nil, err
	}

	return b, nil
}

// addToArchive adds a file to an archive. Because go zip package does not support adding a file to existing (closed) archive file, the way to do it
// is to create a new archive copy the raw content of the archive to the new one and add the new file before closing the archive.
func addFileToArchive(zipReader *zip.Reader, filePath string) (*bytes.Buffer, error) {
	file, err := os.Open(filePath) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	return addToArchive(zipReader, file, filepath.Base(filePath))
}

// addToArchive adds data from a reader to a file in an archive.
func addToArchive(zipReader *zip.Reader, reader io.Reader, filename string) (*bytes.Buffer, error) {
	buffer := bytes.NewBuffer([]byte{})
	zipWriter := zip.NewWriter(buffer)

	if err := copyZipContent(zipReader, zipWriter); err != nil {
		return nil, err
	}

	f, err := zipWriter.Create(filename)
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(f, reader); err != nil {
		return nil, err
	}

	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	return buffer, nil
}

// copyZipContent copies the content of a zip to another without recompression.
func copyZipContent(zipReader *zip.Reader, zipWriter *zip.Writer) error {
	for _, zipItem := range zipReader.File {
		itemReader, err := zipItem.OpenRaw()
		if err != nil {
			return err
		}

		header := zipItem.FileHeader
		targetItem, err := zipWriter.CreateRaw(&header)
		if err != nil {
			return err
		}

		if _, err := io.Copy(targetItem, itemReader); err != nil {
			return err
		}
	}

	return nil
}
