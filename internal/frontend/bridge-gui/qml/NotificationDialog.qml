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

import QtQml
import QtQuick
import QtQuick.Layouts
import QtQuick.Controls

import Proton
import Notifications

Dialog {
    id: root

    property var notification

    shouldShow: notification && notification.active && !notification.dismissed
    modal: true

    default property alias data: additionalChildrenContainer.children

    ColumnLayout {
        spacing: 0

        Image {
            Layout.alignment: Qt.AlignHCenter

            sourceSize.width: 64
            sourceSize.height: 64

            Layout.preferredHeight: 64
            Layout.preferredWidth: 64

            Layout.bottomMargin: 16

            visible: source != ""

            source: {
                if (!root.notification) {
                    return ""
                }

                switch (root.notification.type) {
                    case Notification.NotificationType.Info:
                    return "/qml/icons/ic-info.svg"
                    case Notification.NotificationType.Success:
                    return "/qml/icons/ic-success.svg"
                    case Notification.NotificationType.Warning:
                    case Notification.NotificationType.Danger:
                    return "/qml/icons/ic-alert.svg"
                }
            }
        }

        Label {
            Layout.alignment: Qt.AlignHCenter
            Layout.bottomMargin: 8
            colorScheme: root.colorScheme
            text: root.notification.title
            type: Label.LabelType.Title
        }

        Label {
            Layout.fillWidth: true
            Layout.preferredWidth: 240
            Layout.bottomMargin: 16

            colorScheme: root.colorScheme
            text: root.notification.description
            wrapMode: Text.WordWrap
            horizontalAlignment: Text.AlignHCenter
            type: Label.LabelType.Body
            onLinkActivated: function(link) { Qt.openUrlExternally(link) }
        }

        Item {
            id: additionalChildrenContainer

            Layout.fillWidth: true
            Layout.bottomMargin: 16

            visible: children.length > 0

            implicitHeight: additionalChildrenContainer.childrenRect.height
            implicitWidth: additionalChildrenContainer.childrenRect.width
        }

        ColumnLayout {
            spacing: 8
            Repeater {
                model: root.notification.action
                delegate: Button {
                    Layout.fillWidth: true

                    colorScheme: root.colorScheme
                    action: modelData

                    secondary: index > 0

                    loading: modelData.loading
                }
            }
        }
    }
}
