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
    property ColorScheme colorScheme

    property var backend
    property var notifications

    signal showSetupGuide(var user, string address)

    RowLayout {
        anchors.fill: parent
        spacing: 0

        Rectangle {
            id: leftBar
            property ColorScheme colorScheme: root.colorScheme.prominent

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

                    Status {
                        colorScheme: leftBar.colorScheme
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
                        colorScheme: leftBar.colorScheme
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

                        onClicked: rightContent.showHelpView()
                    }

                    Button {
                        colorScheme: leftBar.colorScheme
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

                        onClicked: rightContent.showGeneralSettings()
                    }
                }

                Item {implicitHeight:10}

                // Separator line
                Rectangle {
                    Layout.fillWidth: true
                    Layout.minimumHeight: 1
                    Layout.maximumHeight: 1
                    color: leftBar.colorScheme.border_weak
                }

                ListView {
                    id: accounts

                    property var _topBottomMargins: 24
                    property var _leftRightMargins: 16

                    Layout.fillWidth: true
                    Layout.fillHeight: true
                    Layout.leftMargin: accounts._leftRightMargins
                    Layout.rightMargin: accounts._leftRightMargins
                    Layout.topMargin: accounts._topBottomMargins
                    Layout.bottomMargin: accounts._topBottomMargins

                    spacing: 12
                    clip: true
                    boundsBehavior: Flickable.StopAtBounds

                    header: Rectangle {
                        height: headerLabel.height+16
                        // color: ProtonStyle.transparent
                        Label{
                            colorScheme: leftBar.colorScheme
                            id: headerLabel
                            text: qsTr("Accounts")
                            type: Label.LabelType.Body
                        }
                    }

                    highlight: Rectangle {
                        color: leftBar.colorScheme.interaction_default_active
                        radius: 4
                    }

                    model: root.backend.users
                    delegate: AccountDelegate{
                        width: leftBar.width - 2*accounts._leftRightMargins

                        id: accountDelegate
                        colorScheme: leftBar.colorScheme
                        user: root.backend.users.get(index)
                        onClicked: {
                            var user = root.backend.users.get(index)
                            accounts.currentIndex = index
                            if (!user) return
                            if (user.loggedIn) {
                                rightContent.showAccount()
                            } else {
                                signIn.username = user.username
                                rightContent.showSignIn()
                            }
                        }
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

                    Button {
                        colorScheme: leftBar.colorScheme
                        width: 36
                        height: 36

                        anchors.left: parent.left
                        anchors.top: parent.top

                        anchors.leftMargin: 16
                        anchors.topMargin: 7

                        horizontalPadding: 0

                        icon.source: "./icons/ic-plus.svg"

                        onClicked: {
                            signIn.username = ""
                            rightContent.showSignIn()
                        }
                    }
                }
            }
        }

        Rectangle { // right content background
            Layout.fillWidth: true
            Layout.fillHeight: true

            color: colorScheme.background_norm

            StackLayout {
                id: rightContent
                anchors.fill: parent

                AccountView { // 0
                    colorScheme: root.colorScheme
                    backend: root.backend
                    notifications: root.notifications
                    user: {
                        if (accounts.currentIndex < 0) return undefined
                        if (root.backend.users.count == 0) return undefined
                        return root.backend.users.get(accounts.currentIndex)
                    }
                    onShowSignIn: {
                        signIn.username = this.user.username
                        rightContent.showSignIn()
                    }
                    onShowSetupGuide: {
                        root.showSetupGuide(user,address)
                    }
                }

                GridLayout { // 1 Sign In
                    columns: 2

                    Button {
                        id: backButton
                        Layout.leftMargin: 18
                        Layout.topMargin: 10
                        Layout.alignment: Qt.AlignTop

                        colorScheme: root.colorScheme
                        onClicked: {
                            signIn.abort()
                            rightContent.showAccount()
                        }
                        icon.source: "icons/ic-arrow-left.svg"
                        secondary: true
                        horizontalPadding: 8
                    }

                    SignIn {
                        id: signIn
                        Layout.topMargin: 68
                        Layout.leftMargin: 80 - backButton.width - 18
                        Layout.rightMargin: 80
                        Layout.bottomMargin: 68
                        Layout.preferredWidth: 320
                        Layout.fillWidth: true
                        Layout.fillHeight: true

                        colorScheme: root.colorScheme
                        backend: root.backend
                    }
                }

                GeneralSettings { // 2
                    colorScheme: root.colorScheme
                    backend: root.backend
                    notifications: root.notifications
                }

                PortSettings { // 3
                    colorScheme: root.colorScheme
                    backend: root.backend
                }

                SMTPSettings { // 4
                    colorScheme: root.colorScheme
                    backend: root.backend
                }

                LocalCacheSettings { // 5
                    colorScheme: root.colorScheme
                    backend: root.backend
                    notifications: root.notifications
                }

                HelpView { // 6
                    colorScheme: root.colorScheme
                    backend: root.backend
                }

                BugReportView { // 7
                    colorScheme: root.colorScheme
                    backend: root.backend
                    selectedAddress: {
                        if (accounts.currentIndex < 0) return ""
                        if (root.backend.users.count == 0) return ""
                        var user = root.backend.users.get(accounts.currentIndex)
                        if (!user) return ""
                        return user.addresses[0]
                    }
                }

                function showAccount            () { rightContent.currentIndex = 0 }
                function showSignIn             () { rightContent.currentIndex = 1 }
                function showGeneralSettings    () { rightContent.currentIndex = 2 }
                function showPortSettings       () { rightContent.currentIndex = 3 }
                function showSMTPSettings       () { rightContent.currentIndex = 4 }
                function showLocalCacheSettings () { rightContent.currentIndex = 5 }
                function showHelpView           () { rightContent.currentIndex = 6 }
                function showBugReport          () { rightContent.currentIndex = 7 }

                Connections {
                    target: root.backend

                    onLoginFinished: rightContent.showAccount()
                }
            }
        }
    }

    function showLocalCacheSettings(){rightContent.showLocalCacheSettings() }
    function showSettings(){rightContent.showGeneralSettings() }
    function showHelp(){rightContent.showHelpView() }
    function showSignIn(username){
        signIn.username = username
        rightContent.showSignIn()
    }
}
