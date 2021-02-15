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
	"runtime"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

const (
	systrayNormal  = ""
	systrayWarning = "-warn"
	systrayError   = "-error"
)

func min(a, b int) int {
	if b < a {
		return b
	}
	return a
}

func max(a, b int) int {
	if b > a {
		return b
	}
	return a
}

var systray *widgets.QSystemTrayIcon

func SetupSystray(frontend *FrontendQt) {
	systray = widgets.NewQSystemTrayIcon(nil)
	NormalSystray()
	systray.SetToolTip(frontend.programName)
	systray.SetContextMenu(createMenu(frontend, systray))

	if runtime.GOOS != "darwin" {
		systray.ConnectActivated(func(reason widgets.QSystemTrayIcon__ActivationReason) {
			switch reason {
			case widgets.QSystemTrayIcon__Trigger, widgets.QSystemTrayIcon__DoubleClick:
				frontend.Qml.ShowWindow()
			default:
				systray.ContextMenu().Exec2(menuPosition(systray), nil)
			}
		})
	}

	systray.Show()
}

func qsTr(msg string) string {
	return systray.Tr(msg, "Systray menu", -1)
}

func createMenu(frontend *FrontendQt, systray *widgets.QSystemTrayIcon) *widgets.QMenu {
	menu := widgets.NewQMenu(nil)
	menu.AddAction(qsTr("Open")).ConnectTriggered(func(ok bool) { frontend.Qml.ShowWindow() })
	menu.AddAction(qsTr("Help")).ConnectTriggered(func(ok bool) { frontend.Qml.ShowHelp() })
	menu.AddAction(qsTr("Quit")).ConnectTriggered(func(ok bool) { frontend.Qml.ShowQuit() })
	return menu
}

func menuPosition(systray *widgets.QSystemTrayIcon) *core.QPoint {
	var availRec = gui.QGuiApplication_PrimaryScreen().AvailableGeometry()
	var trayRec = systray.Geometry()
	var x = max(availRec.Left(), min(trayRec.X(), availRec.Right()-trayRec.Width()))
	var y = max(availRec.Top(), min(trayRec.Y(), availRec.Bottom()-trayRec.Height()))
	return core.NewQPoint2(x, y)
}

func showSystray(systrayType string) {
	path := ":/ProtonUI/images/systray" + systrayType
	if runtime.GOOS == "darwin" {
		path += "-mono"
	}
	path += ".png"
	icon := gui.NewQIcon5(path)
	icon.SetIsMask(true)
	systray.SetIcon(icon)
}

func NormalSystray() {
	showSystray(systrayNormal)
}

func HighlightSystray() {
	showSystray(systrayWarning)
}

func ErrorSystray() {
	showSystray(systrayError)
}

func HideSystray() {
	systray.Hide()
}
