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
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"testing"

	r "github.com/stretchr/testify/require"
)

const (
	str1 = "Lorem ipsum dolor sit amet"
	str2 = "consectetur adipisicing elit"
)

// tempFileWithContent() creates a temporary file in folderPath containing the string content.
// Returns the path of the created file.
func tempFileWithContent(folderPath, content string) (string, error) {
	file, err := os.CreateTemp(folderPath, "")
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()
	_, err = file.WriteString(content)
	return file.Name(), err
}

// itemCountInFolder() counts the number of items (files, folders, etc) in a folder.
// Returns -1 if an error occurred.
func itemCountInFolder(path string) int {
	files, err := os.ReadDir(path)
	if err != nil {
		return -1
	}
	return len(files)
}

// hashForFile returns the sha1 hash for the given file.
func hashForFile(path string) (string, error) {
	hash := sha1.New()
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()
	if _, err = io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// filesAreIdentical() returns true if the two given files exist and have the same content.
func filesAreIdentical(path1, path2 string) bool {
	hash1, err := hashForFile(path1)
	if err != nil {
		return false
	}
	hash2, err := hashForFile(path2)
	return (err == nil) && hash1 == hash2
}

func TestCache_IsFolderEmpty(t *testing.T) {
	_, err := isFolderEmpty("")
	r.Error(t, err)
	tempDirPath, err := os.MkdirTemp("", "")
	defer func() { r.NoError(t, os.Remove(tempDirPath)) }()
	r.NoError(t, err)
	result, err := isFolderEmpty(tempDirPath)
	r.NoError(t, err)
	r.True(t, result)
	tempFile, err := os.CreateTemp(tempDirPath, "")
	r.NoError(t, err)
	defer func() { r.NoError(t, os.Remove(tempFile.Name())) }()
	r.NoError(t, tempFile.Close())
	_, err = isFolderEmpty(tempFile.Name())
	r.Error(t, err)
	result, err = isFolderEmpty(tempDirPath)
	r.NoError(t, err)
	r.False(t, result)
}

func TestCache_CheckFolderIsSuitableDestinationForCache(t *testing.T) {
	tempDirPath, err := os.MkdirTemp("", "")
	defer func() { _ = os.Remove(tempDirPath) }() // cleanup in case we fail before removing it.
	r.NoError(t, err)
	tempFile, err := os.CreateTemp(tempDirPath, "")
	r.NoError(t, err)
	defer func() { _ = os.Remove(tempFile.Name()) }() // cleanup in case we fail before removing it.
	r.NoError(t, tempFile.Close())
	r.Error(t, checkFolderIsSuitableDestinationForCache(tempDirPath))
	r.NoError(t, os.Remove(tempFile.Name()))
	r.NoError(t, checkFolderIsSuitableDestinationForCache(tempDirPath))
	r.NoDirExists(t, tempDirPath) // previous call to checkFolderIsSuitableDestinationForCache should have removed the folder
	r.NoError(t, checkFolderIsSuitableDestinationForCache(tempDirPath))
}

func TestCache_CopyFolder(t *testing.T) {
	// create a simple tree structure
	// srcDir/
	// |-file1
	// |-srcSubDir/
	//    |-file2

	srcDir, err := os.MkdirTemp("", "")
	defer func() { r.NoError(t, os.RemoveAll(srcDir)) }()
	r.NoError(t, err)
	srcSubDir, err := os.MkdirTemp(srcDir, "")
	r.NoError(t, err)
	subDirName := filepath.Base(srcSubDir)
	file1, err := tempFileWithContent(srcDir, str1)
	r.NoError(t, err)
	file2, err := tempFileWithContent(srcSubDir, str2)
	r.NoError(t, err)

	// copy it
	dstDir := srcDir + "_"
	r.NoDirExists(t, dstDir)
	r.NoFileExists(t, dstDir)
	r.Error(t, copyFolder(srcDir, srcDir))
	r.NoError(t, copyFolder(srcDir, dstDir))
	defer func() { r.NoError(t, os.RemoveAll(dstDir)) }()

	// check copy and original
	r.DirExists(t, srcDir)
	r.DirExists(t, srcSubDir)
	r.FileExists(t, file1)
	r.FileExists(t, file2)
	r.True(t, itemCountInFolder(srcDir) == 2)
	r.True(t, itemCountInFolder(srcSubDir) == 1)
	r.DirExists(t, dstDir)
	dstSubDir := filepath.Join(dstDir, subDirName)
	r.DirExists(t, dstSubDir)
	dstFile1 := filepath.Join(dstDir, filepath.Base(file1))
	r.FileExists(t, dstFile1)
	dstFile2 := filepath.Join(dstDir, subDirName, filepath.Base(file2))
	r.FileExists(t, dstFile2)
	r.True(t, itemCountInFolder(dstDir) == 2)
	r.True(t, itemCountInFolder(dstSubDir) == 1)
	r.True(t, filesAreIdentical(file1, dstFile1))
	r.True(t, filesAreIdentical(file2, dstFile2))
}

func TestCache_IsSubfolderOf(t *testing.T) {
	dir, err := os.MkdirTemp("", "")
	defer func() { r.NoError(t, os.Remove(dir)) }()
	r.NoError(t, err)
	r.True(t, isSubfolderOf(dir, dir))
	fakeDir := dir + "_"
	r.False(t, isSubfolderOf(dir, fakeDir+"_"))
	subDir := filepath.Join(dir, "A", "B")
	r.True(t, isSubfolderOf(subDir, dir))
	r.True(t, isSubfolderOf(filepath.Dir(subDir), dir))
	r.False(t, isSubfolderOf(dir, subDir))
}

func TestCache_CopyFile(t *testing.T) {
	file1, err := tempFileWithContent("", str1)
	r.NoError(t, err)
	defer func() { r.NoError(t, os.Remove(file1)) }()
	file2, err := tempFileWithContent("", str2)
	r.NoError(t, err)
	defer func() { r.NoError(t, os.Remove(file2)) }()
	r.Error(t, copyFile(file1, file1))
	r.Error(t, copyFile(file1, filepath.Dir(file1)))
	r.Error(t, copyFile(file1, file1))
	r.NoError(t, copyFile(file1, file2))
	file3 := file2 + "_"
	r.NoFileExists(t, file3)
	r.NoError(t, copyFile(file1, file3))
	defer func() { r.NoError(t, os.Remove(file3)) }()
}
