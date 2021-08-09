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

import QtQuick 2.13
import QtQuick.Layouts 1.12
import QtQuick.Controls 2.12

import Proton 4.0

ScrollView {
    id: root
    property ColorScheme colorScheme
    property var backend
    property var notifications
    property var user

    clip: true
    contentWidth: pane.width
    contentHeight: pane.height

    property int _leftRightMargins: 64
    property int _topBottomMargins: 68
    property int _spacing: 22

    Rectangle {
        anchors {
            bottom: pane.bottom
        }
        color: root.colorScheme.background_weak
        width: root.width
        height: configuration.height + root._topBottomMargins
    }

    signal showSignIn()
    signal showSetupGuide(var user, string address)

    ColumnLayout {
        id: pane

        width: root.width

        ColumnLayout {
            spacing: root._spacing
            Layout.topMargin: root._topBottomMargins
            Layout.leftMargin: root._leftRightMargins
            Layout.rightMargin: root._leftRightMargins
            Layout.maximumWidth: root.width - 2*root._leftRightMargins

            RowLayout { // account delegate with action buttons
                Layout.fillWidth: true

                AccountDelegate {
                    Layout.fillWidth: true
                    colorScheme: root.colorScheme
                    user: root.user
                    type: AccountDelegate.LargeView
                    enabled: root.user.loggedIn
                }

                Button {
                    Layout.alignment: Qt.AlignTop
                    colorScheme: root.colorScheme
                    text: qsTr("Sign out")
                    secondary: true
                    visible: root.user.loggedIn
                    onClicked: root.user.logout()
                }

                Button {
                    Layout.alignment: Qt.AlignTop
                    colorScheme: root.colorScheme
                    icon.source: "icons/ic-trash.svg"
                    secondary: true
                    visible: root.user.loggedIn
                    onClicked: root.user.remove()
                }

                Button {
                    Layout.alignment: Qt.AlignTop
                    colorScheme: root.colorScheme
                    text: qsTr("Sign in")
                    secondary: true
                    visible: !root.user.loggedIn
                    onClicked: root.parent.rightContent.showSignIn()
                }
            }

            Rectangle {
                Layout.fillWidth: true
                height: 1
                color: root.colorScheme.border_weak
            }

            SettingsItem {
                colorScheme: root.colorScheme
                text: qsTr("Email clients")
                actionText: qsTr("Configure")
                description: "MISSING WIREFRAME" // TODO
                type: SettingsItem.Button
                enabled: root.user.loggedIn
                visible: !root.user.splitMode
                onClicked: root.showSetupGuide(root.user,user.addresses[0])
            }

            SettingsItem {
                id: splitMode
                colorScheme: root.colorScheme
                text: qsTr("Split addresses")
                description: qsTr("Split addresses allows you to configure multiple email addresses individually. Changing its mode will require you to delete your accounts(s) from your email client and begin the setup process from scratch.")
                type: SettingsItem.Toggle
                checked: root.user.splitMode
                visible: root.user.addresses.length > 1
                enabled: root.user.loggedIn
                onClicked: {
                    if (!splitMode.checked){
                        root.notifications.askEnableSplitMode(user)
                    } else {
                        root.user.toggleSplitMode(!splitMode.checked)
                    }
                }
            }

            RowLayout {
                Layout.fillWidth: true
                enabled: root.user.loggedIn

                visible: root.user.splitMode

                ComboBox {
                    id: addressSelector
                    Layout.fillWidth: true
                    model: root.user.addresses

                    property var _topBottomMargins : 8
                    property var _leftRightMargins : 16

                    background: RoundedRectangle {
                        radiusTopLeft     : 6
                        radiusTopRight    : 6
                        radiusBottomLeft  : addressSelector.down ? 0 : 6
                        radiusBottomRight : addressSelector.down ? 0 : 6

                        height: addressSelector.contentItem.height
                        //width: addressSelector.contentItem.width

                        fillColor   : root.colorScheme.background_norm
                        strokeColor : root.colorScheme.border_norm
                        strokeWidth : 1
                    }

                    delegate: Rectangle {
                        id: listItem
                        width: root.width
                        height: children[0].height + 4 + 2*addressSelector._topBottomMargins

                        Label {
                            anchors {
                                top        : parent.top
                                left       : parent.left
                                topMargin  : addressSelector._topBottomMargins + 4
                                leftMargin : addressSelector._leftRightMargins
                            }

                            colorScheme: root.colorScheme
                            text: modelData
                            elide: Text.ElideMiddle
                        }

                        property bool isOver: false
                        color: {
                            if (listItem.isOver) return root.colorScheme.interaction_weak_hover
                            if (addressSelector.highlightedIndex === index) return root.colorScheme.interaction_weak
                            return root.colorScheme.background_norm
                        }

                        MouseArea {
                            anchors.fill: parent
                            hoverEnabled: true
                            onEntered: listItem.isOver = true
                            onExited: listItem.isOver = false
                            onClicked : {
                                addressSelector.currentIndex = index
                                addressSelector.popup.close()
                            }
                        }
                    }

                    contentItem: Label {
                        topPadding    : addressSelector._topBottomMargins+4
                        bottomPadding : addressSelector._topBottomMargins
                        leftPadding   : addressSelector._leftRightMargins
                        rightPadding  : addressSelector._leftRightMargins

                        colorScheme: root.colorScheme
                        text: addressSelector.displayText
                        elide: Text.ElideMiddle
                    }
                }

                Button {
                    colorScheme: root.colorScheme
                    text: qsTr("Configure")
                    secondary: true
                    onClicked: root.showSetupGuide(root.user, addressSelector.displayText)
                }
            }

            Item {implicitHeight: 1}
        }

        ColumnLayout {
            id: configuration
            Layout.bottomMargin: root._topBottomMargins
            Layout.leftMargin: root._leftRightMargins
            Layout.rightMargin: root._leftRightMargins
            Layout.maximumWidth: root.width - 2*root._leftRightMargins
            spacing: root._spacing
            visible: root.user.loggedIn

            property string currentAddress: addressSelector.displayText

            Item {height: 1}

            Label {
                colorScheme: root.colorScheme
                text: qsTr("Mailbox details")
                type: Label.Body_semibold
            }

            Configuration {
                colorScheme: root.colorScheme
                title: qsTr("IMAP")
                hostname:   root.backend.hostname
                port:       root.backend.portIMAP.toString()
                username:   configuration.currentAddress
                password:   root.user.password
                security:   "STARTTLS"
            }

            Configuration {
                colorScheme: root.colorScheme
                title: qsTr("SMTP")
                hostname : root.backend.hostname
                port     : root.backend.portSMTP.toString()
                username : configuration.currentAddress
                password : root.user.password
                security : root.backend.useSSLforSMTP ? "SSL" : "STARTTLS"
            }
        }
    }
}
