// Copyright (c) 2023 Proton AG
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

import Proton

Item {
    id: root
    property ColorScheme colorScheme
    property var notifications
    property var user

    signal showSignIn

    signal showSetupGuide(var user, string address)

    property int _contentWidth: 640
    property int _topMargin: 32
    property int _detailsMargin: 25
    property int _spacing: 20
    property int _lineThickness: 1
    property bool _connected: root.user ? root.user.state === EUserState.Connected : false

    Rectangle {
        anchors.fill: parent
        color: root.colorScheme.background_weak

        ScrollView {
            id: scrollView
            anchors.fill: parent
            Component.onCompleted: contentItem.boundsBehavior = Flickable.StopAtBounds

            ColumnLayout {
                id: topLevelColumnLayout
                anchors.fill: parent
                spacing: 0

                Rectangle {
                    id: topArea
                    color: root.colorScheme.background_norm
                    clip: true
                    Layout.fillWidth: true
                    implicitHeight: childrenRect.height

                    ColumnLayout {
                        id: topLayout
                        width: _contentWidth
                        anchors.horizontalCenter: parent.horizontalCenter
                        spacing: _spacing

                        RowLayout {
                            // account delegate with action buttons
                            Layout.fillWidth: true
                            Layout.topMargin: _topMargin

                            AccountDelegate {
                                Layout.fillWidth: true
                                colorScheme: root.colorScheme
                                user: root.user
                                type: AccountDelegate.LargeView
                                enabled: _connected
                            }

                            Button {
                                Layout.alignment: Qt.AlignTop
                                colorScheme: root.colorScheme
                                text: qsTr("Sign out")
                                secondary: true
                                visible: _connected
                                onClicked: {
                                    if (!root.user)
                                        return;
                                    root.user.logout();
                                }
                            }

                            Button {
                                Layout.alignment: Qt.AlignTop
                                colorScheme: root.colorScheme
                                text: qsTr("Sign in")
                                secondary: true
                                visible: root.user ? (root.user.state === EUserState.SignedOut) : false
                                onClicked: {
                                    if (!root.user)
                                        return;
                                    root.showSignIn();
                                }
                            }

                            Button {
                                Layout.alignment: Qt.AlignTop
                                colorScheme: root.colorScheme
                                icon.source: "/qml/icons/ic-trash.svg"
                                secondary: true
                                onClicked: {
                                    if (!root.user)
                                        return;
                                    root.notifications.askDeleteAccount(root.user);
                                }
                                visible: root.user ? root.user.state !== EUserState.Locked : false
                            }
                        }

                        Rectangle {
                            Layout.fillWidth: true
                            height: root._lineThickness
                            color: root.colorScheme.border_weak
                        }

                        SettingsItem {
                            colorScheme: root.colorScheme
                            text: qsTr("Email clients")
                            actionText: qsTr("Configure")
                            description: qsTr("Using the mailbox details below (re)configure your client.")
                            type: SettingsItem.Button
                            visible: _connected && (!root.user.splitMode) || (root.user.addresses.length === 1)
                            showSeparator: splitMode.visible
                            onClicked: {
                                if (!root.user)
                                    return;
                                root.showSetupGuide(root.user, user.addresses[0]);
                            }

                            Layout.fillWidth: true
                        }

                        SettingsItem {
                            id: splitMode
                            colorScheme: root.colorScheme
                            text: qsTr("Split addresses")
                            description: qsTr("Setup multiple email addresses individually.")
                            type: SettingsItem.Toggle
                            checked: root.user ? root.user.splitMode : false
                            visible: _connected && root.user.addresses.length > 1
                            showSeparator: addressSelector.visible
                            onClicked: {
                                if (!splitMode.checked) {
                                    root.notifications.askEnableSplitMode(user);
                                } else {
                                    addressSelector.currentIndex = 0;
                                    root.user.toggleSplitMode(!splitMode.checked);
                                }
                            }

                            Layout.fillWidth: true
                        }

                        RowLayout {
                            Layout.fillWidth: true
                            Layout.bottomMargin: _spacing
                            visible: _connected && root.user.splitMode

                            ComboBox {
                                id: addressSelector
                                colorScheme: root.colorScheme
                                Layout.fillWidth: true
                                model: root.user ? root.user.addresses : null
                            }

                            Button {
                                colorScheme: root.colorScheme
                                text: qsTr("Configure")
                                secondary: true
                                onClicked: {
                                    if (!root.user)
                                        return;
                                    root.showSetupGuide(root.user, addressSelector.displayText);
                                }
                            }
                        }

                        Rectangle {
                            height: 0
                        } // just for some extra space before separator
                    }
                }

                Rectangle {
                    id: bottomArea
                    Layout.fillWidth: true
                    implicitHeight: bottomLayout.implicitHeight
                    color: root.colorScheme.background_weak

                    ColumnLayout {
                        id: bottomLayout
                        width: _contentWidth
                        anchors.horizontalCenter: parent.horizontalCenter
                        spacing: _spacing
                        visible: _connected

                        Label {
                            Layout.topMargin: _detailsMargin
                            colorScheme: root.colorScheme
                            text: qsTr("Mailbox details")
                            type: Label.Body_semibold
                        }

                        RowLayout {
                            id: configuration
                            spacing: _spacing
                            Layout.fillWidth: true
                            Layout.fillHeight: true

                            property string currentAddress: addressSelector.displayText

                            Configuration {
                                Layout.fillWidth: true
                                colorScheme: root.colorScheme
                                title: qsTr("IMAP")
                                hostname: Backend.hostname
                                port: Backend.imapPort.toString()
                                username: configuration.currentAddress
                                password: root.user ? root.user.password : ""
                                security: Backend.useSSLForIMAP ? "SSL" : "STARTTLS"
                            }

                            Configuration {
                                Layout.fillWidth: true
                                colorScheme: root.colorScheme
                                title: qsTr("SMTP")
                                hostname: Backend.hostname
                                port: Backend.smtpPort.toString()
                                username: configuration.currentAddress
                                password: root.user ? root.user.password : ""
                                security: Backend.useSSLForSMTP ? "SSL" : "STARTTLS"
                            }
                        }
                    }
                }
            }
        }
    }
}
