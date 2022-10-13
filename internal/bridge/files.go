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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package bridge

import (
	"os"
	"path/filepath"
)

func moveDir(from, to string) error {
	entries, err := os.ReadDir(from)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if err := os.Mkdir(filepath.Join(to, entry.Name()), 0700); err != nil {
				return err
			}

			if err := moveDir(filepath.Join(from, entry.Name()), filepath.Join(to, entry.Name())); err != nil {
				return err
			}

			if err := os.RemoveAll(filepath.Join(from, entry.Name())); err != nil {
				return err
			}
		} else {
			if err := move(filepath.Join(from, entry.Name()), filepath.Join(to, entry.Name())); err != nil {
				return err
			}
		}
	}

	return os.Remove(from)
}

func move(from, to string) error {
	if err := os.MkdirAll(filepath.Dir(to), 0700); err != nil {
		return err
	}

	f, err := os.Open(from) // nolint:gosec
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	c, err := os.Create(to) // nolint:gosec
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	if err := os.Chmod(to, 0600); err != nil {
		return err
	}

	if _, err := c.ReadFrom(f); err != nil {
		return err
	}

	if err := os.Remove(from); err != nil {
		return err
	}

	return nil
}
