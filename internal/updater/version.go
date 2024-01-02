// Copyright (c) 2024 Proton AG
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

	// LandingPage is the address of the app landing page on proton.me
	LandingPage string

	// ReleaseNotesPage is the address of the page containing the release notes.
	ReleaseNotesPage string

	// RolloutProportion indicates the proportion (0,1] of users that should update to this version.
	RolloutProportion float64
}

// VersionMap represents the structure of the version.json file.
// It looks like this:
//
//	{
//	  "stable": {
//	    "Version": "2.3.4",
//	    "Package": "https://proton.me/.../bridge_2.3.4_linux.tgz",
//	    "Installers": [
//	      "https://proton.me/.../something.deb",
//	      "https://proton.me/.../something.rpm",
//	      "https://proton.me/.../PKGBUILD"
//	    ],
//	    "LandingPage": "https://proton.me/mail/bridge#download",
//	    "ReleaseNotesPage": "https://proton.me/download/{ie,bridge}/{stable,early}_releases.html",
//	    "RolloutProportion": 0.5
//	  },
//	  "early": {
//	    "Version": "2.4.0",
//	    "Package": "https://proton.me/.../bridge_2.4.0_linux.tgz",
//	    "Installers": [
//	      "https://proton.me/.../something.deb",
//	      "https://proton.me/.../something.rpm",
//	      "https://proton.me/.../PKGBUILD"
//	    ],
//	    "LandingPage": "https://proton.me/mail/bridge#download",
//	    "ReleaseNotesPage": "https://proton.me/download/{ie,bridge}/{stable,early}_releases.html",
//	    "RolloutProportion": 0.5
//	  },
//	  "...": {
//	    ...
//	  }
//	}.

type VersionMap map[Channel]VersionInfo
