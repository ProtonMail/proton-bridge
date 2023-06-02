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

func syncFolders(localPath, updatePath string) error {
	backupDir := filepath.Join(filepath.Dir(updatePath), "backup")
	if err := createBackup(localPath, backupDir); err != nil {
		logrus.WithField("dir", backupDir).WithError(err).Error("Cannot create backup")
		return err
	}

	if err := removeMissing(localPath, updatePath); err != nil {
		logrus.WithError(err).Error("Sync folders: failed to remove missing.")
		restoreFromBackup(backupDir, localPath)
		return err
	}

	if err := copyRecursively(updatePath, localPath); err != nil {
		logrus.WithError(err).Error("Sync folders: failed to copy.")
		restoreFromBackup(backupDir, localPath)
		return err
	}

	return nil
}

func removeMissing(folderToCleanPath, itemsToKeepPath string) error {
	logrus.WithField("dir", folderToCleanPath).Debug("Remove missing")

	// Create list of files.
	existingRelPaths := map[string]bool{}
	if err := filepath.Walk(itemsToKeepPath, func(keepThis string, _ os.FileInfo, walkErr error) error {
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
	}); err != nil {
		return err
	}

	delList := []string{}
	if err := filepath.Walk(folderToCleanPath, func(fullPath string, _ os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relPath, walkErr := filepath.Rel(folderToCleanPath, fullPath)
		if walkErr != nil {
			logrus.WithField("full", fullPath).WithError(walkErr).Error("Failed to get relative path")
			return walkErr
		}

		l := logrus.WithField("path", relPath)
		l.Debug("Check")

		if !existingRelPaths[relPath] {
			l.WithField("remove", fullPath).Debug("Path not in list, removing")
			delList = append(delList, fullPath)
		}

		return nil
	}); err != nil {
		return err
	}

	for _, removeThis := range delList {
		if err := os.RemoveAll(removeThis); err != nil && !errors.Is(err, fs.ErrNotExist) {
			logrus.WithField("path", removeThis).WithError(err).Error("Cannot remove")
			return err
		}
	}

	return nil
}

func restoreFromBackup(backupDir, localPath string) {
	l := logrus.WithField("from", backupDir).
		WithField("to", localPath)
	l.Warning("Recovering")

	if err := copyRecursively(backupDir, localPath); err != nil {
		l.WithError(err).Error("Not able to recover")
	}
}

func createBackup(srcFile, dstDir string) error {
	l := logrus.WithField("from", srcFile).WithField("to", dstDir)

	l.Debug("Create backup")
	if err := mkdirAllClear(dstDir); err != nil {
		l.WithError(err).Error("Cannot create backup folder")
		return err
	}

	if err := copyRecursively(srcFile, dstDir); err != nil {
		l.WithError(err).Error("Cannot copy to backup folder")
		return err
	}

	return nil
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
		logrus.WithError(err).WithField("path", path).Error("Cannot open file for checksum")
		return
	}
	defer file.Close() //nolint:errcheck,gosec

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		logrus.WithError(err).WithField("path", path).Error("Cannot read file for checksum")
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

		l := logrus.WithField("source", srcPath)

		// Non regular source (e.g. named pipes, sockets, devices...).
		if !srcIsLink && !srcIsDir && !srcInfo.Mode().IsRegular() {
			err := errors.New("irregular source file: copy not implemented")
			l.WithField("mode", srcInfo.Mode()).WithError(err).Error("Source with iregular mode")
			return err
		}

		// Destination path.
		srcRelPath, err := filepath.Rel(srcDir, srcPath)
		if err != nil {
			l.WithField("dir", srcDir).WithError(err).Error("Failed to get relative source path")
			return err
		}

		dstPath := filepath.Join(dstDir, srcRelPath)
		l = l.WithField("destination", dstPath)

		// Destination exists.
		dstInfo, err := os.Lstat(dstPath)
		l.WithError(err).Debug("Destination check")

		if err == nil {
			dstIsLink := dstInfo.Mode()&os.ModeSymlink == os.ModeSymlink
			dstIsDir := dstInfo.IsDir()

			// Non regular destination (e.g. named pipes, sockets, devices...).
			if !dstIsLink && !dstIsDir && !dstInfo.Mode().IsRegular() {
				err := errors.New("irregular target file: copy not implemented")
				l.WithError(err).WithField("mode", dstInfo.Mode()).Error("Destination with irregular mode")
				return err
			}

			if dstIsLink {
				if err = os.Remove(dstPath); err != nil {
					l.WithError(err).Error("Cannot remove destination link")
					return err
				}
			}

			if !dstIsLink && dstIsDir && !srcIsDir {
				if err = os.RemoveAll(dstPath); err != nil {
					l.WithError(err).Error("Cannot remove destination folder")
					return err
				}
			}

			// NOTE: Do not return if !dstIsLink && dstIsDir && srcIsDir: the permissions might change.

			if dstInfo.Mode().IsRegular() && !srcInfo.Mode().IsRegular() {
				if err = os.Remove(dstPath); err != nil {
					l.WithError(err).Error("Cannot remove destination file")
					return err
				}
			}
		} else if !errors.Is(err, fs.ErrNotExist) {
			l.WithError(err).Error("Destination error")
			return err
		}

		// Create symbolic link and return.
		if srcIsLink {
			l.Debug("Source is a symlink")
			linkPath, err := os.Readlink(srcPath)
			if err != nil {
				l.WithError(err).Error("Failed to read link")
				return err
			}
			l.WithField("linkPath", linkPath).Debug("Creating symlink")
			return os.Symlink(linkPath, dstPath)
		}

		// Create dir and return.
		if srcIsDir {
			l.Debug("Source is a dir")
			err := os.MkdirAll(dstPath, srcInfo.Mode())
			if err != nil {
				l.WithError(err).Error("Failed to create dir")
			}
			return err
		}

		// Regular files only.
		// If files are same return.
		if os.SameFile(srcInfo, dstInfo) || checksum(srcPath) == checksum(dstPath) {
			l.Debug("Same files, skip copy")
			return nil
		}

		// Create/overwrite regular file.
		srcReader, err := os.Open(filepath.Clean(srcPath))
		if err != nil {
			l.WithError(err).Error("Failed to open source")
			return err
		}
		defer srcReader.Close() //nolint:errcheck,gosec

		return copyToTmpFileRename(srcReader, dstPath, srcInfo.Mode())
	})
}

func copyToTmpFileRename(srcReader io.Reader, dstPath string, dstMode os.FileMode) error {
	tmpPath := dstPath + ".tmp"
	l := logrus.WithField("dstPath", dstPath)
	l.Debug("Create tmp and rename")

	if err := copyToFileTruncate(srcReader, tmpPath, dstMode); err != nil {
		l.WithError(err).Error("Failed to copy and truncate")
		return err
	}

	if err := os.Rename(tmpPath, dstPath); err != nil {
		l.WithError(err).Error("Failed to rename")
		return err
	}

	return nil
}

func copyToFileTruncate(srcReader io.Reader, dstPath string, dstMode os.FileMode) error {
	l := logrus.WithField("dstPath", dstPath)
	l.Debug("Copy and truncate")

	dstWriter, err := os.OpenFile(filepath.Clean(dstPath), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, dstMode) //nolint:gosec // Cannot guess the safe part of path
	if err != nil {
		l.WithError(err).Error("Failed to open destination")
		return err
	}
	defer dstWriter.Close() //nolint:errcheck,gosec

	if _, err := io.Copy(dstWriter, srcReader); err != nil {
		l.WithError(err).Error("Failed to open destination")
		return err
	}

	return nil
}
