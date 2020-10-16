// Copyright (c) 2020 Proton Technologies AG
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

package transfer

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"net/mail"
	"net/textproto"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

// getFolderNames collects all folder names under `root`.
// Folder names will be without a path.
func getFolderNames(root string) ([]string, error) {
	return getFolderNamesWithFileSuffix(root, "")
}

// getFolderNamesWithFileSuffix collects all folder names under `root`, which
// contains some file with a give `fileSuffix`. Names will be without a path.
func getFolderNamesWithFileSuffix(root, fileSuffix string) ([]string, error) {
	folders := []string{}

	files, err := ioutil.ReadDir(root)
	if err != nil {
		return nil, err
	}

	hasFileWithSuffix := fileSuffix == ""
	for _, file := range files {
		if file.IsDir() {
			subfolders, err := getFolderNamesWithFileSuffix(filepath.Join(root, file.Name()), fileSuffix)
			if err != nil {
				return nil, err
			}
			for _, subfolder := range subfolders {
				match := false
				for _, folder := range folders {
					if folder == subfolder {
						match = true
						break
					}
				}
				if !match {
					folders = append(folders, subfolder)
				}
			}
		} else if fileSuffix == "" || strings.HasSuffix(file.Name(), fileSuffix) {
			hasFileWithSuffix = true
		}
	}

	if hasFileWithSuffix {
		folders = append(folders, filepath.Base(root))
	}

	sort.Strings(folders)
	return folders, nil
}

// getFilePathsWithSuffix collects all file names with `suffix` under `root`.
// File names will be with relative path based to `root`.
func getFilePathsWithSuffix(root, suffix string) ([]string, error) {
	fileNames, err := getFilePathsWithSuffixInner("", root, suffix)
	if err != nil {
		return nil, err
	}
	sort.Strings(fileNames)
	return fileNames, err
}

func getFilePathsWithSuffixInner(prefix, root, suffix string) ([]string, error) {
	fileNames := []string{}

	files, err := ioutil.ReadDir(root)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() {
			if strings.HasSuffix(file.Name(), suffix) {
				fileNames = append(fileNames, filepath.Join(prefix, file.Name()))
			}
		} else {
			subfolderFileNames, err := getFilePathsWithSuffixInner(
				filepath.Join(prefix, file.Name()),
				filepath.Join(root, file.Name()),
				suffix,
			)
			if err != nil {
				return nil, err
			}
			fileNames = append(fileNames, subfolderFileNames...)
		}
	}

	return fileNames, nil
}

// getMessageTime returns time of the message specified in the message header.
func getMessageTime(body []byte) (int64, error) {
	mailHeader, err := getMessageHeader(body)
	if err != nil {
		return 0, err
	}
	if t, err := mailHeader.Date(); err == nil && !t.IsZero() {
		return t.Unix(), nil
	}
	return 0, nil
}

// getMessageHeader returns headers of the message body.
func getMessageHeader(body []byte) (mail.Header, error) {
	tpr := textproto.NewReader(bufio.NewReader(bytes.NewBuffer(body)))
	header, err := tpr.ReadMIMEHeader()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read headers")
	}
	return mail.Header(header), nil
}

// sanitizeFileName replaces problematic special characters with underscore.
func sanitizeFileName(fileName string) string {
	if len(fileName) == 0 {
		return fileName
	}
	if runtime.GOOS != "windows" && (fileName[0] == '-' || fileName[0] == '.') { //nolint[goconst]
		fileName = "_" + fileName[1:]
	}
	return strings.Map(func(r rune) rune {
		switch r {
		case '\\', '/', ':', '*', '?', '"', '<', '>', '|':
			return '_'
		case '[', ']', '(', ')', '{', '}', '^', '#', '%', '&', '!', '@', '+', '=', '\'', '~':
			if runtime.GOOS != "windows" {
				return '_'
			}
		}
		return r
	}, fileName)
}
