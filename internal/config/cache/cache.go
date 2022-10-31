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

// Package cache provides access to contents inside a cache directory.
package cache

import (
	"os"
	"path/filepath"

	"github.com/ProtonMail/proton-bridge/v2/pkg/files"
)

type Cache struct {
	dir, version string
}

func New(dir, version string) (*Cache, error) {
	if err := os.MkdirAll(filepath.Join(dir, version), 0o700); err != nil {
		return nil, err
	}

	return &Cache{
		dir:     dir,
		version: version,
	}, nil
}

// GetDBDir returns folder for db files.
func (c *Cache) GetDBDir() string {
	return c.getCurrentCacheDir()
}

// GetDefaultMessageCacheDir returns folder for cached messages files.
func (c *Cache) GetDefaultMessageCacheDir() string {
	return filepath.Join(c.getCurrentCacheDir(), "messages")
}

// GetIMAPCachePath returns path to file with IMAP status.
func (c *Cache) GetIMAPCachePath() string {
	return filepath.Join(c.getCurrentCacheDir(), "user_info.json")
}

// GetTransferDir returns folder for import-export rules files.
func (c *Cache) GetTransferDir() string {
	return c.getCurrentCacheDir()
}

// RemoveOldVersions removes any cache dirs that are not the current version.
func (c *Cache) RemoveOldVersions() error {
	return files.Remove(c.dir).Except(c.getCurrentCacheDir()).Do()
}

func (c *Cache) getCurrentCacheDir() string {
	return filepath.Join(c.dir, c.version)
}
