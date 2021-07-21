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

package qtcommon

//#include "common.h"
import "C"

import (
	"bufio"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/therecipe/qt/core"
)

var log = logrus.WithField("pkg", "frontend/qt-common")
var logQML = logrus.WithField("pkg", "frontend/qml")

// RegisterTypes for vector of ints
func RegisterTypes() { // need to fix test message
	C.RegisterTypes()
}

func installMessageHandler() {
	C.InstallMessageHandler()
}

//export logMsgPacked
func logMsgPacked(data *C.char, len C.int) {
	logQML.Warn(C.GoStringN(data, len))
}

// QtSetupCoreAndControls hanldes global setup of Qt.
// Should be called once per program. Probably once per thread is fine.
func QtSetupCoreAndControls(programName, programVersion string) {
	installMessageHandler()
	// Core setup.
	core.QCoreApplication_SetApplicationName(programName)
	core.QCoreApplication_SetApplicationVersion(programVersion)
	// High DPI scaling for windows.
	core.QCoreApplication_SetAttribute(core.Qt__AA_EnableHighDpiScaling, false)
	// Software OpenGL: to avoid dedicated GPU.
	core.QCoreApplication_SetAttribute(core.Qt__AA_UseSoftwareOpenGL, true)
	// Basic style for QuickControls2 objects.
	//quickcontrols2.QQuickStyle_SetStyle("material")
}

// NewQByteArrayFromString is wrapper for new QByteArray from string
func NewQByteArrayFromString(name string) *core.QByteArray {
	return core.NewQByteArray2(name, -1)
}

// NewQVariantString is wrapper for QVariant alocator String
func NewQVariantString(data string) *core.QVariant {
	return core.NewQVariant1(data)
}

// NewQVariantStringArray is wrapper for QVariant alocator String Array
func NewQVariantStringArray(data []string) *core.QVariant {
	return core.NewQVariant1(data)
}

// NewQVariantBool is wrapper for QVariant alocator Bool
func NewQVariantBool(data bool) *core.QVariant {
	return core.NewQVariant1(data)
}

// NewQVariantInt is wrapper for QVariant alocator Int
func NewQVariantInt(data int) *core.QVariant {
	return core.NewQVariant1(data)
}

// NewQVariantLong is wrapper for QVariant alocator Int64
func NewQVariantLong(data int64) *core.QVariant {
	return core.NewQVariant1(data)
}

// Pause used to show GUI tests
func Pause() {
	time.Sleep(500 * time.Millisecond)
}

// Longer pause used to diplay GUI tests
func PauseLong() {
	time.Sleep(3 * time.Second)
}

// FIXME: Not working in test...
func WaitForEnter() {
	log.Print("Press 'Enter' to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

type Listener interface {
	Add(string, chan<- string)
	RetryEmit(string)
}
