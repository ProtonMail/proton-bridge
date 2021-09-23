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
import Qt.labs.platform 1.1

import Notifications 1.0

QtObject {
    id: root

    property var backend: go

    property Notifications _notifications: Notifications {
        id: notifications
        backend: root.backend
        frontendMain: mainWindow
        frontendStatus: statusWindow
        frontendTray: trayIcon
    }

    property MainWindow _mainWindow: MainWindow {
        id: mainWindow
        visible: false

        backend: root.backend
        notifications: root._notifications

        onLogin: {
            backend.login(username, password)
        }
        onLogin2FA: {
            backend.login2FA(username, code)
        }
        onLogin2Password: {
            backend.login2Password(username, password)
        }
        onLoginAbort: {
            backend.loginAbort(username)
        }

        onVisibleChanged: {
            backend.dockIconVisible = visible
        }
    }

    property StatusWindow _statusWindow: StatusWindow {
        id: statusWindow
        visible: false

        backend: root.backend
        notifications: root._notifications

        property var x_center: 10
        property var x_min: 0
        property var x_max: 100
        property var y_center: 1000
        property var y_min: 0
        property var y_max: 10000

        x: bound(x_center,x_min, x_max-statusWindow.width)
        y: bound(y_center,y_min, y_max-statusWindow.height)


        onShowMainWindow: {
            mainWindow.showAndRise()
        }

        onShowHelp: {
            mainWindow.showHelp()
            mainWindow.showAndRise()
        }

        onShowSettings: {
            mainWindow.showSettings()
            mainWindow.showAndRise()
        }

        onShowSignIn: {
            mainWindow.showSignIn(username)
            mainWindow.showAndRise()
        }

        onQuit: {
            backend.quit()
        }

        function bound(num, lower_limit, upper_limit) {
            return Math.max(lower_limit, Math.min(upper_limit, num))
        }
    }

    property SystemTrayIcon _trayIcon: SystemTrayIcon {
        id: trayIcon
        visible: true
        icon.source: "./icons/systray-mono.png"
        icon.mask: true // make sure that systems like macOS will use proper color
        tooltip: `Proton Mail Bridge v${go.version}`
        onActivated: {
            function calcStatusWindowPosition() {
                function isInInterval(num, lower_limit, upper_limit) {
                    return lower_limit <= num && num <= upper_limit
                }

                // First we get icon center position.
                // On some platforms (X11 / Wayland) Qt does not provide icon geometry info.
                // In this case we rely on cursor position
                var iconWidth = geometry.width *1.2
                var iconHeight = geometry.height *1.2
                var iconCenter = Qt.point(geometry.x + (geometry.width / 2), geometry.y + (geometry.height / 2))
                if (geometry.width == 0 && geometry.height == 0) {
                    iconCenter = backend.getCursorPos()
                    // fallback: simple guess, no data to estimate
                    iconWidth = 25
                    iconHeight = 25
                }

                // Find screen
                var screen = Qt.application.screens[0]

                for (var i in Qt.application.screens) {
                    screen = Qt.application.screens[i]
                    if (
                        isInInterval(iconCenter.x, screen.virtualX, screen.virtualX+screen.width) &&
                        isInInterval(iconCenter.y, screen.virtualY, screen.virtualY+screen.heigh)
                    ) {
                        return
                    }
                }

                // Calculate allowed square where status window top left corner can be positioned
                statusWindow.x_center = iconCenter.x
                statusWindow.y_center = iconCenter.y
                statusWindow.x_min = screen.virtualX + iconWidth
                statusWindow.x_max = screen.virtualX + screen.width - iconWidth
                statusWindow.y_min = screen.virtualY + iconHeight
                statusWindow.y_max = screen.virtualY + screen.height - iconHeight
            }

            function toggleWindow(win) {
                if (win.visible) {
                    win.close()
                } else {
                    win.showAndRise()
                }
            }


            switch (reason) {
                case SystemTrayIcon.Unknown:
                break;
                case SystemTrayIcon.Context:
                case SystemTrayIcon.Trigger:
                calcStatusWindowPosition()
                toggleWindow(statusWindow)
                break
                case SystemTrayIcon.DoubleClick:
                case SystemTrayIcon.MiddleClick:
                toggleWindow(mainWindow)
                break;
                default:
                break;
            }
        }
    }

    Component.onCompleted: {
        if (root.backend.users.count === 0) {
            mainWindow.showAndRise()
        }

        if (root.backend.users.count === 1 && root.backend.users.get(0).loggedIn === false) {
            mainWindow.showAndRise()
        }

        if (root.backend.showOnStartup) {
            mainWindow.showAndRise()
        }

        root.backend.guiReady()
    }
}
