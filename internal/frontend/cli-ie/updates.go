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

package cliie

import (
	"strings"

	"github.com/ProtonMail/proton-bridge/internal/importexport"
	"github.com/ProtonMail/proton-bridge/internal/updates"
	"github.com/abiosoft/ishell"
)

func (f *frontendCLI) checkUpdates(c *ishell.Context) {
	isUpToDate, latestVersionInfo, err := f.updates.CheckIsUpToDate()
	if err != nil {
		f.printAndLogError("Cannot retrieve version info: ", err)
		f.checkInternetConnection(c)
		return
	}
	if isUpToDate {
		f.Println("Your version is up to date.")
	} else {
		f.notifyNeedUpgrade()
		f.Println("")
		f.printReleaseNotes(latestVersionInfo)
	}
}

func (f *frontendCLI) printLocalReleaseNotes(c *ishell.Context) {
	localVersion := f.updates.GetLocalVersion()
	f.printReleaseNotes(localVersion)
}

func (f *frontendCLI) printReleaseNotes(versionInfo updates.VersionInfo) {
	f.Println(bold("ProtonMail Import-Export app "+versionInfo.Version), "\n")
	if versionInfo.ReleaseNotes != "" {
		f.Println(bold("Release Notes"))
		f.Println(versionInfo.ReleaseNotes)
	}
	if versionInfo.ReleaseFixedBugs != "" {
		f.Println(bold("Fixed bugs"))
		f.Println(versionInfo.ReleaseFixedBugs)
	}
}

func (f *frontendCLI) printCredits(c *ishell.Context) {
	for _, pkg := range strings.Split(importexport.Credits, ";") {
		f.Println(pkg)
	}
}
