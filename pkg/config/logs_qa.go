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

// +build build_qa

package config

import (
	"github.com/sirupsen/logrus"
)

// getLogLevelAndFile for QA build is altered in a way even decrypted data are stored
// in the log file when forced with `debug-client-json` or `debug-server-json`.
func getLogLevelAndFile(levelFlag string) (level logrus.Level, useFile bool) {
	useFile = true
	switch levelFlag {
	case "panic":
		level = logrus.PanicLevel
	case "fatal":
		level = logrus.FatalLevel
	case "error":
		level = logrus.ErrorLevel
	case "warn":
		level = logrus.WarnLevel
	case "info":
		level = logrus.InfoLevel
	case "debug-client-json", "debug-server-json":
		level = logrus.DebugLevel
	case "debug", "debug-client", "debug-server":
		level = logrus.DebugLevel
		useFile = false
	default:
		level = logrus.InfoLevel
	}
	return
}
