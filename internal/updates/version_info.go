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

package updates

import (
	"runtime"
	"strings"
)

type VersionInfo struct {
	Version          string
	Revision         string
	ReleaseDate      string   // Timestamp generated automatically
	ReleaseNotes     string   // List of features, new line separated with leading dot e.g. `• example\n`
	ReleaseFixedBugs string   // List of fixed bugs, same usage as release notes
	FixedBugs        []string // Deprecated list of fixed bugs keeping for backward compatibility (mandatory for working versions up to 1.1.5)
	URL              string   // Open browser and download (obsolete replaced by InstallerFile)

	LandingPage   string // landing page for manual download
	UpdateFile    string // automatic update file
	InstallerFile string `json:",omitempty"` // manual update file
	DebFile       string `json:",omitempty"` // debian package file
	RpmFile       string `json:",omitempty"` // red hat package file
	PkgFile       string `json:",omitempty"` // arch PKGBUILD file
}

func (info *VersionInfo) GetDownloadLink() string {
	switch runtime.GOOS {
	case "linux":
		return strings.Join([]string{info.DebFile, info.RpmFile, info.PkgFile}, "\n")
	default:
		return info.InstallerFile
	}
}
