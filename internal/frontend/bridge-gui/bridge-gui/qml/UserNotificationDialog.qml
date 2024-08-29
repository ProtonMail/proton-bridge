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

Dialog {
    id: root

    property var notification
    property bool isUserNotification: true
    padding: 40

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
                    case Notification.NotificationType.UserNotification:
                        return "/qml/icons/ic-notification-bell.svg"
                }
            }
            sourceSize.height: 64
            sourceSize.width: 64
            visible: source != ""
        }
        // Title Label
        Label {
            Layout.alignment: Qt.AlignHCenter
            Layout.bottomMargin: 4
            Layout.preferredWidth: 320
            colorScheme: root.colorScheme
            horizontalAlignment: Text.AlignHCenter
            text: root.notification.title
            wrapMode: Text.WordWrap
            type: Label.LabelType.Title
        }
        // Username or primary email
        Label {
            Layout.alignment: Qt.AlignHCenter
            Layout.bottomMargin: 24
            Layout.preferredWidth: 320
            colorScheme: root.colorScheme
            horizontalAlignment: Text.AlignHCenter
            text: root.notification.username
            wrapMode: Text.WordWrap
            visible: root.notification.username.length > 0
            type: Label.LabelType.Caption
        }
        // Subtitle
        Label {
            Layout.alignment: Qt.AlignHCenter
            Layout.bottomMargin: 24
            Layout.fillWidth: true
            Layout.preferredWidth: 320
            colorScheme: root.colorScheme
            horizontalAlignment: Text.AlignHCenter
            text: root.notification.subtitle
            wrapMode: Text.WordWrap
            visible: root.notification.subtitle.length > 0
            type: Label.LabelType.Lead
            color: root.colorScheme.text_weak
        }
        Label {
            Layout.bottomMargin: 24
            Layout.fillWidth: true
            Layout.preferredWidth: 320
            colorScheme: root.colorScheme
            horizontalAlignment: Text.AlignHCenter
            text: root.notification.description
            type: Label.LabelType.Body
            wrapMode: Text.WordWrap

            onLinkActivated: function (link) {
                Backend.openExternalLink(link);
            }
        }


        ColumnLayout {
            spacing: 40

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
