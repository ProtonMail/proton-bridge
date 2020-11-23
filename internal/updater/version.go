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

	// Landing is the address of the app landing page on protonmail.com.
	Landing string

	// Rollout is the current progress of the rollout of this release.
	Rollout float64
}

// VersionMap represents the structure of the version.json file.
// It looks like this:
// {
//   "live": {
//     "Version": "2.3.4",
//     "Package": "https://protonmail.com/.../bridge_2.3.4_linux.tgz",
//     "Installers": [
//       "https://protonmail.com/.../something.deb",
//       "https://protonmail.com/.../something.rpm",
//       "https://protonmail.com/.../PKGBUILD"
//     ],
//     "Landing "https://protonmail.com/bridge",
//     "Rollout": 0.5
//   },
//   "beta": {
//     "Version": "2.4.0-beta",
//     "Package": "https://protonmail.com/.../bridge_2.4.0-beta_linux.tgz",
//     "Installers": [
//       "https://protonmail.com/.../something.deb",
//       "https://protonmail.com/.../something.rpm",
//       "https://protonmail.com/.../PKGBUILD"
//     ],
//     "Landing "https://protonmail.com/bridge",
//     "Rollout": 0.5
//   },
//   "...": {
//     ...
//   }
// }
type VersionMap map[string]VersionInfo

// getVersionFileURL returns the URL of the version file.
// For example:
//  - https://protonmail.com/download/bridge/version_linux.json
//  - https://protonmail.com/download/ie/version_linux.json
func (u *Updater) getVersionFileURL() string {
	return fmt.Sprintf("%v/%v/version_%v.json", Host, u.updateURLName, u.platform)
}
