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

package users

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// isFolderEmpty checks whether a folder is empty.
// path must point to an existing folder.
func isFolderEmpty(path string) (bool, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return true, err
	}
	return len(files) == 0, nil
}

// checkFolderIsSuitableDestinationForCache determine if a folder is a suitable destination as a cache
// if it is suitable (non existing, or empty and deletable) the folder is deleted.
func checkFolderIsSuitableDestinationForCache(path string) error {
	// Ensure the parent directory exists.
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	// if the folder does not exists, its suitable
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if !fileInfo.IsDir() {
		return errors.New("the destination folder for message cache exists and is a file")
	}

	empty, err := isFolderEmpty(path)
	if err != nil {
		return err
	}

	if !empty {
		return errors.New("the destination folder is not empty")
	}
	return os.Remove(path)
}

// copyFolder recursively copy folder at srcPath to dstPath.
// srcPath must be an existing folder.
// dstPath must point to a non-existing folder.
func copyFolder(srcPath, dstPath string) error {
	fiFrom, err := os.Stat(srcPath)
	if err != nil {
		return err
	}

	_, err = os.Stat(dstPath)
	if !os.IsNotExist(err) {
		return errors.New("the destination folder already exists")
	}

	if !fiFrom.IsDir() {
		return errors.New("source is not an existing folder")
	}

	if err = os.MkdirAll(dstPath, 0o700); err != nil {
		return err
	}
	files, err := ioutil.ReadDir(srcPath)
	if err != nil {
		return err
	}
	// copy only regular files and folders
	for _, fileInfo := range files {
		mode := fileInfo.Mode()
		if mode&os.ModeSymlink != 0 {
			continue // we skip symbolic links to avoid potential endless recursion
		}
		srcSubPath := srcPath + "/" + fileInfo.Name()
		dstSubPath := dstPath + "/" + fileInfo.Name()

		if mode.IsDir() {
			if err = copyFolder(srcSubPath, dstSubPath); err != nil {
				return err
			}
			continue
		}

		if mode.IsRegular() {
			if err = copyFile(srcSubPath, dstSubPath); err != nil {
				return err
			}
			continue // unnecessary but safer if we had code below
		}
	}
	return nil
}

// isSubfolderOf check whether path is subfolder of refPath or is the same.
// RefPath must exist otherwise the function returns false.
func isSubfolderOf(path, refPath string) bool {
	refInfo, err := os.Stat(refPath)
	if (err != nil) || (!refInfo.IsDir()) {
		return false // refpath does not exist. Not acceptable as we use os.SameFile for testing identity
	}

	// we check path and all its parent folder to verify if it is refPath.
	prevPath := ""
	for path != prevPath {
		pathInfo, err := os.Stat(path) // path may not exist, and it's acceptable, so wo keep going event if err != nil
		if err == nil && os.SameFile(pathInfo, refInfo) {
			return true
		}
		prevPath = path
		path = filepath.Dir(path)
	}
	return false
}

// copyFile copies file srcPath to dstPath. both path are files names. srcPath must exist, dstPath will be overwritten
// if it exists and is a file.
func copyFile(srcPath, dstPath string) error {
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return errors.New("could not open source file")
	}
	if !srcInfo.Mode().IsRegular() {
		return errors.New("source file is not a regular file")
	}

	dstInfo, err := os.Stat(dstPath)
	if err == nil {
		if !dstInfo.Mode().IsRegular() {
			return errors.New("destination exists and is not a regular file")
		}
		if os.SameFile(srcInfo, dstInfo) {
			return errors.New("source and destination are the same")
		}
	}

	src, err := os.Open(filepath.Clean(srcPath))
	if err != nil {
		return err
	}
	defer func() {
		err = src.Close()
	}()

	dst, err := os.OpenFile(filepath.Clean(dstPath), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer func() {
		err = dst.Close()
	}()
	_, err = io.Copy(dst, src)
	return err
}

func (u *Users) EnableCache() error {
	// NOTE(GODT-1158): Check for available size before enabling.

	return nil
}

func (u *Users) DisableCache() error {
	for _, user := range u.users {
		if err := user.store.RemoveCache(); err != nil {
			logrus.WithError(err).Error("Failed to remove user's message cache")
		}
	}

	return nil
}

// MigrateCache moves the message cache folder from folder srcPath to folder dstPath.
// srcPath must point to an existing folder. dstPath must be an empty folder or not exist.
func (u *Users) MigrateCache(srcPath, dstPath string) error {
	fiSrc, err := os.Stat(srcPath)
	if os.IsNotExist(err) {
		logrus.WithError(err).Warn("Skipping migration: unknown source for cache migration")
		return nil
	}
	if !fiSrc.IsDir() {
		logrus.WithError(err).Warn("Skipping migration: srcPath is not a dir")
		return nil
	}

	if isSubfolderOf(dstPath, srcPath) {
		return errors.New("destination folder is a subfolder of the source folder")
	}

	if err = checkFolderIsSuitableDestinationForCache(dstPath); err != nil {
		logrus.WithError(err).Error("The destination folder is not suitable for cache migration")
		return err
	}

	for _, user := range u.users {
		if err := user.closeStore(); err != nil {
			logrus.WithError(err).Error("Failed to close user's store")
		}
	}

	// GODT-1381 Edge case: read-only source migration: prevent re-naming
	// (read-only is conserved). Do copy instead.
	tmp, err := ioutil.TempFile(srcPath, "tmp")
	if err == nil {
		defer func() {
			tmp.Close()           //nolint:errcheck,gosec
			os.Remove(tmp.Name()) //nolint:errcheck,gosec
		}()

		if err := os.Rename(srcPath, dstPath); err == nil {
			return nil
		}
	} else {
		logrus.WithError(err).Warn("Cannot write to source: do copy to new destination instead of rename")
	}

	// Rename failed let's try an actual copy/delete
	if err = copyFolder(srcPath, dstPath); err != nil {
		return err
	}

	if err = os.RemoveAll(srcPath); err != nil { // we don't care much about error there.
		logrus.WithError(err).Warn("Original cache folder could not be entirely removed")
	}

	return nil
}
