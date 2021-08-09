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

import (
	"errors"
	qmlLog "github.com/ProtonMail/proton-bridge/internal/frontend/qt/log"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/qml"
	"github.com/therecipe/qt/quickcontrols2"
	"github.com/therecipe/qt/widgets"
	"os"
)

func (f *FrontendQt) initiateQtApplication() error {
	qmlLog.InstallMessageHandler()

	f.app = widgets.NewQApplication(len(os.Args), os.Args)

	core.QCoreApplication_SetApplicationName(f.programName)
	core.QCoreApplication_SetApplicationVersion(f.programVersion)

	// High DPI scaling for windows.
	core.QCoreApplication_SetAttribute(core.Qt__AA_EnableHighDpiScaling, false)
	// Software OpenGL: to avoid dedicated GPU.
	core.QCoreApplication_SetAttribute(core.Qt__AA_UseSoftwareOpenGL, true)

	// Bridge runs background, no window is needed to be opened.
	f.app.SetQuitOnLastWindowClosed(false)

	// QML Engine and path
	f.engine = qml.NewQQmlApplicationEngine(f.app)

	f.qml = NewQMLBackend(nil)
	f.qml.setup(f)
	f.engine.RootContext().SetContextProperty("go", f.qml)

	f.engine.AddImportPath("qrc:/qml/")
	f.engine.AddPluginPath("qrc:/qml/")

	// Add style: if colorScheme / style is forgotten we should fallback to
	// default style and should be Proton
	quickcontrols2.QQuickStyle_AddStylePath("qrc:/qml/")
	quickcontrols2.QQuickStyle_SetStyle("Proton")

	f.engine.Load(core.NewQUrl3("qrc:/qml/Bridge.qml", 0))

	// Check QML is loaded properly.
	if len(f.engine.RootObjects()) == 0 {
		return errors.New("QML not loaded properly")
	}

	return nil
}
