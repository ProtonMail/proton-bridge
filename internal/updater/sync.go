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

package updater

import (
	"crypto/sha256"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func syncFolders(localPath, updatePath string) (err error) {
	backupDir := filepath.Join(filepath.Dir(updatePath), "backup")
	if err = createBackup(localPath, backupDir); err != nil {
		return
	}

	if err = removeMissing(localPath, updatePath); err != nil {
		restoreFromBackup(backupDir, localPath)
		return
	}

	if err = copyRecursively(updatePath, localPath); err != nil {
		restoreFromBackup(backupDir, localPath)
		return
	}

	return nil
}

//nolint:nakedret
func removeMissing(folderToCleanPath, itemsToKeepPath string) (err error) {
	logrus.WithField("from", folderToCleanPath).Debug("Remove missing")
	// Create list of files.
	existingRelPaths := map[string]bool{}
	err = filepath.Walk(itemsToKeepPath, func(keepThis string, _ os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relPath, walkErr := filepath.Rel(itemsToKeepPath, keepThis)
		if walkErr != nil {
			return walkErr
		}
		logrus.WithField("path", relPath).Trace("Keep the path")
		existingRelPaths[relPath] = true
		return nil
	})
	if err != nil {
		return
	}

	delList := []string{}
	err = filepath.Walk(folderToCleanPath, func(removeThis string, _ os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relPath, walkErr := filepath.Rel(folderToCleanPath, removeThis)
		if walkErr != nil {
			return walkErr
		}
		logrus.Debug("check path ", relPath)
		if !existingRelPaths[relPath] {
			logrus.Debug("path not in list, removing ", removeThis)
			delList = append(delList, removeThis)
		}
		return nil
	})
	if err != nil {
		return
	}

	for _, removeThis := range delList {
		if err = os.RemoveAll(removeThis); err != nil && !errors.Is(err, fs.ErrNotExist) {
			logrus.Error("remove error ", err)
			return
		}
	}

	return nil
}

func restoreFromBackup(backupDir, localPath string) {
	logrus.WithField("from", backupDir).
		WithField("to", localPath).
		Error("recovering")
	if err := copyRecursively(backupDir, localPath); err != nil {
		logrus.WithField("from", backupDir).
			WithField("to", localPath).
			Error("Not able to recover.")
	}
}

func createBackup(srcFile, dstDir string) (err error) {
	logrus.WithField("from", srcFile).WithField("to", dstDir).Debug("Create backup")
	if err = mkdirAllClear(dstDir); err != nil {
		return
	}

	return copyRecursively(srcFile, dstDir)
}

func mkdirAllClear(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return os.MkdirAll(path, 0o750)
}

// checksum assumes the file is a regular file and that it exists.
func checksum(path string) (hash string) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return
	}
	defer file.Close() //nolint:errcheck,gosec

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return
	}

	return string(hasher.Sum(nil))
}

// srcDir including app folder.
// dstDir including app folder.
func copyRecursively(srcDir, dstDir string) error {
	return filepath.Walk(srcDir, func(srcPath string, srcInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		srcIsLink := srcInfo.Mode()&os.ModeSymlink == os.ModeSymlink
		srcIsDir := srcInfo.IsDir()

		// Non regular source (e.g. named pipes, sockets, devices...).
		if !srcIsLink && !srcIsDir && !srcInfo.Mode().IsRegular() {
			logrus.Error("File ", srcPath, " with mode ", srcInfo.Mode())
			return errors.New("irregular source file. Copy not implemented")
		}

		// Destination path.
		srcRelPath, err := filepath.Rel(srcDir, srcPath)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dstDir, srcRelPath)
		logrus.Debug("src: ", srcPath, " dst: ", dstPath)

		// Destination exists.
		dstInfo, err := os.Lstat(dstPath)
		if err == nil {
			dstIsLink := dstInfo.Mode()&os.ModeSymlink == os.ModeSymlink
			dstIsDir := dstInfo.IsDir()

			// Non regular destination (e.g. named pipes, sockets, devices...).
			if !dstIsLink && !dstIsDir && !dstInfo.Mode().IsRegular() {
				logrus.Error("File ", dstPath, " with mode ", dstInfo.Mode())
				return errors.New("irregular target file. Copy not implemented")
			}

			if dstIsLink {
				if err = os.Remove(dstPath); err != nil {
					return err
				}
			}

			if !dstIsLink && dstIsDir && !srcIsDir {
				if err = os.RemoveAll(dstPath); err != nil {
					return err
				}
			}

			// NOTE: Do not return if !dstIsLink && dstIsDir && srcIsDir: the permissions might change.

			if dstInfo.Mode().IsRegular() && !srcInfo.Mode().IsRegular() {
				if err = os.Remove(dstPath); err != nil {
					return err
				}
			}
		} else if !errors.Is(err, fs.ErrNotExist) {
			return err
		}

		// Create symbolic link and return.
		if srcIsLink {
			logrus.Debug("It is a symlink")
			linkPath, err := os.Readlink(srcPath)
			if err != nil {
				return err
			}
			logrus.Debug("link to ", linkPath)
			return os.Symlink(linkPath, dstPath)
		}

		// Create dir and return.
		if srcIsDir {
			logrus.Debug("It is a dir")
			return os.MkdirAll(dstPath, srcInfo.Mode())
		}

		// Regular files only.
		// If files are same return.
		if os.SameFile(srcInfo, dstInfo) || checksum(srcPath) == checksum(dstPath) {
			logrus.Debug("Same files, skip copy")
			return nil
		}

		// Create/overwrite regular file.
		srcReader, err := os.Open(filepath.Clean(srcPath))
		if err != nil {
			return err
		}
		defer srcReader.Close() //nolint:errcheck,gosec
		return copyToTmpFileRename(srcReader, dstPath, srcInfo.Mode())
	})
}

func copyToTmpFileRename(srcReader io.Reader, dstPath string, dstMode os.FileMode) error {
	logrus.Debug("Tmp and rename ", dstPath)
	tmpPath := dstPath + ".tmp"
	if err := copyToFileTruncate(srcReader, tmpPath, dstMode); err != nil {
		return err
	}
	return os.Rename(tmpPath, dstPath)
}

func copyToFileTruncate(srcReader io.Reader, dstPath string, dstMode os.FileMode) error {
	logrus.Debug("Copy and truncate ", dstPath)
	dstWriter, err := os.OpenFile(filepath.Clean(dstPath), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, dstMode) //nolint:gosec // Cannot guess the safe part of path
	if err != nil {
		return err
	}
	defer dstWriter.Close() //nolint:errcheck,gosec
	_, err = io.Copy(dstWriter, srcReader)
	return err
}
