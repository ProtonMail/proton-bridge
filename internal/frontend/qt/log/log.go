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

// Package log redirects QML logs to logrus
package log

//#include "log.h"
import "C"

import (
	"github.com/sirupsen/logrus"
	"github.com/therecipe/qt/core"
)

var logQML = logrus.WithField("pkg", "frontent/qml")

// InstallMessageHandler is registering logQML as logger for QML calls.
func InstallMessageHandler() {
	C.InstallMessageHandler()
}

//export logMsgPacked
func logMsgPacked(data *C.char, len C.int) {
	logQML.Warn(C.GoStringN(data, len))
}

// logDummy is here to trigger qtmoc to create cgo instructions
type logDummy struct {
	core.QObject
}
