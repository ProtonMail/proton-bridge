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

package cli

import (
	"github.com/abiosoft/ishell"
)

var (
	currentPort = "" //nolint[gochecknoglobals]
)

func (f *frontendCLI) restart(c *ishell.Context) {
	if f.yesNoQuestion("Are you sure you want to restart the Import/Export") {
		f.Println("Restarting Import/Export...")
		f.appRestart = true
		f.Stop()
	}
}

func (f *frontendCLI) checkInternetConnection(c *ishell.Context) {
	if f.ie.CheckConnection() == nil {
		f.Println("Internet connection is available.")
	} else {
		f.Println("Can not contact the server, please check you internet connection.")
	}
}

func (f *frontendCLI) printLogDir(c *ishell.Context) {
	f.Println("Log files are stored in\n\n ", f.config.GetLogDir())
}

func (f *frontendCLI) printManual(c *ishell.Context) {
	f.Println("More instructions about the Import/Export can be found at\n\n  https://protonmail.com/support/categories/import-export/")
}
