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
			checkCopyWorks(t, srcType, dstType)
			logrus.Warn("OK: from ", srcType, " to ", dstType)
		}
	}
}

func checkCopyWorks(tb testing.TB, srcType, dstType string) {
	r := require.New(tb)
	dirName := "from_" + srcType + "_to_" + dstType
	AppCacheDir := "/tmp"
	srcDir := filepath.Join(AppCacheDir, "sync_src", dirName)
	destDir := filepath.Join(AppCacheDir, "sync_dst", dirName)

	// clear before
	r.NoError(os.RemoveAll(srcDir))
	r.NoError(os.RemoveAll(destDir))

	// create
	r.NoError(createTestFolder(srcDir, srcType))
	r.NoError(createTestFolder(destDir, dstType))

	// copy
	r.NoError(syncFolders(destDir, srcDir))

	// Check
	checkThatFilesAreSame(r, srcDir, destDir)

	// clear after
	r.NoError(os.RemoveAll(srcDir))
	r.NoError(os.RemoveAll(destDir))
}

func checkThatFilesAreSame(r *require.Assertions, src, dst string) {
	srcFiles, srcDirs, err := walkDir(src)
	r.NoError(err)

	dstFiles, dstDirs, err := walkDir(dst)
	r.NoError(err)

	r.ElementsMatch(srcFiles, dstFiles)
	r.ElementsMatch(srcDirs, dstDirs)

	for _, relPath := range srcFiles {
		srcPath := filepath.Join(src, relPath)
		r.FileExists(srcPath)

		dstPath := filepath.Join(dst, relPath)
		r.FileExists(dstPath)

		srcInfo, err := os.Lstat(srcPath)
		r.NoError(err)

		dstInfo, err := os.Lstat(dstPath)
		r.NoError(err)

		r.Equal(srcInfo.Mode(), dstInfo.Mode())

		if srcInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
			srcLnk, err := os.Readlink(srcPath)
			r.NoError(err)

			dstLnk, err := os.Readlink(dstPath)
			r.NoError(err)

			r.Equal(srcLnk, dstLnk)
		} else {
			srcContent, err := ioutil.ReadFile(srcPath)
			r.NoError(err)

			dstContent, err := ioutil.ReadFile(dstPath)
			r.NoError(err)

			r.Equal(srcContent, dstContent)
		}
	}
}

func walkDir(dir string) (files, dirs []string, err error) {
	err = filepath.Walk(dir, func(path string, info os.FileInfo, errWalk error) error {
		if errWalk != nil {
			return errWalk
		}

		relPath, errRel := filepath.Rel(dir, path)
		if errRel != nil {
			return errRel
		}

		if info.IsDir() {
			dirs = append(dirs, relPath)
		} else {
			files = append(files, relPath)
		}

		return nil
	})
	return
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
		err = ioutil.WriteFile(path, []byte("This is a test"), 0o640)
		if err != nil {
			return err
		}

	case SymlinkType:
		err = os.Symlink("../../", path)
		if err != nil {
			return err
		}

	case DirType:
		err = os.MkdirAll(path, 0o750)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(filepath.Join(path, "another_file"), []byte("This is a test"), 0o640)
		if err != nil {
			return err
		}
	}

	return nil
}
