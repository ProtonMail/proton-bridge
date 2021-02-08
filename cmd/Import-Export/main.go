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

package main

import (
	"os"

	"github.com/ProtonMail/proton-bridge/internal/app/base"
	"github.com/ProtonMail/proton-bridge/internal/app/ie"
	"github.com/sirupsen/logrus"
)

const (
	appName       = "ProtonMail Import-Export app"
	appUsage      = "Import and export messages to/from your ProtonMail account"
	configName    = "importExport"
	updateURLName = "ie"
	keychainName  = "import-export-app"
	cacheVersion  = "c11"
)

func main() {
	base, err := base.New(
		appName,
		appUsage,
		configName,
		updateURLName,
		keychainName,
		cacheVersion,
	)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create app base")
	}
	// Other instance already running.
	if base == nil {
		return
	}

	if err := ie.New(base).Run(os.Args); err != nil {
		logrus.WithError(err).Fatal("IE exited with error")
	}
}
