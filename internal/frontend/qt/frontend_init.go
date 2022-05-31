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

package qt

import (
	"errors"
	"os"
	"runtime"

	"github.com/Masterminds/semver/v3"
	qmlLog "github.com/ProtonMail/proton-bridge/v2/internal/frontend/qt/log"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/qml"
	"github.com/therecipe/qt/quickcontrols2"
	"github.com/therecipe/qt/widgets"
)

func (f *FrontendQt) initiateQtApplication() error {
	qmlLog.InstallMessageHandler()

	f.app = widgets.NewQApplication(len(os.Args), os.Args)

	if os.Getenv("QSG_INFO") != "" && os.Getenv("QSG_INFO") != "0" {
		core.QLoggingCategory_SetFilterRules("qt.scenegraph.general=true")
	}

	core.QCoreApplication_SetApplicationName(f.programName)
	core.QCoreApplication_SetApplicationVersion(f.programVersion)
	core.QCoreApplication_SetOrganizationName("Proton AG")
	core.QCoreApplication_SetOrganizationDomain("proton.ch")

	// High DPI scaling for windows.
	core.QCoreApplication_SetAttribute(core.Qt__AA_EnableHighDpiScaling, false)

	// Use software OpenGL to avoid dedicated GPU on darwin. It cause no
	// problems on linux, but it can cause initializaion issues on windows
	// for some specific GPU / driver combination.
	if runtime.GOOS != "windows" {
		core.QCoreApplication_SetAttribute(core.Qt__AA_UseSoftwareOpenGL, true)
	}

	// Bridge runs background, no window is needed to be opened.
	f.app.SetQuitOnLastWindowClosed(false)

	// QML Engine and path
	f.engine = qml.NewQQmlEngine(f.app)
	rootComponent := qml.NewQQmlComponent2(f.engine, f.engine)

	f.qml = NewQMLBackend(f.engine)
	f.qml.setup(f)

	f.engine.AddImportPath("qrc:/qml/")
	f.engine.AddPluginPath("qrc:/qml/")

	// Add style: if colorScheme / style is forgotten we should fallback to
	// default style and should be Proton
	quickcontrols2.QQuickStyle_AddStylePath("qrc:/qml/")
	quickcontrols2.QQuickStyle_SetStyle("Proton")

	// Before loading a component we should load translations.
	// See https://github.com/qt/qtdeclarative/blob/bedef212a74a62452ed31d7f65536a6c67881fc4/src/qml/qml/qqmlapplicationengine.cpp#L69 as example.

	rootComponent.LoadUrl(core.NewQUrl3("qrc:/qml/Bridge.qml", 0))

	if rootComponent.Status() != qml.QQmlComponent__Ready {
		return errors.New("QML not loaded properly")
	}

	// Instead of creating component right away we use BeginCreate to stop right before binding evaluation.
	// That is needed to set backend property so all bindings will be calculated properly.
	rootObject := rootComponent.BeginCreate(f.engine.RootContext())
	// Check QML is loaded properly.
	if rootObject == nil {
		return errors.New("QML not created properly")
	}

	rootObject.SetProperty("backend", f.qml.ToVariant())
	rootComponent.CompleteCreate()

	return nil
}

func (f *FrontendQt) setShowSplashScreen() {
	f.qml.SetShowSplashScreen(false)

	// Splash screen should not be shown to new users or after factory reset.
	if f.bridge.IsFirstStart() {
		return
	}

	ver, err := semver.NewVersion(f.bridge.GetLastVersion())
	if err != nil {
		f.log.WithError(err).WithField("last", f.bridge.GetLastVersion()).Debug("Cannot parse last version")
		return
	}

	// Current splash screen contains update on rebranding. Therefore, it
	// should be shown only if the last used version was less than 2.2.0.
	if ver.LessThan(semver.MustParse("2.2.0")) {
		f.qml.SetShowSplashScreen(true)
	}
}
