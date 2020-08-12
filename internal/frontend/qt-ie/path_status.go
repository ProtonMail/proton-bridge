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

// +build !nogui

package qtie

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// PathStatus maps folder properties to flag
type PathStatus int

// Definition of PathStatus flags
const (
	PathOK PathStatus = 1 << iota
	PathEmptyPath
	PathWrongPath
	PathNotADir
	PathWrongPermissions
	PathDirEmpty
)

// CheckPathStatus return PathStatus flag as int
func CheckPathStatus(path string) int {
	stat := PathStatus(0)
	// path is not empty
	if path == "" {
		stat |= PathEmptyPath
		return int(stat)
	}
	// is dir
	fi, err := os.Lstat(path)
	if err != nil {
		stat |= PathWrongPath
		return int(stat)
	}
	if fi.IsDir() {
		// can open
		files, err := ioutil.ReadDir(path)
		if err != nil {
			stat |= PathWrongPermissions
			return int(stat)
		}
		// empty folder
		if len(files) == 0 {
			stat |= PathDirEmpty
		}
		// can write
		tmpFile := filepath.Join(path, "tmp")
		for err == nil {
			tmpFile += "tmp"
			_, err = os.Lstat(tmpFile)
		}
		err = os.Mkdir(tmpFile, 0750)
		if err != nil {
			stat |= PathWrongPermissions
			return int(stat)
		}
		_ = os.Remove(tmpFile)
	} else {
		stat |= PathNotADir
	}
	stat |= PathOK
	return int(stat)
}
