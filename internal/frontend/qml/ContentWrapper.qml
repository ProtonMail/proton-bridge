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

Item {
    id: root
    property var colorScheme: parent.colorScheme

    property var window

    RowLayout {
        anchors.fill: parent
        spacing: 0

        Rectangle {
            id: leftBar
            property var colorScheme: ProtonStyle.prominentStyle

            Layout.minimumWidth: 264
            Layout.maximumWidth: 320
            Layout.preferredWidth: 320
            Layout.fillHeight: true

            color: colorScheme.background_norm

            ColumnLayout {
                anchors.fill: parent
                spacing: 0

                RowLayout {
                    id:topLeftBar

                    Layout.fillWidth: true
                    Layout.minimumHeight: 60
                    Layout.maximumHeight: 60
                    Layout.preferredHeight: 60
                    spacing: 0

                    property var colorScheme: leftBar.colorScheme

                    Status {
                        Layout.leftMargin: 16
                        Layout.topMargin: 24
                        Layout.bottomMargin: 17

                        Layout.alignment: Qt.AlignHCenter
                    }

                    // just a placeholder
                    Item {
                        Layout.fillHeight: true
                        Layout.fillWidth: true
                    }

                    Button {
                        Layout.minimumHeight: 36
                        Layout.maximumHeight: 36
                        Layout.preferredHeight: 36
                        Layout.minimumWidth: 36
                        Layout.maximumWidth: 36
                        Layout.preferredWidth: 36

                        Layout.topMargin: 16
                        Layout.bottomMargin: 9
                        Layout.rightMargin: 4

                        horizontalPadding: 0

                        icon.source: "./icons/ic-question-circle.svg"
                    }

                    Button {
                        Layout.minimumHeight: 36
                        Layout.maximumHeight: 36
                        Layout.preferredHeight: 36
                        Layout.minimumWidth: 36
                        Layout.maximumWidth: 36
                        Layout.preferredWidth: 36

                        Layout.topMargin: 16
                        Layout.bottomMargin: 9
                        Layout.rightMargin: 16

                        horizontalPadding: 0

                        icon.source: "./icons/ic-cog-wheel.svg"
                    }
                }

                // Separator
                Rectangle {
                    Layout.fillWidth: true
                    Layout.minimumHeight: 1
                    Layout.maximumHeight: 1
                    color: leftBar.colorScheme.border_weak
                }

                ListView {
                    id: accounts
                    Layout.fillWidth: true
                    Layout.fillHeight: true
                    Layout.leftMargin: 16
                    Layout.rightMargin: 16
                    Layout.topMargin: 24
                    Layout.bottomMargin: 24

                    spacing: 12

                    header: Rectangle {
                        height: headerLabel.height+16
                        color: ProtonStyle.transparent
                        ProtonLabel{
                            id:headerLabel
                            text: qsTr("Accounts")
                            color: leftBar.colorScheme.text_norm
                            state: "body"
                        }
                    }

                    model: window.backend.users
                    delegate: AccountDelegate{
                        id: accountDelegate
                        colorScheme: leftBar.colorScheme
                        text: modelData.username
                    }
                }

                // Separator
                Rectangle {
                    Layout.fillWidth: true
                    Layout.minimumHeight: 1
                    Layout.maximumHeight: 1
                    color: leftBar.colorScheme.border_weak
                }

                Item {
                    id: bottomLeftBar

                    Layout.fillWidth: true
                    Layout.minimumHeight: 52
                    Layout.maximumHeight: 52
                    Layout.preferredHeight: 52

                    property var colorScheme: leftBar.colorScheme

                    Button {
                        width: 36
                        height: 36

                        anchors.left: parent.left
                        anchors.top: parent.top

                        anchors.leftMargin: 16
                        anchors.topMargin: 7

                        horizontalPadding: 0

                        icon.source: "./icons/ic-plus.svg"

                        onClicked: root.showSignIn()
                    }
                }
            }
        }

        Rectangle {
            id: rightPlane

            Layout.fillWidth: true
            Layout.fillHeight: true

            color: colorScheme.background_norm

            StackLayout {
                id: rightContent
                anchors.fill: parent

                AccountView {
                    colorScheme: root.colorScheme
                }

                GridLayout {
                    SignIn {
                        Layout.topMargin: 68
                        Layout.leftMargin: 80
                        Layout.rightMargin: 80
                        Layout.bottomMargin: 68
                        Layout.preferredWidth: 320
                        Layout.fillWidth: true
                        Layout.fillHeight: true

                        colorScheme: root.colorScheme
                        user: (root.window.backend.users.count === 1 && root.window.backend.users.get(0).loggedIn === false) ? root.window.backend.users.get(0) : undefined
                        backend: root.window.backend
                        window: root.window

                        onLogin          : { root.window.login          ( username , password ) }
                        onLogin2FA       : { root.window.login2FA       ( username , code     ) }
                        onLogin2Password : { root.window.login2Password ( username , password ) }
                        onLoginAbort     : { root.window.loginAbort     ( username ) }
                    }
                }
            }
        }
    }


    function showSignIn() {
        rightContent.currentIndex = 1
    }
}
