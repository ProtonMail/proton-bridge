// Copyright (c) 2020 Proton Technologies AG
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
	"fmt"
	"os/exec"
	"runtime"
)

// IsCatalinaOrNewer checks that host is MacOS Catalina 10.14.xx or higher.
func IsCatalinaOrNewer() bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	major, minor, _ := getMacVersion()
	return isVersionCatalinaOrNewer(major, minor)
}

func getMacVersion() (major, minor, tiny int) {
	major, minor, tiny = 10, 0, 0
	out, err := exec.Command("sw_vers", "-productVersion").Output()
	if err != nil {
		return
	}
	return parseMacVersion(string(out))
}

func parseMacVersion(version string) (major, minor, tiny int) {
	_, _ = fmt.Sscanf(version, "%d.%d.%d", &major, &minor, &tiny)
	return
}

func isVersionCatalinaOrNewer(major, minor int) bool {
	if major != 10 {
		return false
	}
	if minor < 15 {
		return false
	}
	return true
}
