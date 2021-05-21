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

import QtQuick 2.13
import QtQuick.Window 2.15
import Qt.labs.platform 1.0

QtObject {

    property var _mainWindow: Window {
        id: mainWindow
        title: "window 1"
        visible: false
    }

    property var _trayMenu: Window {
        id: trayMenu
        title: "window 2"
        visible: false
        flags: Qt.Dialog

        width: 448
    }

    property var _trayIcon: SystemTrayIcon {
        id: trayIcon
        visible: true
        iconSource: "./icons/rectangle-systray.png"
        onActivated: {
            switch (reason) {
            case SystemTrayIcon.Unknown:
                break;
            case SystemTrayIcon.Context:
                break
            case SystemTrayIcon.DoubleClick:
                break
            case SystemTrayIcon.Trigger:
                trayMenu.x = (Screen.desktopAvailableWidth - trayMenu.width) / 2
                trayMenu.visible = !trayMenu.visible
                break;
            case SystemTrayIcon.MiddleClick:
                mainWindow.visible = !mainWindow.visible
                break;
            default:
                break;
            }
        }
    }
}
