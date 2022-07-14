
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

import QtQuick
import QtQuick.Layouts
import QtQuick.Controls
import QtQuick.Controls.impl

import Proton
import Notifications

Item {
    id: root

    property var notifications
    property ColorScheme colorScheme

    property int notificationWhitelist: NotificationFilter.FilterConsts.All
    property int notificationBlacklist: NotificationFilter.FilterConsts.None

    readonly property Notification activeNotification: notificationFilter.topmost

    implicitHeight: children[0].implicitHeight + children[0].anchors.topMargin + children[0].anchors.bottomMargin
    implicitWidth: children[0].implicitWidth + children[0].anchors.leftMargin + children[0].anchors.rightMargin

    NotificationFilter {
        id: notificationFilter

        source: root.notifications ? root.notifications.all : undefined
        whitelist: root.notificationWhitelist
        blacklist: root.notificationBlacklist

        onTopmostChanged: {
            if (!topmost) {
                image.source = "/qml/icons/ic-connected.svg"
                image.color = root.colorScheme.signal_success
                label.text = qsTr("Connected")
                label.color = root.colorScheme.signal_success
                return;
            }

            image.source = topmost.icon
            label.text = topmost.brief

            switch (topmost.type) {
                case Notification.NotificationType.Danger:
                image.color = root.colorScheme.signal_danger
                label.color = root.colorScheme.signal_danger
                break;
                case Notification.NotificationType.Warning:
                image.color = root.colorScheme.signal_warning
                label.color = root.colorScheme.signal_warning
                break;
                case Notification.NotificationType.Success:
                image.color = root.colorScheme.signal_success
                label.color = root.colorScheme.signal_success
                break;
                case Notification.NotificationType.Info:
                image.color = root.colorScheme.signal_info
                label.color = root.colorScheme.signal_info
                break;
            }
        }
    }

    RowLayout {
        anchors.fill: parent
        spacing: 8

        ColorImage {
            id: image
            width: 16
            height: 16
            sourceSize.width: width
            sourceSize.height: height
            source: "/qml/icons/ic-connected.svg"
            color: root.colorScheme.signal_success
        }

        Label {
            colorScheme: root.colorScheme
            id: label

            Layout.fillHeight: true
            Layout.fillWidth: true

            wrapMode: Text.WordWrap

            horizontalAlignment: Text.AlignLeft
            verticalAlignment: Text.AlignVCenter

            text: qsTr("Connected")
            color: root.colorScheme.signal_success
        }
    }
}
