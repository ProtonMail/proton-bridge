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

    property var backend

    signal login(string username, string password)
    signal login2FA(string username, string code)
    signal login2Password(string username, string password)
    signal loginAbort(string username)

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
        notifications: notifications

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

    property StatusWindow _statusWindow: StatusWindow {
        id: statusWindow
        visible: false

        backend: root.backend
        notifications: notifications

        onShowMainWindow: {
            mainWindow.visible = true
        }
        onShowHelp: {

        }
        onShowSettings: {

        }
        onQuit: {
            backend.quit()
        }
    }

    property SystemTrayIcon _trayIcon: SystemTrayIcon {
        id: trayIcon
        visible: true
        iconSource: "./icons/ic-systray.svg"
        onActivated: {
            function calcStatusWindowPosition(statusWidth, statusHeight) {
                function bound(num, lower_limit, upper_limit) {
                    return Math.max(lower_limit, Math.min(upper_limit, num))
                }
                // checks if rect1 fits within rect2
                function isRectFit(rect1, rect2) {
                    //if (rect2.)
                    if ((rect2.left > rect1.left) ||
                            (rect2.right < rect1.right) ||
                            (rect2.top > rect1.top) ||
                            (rect2.bottom < rect1.bottom)) {
                        return false
                    }

                    return true
                }

                // First we get icon center position.
                // On some platforms (X11 / Wayland) Qt does not provide icon geometry info.
                // In this case we rely on cursor position
                var iconCenter = Qt.point(geometry.x + (geometry.width / 2), geometry.y + (geometry.height / 2))

                if (geometry.width == 0 && geometry.height == 0) {
                    iconCenter = backend.getCursorPos()
                }

                // Now bound this position to virtual screen available rect
                // TODO: here we should detect which screen mouse is on and use that screen available geometry to bound
                iconCenter.x = bound(iconCenter.x, 0, Qt.application.screens[0].desktopAvailableWidth)
                iconCenter.y = bound(iconCenter.y, 0, Qt.application.screens[0].desktopAvailableHeight)

                var x = 0
                var y = 0

                // Check if window may fit above
                x = iconCenter.x - statusWidth / 2
                y = iconCenter.y - statusHeight
                if (isRectFit(
                            Qt.rect(x, y, statusWidth, statusHeight),
                            // TODO: we should detect which screen mouse is on and use that screen available geometry to bound
                            Qt.rect(0, 0, Qt.application.screens[0].desktopAvailableWidth, Qt.application.screens[0].desktopAvailableHeight)
                            )) {
                    return Qt.point(x, y)
                }

                // Check if window may fit below
                x = iconCenter.x - statusWidth / 2
                y = iconCenter.y
                if (isRectFit(
                            Qt.rect(x, y, statusWidth, statusHeight),
                            // TODO: we should detect which screen mouse is on and use that screen available geometry to bound
                            Qt.rect(0, 0, Qt.application.screens[0].desktopAvailableWidth, Qt.application.screens[0].desktopAvailableHeight)
                            )) {
                    return Qt.point(x, y)
                }

                // Check if window may fit to the left
                x = iconCenter.x - statusWidth
                y = iconCenter.y - statusHeight / 2
                if (isRectFit(
                            Qt.rect(x, y, statusWidth, statusHeight),
                            // TODO: we should detect which screen mouse is on and use that screen available geometry to bound
                            Qt.rect(0, 0, Qt.application.screens[0].desktopAvailableWidth, Qt.application.screens[0].desktopAvailableHeight)
                            )) {
                    return Qt.point(x, y)
                }

                // Check if window may fit to the right
                x = iconCenter.x
                y = iconCenter.y - statusHeight / 2
                if (isRectFit(
                            Qt.rect(x, y, statusWidth, statusHeight),
                            // TODO: we should detect which screen mouse is on and use that screen available geometry to bound
                            Qt.rect(0, 0, Qt.application.screens[0].desktopAvailableWidth, Qt.application.screens[0].desktopAvailableHeight)
                            )) {
                    return Qt.point(x, y)
                }

                // TODO: add fallback
            }

            switch (reason) {
            case SystemTrayIcon.Unknown:
                break;
            case SystemTrayIcon.Context:
            case SystemTrayIcon.Trigger:!statusWindow.visible
                if (!statusWindow.visible) {
                    var point = calcStatusWindowPosition(statusWindow.width, statusWindow.height)
                    statusWindow.x = point.x
                    statusWindow.y = point.y
                }
                statusWindow.visible = !statusWindow.visible
                break
            case SystemTrayIcon.DoubleClick:
            case SystemTrayIcon.MiddleClick:
                mainWindow.visible = !mainWindow.visible
                break;
            default:
                break;
            }
        }
    }

    Component.onCompleted: {
        if (root.backend.users.count === 0) {
            mainWindow.show()
        }

        if (root.backend.users.count === 1 && root.backend.users.get(0).loggedIn === false) {
            mainWindow.show()
        }
    }
}
