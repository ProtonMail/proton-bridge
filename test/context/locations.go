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
	"io/ioutil"
	"os"
)

type fakeLocations struct {
	dir string
}

func newFakeLocations() *fakeLocations {
	dir, err := ioutil.TempDir("", "test-cache")
	if err != nil {
		panic(err)
	}

	return &fakeLocations{
		dir: dir,
	}
}

func (l *fakeLocations) ProvideLogsPath() (string, error) {
	return l.dir, nil
}

func (l *fakeLocations) ProvideSettingsPath() (string, error) {
	return l.dir, nil
}

func (l *fakeLocations) GetDependencyLicensesLink() string {
	return "/path/to/dependencies"
}

func (l *fakeLocations) GetLicenseFilePath() string {
	return "/path/to/license"
}

func (l *fakeLocations) Clear() error {
	return os.RemoveAll(l.dir)
}

func (l *fakeLocations) ClearUpdates() error {
	return nil
}
