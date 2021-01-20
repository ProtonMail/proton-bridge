// Copyright (c) 2021 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package tar

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
)

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

		target := filepath.Join(dir, header.Name)

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
			f, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil { // nolint[gosec]
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
