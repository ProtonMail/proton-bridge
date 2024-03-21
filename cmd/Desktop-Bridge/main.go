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

package main

import (
	"os"
	"runtime"
	"strings"

	"github.com/ProtonMail/proton-bridge/v3/internal/constants"
	"github.com/ProtonMail/proton-bridge/v3/internal/locations"
	"github.com/ProtonMail/proton-bridge/v3/internal/logging"
	"github.com/ProtonMail/proton-bridge/v3/internal/sentry"
	"github.com/sirupsen/logrus"

	"github.com/ProtonMail/proton-bridge/v3/internal/app"
	"github.com/bradenaw/juniper/xslices"
)

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

func main() {
	appErr := app.New().Run(xslices.Filter(os.Args, func(arg string) bool { return !strings.Contains(arg, "-psn_") }))
	if appErr != nil {
		_ = app.WithLocations(func(l *locations.Locations) error {
			logsPath, err := l.ProvideLogsPath()
			if err != nil {
				return err
			}

			// Get the session ID if its specified
			var sessionID logging.SessionID
			if flagVal, found := getFlagValue(os.Args, app.FlagSessionID); found {
				sessionID = logging.SessionID(flagVal)
			} else {
				sessionID = logging.NewSessionID()
			}

			closer, err := logging.Init(
				logsPath,
				sessionID,
				logging.BridgeShortAppName,
				logging.DefaultMaxLogFileSize,
				logging.DefaultPruningSize,
				"",
			)
			if err != nil {
				return err
			}

			defer func() {
				_ = logging.Close(closer)
			}()

			logrus.
				WithField("appName", constants.FullAppName).
				WithField("version", constants.Version).
				WithField("revision", constants.Revision).
				WithField("tag", constants.Tag).
				WithField("build", constants.BuildTime).
				WithField("runtime", runtime.GOOS).
				WithField("args", os.Args).
				WithField("SentryID", sentry.GetProtectedHostname()).WithError(appErr).Error("Failed to initialize bridge")
			return nil
		})
	}
}

// getFlagValue - obtains the value of a specified tag
// The flag can be of the following form `-flag value`, `--flag value`, `-flag=value` or `--flags=value`.
func getFlagValue(argList []string, flag string) (string, bool) {
	eqPrefix1 := "-" + flag + "="
	eqPrefix2 := "--" + flag + "="

	for i := 0; i < len(argList); i++ {
		arg := argList[i]
		if strings.HasPrefix(arg, eqPrefix1) {
			val := strings.TrimPrefix(arg, eqPrefix1)
			return val, len(val) > 0
		}
		if strings.HasPrefix(arg, eqPrefix2) {
			val := strings.TrimPrefix(arg, eqPrefix2)
			return val, len(val) > 0
		}
		if (arg == "-"+flag || arg == "--"+flag) && i+1 < len(argList) {
			return argList[i+1], true
		}
	}

	return "", false
}
