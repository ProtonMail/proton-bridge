// Copyright (c) 2023 Proton AG
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

import QtQml
import QtQuick
import QtQuick.Window
import Qt.labs.platform

import Proton
import Notifications

QtObject {
    id: root

    function bound(num, lowerLimit, upperLimit) {
        return Math.max(lowerLimit, Math.min(upperLimit, num))
    }

    property var title: Backend.appname

    property Notifications _notifications: Notifications {
        id: notifications
        frontendMain: mainWindow
    }

    property NotificationFilter _trayNotificationFilter: NotificationFilter {
    id: trayNotificationFilter
    source: root._notifications ? root._notifications.all : undefined
    onTopmostChanged: {
        if (topmost) {
            switch (topmost.type) {
                    case Notification.NotificationType.Danger:
                        Backend.setErrorTrayIcon(topmost.brief, topmost.icon)
                        return
                    case Notification.NotificationType.Warning:
                        Backend.setWarnTrayIcon(topmost.brief, topmost.icon)
                        return
                    case Notification.NotificationType.Info:
                        Backend.setUpdateTrayIcon(topmost.brief, topmost.icon)
                        return
                }
            }
            Backend.setNormalTrayIcon()
        }
    }


    property MainWindow _mainWindow: MainWindow {
        id: mainWindow
        visible: false

        title: root.title
        notifications: root._notifications

        onVisibleChanged: {
            Backend.dockIconVisible = visible
        }

        Connections {
            target: Backend
            function onDiskCacheUnavailable() {
                mainWindow.showAndRise()
            }
            function onColorSchemeNameChanged(scheme) { root.setColorScheme() }

            function onHideMainWindow() {
                mainWindow.hide();
            }
        }
    }


    Component.onCompleted: {
        if (!Backend) {
            console.log("Backend not loaded")
        }

        root.setColorScheme()


        if (!Backend.users) {
            console.log("users not loaded")
        }

        var c = Backend.users.count
        var u = Backend.users.get(0)
        // DEBUG
        if (c !== 0) {
            console.log("users non zero", c)
            console.log("first user", u )
        }

        if (c === 0) {
            mainWindow.showAndRise()
        }

        if (u) {
            if (c === 1 && (u.state === EUserState.SignedOut)) {
                mainWindow.showAndRise()
            }
        }

        Backend.guiReady()

         if (Backend.showOnStartup || Backend.showSplashScreen) {
            mainWindow.showAndRise()
        }

    }

    function setColorScheme() {
        if (Backend.colorSchemeName === "light") ProtonStyle.currentStyle = ProtonStyle.lightStyle
        if (Backend.colorSchemeName === "dark") ProtonStyle.currentStyle = ProtonStyle.darkStyle
    }
}
