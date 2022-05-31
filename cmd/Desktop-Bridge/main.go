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

package main

/*
                             ___....___
   ^^                __..-:'':__:..:__:'':-..__
                 _.-:__:.-:'':  :  :  :'':-.:__:-._
               .':.-:  :  :  :  :  :  :  :  :  :._:'.
            _ :.':  :  :  :  :  :  :  :  :  :  :  :'.: _
           [ ]:  :  :  :  :  :  :  :  :  :  :  :  :  :[ ]
           [ ]:  :  :  :  :  :  :  :  :  :  :  :  :  :[ ]
  :::::::::[ ]:__:__:__:__:__:__:__:__:__:__:__:__:__:[ ]:::::::::::
  !!!!!!!!![ ]!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!![ ]!!!!!!!!!!!
  ^^^^^^^^^[ ]^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^[ ]^^^^^^^^^^^
           [ ]                                        [ ]
           [ ]                                        [ ]
     jgs   [ ]                                        [ ]
   ~~^_~^~/   \~^-~^~ _~^-~_^~-^~_^~~-^~_~^~-~_~-^~_^/   \~^ ~~_ ^
*/

import (
	"os"

	"github.com/ProtonMail/proton-bridge/v2/internal/app/base"
	"github.com/ProtonMail/proton-bridge/v2/internal/app/bridge"
	"github.com/sirupsen/logrus"
)

const (
	appName       = "Proton Mail Bridge"
	appUsage      = "Proton Mail IMAP and SMTP Bridge"
	configName    = "bridge"
	updateURLName = "bridge"
	keychainName  = "bridge"
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

	if err := bridge.New(base).Run(os.Args); err != nil {
		logrus.WithError(err).Fatal("Bridge exited with error")
	}
}
