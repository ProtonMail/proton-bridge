// Copyright (c) 2023 Proton AG
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
)

const VendorName = "protonmail"

//nolint:gochecknoglobals
var (
	// FullAppName is the full app name (to show to the user).
	FullAppName = ""

	// Version of the build.
	Version = "0.0.0"

	// Revision is build time commit hash.
	Revision = ""

	// Tag is build time git describe.
	Tag = ""

	// BuildTime stamp of the build.
	BuildTime = ""

	// BuildVersion is derived from LongVersion and BuildTime.
	BuildVersion = fmt.Sprintf("%v (%v) %v", Version, Revision, BuildTime)

	// DSNSentry client keys to be able to report crashes to Sentry.
	DSNSentry = ""

	// BuildEnv tags used at build time.
	BuildEnv = ""
)

const (
	// AppName is the name of the product appearing in the request headers.
	AppName = "bridge"

	// UpdateName is the name of the product appearing in the update URL.
	UpdateName = "bridge"

	// ConfigName determines the name of the location where bridge stores config/cache files.
	ConfigName = "bridge-v3"

	// KeyChainName is the name of the entry in the OS keychain.
	KeyChainName = "bridge-v3"

	// Host is the hostname of the bridge server.
	Host = "127.0.0.1"
)

// nolint:goconst
func getAPIOS() string {
	switch runtime.GOOS {
	case "darwin":
		return "macos"

	case "linux":
		return "linux"

	case "windows":
		return "windows"

	default:
		return "linux"
	}
}
