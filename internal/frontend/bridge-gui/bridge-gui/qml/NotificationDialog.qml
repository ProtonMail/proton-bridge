// Copyright (c) 2023 Proton AG
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

Dialog {
    id: root

    default property alias data: additionalChildrenContainer.children
    property var notification

    modal: true
    shouldShow: notification && notification.active && !notification.dismissed

    ColumnLayout {
        spacing: 0

        Image {
            Layout.alignment: Qt.AlignHCenter
            Layout.bottomMargin: 16
            Layout.preferredHeight: 64
            Layout.preferredWidth: 64
            source: {
                if (!root.notification) {
                    return "";
                }
                switch (root.notification.type) {
                case Notification.NotificationType.Info:
                    return "/qml/icons/ic-info.svg";
                case Notification.NotificationType.Success:
                    return "/qml/icons/ic-success.svg";
                case Notification.NotificationType.Warning:
                case Notification.NotificationType.Danger:
                    return "/qml/icons/ic-alert.svg";
                }
            }
            sourceSize.height: 64
            sourceSize.width: 64
            visible: source != ""
        }
        Label {
            Layout.alignment: Qt.AlignHCenter
            Layout.bottomMargin: 8
            colorScheme: root.colorScheme
            horizontalAlignment: Text.AlignHCenter
            text: root.notification.title
            type: Label.LabelType.Title
        }
        Label {
            Layout.bottomMargin: 16
            Layout.fillWidth: true
            Layout.preferredWidth: 240
            colorScheme: root.colorScheme
            horizontalAlignment: Text.AlignHCenter
            text: root.notification.description
            type: Label.LabelType.Body
            wrapMode: Text.WordWrap

            onLinkActivated: function (link) {
                Backend.openExternalLink(link);
            }
        }
        Item {
            id: additionalChildrenContainer
            Layout.bottomMargin: 16
            Layout.fillWidth: true
            implicitHeight: additionalChildrenContainer.childrenRect.height
            implicitWidth: additionalChildrenContainer.childrenRect.width
            visible: children.length > 0
        }
        LinkLabel {
            Layout.alignment: Qt.AlignHCenter
            Layout.bottomMargin: 32
            colorScheme: root.colorScheme
            external: true
            link: notification.linkUrl
            text: notification.linkText
            visible: notification.linkUrl.length > 0

        }

        ColumnLayout {
            spacing: 8

            Repeater {
                model: root.notification.action

                delegate: Button {
                    Layout.fillWidth: true
                    action: modelData
                    colorScheme: root.colorScheme
                    loading: modelData.loading
                    secondary: index > 0
                }
            }
        }
    }
}
