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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

//go:build darwin
// +build darwin

package versioner

import "github.com/Masterminds/semver/v3"

// RemoveOldVersions removes all but the latest app version.
func (v *Versioner) RemoveOldVersions() error {
	// darwin does not use the versioner; removal is a noop.
	return nil
}

// RemoveOtherVersions removes all but the specific provided app version.
func (v *Versioner) RemoveOtherVersions(_ *semver.Version) error {
	// darwin does not use the versioner; removal is a noop.
	return nil
}

// RemoveCurrentVersion removes current app version unless it is base installed version.
func (v *Versioner) RemoveCurrentVersion() error {
	// darwin does not use the versioner; removal is a noop.
	return nil
}
