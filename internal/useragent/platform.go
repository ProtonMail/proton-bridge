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

package useragent

import (
	"runtime"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// IsCatalinaOrNewer checks whether the host is macOS Catalina 10.15.x or higher.
func IsCatalinaOrNewer() bool {
	return isThisDarwinNewerOrEqual(getMinCatalina())
}

// IsBigSurOrNewer checks whether the host is macOS BigSur 10.16.x or higher.
func IsBigSurOrNewer() bool {
	return isThisDarwinNewerOrEqual(getMinBigSur())
}

// IsVenturaOrNewer checks whether the host is macOS BigSur 13.x or higher.
func IsVenturaOrNewer() bool {
	return isThisDarwinNewerOrEqual(getMinVentura())
}

func getMinCatalina() *semver.Version { return semver.MustParse("19.0.0") }
func getMinBigSur() *semver.Version   { return semver.MustParse("20.0.0") }
func getMinVentura() *semver.Version  { return semver.MustParse("22.0.0") }

func isThisDarwinNewerOrEqual(minVersion *semver.Version) bool {
	if runtime.GOOS != "darwin" {
		return false
	}

	rawVersion, err := getDarwinVersion()
	if err != nil {
		return false
	}

	return isVersionEqualOrNewer(minVersion, strings.TrimSpace(rawVersion))
}

// isVersionEqualOrNewer is separated to be able to run test on other than darwin.
func isVersionEqualOrNewer(minVersion *semver.Version, rawVersion string) bool {
	semVersion, err := semver.NewVersion(rawVersion)
	if err != nil {
		return false
	}
	return semVersion.GreaterThan(minVersion) || semVersion.Equal(minVersion)
}
