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
import QtQuick 2.12
import QtQuick.Layouts 1.12
import QtQuick.Controls 2.12

import Proton 4.0
import Notifications 1.0

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
                    // TODO: Add info icon?
                    return ""
                case Notification.NotificationType.Success:
                    return "./icons/ic-success.svg"
                case Notification.NotificationType.Warning:
                case Notification.NotificationType.Danger:
                    return "./icons/ic-alert.svg"
                }
            }
        }

        Label {
            Layout.alignment: Qt.AlignHCenter
            Layout.bottomMargin: 8
            colorScheme: root.colorScheme
            text: root.notification.text
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
                }
            }
        }
    }
}
