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

package pmapi

import (
	"runtime"
	"strings"
	"time"
)

// rootURL is the API root URL.
// It must not contain the protocol! The protocol should be in rootScheme.
var rootURL = "api.protonmail.ch" //nolint[gochecknoglobals]

// rootScheme is the scheme to use for connections to the root URL.
var rootScheme = "https" //nolint[gochecknoglobals]

func GetAPIConfig(configName, appVersion string) *ClientConfig {
	return &ClientConfig{
		AppVersion:        getAPIOS() + strings.Title(configName) + "_" + appVersion,
		ClientID:          configName,
		Timeout:           25 * time.Minute, // Overall request timeout (~25MB / 25 mins => ~16kB/s, should be reasonable).
		FirstReadTimeout:  30 * time.Second, // 30s to match 30s response header timeout.
		MinBytesPerSecond: 1 << 10,          // Enforce minimum download speed of 1kB/s.
	}
}

// getAPIOS returns actual operating system.
func getAPIOS() string {
	switch os := runtime.GOOS; os {
	case "darwin": // nolint: goconst
		return "macOS"
	case "linux":
		return "Linux"
	case "windows":
		return "Windows"
	}

	return "Linux"
}
