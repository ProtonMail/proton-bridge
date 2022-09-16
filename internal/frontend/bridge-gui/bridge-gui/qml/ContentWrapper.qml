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
import Notifications

Item {
    id: root
    property ColorScheme colorScheme

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
                        Layout.leftMargin: 16
                        Layout.topMargin: 24
                        Layout.bottomMargin: 17
                        Layout.alignment: Qt.AlignHCenter

                        colorScheme: leftBar.colorScheme
                        notifications: root.notifications

                        notificationWhitelist: Notifications.Group.Connection | Notifications.Group.ForceUpdate
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

                        icon.source: "/qml/icons/ic-question-circle.svg"

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

                        icon.source: "/qml/icons/ic-cog-wheel.svg"

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
                        radius: ProtonStyle.account_row_radius
                    }

                    model: Backend.users
                    delegate: Item {
                        width: leftBar.width - 2*accounts._leftRightMargins
                        implicitHeight: children[0].implicitHeight + children[0].anchors.topMargin + children[0].anchors.bottomMargin
                        implicitWidth: children[0].implicitWidth + children[0].anchors.leftMargin + children[0].anchors.rightMargin

                        AccountDelegate {
                            id: accountDelegate

                            anchors.fill: parent
                            anchors.topMargin: 8
                            anchors.bottomMargin: 8
                            anchors.leftMargin: 12
                            anchors.rightMargin: 12

                            colorScheme: leftBar.colorScheme
                            user: Backend.users.get(index)
                        }

                        MouseArea {
                            anchors.fill: parent
                            onClicked: {
                                var user = Backend.users.get(index)
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

                        icon.source: "/qml/icons/ic-plus.svg"

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
                    notifications: root.notifications
                    user: {
                        if (accounts.currentIndex < 0) return undefined
                        if (Backend.users.count == 0) return undefined
                        return Backend.users.get(accounts.currentIndex)
                    }
                    onShowSignIn: {
                        signIn.username = this.user.username
                        rightContent.showSignIn()
                    }
                    onShowSetupGuide: function(user, address) {
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
                        icon.source: "/qml/icons/ic-arrow-left.svg"
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
                    }
                }

                GeneralSettings { // 2
                    colorScheme: root.colorScheme
                    notifications: root.notifications

                    onBack: {
                        rightContent.showAccount()
                    }
                }

                KeychainSettings { // 3
                    colorScheme: root.colorScheme

                    onBack: {
                        rightContent.showGeneralSettings()
                    }
                }

                PortSettings { // 4
                    colorScheme: root.colorScheme

                    onBack: {
                        rightContent.showGeneralSettings()
                    }
                }

                SMTPSettings { // 5
                    colorScheme: root.colorScheme

                    onBack: {
                        rightContent.showGeneralSettings()
                    }
                }

                LocalCacheSettings { // 6
                    colorScheme: root.colorScheme
                    notifications: root.notifications

                    onBack: {
                        rightContent.showGeneralSettings()
                    }
                }

                HelpView { // 7
                    colorScheme: root.colorScheme

                    onBack: {
                        rightContent.showAccount()
                    }
                }

                BugReportView { // 8
                    colorScheme: root.colorScheme
                    selectedAddress: {
                        if (accounts.currentIndex < 0) return ""
                        if (Backend.users.count == 0) return ""
                        var user = Backend.users.get(accounts.currentIndex)
                        if (!user) return ""
                        return user.addresses[0]
                    }

                    onBack: {
                        rightContent.showHelpView()
                    }
                }

                function showAccount(index) {
                    if (index !== undefined && index >= 0){
                        accounts.currentIndex = index
                    }
                    rightContent.currentIndex = 0
                }

                function showSignIn             () { rightContent.currentIndex = 1; signIn.focus = true }
                function showGeneralSettings    () { rightContent.currentIndex = 2 }
                function showKeychainSettings   () { rightContent.currentIndex = 3 }
                function showPortSettings       () { rightContent.currentIndex = 4 }
                function showSMTPSettings       () { rightContent.currentIndex = 5 }
                function showLocalCacheSettings () { rightContent.currentIndex = 6 }
                function showHelpView           () { rightContent.currentIndex = 7 }
                function showBugReport          () { rightContent.currentIndex = 8 }

                Connections {
                    target: Backend

                    function onLoginFinished(index) { rightContent.showAccount(index) }
                    function onLoginAlreadyLoggedIn(index) { rightContent.showAccount(index) }
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
