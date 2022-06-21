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

package updater

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

// VersionInfo is information about one version of the app.
type VersionInfo struct {
	// Version is the semantic version of the release.
	Version *semver.Version

	// MinAuto is the earliest version that is able to autoupdate to this version.
	// Apps older than this version must run the manual installer and cannot autoupdate.
	MinAuto *semver.Version

	// Package is the location of the update package.
	Package string

	// Installers are the locations of installer files (for manual installation).
	Installers []string

	// LandingPage is the address of the app landing page on protonmail.com.
	LandingPage string

	// ReleaseNotesPage is the address of the page containing the release notes.
	ReleaseNotesPage string

	// RolloutProportion indicates the proportion (0,1] of users that should update to this version.
	RolloutProportion float64
}

// VersionMap represents the structure of the version.json file.
// It looks like this:
// {
//   "stable": {
//     "Version": "2.3.4",
//     "Package": "https://protonmail.com/.../bridge_2.3.4_linux.tgz",
//     "Installers": [
//       "https://protonmail.com/.../something.deb",
//       "https://protonmail.com/.../something.rpm",
//       "https://protonmail.com/.../PKGBUILD"
//     ],
//     "LandingPage": "https://protonmail.com/bridge",
//     "ReleaseNotesPage": "https://protonmail.com/.../release_notes.html",
//     "RolloutProportion": 0.5
//   },
//   "early": {
//     "Version": "2.4.0",
//     "Package": "https://protonmail.com/.../bridge_2.4.0_linux.tgz",
//     "Installers": [
//       "https://protonmail.com/.../something.deb",
//       "https://protonmail.com/.../something.rpm",
//       "https://protonmail.com/.../PKGBUILD"
//     ],
//     "LandingPage": "https://protonmail.com/bridge",
//     "ReleaseNotesPage": "https://protonmail.com/.../release_notes.html",
//     "RolloutProportion": 0.5
//   },
//   "...": {
//     ...
//   }
// }.
type VersionMap map[string]VersionInfo

// getVersionFileURL returns the URL of the version file.
// For example:
//  - https://protonmail.com/download/bridge/version_linux.json
//  - https://protonmail.com/download/ie/version_linux.json
func (u *Updater) getVersionFileURL() string {
	return fmt.Sprintf("%v/%v/version_%v.json", Host, u.updateURLName, u.platform)
}
