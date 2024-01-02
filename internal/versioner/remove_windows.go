// Copyright (c) 2024 Proton AG
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

//go:build windows
// +build windows

package versioner

import (
	"os"
	"path/filepath"
	"strings"
)

func (v *Versioner) RemoveCurrentVersion() error {
	// get current executable
	exec, err := os.Executable()
	if err != nil {
		return err
	}

	// Check that current executtable is update package so we won't
	// delete base version (that is controlled by package manager).
	// Get absolute paths to ensure there is no crazy stuff there.
	absExec, err := filepath.Abs(exec)
	if err != nil {
		return err
	}
	absRoot, err := filepath.Abs(v.root)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(absExec, absRoot) {
		return ErrNoRemoveBase
	}

	// It is impossible delete running executable on Windows, so instead
	// we rename it. Next time launcher will start it will remove this version
	// as checksum won't match.
	return os.Rename(absExec, filepath.Join(filepath.Dir(absExec), "_"+filepath.Base(absExec)))
}
