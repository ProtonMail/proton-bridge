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
import QtQuick.Window
import QtQuick.Layouts
import QtQuick.Controls

import Proton
import Notifications

Window {
    id: root

    height: contentLayout.implicitHeight
    width: contentLayout.implicitWidth

    flags: (Qt.platform.os === "linux" ? Qt.Tool : 0) | Qt.FramelessWindowHint | Qt.NoDropShadowWindowHint | Qt.WindowStaysOnTopHint | Qt.WA_TranslucentBackground
    color: "transparent"

    property ColorScheme colorScheme: ProtonStyle.currentStyle

    property var notifications

    signal showMainWindow()
    signal showHelp()
    signal showSettings()
    signal showSignIn(string username)
    signal quit()

    onVisibleChanged: {
        if (visible) { // GODT-1479 restore the hover-able status that may have been disabled when clicking on the 'Open Bridge' button.
            openBridgeButton.hoverEnabled = true
            openBridgeButton.focus = false
        } else {
            menu.close()
        }
    }

    ColumnLayout {
        id: contentLayout

        Layout.minimumHeight: 201

        anchors.fill: parent
        spacing: 0

        ColumnLayout {
            Layout.minimumWidth: 448
            Layout.fillWidth: true
            spacing: 0

            Item {
                implicitHeight: 12
                Layout.fillWidth: true
                clip: true
                Rectangle {
                    anchors.top: parent.top
                    anchors.left: parent.left
                    anchors.right: parent.right
                    height: parent.height * 2
                    radius: ProtonStyle.dialog_radius

                    color: {
                        if (!statusItem.activeNotification) {
                            return root.colorScheme.signal_success
                        }

                        switch (statusItem.activeNotification.type) {
                            case Notification.NotificationType.Danger:
                            return root.colorScheme.signal_danger
                            case Notification.NotificationType.Warning:
                            return root.colorScheme.signal_warning
                            case Notification.NotificationType.Success:
                            return root.colorScheme.signal_success
                            case Notification.NotificationType.Info:
                            return root.colorScheme.signal_info
                        }
                    }
                }
            }

            Rectangle {
                Layout.fillWidth: true

                implicitHeight: children[0].implicitHeight + children[0].anchors.topMargin + children[0].anchors.bottomMargin
                implicitWidth: children[0].implicitWidth + children[0].anchors.leftMargin + children[0].anchors.rightMargin

                color: colorScheme.background_norm

                RowLayout {
                    anchors.fill: parent

                    anchors.topMargin: 8
                    anchors.bottomMargin: 8
                    anchors.leftMargin: 24
                    anchors.rightMargin: 24

                    spacing: 8

                    Status {
                        id: statusItem

                        Layout.fillWidth: true

                        Layout.topMargin: 12
                        Layout.bottomMargin: 12

                        colorScheme: root.colorScheme
                        notifications: root.notifications

                        notificationWhitelist: Notifications.Group.Connection | Notifications.Group.Update | Notifications.Group.Configuration
                    }

                    Button {
                        colorScheme: root.colorScheme
                        secondary: true

                        Layout.topMargin: 12
                        Layout.bottomMargin: 12

                        visible: statusItem.activeNotification && statusItem.activeNotification.action.length > 0
                        action: statusItem.activeNotification && statusItem.activeNotification.action.length > 0 ? statusItem.activeNotification.action[0] : null
                    }
                }
            }

            Rectangle {
                Layout.fillWidth: true
                height: 1
                color: root.colorScheme.background_norm

                Rectangle {
                    anchors.fill: parent
                    anchors.leftMargin: 24
                    anchors.rightMargin: 24
                    color: root.colorScheme.border_norm
                }
            }
        }

        Rectangle {
            Layout.fillWidth: true
            Layout.fillHeight: true

            Layout.maximumHeight: accountListView.count ?
            accountListView.contentHeight / accountListView.count * 3 + accountListView.anchors.topMargin + accountListView.anchors.bottomMargin :
            Number.POSITIVE_INFINITY

            color: root.colorScheme.background_norm
            clip: true

            implicitHeight: children[0].contentHeight + children[0].anchors.topMargin + children[0].anchors.bottomMargin
            implicitWidth: children[0].contentWidth + children[0].anchors.leftMargin + children[0].anchors.rightMargin

            ListView {
                id: accountListView

                model: Backend.users
                anchors.fill: parent

                anchors.topMargin: 8
                anchors.bottomMargin: 8
                anchors.leftMargin: 24
                anchors.rightMargin: 24

                interactive: contentHeight > parent.height
                snapMode: ListView.SnapToItem
                boundsBehavior: Flickable.StopAtBounds

                spacing: 4

                delegate: Item {
                    id: viewItem
                    width: ListView.view.width

                    implicitHeight: children[0].implicitHeight
                    implicitWidth: children[0].implicitWidth

                    property var user: Backend.users.get(index)

                    RowLayout {
                        spacing: 0
                        anchors.fill: parent

                        AccountDelegate {
                            Layout.fillWidth: true

                            Layout.topMargin: 12
                            Layout.bottomMargin: 12

                            user: viewItem.user
                            colorScheme: root.colorScheme
                        }

                        Button {
                            Layout.topMargin: 12
                            Layout.bottomMargin: 12

                            colorScheme: root.colorScheme
                            visible: viewItem.user ? !viewItem.user.loggedIn : false
                            text: qsTr("Sign in")
                            onClicked: {
                                root.showSignIn(viewItem.username)
                                root.close()
                            }
                        }
                    }
                }
            }
        }

        Item {
            Layout.fillWidth: true

            implicitHeight: children[1].implicitHeight + children[1].anchors.topMargin + children[1].anchors.bottomMargin
            implicitWidth: children[1].implicitWidth + children[1].anchors.leftMargin + children[1].anchors.rightMargin

            // background:
            clip: true
            Rectangle {
                anchors.bottom: parent.bottom
                anchors.left: parent.left
                anchors.right: parent.right
                height: parent.height * 2
                radius: ProtonStyle.dialog_radius

                color: root.colorScheme.background_weak
            }

            RowLayout {
                anchors.fill: parent
                anchors.margins: 8
                spacing: 0

                Button {
                    id: openBridgeButton
                    colorScheme: root.colorScheme
                    secondary: true
                    text: qsTr("Open Bridge")

                    borderless: true
                    labelType: Label.LabelType.Caption_semibold

                    onClicked: {
                        // GODT-1479: we disable hover for the button to avoid a visual glitch where the button is
                        // wrongly hovered when re-opening the status window after clicking
                        hoverEnabled = false;
                        root.showMainWindow()
                        root.close()
                    }
                }

                Item {
                    Layout.fillWidth: true
                }

                Button {
                    colorScheme: root.colorScheme
                    secondary: true
                    icon.source: "/qml/icons/ic-three-dots-vertical.svg"
                    borderless: true
                    checkable: true

                    onClicked: {
                        menu.open()
                    }

                    Menu {
                        id: menu
                        colorScheme: root.colorScheme
                        modal: true

                        y: 0 - height

                        MenuItem {
                            colorScheme: root.colorScheme
                            text: qsTr("Help")
                            onClicked: {
                                root.showHelp()
                                root.close()
                            }
                        }
                        MenuItem {
                            colorScheme: root.colorScheme
                            text: qsTr("Settings")
                            onClicked: {
                                root.showSettings()
                                root.close()
                            }
                        }
                        MenuItem {
                            colorScheme: root.colorScheme
                            text: qsTr("Quit Bridge")
                            onClicked: {
                                root.close()
                                root.quit()
                            }
                        }

                        onClosed: {
                            parent.checked = false
                        }
                        onOpened: {
                            parent.checked = true
                        }
                    }
                }
            }
        }
    }

    onActiveChanged: {
        if (!active) root.close()
    }

    function showAndRise() {
        root.show()
        root.raise()
        if (!root.active) {
            root.requestActivate()
        }
    }
}
