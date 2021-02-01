// Copyright (c) 2021 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package useragent

import (
	"os/exec"
	"runtime"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// IsCatalinaOrNewer checks that host is MacOS Catalina 10.15.x or higher.
func IsCatalinaOrNewer() bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	return isVersionCatalinaOrNewer(getMacVersion())
}

func getMacVersion() string {
	out, err := exec.Command("sw_vers", "-productVersion").Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(out))
}

func isVersionCatalinaOrNewer(version string) bool {
	v, err := semver.NewVersion(version)
	if err != nil {
		return false
	}

	catalina := semver.MustParse("10.15.0")
	return v.GreaterThan(catalina) || v.Equal(catalina)
}
