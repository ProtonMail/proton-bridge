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

// +build build_qt

package qt

func (f *FrontendQt) setVersion() {
	f.qml.SetVersion(f.programVersion)
}

func (f *FrontendQt) setLogsPath() {
	path, err := f.locations.ProvideLogsPath()
	if err != nil {
		f.log.WithError(err).Error("Cannot update path folder")
		return
	}
	f.qml.SetLogsPath(path)
}

func (f *FrontendQt) setLicensePath() {
	f.qml.SetLicensePath(f.locations.GetLicenseFilePath())
}

func (f *FrontendQt) setCurrentEmailClient() {
	f.qml.SetCurrentEmailClient(f.userAgent.String())
}

func (f *FrontendQt) reportBug(description, address, emailClient string, includeLogs bool) {
	//TODO
}
