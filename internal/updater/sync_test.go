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

package updater

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const (
	FileType    = "File"
	SymlinkType = "Symlink"
	DirType     = "Dir"
	EmptyType   = "Empty"
	NewType     = "New"
)

func TestSyncFolder(t *testing.T) {
	for _, srcType := range []string{EmptyType, FileType, SymlinkType, DirType} {
		for _, dstType := range []string{EmptyType, FileType, SymlinkType, DirType} {
			require.NoError(t, checkCopyWorks(srcType, dstType))
			logrus.Warn("OK: from ", srcType, " to ", dstType)
		}
	}
}

func checkCopyWorks(srcType, dstType string) error {
	dirName := "from_" + srcType + "_to_" + dstType
	AppCacheDir := "/tmp"
	srcDir := filepath.Join(AppCacheDir, "sync_src", dirName)
	destDir := filepath.Join(AppCacheDir, "sync_dst", dirName)

	// clear before
	logrus.Info("remove all ", srcDir)
	err := os.RemoveAll(srcDir)
	if err != nil {
		return err
	}

	logrus.Info("remove all ", destDir)
	err = os.RemoveAll(destDir)
	if err != nil {
		return err
	}

	// create
	err = createTestFolder(srcDir, srcType)
	if err != nil {
		return err
	}

	err = createTestFolder(destDir, dstType)
	if err != nil {
		return err
	}

	// copy
	logrus.Info("Sync from ", srcDir, " to ", destDir)
	err = syncFolders(destDir, srcDir)
	if err != nil {
		return err
	}

	// Check
	logrus.Info("check ", srcDir, " and ", destDir)
	err = checkThatFilesAreSame(srcDir, destDir)
	if err != nil {
		return err
	}

	// clear after
	logrus.Info("remove all ", srcDir)
	err = os.RemoveAll(srcDir)
	if err != nil {
		return err
	}

	logrus.Info("remove all ", destDir)
	err = os.RemoveAll(destDir)
	if err != nil {
		return err
	}

	return err
}

func checkThatFilesAreSame(src, dst string) error {
	cmd := exec.Command("diff", "-qr", src, dst) //nolint:gosec
	cmd.Stderr = logrus.StandardLogger().WriterLevel(logrus.ErrorLevel)
	cmd.Stdout = logrus.StandardLogger().WriterLevel(logrus.InfoLevel)
	return cmd.Run()
}

func createTestFolder(dirPath, dirType string) error {
	logrus.Info("creating folder ", dirPath, " type ", dirType)
	if dirType == NewType {
		return nil
	}

	err := mkdirAllClear(dirPath)
	if err != nil {
		return err
	}

	if dirType == EmptyType {
		return nil
	}

	path := filepath.Join(dirPath, "testpath")
	switch dirType {
	case FileType:
		err = ioutil.WriteFile(path, []byte("This is a test"), 0640)
		if err != nil {
			return err
		}

	case SymlinkType:
		err = os.Symlink("../../", path)
		if err != nil {
			return err
		}

	case DirType:
		err = os.MkdirAll(path, 0750)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(filepath.Join(path, "another_file"), []byte("This is a test"), 0640)
		if err != nil {
			return err
		}
	}

	return nil
}
