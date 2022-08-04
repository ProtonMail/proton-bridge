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

import QtQuick
import QtQuick.Layouts
import QtQuick.Controls

import Proton

Item {
    id: root
    property ColorScheme colorScheme
    property var notifications
    property var user

    signal showSignIn()
    signal showSetupGuide(var user, string address)

    property int _leftMargin: 64
    property int _rightMargin: 64
    property int _topMargin: 32
    property int _detailsTopMargin: 25
    property int _bottomMargin: 12
    property int _spacing: 20
    property int _lineWidth: 1

    ScrollView {
        id: scrollView
        clip: true

        anchors.fill: parent

        Item {
            // can't use parent here because parent is not ScrollView (Flickable inside contentItem inside ScrollView)
            width: scrollView.availableWidth
            height: scrollView.availableHeight

            implicitHeight: children[0].implicitHeight + children[0].anchors.topMargin + children[0].anchors.bottomMargin
            // do not set implicitWidth because implicit width of ColumnLayout will be equal to maximum implicit width of
            // internal items. And if one of internal items would be a Text or Label - implicit width of those is always
            // equal to non-wrapped text (i.e. one line only). That will lead to enabling horizontal scroll when not needed
            implicitWidth: width

            ColumnLayout {
                spacing: 0

                anchors.fill: parent

                Rectangle {
                    id: topRectangle
                    color: root.colorScheme.background_norm

                    implicitHeight: children[0].implicitHeight + children[0].anchors.topMargin + children[0].anchors.bottomMargin
                    implicitWidth: children[0].implicitWidth + children[0].anchors.leftMargin + children[0].anchors.rightMargin

                    Layout.fillWidth: true

                    ColumnLayout {
                        spacing: root._spacing

                        anchors.fill: parent
                        anchors.leftMargin: root._leftMargin
                        anchors.rightMargin: root._rightMargin
                        anchors.topMargin: root._topMargin
                        anchors.bottomMargin: root._bottomMargin

                        RowLayout { // account delegate with action buttons
                            Layout.fillWidth: true

                            AccountDelegate {
                                Layout.fillWidth: true
                                colorScheme: root.colorScheme
                                user: root.user
                                type: AccountDelegate.LargeView
                                enabled: root.user ? root.user.loggedIn : false
                            }

                            Button {
                                Layout.alignment: Qt.AlignTop
                                colorScheme: root.colorScheme
                                text: qsTr("Sign out")
                                secondary: true
                                visible: root.user ? root.user.loggedIn : false
                                onClicked: {
                                    if (!root.user) return
                                    root.user.logout()
                                }
                            }

                            Button {
                                Layout.alignment: Qt.AlignTop
                                colorScheme: root.colorScheme
                                text: qsTr("Sign in")
                                secondary: true
                                visible: root.user ? !root.user.loggedIn : false
                                onClicked: {
                                    if (!root.user) return
                                    root.showSignIn()
                                }
                            }

                            Button {
                                Layout.alignment: Qt.AlignTop
                                colorScheme: root.colorScheme
                                icon.source: "/qml/icons/ic-trash.svg"
                                secondary: true
                                onClicked: {
                                    if (!root.user) return
                                    root.notifications.askDeleteAccount(root.user)
                                }
                            }
                        }

                        Rectangle {
                            Layout.fillWidth: true
                            height: root._lineWidth
                            color: root.colorScheme.border_weak
                        }

                        SettingsItem {
                            colorScheme: root.colorScheme
                            text: qsTr("Email clients")
                            actionText: qsTr("Configure")
                            description: qsTr("Using the mailbox details below (re)configure your client.")
                            type: SettingsItem.Button
                            enabled: root.user ? root.user.loggedIn : false
                            visible: root.user ? !root.user.splitMode || root.user.addresses.length==1 : false
                            showSeparator: splitMode.visible
                            onClicked: {
                                if (!root.user) return
                                root.showSetupGuide(root.user, user.addresses[0])
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
                            visible: root.user ? root.user.addresses.length > 1 : false
                            enabled: root.user ? root.user.loggedIn : false
                            showSeparator: addressSelector.visible
                            onClicked: {
                                if (!splitMode.checked){
                                    root.notifications.askEnableSplitMode(user)
                                } else {
                                    addressSelector.currentIndex = 0
                                    root.user.toggleSplitMode(!splitMode.checked)
                                }
                            }

                            Layout.fillWidth: true
                        }

                        RowLayout {
                            Layout.fillWidth: true
                            enabled: root.user ? root.user.loggedIn : false
                            visible: root.user ? root.user.splitMode : false

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
                                    if (!root.user) return
                                    root.showSetupGuide(root.user, addressSelector.displayText)
                                }
                            }
                        }
                    }
                }

                Rectangle {
                    color: root.colorScheme.background_weak

                    implicitHeight: children[0].implicitHeight + children[0].anchors.topMargin + children[0].anchors.bottomMargin
                    implicitWidth: children[0].implicitWidth + children[0].anchors.leftMargin + children[0].anchors.rightMargin

                    Layout.fillWidth: true

                    ColumnLayout {
                        id: configuration

                        anchors.fill: parent
                        anchors.leftMargin: root._leftMargin
                        anchors.rightMargin: root._rightMargin
                        anchors.topMargin: root._detailsTopMargin
                        anchors.bottomMargin: root._spacing

                        spacing: root._spacing
                        visible: root.user ? root.user.loggedIn : false

                        property string currentAddress: addressSelector.displayText

                        Label {
                            colorScheme: root.colorScheme
                            text: qsTr("Mailbox details")
                            type: Label.Body_semibold
                        }

                        Configuration {
                            colorScheme: root.colorScheme
                            title: qsTr("IMAP")
                            hostname:   Backend.hostname
                            port:       Backend.portIMAP.toString()
                            username:   configuration.currentAddress
                            password:   root.user ? root.user.password : ""
                            security:   "STARTTLS"
                        }

                        Configuration {
                            colorScheme: root.colorScheme
                            title: qsTr("SMTP")
                            hostname : Backend.hostname
                            port     : Backend.portSMTP.toString()
                            username : configuration.currentAddress
                            password : root.user ? root.user.password : ""
                            security : Backend.useSSLforSMTP ? "SSL" : "STARTTLS"
                        }
                    }
                }
            }
        }
    }
}
