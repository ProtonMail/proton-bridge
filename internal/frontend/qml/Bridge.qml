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

import QtQml 2.12
import QtQuick 2.13
import QtQuick.Window 2.13
import Qt.labs.platform 1.0

QtObject {
    id: root

    property var backend
    property var users

    signal login(string username, string password)
    signal login2FA(string username, string code)
    signal login2Password(string username, string password)
    signal loginAbort(string username)

    property var mainWindow: MainWindow {
        id: mainWindow
        visible: true

        backend: root.backend
        users: root.users

        onLogin: {
            root.login(username, password)
        }
        onLogin2FA: {
            root.login2FA(username, code)
        }
        onLogin2Password: {
            root.login2Password(username, password)
        }
        onLoginAbort: {
            root.loginAbort(username)
        }
    }

    property var _trayMenu: Window {
        id: trayMenu
        title: "window 2"
        visible: false
        flags: Qt.Dialog
    }

    property var _trayIcon: SystemTrayIcon {
        id: trayIcon
        visible: true
        iconSource: "./icons/ic-systray.svg"
        onActivated: {
            switch (reason) {
                case SystemTrayIcon.Unknown:
                break;
                case SystemTrayIcon.Context:
                trayMenu.x = (Screen.desktopAvailableWidth - trayMenu.width) / 2
                trayMenu.visible = !trayMenu.visible
                break
                case SystemTrayIcon.DoubleClick:
                mainWindow.visible = !mainWindow.visible
                break;
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
