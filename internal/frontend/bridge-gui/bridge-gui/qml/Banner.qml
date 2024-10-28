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

Popup {
    id: root

    property ColorScheme colorScheme
    property var mainWindow
    property Notification notification

    implicitHeight: contentLayout.implicitHeight + contentLayout.anchors.topMargin + contentLayout.anchors.bottomMargin
    implicitWidth: 600 // contentLayout.implicitWidth + contentLayout.anchors.leftMargin + contentLayout.anchors.rightMargin
    leftMargin: (mainWindow.width - root.implicitWidth) / 2
    modal: false
    popupPriority: ApplicationWindow.PopupPriority.Banner
    shouldShow: notification ? (notification.active && !notification.dismissed) : false
    topMargin: 37

    Action {
        id: defaultDismissAction
        text: qsTr("OK")

        onTriggered: {
            if (!root.notification) {
                return;
            }
            root.notification.dismissed = true;
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
                anchors.bottom: parent.bottom
                anchors.left: parent.left
                anchors.top: parent.top
                color: {
                    if (!root.notification) {
                        return "transparent";
                    }
                    switch (root.notification.type) {
                    case Notification.NotificationType.Info:
                        return root.colorScheme.signal_info;
                    case Notification.NotificationType.Success:
                        return root.colorScheme.signal_success;
                    case Notification.NotificationType.Warning:
                        return root.colorScheme.signal_warning;
                    case Notification.NotificationType.Danger:
                        return root.colorScheme.signal_danger;
                    }
                }
                radius: ProtonStyle.banner_radius
                width: parent.width + 10
            }
            RowLayout {
                anchors.bottomMargin: 14
                anchors.fill: parent
                anchors.leftMargin: 16
                anchors.topMargin: 14
                spacing: 8


                ColorImage {
                    Layout.preferredHeight: 24
                    Layout.preferredWidth: 24
                    color: root.colorScheme.text_invert
                    height: 24
                    source: {
                        if (!root.notification) {
                            return "";
                        }
                        switch (root.notification.type) {
                        case Notification.NotificationType.Info:
                            return "/qml/icons/ic-info-circle-filled.svg";
                        case Notification.NotificationType.Success:
                            return "/qml/icons/ic-info-circle-filled.svg";
                        case Notification.NotificationType.Warning:
                            return "/qml/icons/ic-exclamation-circle-filled.svg";
                        case Notification.NotificationType.Danger:
                            return "/qml/icons/ic-exclamation-circle-filled.svg";
                        }
                    }
                    sourceSize.height: 24
                    sourceSize.width: 24
                    width: 24
                }
                ColumnLayout {
                    Layout.alignment: Qt.AlignTop
                    Layout.fillWidth: true
                    Layout.leftMargin: 16
                    Label {
                        id: messageLabel
                        Layout.alignment: Qt.AlignTop
                        Layout.fillWidth: true
                        color: root.colorScheme.text_invert
                        colorScheme: root.colorScheme
                        text: root.notification ? root.notification.description : ""
                        wrapMode: Text.WordWrap
                    }
                    LinkLabel {
                        Layout.alignment: Qt.AlignTop | Qt.AlignLeft
                        Layout.fillWidth: true
                        colorScheme: root.colorScheme
                        color: messageLabel.color
                        external: true
                        link: root.notification ? root.notification.linkUrl : ""
                        text: root.notification ? root.notification.linkText : ""
                        visible: root.notification && root.notification.linkUrl.length > 0
                    }
                }
            }
        }
        Rectangle {
            Layout.fillHeight: true
            color: {
                if (!root.notification) {
                    return "transparent";
                }
                switch (root.notification.type) {
                case Notification.NotificationType.Info:
                    return root.colorScheme.signal_info_active;
                case Notification.NotificationType.Success:
                    return root.colorScheme.signal_success_active;
                case Notification.NotificationType.Warning:
                    return root.colorScheme.signal_warning_active;
                case Notification.NotificationType.Danger:
                    return root.colorScheme.signal_danger_active;
                }
            }
            width: 1
        }
        Button {
            id: actionButton
            Layout.fillHeight: true
            action: (root.notification && root.notification.action.length > 0) ? root.notification.action[0] : defaultDismissAction
            colorScheme: root.colorScheme

            background: Item {
                clip: true

                Rectangle {
                    anchors.bottom: parent.bottom
                    anchors.right: parent.right
                    anchors.top: parent.top
                    color: {
                        if (!root.notification) {
                            return "transparent";
                        }
                        let norm;
                        let hover;
                        let active;
                        switch (root.notification.type) {
                        case Notification.NotificationType.Info:
                            norm = root.colorScheme.signal_info;
                            hover = root.colorScheme.signal_info_hover;
                            active = root.colorScheme.signal_info_active;
                            break;
                        case Notification.NotificationType.Success:
                            norm = root.colorScheme.signal_success;
                            hover = root.colorScheme.signal_success_hover;
                            active = root.colorScheme.signal_success_active;
                            break;
                        case Notification.NotificationType.Warning:
                            norm = root.colorScheme.signal_warning;
                            hover = root.colorScheme.signal_warning_hover;
                            active = root.colorScheme.signal_warning_active;
                            break;
                        case Notification.NotificationType.Danger:
                            norm = root.colorScheme.signal_danger;
                            hover = root.colorScheme.signal_danger_hover;
                            active = root.colorScheme.signal_danger_active;
                            break;
                        }
                        if (actionButton.down) {
                            return active;
                        }
                        if (actionButton.enabled && (actionButton.highlighted || actionButton.hovered || actionButton.checked)) {
                            return hover;
                        }
                        if (actionButton.loading) {
                            return hover;
                        }
                        return norm;
                    }
                    radius: ProtonStyle.banner_radius
                    width: parent.width + 10
                }
            }
        }
    }
}
