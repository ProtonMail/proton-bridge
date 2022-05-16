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

import QtQuick 2.13
import QtQuick.Layouts 1.12
import QtQuick.Controls 2.12
import QtQuick.Controls.impl 2.12

import Proton 4.0
import Notifications 1.0

Popup {
    id: root

    property ColorScheme colorScheme
    property Notification notification
    property var mainWindow

    topMargin:  37
    leftMargin:  (mainWindow.width - root.implicitWidth)/2

    implicitHeight: contentLayout.implicitHeight + contentLayout.anchors.topMargin + contentLayout.anchors.bottomMargin
    implicitWidth: 600 // contentLayout.implicitWidth + contentLayout.anchors.leftMargin + contentLayout.anchors.rightMargin

    popupType: ApplicationWindow.PopupType.Banner

    shouldShow: notification ? (notification.active && !notification.dismissed) : false

    modal: false

    Action {
        id: defaultDismissAction

        text: qsTr("OK")
        onTriggered: {
            if (!root.notification) {
                return
            }

            root.notification.dismissed = true
        }
    }

    RowLayout {
        id: contentLayout
        anchors.fill: parent
        spacing: 0

        Item {
            Layout.fillHeight: true
            Layout.fillWidth: true

            clip: true
            implicitHeight: children[1].implicitHeight + children[1].anchors.topMargin + children[1].anchors.bottomMargin
            implicitWidth: children[1].implicitWidth + children[1].anchors.leftMargin + children[1].anchors.rightMargin

            Rectangle {
                anchors.top: parent.top
                anchors.bottom: parent.bottom
                anchors.left: parent.left
                width: parent.width + 10
                radius: ProtonStyle.banner_radius
                color: {
                    if (!root.notification) {
                        return "transparent"
                    }

                    switch (root.notification.type) {
                        case Notification.NotificationType.Info:
                        return root.colorScheme.signal_info
                        case Notification.NotificationType.Success:
                        return root.colorScheme.signal_success
                        case Notification.NotificationType.Warning:
                        return root.colorScheme.signal_warning
                        case Notification.NotificationType.Danger:
                        return root.colorScheme.signal_danger
                    }
                }
            }

            RowLayout {
                anchors.fill: parent

                anchors.topMargin: 14
                anchors.bottomMargin: 14
                anchors.leftMargin: 16

                spacing: 8

                ColorImage {
                    color: root.colorScheme.text_invert
                    width: 24
                    height: 24

                    sourceSize.width: 24
                    sourceSize.height: 24

                    Layout.preferredHeight: 24
                    Layout.preferredWidth: 24

                    source: {
                        if (!root.notification) {
                            return ""
                        }

                        switch (root.notification.type) {
                            case Notification.NotificationType.Info:
                            return "./icons/ic-info-circle-filled.svg"
                            case Notification.NotificationType.Success:
                            return "./icons/ic-info-circle-filled.svg"
                            case Notification.NotificationType.Warning:
                            return "./icons/ic-exclamation-circle-filled.svg"
                            case Notification.NotificationType.Danger:
                            return "./icons/ic-exclamation-circle-filled.svg"
                        }
                    }
                }

                Label {
                    colorScheme: root.colorScheme
                    Layout.fillWidth: true
                    Layout.alignment: Qt.AlignVCenter
                    Layout.leftMargin: 16

                    color: root.colorScheme.text_invert
                    text: root.notification ? root.notification.description : ""

                    wrapMode: Text.WordWrap
                }
            }
        }

        Rectangle {
            Layout.fillHeight: true
            width: 1
            color: {
                if (!root.notification) {
                    return "transparent"
                }

                switch (root.notification.type) {
                    case Notification.NotificationType.Info:
                    return root.colorScheme.signal_info_active
                    case Notification.NotificationType.Success:
                    return root.colorScheme.signal_success_active
                    case Notification.NotificationType.Warning:
                    return root.colorScheme.signal_warning_active
                    case Notification.NotificationType.Danger:
                    return root.colorScheme.signal_danger_active
                }
            }
        }

        Button {
            colorScheme: root.colorScheme
            Layout.fillHeight: true

            id: actionButton

            action: (root.notification && root.notification.action.length > 0) ? root.notification.action[0] : defaultDismissAction

            background: Item {
                clip: true
                Rectangle {
                    anchors.top: parent.top
                    anchors.bottom: parent.bottom
                    anchors.right: parent.right
                    width: parent.width + 10
                    radius: ProtonStyle.banner_radius
                    color: {
                        if (!root.notification) {
                            return "transparent"
                        }

                        var norm
                        var hover
                        var active

                        switch (root.notification.type) {
                            case Notification.NotificationType.Info:
                            norm = root.colorScheme.signal_info
                            hover = root.colorScheme.signal_info_hover
                            active = root.colorScheme.signal_info_active
                            break;
                            case Notification.NotificationType.Success:
                            norm = root.colorScheme.signal_success
                            hover = root.colorScheme.signal_success_hover
                            active = root.colorScheme.signal_success_active
                            break;
                            case Notification.NotificationType.Warning:
                            norm = root.colorScheme.signal_warning
                            hover = root.colorScheme.signal_warning_hover
                            active = root.colorScheme.signal_warning_active
                            break;
                            case Notification.NotificationType.Danger:
                            norm = root.colorScheme.signal_danger
                            hover = root.colorScheme.signal_danger_hover
                            active = root.colorScheme.signal_danger_active
                            break;
                        }

                        if (actionButton.down) {
                            return active
                        }

                        if (actionButton.enabled && (actionButton.highlighted || actionButton.hovered || actionButton.checked)) {
                            return hover
                        }

                        if (actionButton.loading) {
                            return hover
                        }

                        return norm
                    }
                }
            }
        }
    }
}
