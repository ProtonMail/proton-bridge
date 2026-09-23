// Copyright (c) 2026 Proton AG
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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

//go:build !windows

package atomicfile

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func replaceFile(from, to string) error {
	if err := os.Rename(from, to); err != nil {
		return fmt.Errorf("failed to replace file: %w", err)
	}

	if err := syncDir(filepath.Dir(to)); err != nil {
		logrus.WithError(err).Warn("File replaced, but failed to sync parent directory")
	}

	return nil
}

func syncDir(path string) error {
	dir, err := os.Open(filepath.Clean(path))
	if err != nil {
		return err
	}

	defer func() {
		_ = dir.Close()
	}()

	return dir.Sync()
}
