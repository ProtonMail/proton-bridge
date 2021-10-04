
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
import QtQuick.Layouts 1.12
import QtQuick.Controls 2.12
import QtQuick.Controls.impl 2.12

import Proton 4.0
import Notifications 1.0

Item {
    id: root

    property var backend
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
                image.source = "./icons/ic-connected.svg"
                image.color = root.colorScheme.signal_success
                label.text = qsTr("Connected")
                label.color = root.colorScheme.signal_success
                return;
            }

            image.source = topmost.icon
            label.text = topmost.text

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
            Layout.fillHeight: true
            sourceSize.height: height
        }

        Label {
            colorScheme: root.colorScheme
            id: label

            Layout.fillHeight: true
            Layout.fillWidth: true

            wrapMode: Text.WordWrap

            horizontalAlignment: Text.AlignLeft
            verticalAlignment: Text.AlignVCenter
        }
    }

    state: "Connected"
    states: [
        State {
            name: "Connected"
            PropertyChanges {
                target: image
                source: "./icons/ic-connected.svg"
                color: ProtonStyle.currentStyle.signal_success
            }
            PropertyChanges {
                target: label
                text: qsTr("Connected")
                color: ProtonStyle.currentStyle.signal_success
            }
        },
        State {
            name: "No connection"
            PropertyChanges {
                target: image
                source: "./icons/ic-no-connection.svg"
                color: ProtonStyle.currentStyle.signal_danger
            }
            PropertyChanges {
                target: label
                text: qsTr("No connection")
                color: ProtonStyle.currentStyle.signal_danger
            }
        },
        State {
            name: "Outdated"
            PropertyChanges {
                target: image
                source: "./icons/ic-exclamation-circle-filled.svg"
                color: ProtonStyle.currentStyle.signal_danger
            }
            PropertyChanges {
                target: label
                text: qsTr("Bridge is outdated")
                color: ProtonStyle.currentStyle.signal_danger
            }
        },
        State {
            name: "Account changed"
            PropertyChanges {
                target: image
                source: "./icons/ic-exclamation-circle-filled.svg"
                color: ProtonStyle.currentStyle.signal_danger
            }
            PropertyChanges {
                target: label
                text: qsTr("The address list for your account has changed")
                color: ProtonStyle.currentStyle.signal_danger
            }
        },
        State {
            name: "Auto update failed"
            PropertyChanges {
                target: image
                source: "./icons/ic-info-circle-filled.svg"
                color: ProtonStyle.currentStyle.signal_info
            }
            PropertyChanges {
                target: label
                text: qsTr("Bridge couldn’t update automatically")
                color: ProtonStyle.currentStyle.signal_info
            }
        },
        State {
            name: "Update ready"
            PropertyChanges {
                target: image
                source: "./icons/ic-info-circle-filled.svg"
                color: ProtonStyle.currentStyle.signal_info
            }
            PropertyChanges {
                target: label
                text: qsTr("Bridge update is ready")
                color: ProtonStyle.currentStyle.signal_info
            }
        }
    ]
}
