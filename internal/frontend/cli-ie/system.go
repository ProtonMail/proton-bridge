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

package cliie

import (
	"github.com/abiosoft/ishell"
)

func (f *frontendCLI) restart(c *ishell.Context) {
	if f.yesNoQuestion("Are you sure you want to restart the Import-Export app") {
		f.Println("Restarting the Import-Export app...")
		f.restarter.SetToRestart()
		f.Stop()
	}
}

func (f *frontendCLI) checkInternetConnection(c *ishell.Context) {
	if f.ie.CheckConnection() == nil {
		f.Println("Internet connection is available.")
	} else {
		f.Println("Can not contact the server, please check your internet connection.")
	}
}

func (f *frontendCLI) printLogDir(c *ishell.Context) {
	if path, err := f.locations.ProvideLogsPath(); err != nil {
		f.Println("Failed to determine location of log files")
	} else {
		f.Println("Log files are stored in\n\n ", path)
	}
}

func (f *frontendCLI) printManual(c *ishell.Context) {
	f.Println("More instructions about the Import-Export app can be found at\n\n  https://protonmail.com/support/categories/import-export/")
}
