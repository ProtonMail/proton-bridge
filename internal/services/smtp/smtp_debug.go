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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

//go:build build_qa

package smtp

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func debugDumpToDisk(b []byte) error {
	if os.Getenv("BRIDGE_SMTP_DEBUG") == "" {
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home dir: %w", err)
	}

	if err := os.WriteFile(filepath.Join(home, getFileName()), b, 0600); err != nil {
		return fmt.Errorf("failed to write message file: %w", err)
	}

	return nil
}

func getFileName() string {
	return fmt.Sprintf("smtp_debug_%v.eml", time.Now().Unix())
}
