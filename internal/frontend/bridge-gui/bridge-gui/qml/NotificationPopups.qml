// Copyright (c) 2024 Proton AG
// This file is part of Proton Mail Bridge.
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.
import QtQml
import QtQuick
import QtQuick.Layouts
import QtQuick.Controls
import Proton
import Notifications

Item {
    id: root

    property ColorScheme colorScheme
    property var mainWindow
    property int notificationBlacklist: NotificationFilter.FilterConsts.None
    property int notificationWhitelist: NotificationFilter.FilterConsts.All
    property var notifications

    NotificationFilter {
        id: bannerNotificationFilter
        blacklist: Notifications.Group.Dialogs
        source: root.notifications.all
    }
    Banner {
        colorScheme: root.colorScheme
        mainWindow: root.mainWindow
        notification: bannerNotificationFilter.topmost
    }
    NotificationDialog {
        colorScheme: root.colorScheme
        notification: root.notifications.updateManualReady

        Switch {
            id: autoUpdate
            checked: Backend.isAutomaticUpdateOn
            colorScheme: root.colorScheme
            text: qsTr("Update automatically in the future")

            onClicked: Backend.toggleAutomaticUpdate(autoUpdate.checked)
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
        notification: root.notifications.cacheCantMove
    }
    NotificationDialog {
        colorScheme: root.colorScheme
        notification: root.notifications.enableSplitMode
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
    NotificationDialog {
        colorScheme: root.colorScheme
        notification: root.notifications.apiCertIssue
    }
    NotificationDialog {
        colorScheme: root.colorScheme
        notification: root.notifications.userBadEvent
    }
    NotificationDialog {
        colorScheme: root.colorScheme
        notification: root.notifications.genericError
    }
    NotificationDialog {
        colorScheme: root.colorScheme
        notification: root.notifications.genericQuestion
    }
    NotificationDialog {
        colorScheme: root.colorScheme
        notification: root.notifications.repairBridge
    }
    UserNotificationDialog {
        colorScheme: root.colorScheme
        notification: root.notifications.userNotification
    }
}
