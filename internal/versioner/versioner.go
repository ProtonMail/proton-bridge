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

package versioner

import (
	"errors"
	"os"
	"path/filepath"
	"sort"

	"github.com/Masterminds/semver/v3"
)

var (
	ErrNoVersions   = errors.New("no available versions")
	ErrNoExecutable = errors.New("no executable found")
	ErrNoRemoveBase = errors.New("can't remove base version")
)

// Versioner manages a directory of versioned app directories.
type Versioner struct {
	root string
}

func New(root string) *Versioner {
	return &Versioner{root: root}
}

// ListVersions returns a collection of all available version numbers, sorted from newest to oldest.
func (v *Versioner) ListVersions() (Versions, error) {
	dirs, err := os.ReadDir(v.root)
	if err != nil {
		return nil, err
	}

	var versions Versions

	for _, dir := range dirs {
		version, err := semver.StrictNewVersion(dir.Name())
		if err != nil {
			continue
		}

		// NOTE: If it's a bad directory, maybe delete it?

		versions = append(versions, &Version{
			version: version,
			path:    filepath.Join(v.root, dir.Name()),
		})
	}

	sort.Sort(sort.Reverse(versions))

	return versions, nil
}

// GetExecutableInDirectory returns the full path to the executable in the given directory, if present.
// It returns an error if the executable is missing or does not have executable permissions set.
func (v *Versioner) GetExecutableInDirectory(name, directory string) (string, error) {
	return getExecutableInDirectory(name, directory)
}

func getExecutableInDirectory(name, directory string) (string, error) {
	exe := filepath.Join(directory, getExeName(name))

	if !fileExists(exe) || !fileIsExecutable(exe) {
		return "", ErrNoExecutable
	}

	return exe, nil
}
