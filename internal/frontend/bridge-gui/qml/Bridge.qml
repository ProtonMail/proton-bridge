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

import QtQml 2.12
import QtQuick 2.13
import QtQuick.Window 2.13
import Qt.labs.platform 1.1

import Proton 4.0
import Notifications 1.0

QtObject {
    id: root

    function isInInterval(num, lower_limit, upper_limit) {
        return lower_limit <= num && num <= upper_limit
    }
    function bound(num, lower_limit, upper_limit) {
        return Math.max(lower_limit, Math.min(upper_limit, num))
    }

    property var backend
    property var title: "Proton Mail Bridge"

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

        title: root.title
        backend: root.backend
        notifications: root._notifications

        onVisibleChanged: {
            backend.dockIconVisible = visible
        }

        Connections {
            target: root.backend
            onCacheUnavailable: {
                mainWindow.showAndRise()
            }
            onColorSchemeNameChanged: root.setColorScheme()
        }
    }

    property StatusWindow _statusWindow: StatusWindow {
        id: statusWindow
        visible: false

        title: root.title
        backend: root.backend
        notifications: root._notifications

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

        property rect screenRect
        property rect iconRect

        // use binding from function with width and height as arguments so it will be recalculated every time width and height are changed
        property point position: getPosition(width, height)
        x: position.x
        y: position.y

        function getPosition(_width, _height) {
            if (screenRect.width === 0 || screenRect.height === 0) {
                return Qt.point(0, 0)
            }

            var _x = 0
            var _y = 0

            // fit above
            _y = iconRect.top - height
            if (isInInterval(_y, screenRect.top, screenRect.bottom - height)) {
                // position preferebly in the horizontal center but bound to the screen rect
                _x = bound(iconRect.left + (iconRect.width - width)/2, screenRect.left, screenRect.right - width)
                return Qt.point(_x, _y)
            }

            // fit below
            _y = iconRect.bottom
            if (isInInterval(_y, screenRect.top, screenRect.bottom - height)) {
                // position preferebly in the horizontal center but bound to the screen rect
                _x = bound(iconRect.left + (iconRect.width - width)/2, screenRect.left, screenRect.right - width)
                return Qt.point(_x, _y)
            }

            // fit to the left
            _x = iconRect.left - width
            if (isInInterval(_x, screenRect.left, screenRect.right - width)) {
                // position preferebly in the vertical center but bound to the screen rect
                _y = bound(iconRect.top + (iconRect.height - height)/2, screenRect.top, screenRect.bottom - height)
                return Qt.point(_x, _y)
            }

            // fit to the right
            _x = iconRect.right
            if (isInInterval(_x, screenRect.left, screenRect.right - width)) {
                // position preferebly in the vertical center but bound to the screen rect
                _y = bound(iconRect.top + (iconRect.height - height)/2, screenRect.top, screenRect.bottom - height)
                return Qt.point(_x, _y)
            }

            // Fallback: position satatus window right above icon and let window manager decide.
            console.warn("Can't position status window: screenRect =", screenRect, "iconRect =", iconRect)
            _x = bound(iconRect.left + (iconRect.width - width)/2, screenRect.left, screenRect.right - width)
            _y = bound(iconRect.top + (iconRect.height - height)/2, screenRect.top, screenRect.bottom - height)
            return Qt.point(_x, _y)
        }
    }

    property SystemTrayIcon _trayIcon: SystemTrayIcon {
        id: trayIcon
        visible: true
        icon.source: getTrayIconPath()
        icon.mask: true // make sure that systems like macOS will use proper color
        tooltip: `${root.title} v${backend.version}`
        onActivated: {
            function calcStatusWindowPosition() {
                // On some platforms (X11 / Plasma) Qt does not provide icon position and geometry info.
                // In this case we rely on cursor position
                var iconRect = Qt.rect(geometry.x, geometry.y, geometry.width, geometry.height)
                if (geometry.width == 0 && geometry.height == 0) {
                    var mousePos = backend.getCursorPos()
                    iconRect.x = mousePos.x
                    iconRect.y = mousePos.y
                    iconRect.width = 0
                    iconRect.height = 0
                }

                // Find screen
                var screen
                for (var i in Qt.application.screens) {
                    var _screen = Qt.application.screens[i]
                    if (
                        isInInterval(iconRect.x, _screen.virtualX, _screen.virtualX + _screen.width) &&
                        isInInterval(iconRect.y, _screen.virtualY, _screen.virtualY + _screen.height)
                    ) {
                        screen = _screen
                        break
                    }
                }
                if (!screen) {
                    // Fallback to primary screen
                    screen = Qt.application.screens[0]
                }

                // In case we used mouse to detect icon position - we want to make a fake icon rectangle from a point
                if (iconRect.width == 0 && iconRect.height == 0) {
                    iconRect.x = bound(iconRect.x - 16, screen.virtualX, screen.virtualX + screen.width - 32)
                    iconRect.y = bound(iconRect.y - 16, screen.virtualY, screen.virtualY + screen.height - 32)
                    iconRect.width = 32
                    iconRect.height = 32
                }

                statusWindow.screenRect = Qt.rect(screen.virtualX, screen.virtualY, screen.width, screen.height)
                statusWindow.iconRect = iconRect
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
                case SystemTrayIcon.DoubleClick:
                case SystemTrayIcon.MiddleClick:
                calcStatusWindowPosition()
                toggleWindow(statusWindow)
                break;
                default:
                break;
            }
        }

        property NotificationFilter _systrayfilter: NotificationFilter {
            source: root._notifications ? root._notifications.all : undefined
        }

        function getTrayIconPath() {
            var color = backend.goos == "darwin" ? "mono" : "color" 

            var level = "norm"
            if (_systrayfilter.topmost) {
                switch (_systrayfilter.topmost.type) {
                    case Notification.NotificationType.Danger:
                    level = "error"
                    break;
                    case Notification.NotificationType.Warning:
                    level = "warn"
                    break;
                    case Notification.NotificationType.Info:
                    level = "update"
                    break;
                }
            }

            return `./icons/systray-${color}-${level}.png`
        }
    }

    Component.onCompleted: {
        if (!root.backend) {
            console.log("backend not loaded")
        }

        root.setColorScheme()


        if (!root.backend.users) {
            console.log("users not loaded")
        }

        var c = root.backend.users.count
        var u = root.backend.users.get(0)
        // DEBUG
        if (c != 0) {
            console.log("users non zero", c)
            console.log("first user", u )
        }

        if (c === 0) {
            mainWindow.showAndRise()
        }

        if (u) {
            if (c === 1 && u.loggedIn === false) {
                mainWindow.showAndRise()
            }
        }

        if (root.backend.showOnStartup) {
            mainWindow.showAndRise()
        }

        root.backend.guiReady()
    }

    function setColorScheme() {
        if (root.backend.colorSchemeName == "light") ProtonStyle.currentStyle = ProtonStyle.lightStyle
        if (root.backend.colorSchemeName == "dark") ProtonStyle.currentStyle = ProtonStyle.darkStyle
    }
}
