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
import QtQuick
import QtQuick.Layouts
import QtQuick.Controls
import QtQuick.Controls.impl
import Proton
import Notifications

Item {
    id: root

    readonly property Notification activeNotification: notificationFilter.topmost
    property ColorScheme colorScheme
    property int notificationBlacklist: NotificationFilter.FilterConsts.None
    property int notificationWhitelist: NotificationFilter.FilterConsts.All
    property var notifications

    implicitHeight: children[0].implicitHeight + children[0].anchors.topMargin + children[0].anchors.bottomMargin
    implicitWidth: children[0].implicitWidth + children[0].anchors.leftMargin + children[0].anchors.rightMargin

    NotificationFilter {
        id: notificationFilter
        blacklist: root.notificationBlacklist
        source: root.notifications ? root.notifications.all : undefined
        whitelist: root.notificationWhitelist

        onTopmostChanged: {
            if (!topmost) {
                image.source = "/qml/icons/ic-connected.svg";
                image.color = root.colorScheme.signal_success;
                label.text = qsTr("Connected");
                label.color = root.colorScheme.signal_success;
                return;
            }
            image.source = topmost.icon;
            label.text = topmost.brief;
            switch (topmost.type) {
            case Notification.NotificationType.Danger:
                image.color = root.colorScheme.signal_danger;
                label.color = root.colorScheme.signal_danger;
                break;
            case Notification.NotificationType.Warning:
                image.color = root.colorScheme.signal_warning;
                label.color = root.colorScheme.signal_warning;
                break;
            case Notification.NotificationType.Success:
                image.color = root.colorScheme.signal_success;
                label.color = root.colorScheme.signal_success;
                break;
            case Notification.NotificationType.Info:
                image.color = root.colorScheme.signal_info;
                label.color = root.colorScheme.signal_info;
                break;
            }
        }
    }
    RowLayout {
        anchors.fill: parent
        spacing: 8

        ColorImage {
            id: image
            color: root.colorScheme.signal_success
            height: 16
            source: "/qml/icons/ic-connected.svg"
            sourceSize.height: height
            sourceSize.width: width
            width: 16
        }
        Label {
            id: label
            Layout.fillHeight: true
            Layout.fillWidth: true
            color: root.colorScheme.signal_success
            colorScheme: root.colorScheme
            horizontalAlignment: Text.AlignLeft
            text: qsTr("Connected")
            verticalAlignment: Text.AlignVCenter
            wrapMode: Text.WordWrap
        }
    }
}
