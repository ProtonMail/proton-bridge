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

// Package constants contains variables that are set via ldflags during build.
package constants

import (
	"fmt"
	"runtime"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const VendorName = "protonmail"

//nolint:gochecknoglobals
var (
	// Full app name (to show to the user).
	FullAppName = ""

	// ConfigName determines the name of the location where bridge stores config files.
	ConfigName = "bridge"

	// UpdateName is the name of the product appearing in the update URL.
	UpdateName = "bridge"

	// KeyChainName is the name of the entry in the OS keychain.
	KeyChainName = "bridge"

	// Version of the build.
	Version = "2.4.1+git"

	// Revision is current hash of the build.
	Revision = ""

	// BuildTime stamp of the build.
	BuildTime = ""

	// BuildVersion is derived from LongVersion and BuildTime.
	BuildVersion = fmt.Sprintf("%v (%v) %v", Version, Revision, BuildTime)

	// DSNSentry client keys to be able to report crashes to Sentry.
	DSNSentry = ""

	// Host is the hostname of the bridge server.
	Host = "127.0.0.1"
)

// AppVersion returns the full rendered version of the app (to be used in request headers).
func AppVersion(version string) string {
	return getAPIOS() + cases.Title(language.Und).String(ConfigName) + "_" + version
}

func getAPIOS() string {
	switch runtime.GOOS {
	case "darwin":
		return "macOS"

	case "linux":
		return "Linux"

	case "windows":
		return "Windows"

	default:
		return "Linux"
	}
}
