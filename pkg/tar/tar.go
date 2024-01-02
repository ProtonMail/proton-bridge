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

package tar

import (
	"archive/tar"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
)

// maxFileSize limit the single file size after decompression is not larger than 1GB.
const maxFileSize = int64(1 * 1024 * 1024 * 1024) // 1 GB

// ErrFileTooLarge returned when decompressed file is too large.
var ErrFileTooLarge = errors.New("trying to decompress file larger than 1GB")

type limitReader struct {
	r io.Reader
	n int64
}

// Read returns error if limit was exceeded. Inspired by io.LimitReader.Read
// implementation.
func (lr *limitReader) Read(p []byte) (n int, err error) {
	if lr.n <= 0 {
		return 0, ErrFileTooLarge
	}
	if int64(len(p)) > lr.n {
		p = p[0:lr.n]
	}
	n, err = lr.r.Read(p)
	lr.n -= int64(n)
	return
}

// UntarToDir decopmress and unarchive the files into directory.
func UntarToDir(r io.Reader, dir string) error {
	tr := tar.NewReader(r)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if header == nil {
			continue
		}

		target := filepath.Join(dir, filepath.Clean(header.Name)) // gosec G305

		switch {
		case header.Typeflag == tar.TypeSymlink:
			if err := os.Symlink(header.Linkname, target); err != nil {
				return err
			}

		case header.FileInfo().IsDir():
			if err := os.MkdirAll(target, header.FileInfo().Mode()); err != nil {
				return err
			}

		default:
			f, err := os.Create(filepath.Clean(target))
			if err != nil {
				return err
			}
			lr := &limitReader{r: tr, n: maxFileSize} // gosec G110
			if _, err := io.Copy(f, lr); err != nil {
				return err
			}
			if runtime.GOOS != "windows" {
				if err := f.Chmod(header.FileInfo().Mode()); err != nil {
					return err
				}
			}
			if err := f.Close(); err != nil {
				logrus.WithError(err).Error("Failed to close file")
			}
		}
	}
}
