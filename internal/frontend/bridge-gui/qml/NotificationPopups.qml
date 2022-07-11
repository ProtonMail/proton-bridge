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
import QtQuick 2.12
import QtQuick.Layouts 1.12
import QtQuick.Controls 2.12

import Proton 4.0
import Notifications 1.0

Item {
    id: root
    property var backend

    property ColorScheme colorScheme
    property var notifications
    property var mainWindow

    property int notificationWhitelist: NotificationFilter.FilterConsts.All
    property int notificationBlacklist: NotificationFilter.FilterConsts.None

    NotificationFilter {
        id: bannerNotificationFilter

        source: root.notifications.all
        blacklist: Notifications.Group.Dialogs
    }

    Banner {
        colorScheme: root.colorScheme
        notification: bannerNotificationFilter.topmost
        mainWindow: root.mainWindow
    }

    NotificationDialog {
        colorScheme: root.colorScheme
        notification: root.notifications.updateManualReady

        Switch {
            id:autoUpdate
            colorScheme: root.colorScheme
            text: qsTr("Update automatically in the future")
            checked: root.backend.isAutomaticUpdateOn
            onClicked: root.backend.toggleAutomaticUpdate(autoUpdate.checked)
        }
    }

    NotificationDialog {
        colorScheme: root.colorScheme
        notification: root.notifications.updateForce
    }

    NotificationDialog {
        colorScheme: root.colorScheme
        notification: root.notifications.updateForceError
    }

    NotificationDialog {
        colorScheme: root.colorScheme
        notification: root.notifications.enableBeta
    }

    NotificationDialog {
        colorScheme: root.colorScheme
        notification: root.notifications.cacheUnavailable
    }

    NotificationDialog {
        colorScheme: root.colorScheme
        notification: root.notifications.cacheCantMove
    }

    NotificationDialog {
        colorScheme: root.colorScheme
        notification: root.notifications.diskFull
    }

    NotificationDialog {
        colorScheme: root.colorScheme
        notification: root.notifications.enableSplitMode
    }

    NotificationDialog {
        colorScheme: root.colorScheme
        notification: root.notifications.disableLocalCache
    }

    NotificationDialog {
        colorScheme: root.colorScheme
        notification: root.notifications.enableLocalCache
    }

    NotificationDialog {
        colorScheme: root.colorScheme
        notification: root.notifications.resetBridge
    }

    NotificationDialog {
        colorScheme: root.colorScheme
        notification: root.notifications.changeAllMailVisibility
    }

    NotificationDialog {
        colorScheme: root.colorScheme
        notification: root.notifications.deleteAccount
    }

    NotificationDialog {
        colorScheme: root.colorScheme
        notification: root.notifications.noKeychain
    }

    NotificationDialog {
        colorScheme: root.colorScheme
        notification: root.notifications.rebuildKeychain
    }
}
