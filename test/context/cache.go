// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.Bridge.
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

package context

import (
	"os"
	"path/filepath"
)

type fakeCache struct {
	dir string
}

// newFakeCache creates a temporary folder for files.
// It's expected the test calls `ClearData` before finish to remove it from the file system.
func newFakeCache() *fakeCache {
	dir, err := os.MkdirTemp("", "test-cache")
	if err != nil {
		panic(err)
	}

	return &fakeCache{
		dir: dir,
	}
}

// GetDBDir returns folder for db files.
func (c *fakeCache) GetDBDir() string {
	return c.dir
}

// GetIMAPCachePath returns path to file with IMAP status.
func (c *fakeCache) GetIMAPCachePath() string {
	return filepath.Join(c.dir, "user_info.json")
}

// GetTransferDir returns folder for import-export rules files.
func (c *fakeCache) GetTransferDir() string {
	return c.dir
}

func (c *fakeCache) GetDefaultMessageCacheDir() string {
	return c.dir
}
