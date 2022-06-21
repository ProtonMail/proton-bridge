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

//go:build build_qt
// +build build_qt

package qt

import (
	"github.com/therecipe/qt/core"
)

func (f *FrontendQt) setVersion() {
	f.qml.SetVersion(f.programVersion)
}

func (f *FrontendQt) setLogsPath() {
	path, err := f.locations.ProvideLogsPath()
	if err != nil {
		f.log.WithError(err).Error("Cannot update path folder")
		return
	}

	f.qml.SetLogsPath(core.QUrl_FromLocalFile(path))
}

func (f *FrontendQt) setLicensePath() {
	f.qml.SetLicensePath(core.QUrl_FromLocalFile(f.locations.GetLicenseFilePath()))
	f.qml.SetDependencyLicensesLink(core.NewQUrl3(f.locations.GetDependencyLicensesLink(), core.QUrl__TolerantMode))
}

func (f *FrontendQt) setCurrentEmailClient() {
	f.qml.SetCurrentEmailClient(f.userAgent.String())
}

func (f *FrontendQt) reportBug(description, address, emailClient string, includeLogs bool) {
	defer f.qml.ReportBugFinished()

	if err := f.bridge.ReportBug(
		core.QSysInfo_ProductType(),
		core.QSysInfo_PrettyProductName(),
		description,
		address,
		address,
		emailClient,
		includeLogs,
	); err != nil {
		f.log.WithError(err).Error("Failed to report bug")
		f.qml.BugReportSendError()
		return
	}

	f.qml.BugReportSendSuccess()
}
