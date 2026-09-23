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

package atomicfile

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteFile writes data to a temporary file with 0600 permissions, flushes it to disk and atomically replaces path with it.
func WriteFile(path string, data []byte) (err error) {
	f, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+"-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	defer func() {
		if err != nil {
			_ = f.Close()
			_ = os.Remove(f.Name())
		}
	}()

	if _, err = f.Write(data); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err = f.Sync(); err != nil {
		return fmt.Errorf("failed to flush temp file to disk: %w", err)
	}

	if err = f.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	return replaceFile(f.Name(), path)
}
