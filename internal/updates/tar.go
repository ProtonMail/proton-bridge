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

package updates

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
)

func createTar(tarPath, sourcePath string) error { //nolint[unused]
	if runtime.GOOS != "linux" {
		return errors.New("tar not implemented only for linux")
	}
	// Check whether it exists and is a directory.
	if _, err := os.Lstat(sourcePath); err != nil {
		return err
	}

	absPath, err := filepath.Abs(tarPath)
	if err != nil {
		return err
	}

	cmd := exec.Command("tar", "-zvcf", absPath, filepath.Base(sourcePath)) //nolint[gosec]
	cmd.Dir = filepath.Dir(sourcePath)
	cmd.Stderr = log.WriterLevel(logrus.ErrorLevel)
	cmd.Stdout = log.WriterLevel(logrus.InfoLevel)
	return cmd.Run()
}

func untarToDir(tarPath, targetDir string, status *Progress) error { //nolint[funlen]
	// Check whether it exists and is a directory.
	if ls, err := os.Lstat(targetDir); err == nil {
		if !ls.IsDir() {
			return errors.New("not a dir")
		}
	} else {
		return err
	}

	tgzReader, err := os.Open(tarPath) //nolint[gosec]
	if err != nil {
		return err
	}
	defer tgzReader.Close() //nolint[errcheck]

	size := uint64(0)
	if info, err := tgzReader.Stat(); err == nil {
		size = uint64(info.Size())
	}

	wc := &WriteCounter{
		Status: status,
		Size:   size,
	}

	tarReader, err := gzip.NewReader(io.TeeReader(tgzReader, wc))
	if err != nil {
		return err
	}

	fileReader := tar.NewReader(tarReader)
	for {
		header, err := fileReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if header == nil {
			continue
		}

		targetFile := filepath.Join(targetDir, header.Name)
		info := header.FileInfo()

		// Create symlink.
		if header.Typeflag == tar.TypeSymlink {
			if header.Linkname == "" {
				return errors.New("missing linkname")
			}
			if err := os.Symlink(header.Linkname, targetFile); err != nil {
				return err
			}
			continue
		}

		// Handle case that it is a directory.
		if info.IsDir() {
			if err := os.MkdirAll(targetFile, info.Mode()); err != nil {
				return err
			}
			continue
		}

		// Handle case that it is a regular file.
		if err := copyToFileTruncate(fileReader, targetFile, info.Mode()); err != nil {
			return err
		}
	}
	return nil
}
