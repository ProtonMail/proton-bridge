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
import Proton

Item {
    id: root

    property bool _connected: root.user ? root.user.state === EUserState.Connected : false
    property int _contentWidth: 640
    property int _detailsMargin: 25
    property int _lineThickness: 1
    property int _spacing: 20
    property int _buttonSpacing: 8
    property int _topMargin: 32
    property ColorScheme colorScheme
    property var notifications
    property var user

    signal showClientConfigurator(var user, string address, bool justLoggedIn)
    signal showLogin(var username)

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
                    Layout.fillWidth: true
                    clip: true
                    color: root.colorScheme.background_norm
                    implicitHeight: childrenRect.height

                    ColumnLayout {
                        id: topLayout
                        anchors.horizontalCenter: parent.horizontalCenter
                        spacing: _spacing
                        width: _contentWidth

                        RowLayout {
                            // account delegate with action buttons
                            Layout.fillWidth: true
                            Layout.topMargin: _topMargin
                            spacing: _buttonSpacing
                            AccountDelegate {
                                Layout.fillWidth: true
                                colorScheme: root.colorScheme
                                enabled: _connected
                                type: AccountDelegate.LargeView
                                user: root.user
                            }
                            Button {
                                Layout.alignment: Qt.AlignTop
                                colorScheme: root.colorScheme
                                secondary: true
                                text: qsTr("Sign out")
                                visible: _connected

                                onClicked: {
                                    if (!root.user)
                                        return;
                                    root.user.logout();
                                }
                            }
                            Button {
                                id: signIn
                                Layout.alignment: Qt.AlignTop
                                colorScheme: root.colorScheme
                                secondary: true
                                text: qsTr("Sign in")
                                visible: root.user ? (root.user.state === EUserState.SignedOut) : false
                                Accessible.name: text

                                onClicked: {
                                    if (user) {
                                        root.showLogin(user.primaryEmailOrUsername());
                                    }
                                }
                            }
                            Button {
                                id: removeAccount
                                Layout.alignment: Qt.AlignTop
                                colorScheme: root.colorScheme
                                icon.source: "/qml/icons/ic-trash.svg"
                                secondary: true
                                visible: root.user ? root.user.state !== EUserState.Locked : false
                                Accessible.name: qsTr("Remove account")

                                onClicked: {
                                    if (!root.user)
                                        return;
                                    root.notifications.askDeleteAccount(root.user);
                                }
                            }
                        }
                        Rectangle {
                            Layout.fillWidth: true
                            color: root.colorScheme.border_weak
                            height: root._lineThickness
                        }
                        SettingsItem {
                            id: configureEmailClient
                            Layout.fillWidth: true
                            actionText: qsTr("Configure email client")
                            colorScheme: root.colorScheme
                            description: qsTr("Using the mailbox details below (re)configure your client.")
                            showSeparator: splitMode.visible
                            text: qsTr("Email clients")
                            type: SettingsItem.PrimaryButton
                            visible: _connected && ((!root.user.splitMode) || (root.user.addresses.length === 1))
                            Accessible.name: actionText

                            onClicked: {
                                if (!root.user)
                                    return;
                                root.showClientConfigurator(root.user, user.addresses[0], false);
                            }
                        }
                        SettingsItem {
                            id: splitMode
                            Layout.fillWidth: true
                            checked: root.user ? root.user.splitMode : false
                            colorScheme: root.colorScheme
                            description: qsTr("Setup multiple email addresses individually.")
                            showSeparator: addressSelector.visible
                            text: qsTr("Split addresses")
                            type: SettingsItem.Toggle
                            visible: _connected && root.user.addresses.length > 1
                            Accessible.name: text

                            onClicked: {
                                if (!splitMode.checked) {
                                    root.notifications.askEnableSplitMode(user);
                                } else {
                                    addressSelector.currentIndex = 0;
                                    root.user.toggleSplitMode(!splitMode.checked);
                                }
                            }
                        }
                        RowLayout {
                            Layout.bottomMargin: _spacing
                            Layout.fillWidth: true
                            visible: _connected && root.user.splitMode

                            ComboBox {
                                id: addressSelector
                                Layout.fillWidth: true
                                colorScheme: root.colorScheme
                                model: root.user ? root.user.addresses : null
                            }
                            Button {
                                colorScheme: root.colorScheme
                                secondary: false
                                text: qsTr("Configure email client")

                                onClicked: {
                                    if (!root.user)
                                        return;
                                    root.showClientConfigurator(root.user, addressSelector.displayText, false);
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
                    color: root.colorScheme.background_weak
                    implicitHeight: bottomLayout.implicitHeight

                    ColumnLayout {
                        id: bottomLayout
                        anchors.horizontalCenter: parent.horizontalCenter
                        spacing: _spacing
                        visible: _connected
                        width: _contentWidth

                        Label {
                            Layout.topMargin: _detailsMargin
                            colorScheme: root.colorScheme
                            text: qsTr("Mailbox details")
                            type: Label.Body_semibold
                        }
                        RowLayout {
                            id: configuration

                            property string currentAddress: addressSelector.displayText

                            Layout.fillHeight: true
                            Layout.fillWidth: true
                            spacing: _spacing

                            Configuration {
                                Layout.fillWidth: true
                                colorScheme: root.colorScheme
                                hostname: Backend.hostname
                                password: root.user ? root.user.password : ""
                                port: Backend.imapPort.toString()
                                security: Backend.useSSLForIMAP ? "SSL" : "STARTTLS"
                                title: qsTr("IMAP")
                                username: configuration.currentAddress
                            }
                            Configuration {
                                Layout.fillWidth: true
                                colorScheme: root.colorScheme
                                hostname: Backend.hostname
                                password: root.user ? root.user.password : ""
                                port: Backend.smtpPort.toString()
                                security: Backend.useSSLForSMTP ? "SSL" : "STARTTLS"
                                title: qsTr("SMTP")
                                username: configuration.currentAddress
                            }
                        }
                    }
                }
            }
        }
    }
}
