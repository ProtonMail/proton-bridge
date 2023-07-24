// Copyright (c) 2023 Proton AG
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
import Notifications

Item {
    id: root

    property ColorScheme colorScheme
    property var notifications

    signal closeWindow
    signal quitBridge
    signal showSetupGuide(var user, string address)
    signal showSetupWizard

    function selectUser(userID) {
        const users = Backend.users;
        for (let i = 0; i < users.count; i++) {
            const user = users.get(i);
            if (user.id !== userID) {
                continue;
            }
            accounts.currentIndex = i;
            if (user.state === EUserState.SignedOut)
                showSignIn(user.primaryEmailOrUsername());
            return;
        }
        console.error("User with ID ", userID, " was not found in the account list");
    }
    function showHelp() {
        rightContent.showHelpView();
    }
    function showLocalCacheSettings() {
        rightContent.showLocalCacheSettings();
    }
    function showSettings() {
        rightContent.showGeneralSettings();
    }
    function showSignIn(username) {
        signIn.username = username;
        rightContent.showSignIn();
    }

    RowLayout {
        anchors.fill: parent
        spacing: 0

        Rectangle {
            id: leftBar

            property ColorScheme colorScheme: root.colorScheme.prominent

            Layout.fillHeight: true
            Layout.maximumWidth: 320
            Layout.minimumWidth: 264
            Layout.preferredWidth: 320
            color: colorScheme.background_norm

            ColumnLayout {
                anchors.fill: parent
                spacing: 0

                RowLayout {
                    id: topLeftBar
                    Layout.fillWidth: true
                    Layout.maximumHeight: 60
                    Layout.minimumHeight: 60
                    Layout.preferredHeight: 60
                    spacing: 0

                    Status {
                        Layout.alignment: Qt.AlignHCenter
                        Layout.bottomMargin: 17
                        Layout.leftMargin: 16
                        Layout.topMargin: 24
                        colorScheme: leftBar.colorScheme
                        notificationWhitelist: Notifications.Group.Connection | Notifications.Group.ForceUpdate
                        notifications: root.notifications
                    }

                    // just a placeholder
                    Item {
                        Layout.fillHeight: true
                        Layout.fillWidth: true
                    }
                    Button {
                        Layout.bottomMargin: 9
                        Layout.maximumHeight: 36
                        Layout.maximumWidth: 36
                        Layout.minimumHeight: 36
                        Layout.minimumWidth: 36
                        Layout.preferredHeight: 36
                        Layout.preferredWidth: 36
                        Layout.rightMargin: 4
                        Layout.topMargin: 16
                        colorScheme: leftBar.colorScheme
                        horizontalPadding: 0
                        icon.source: "/qml/icons/ic-question-circle.svg"

                        onClicked: rightContent.showHelpView()
                    }
                    Button {
                        Layout.bottomMargin: 9
                        Layout.maximumHeight: 36
                        Layout.maximumWidth: 36
                        Layout.minimumHeight: 36
                        Layout.minimumWidth: 36
                        Layout.preferredHeight: 36
                        Layout.preferredWidth: 36
                        Layout.rightMargin: 4
                        Layout.topMargin: 16
                        colorScheme: leftBar.colorScheme
                        horizontalPadding: 0
                        icon.source: "/qml/icons/ic-cog-wheel.svg"

                        onClicked: rightContent.showGeneralSettings()
                    }
                    Button {
                        id: dotMenuButton
                        Layout.bottomMargin: 9
                        Layout.maximumHeight: 36
                        Layout.maximumWidth: 36
                        Layout.minimumHeight: 36
                        Layout.minimumWidth: 36
                        Layout.preferredHeight: 36
                        Layout.preferredWidth: 36
                        Layout.rightMargin: 16
                        Layout.topMargin: 16
                        colorScheme: leftBar.colorScheme
                        horizontalPadding: 0
                        icon.source: "/qml/icons/ic-three-dots-vertical.svg"

                        onClicked: {
                            dotMenu.open();
                        }

                        Menu {
                            id: dotMenu
                            colorScheme: root.colorScheme
                            modal: true
                            y: dotMenuButton.Layout.preferredHeight + dotMenuButton.Layout.bottomMargin

                            onClosed: {
                                parent.checked = false;
                            }
                            onOpened: {
                                parent.checked = true;
                            }

                            MenuItem {
                                colorScheme: root.colorScheme
                                text: qsTr("Close window")

                                onClicked: {
                                    root.closeWindow();
                                }
                            }
                            MenuItem {
                                colorScheme: root.colorScheme
                                text: qsTr("Quit Bridge")

                                onClicked: {
                                    root.quitBridge();
                                }
                            }
                        }
                    }
                }
                Item {
                    implicitHeight: 10
                }

                // Separator line
                Rectangle {
                    Layout.fillWidth: true
                    Layout.maximumHeight: 1
                    Layout.minimumHeight: 1
                    color: leftBar.colorScheme.border_weak
                }
                ListView {
                    id: accounts

                    property var _leftRightMargins: 16
                    property var _topBottomMargins: 24

                    Layout.bottomMargin: accounts._topBottomMargins
                    Layout.fillHeight: true
                    Layout.fillWidth: true
                    Layout.leftMargin: accounts._leftRightMargins
                    Layout.rightMargin: accounts._leftRightMargins
                    Layout.topMargin: accounts._topBottomMargins
                    boundsBehavior: Flickable.StopAtBounds
                    clip: true
                    model: Backend.users
                    spacing: 12

                    delegate: Item {
                        implicitHeight: children[0].implicitHeight + children[0].anchors.topMargin + children[0].anchors.bottomMargin
                        implicitWidth: children[0].implicitWidth + children[0].anchors.leftMargin + children[0].anchors.rightMargin
                        width: leftBar.width - 2 * accounts._leftRightMargins

                        AccountDelegate {
                            id: accountDelegate
                            anchors.bottomMargin: 8
                            anchors.fill: parent
                            anchors.leftMargin: 12
                            anchors.rightMargin: 12
                            anchors.topMargin: 8
                            colorScheme: leftBar.colorScheme
                            user: Backend.users.get(index)
                        }
                        MouseArea {
                            anchors.fill: parent

                            onClicked: {
                                const user = Backend.users.get(index);
                                accounts.currentIndex = index;
                                if (!user)
                                    return;
                                if (user.state !== EUserState.SignedOut) {
                                    rightContent.showAccount();
                                } else {
                                    signIn.username = user.primaryEmailOrUsername();
                                    rightContent.showSignIn();
                                }
                            }
                        }
                    }
                    header: Rectangle {
                        height: headerLabel.height + 16

                        // color: ProtonStyle.transparent
                        Label {
                            id: headerLabel
                            colorScheme: leftBar.colorScheme
                            text: qsTr("Accounts")
                            type: Label.LabelType.Body
                        }
                    }
                    highlight: Rectangle {
                        color: leftBar.colorScheme.interaction_default_active
                        radius: ProtonStyle.account_row_radius
                    }
                }

                // Separator
                Rectangle {
                    Layout.fillWidth: true
                    Layout.maximumHeight: 1
                    Layout.minimumHeight: 1
                    color: leftBar.colorScheme.border_weak
                }
                Item {
                    id: bottomLeftBar
                    Layout.fillWidth: true
                    Layout.maximumHeight: 52
                    Layout.minimumHeight: 52
                    Layout.preferredHeight: 52

                    Button {
                        anchors.left: parent.left
                        anchors.leftMargin: 16
                        anchors.top: parent.top
                        anchors.topMargin: 7
                        colorScheme: leftBar.colorScheme
                        height: 36
                        horizontalPadding: 0
                        icon.source: "/qml/icons/ic-plus.svg"
                        width: 36

                        onClicked: {
                            signIn.username = "";
                            root.showSetupWizard();
                        }
                    }
                }
            }
        }
        Rectangle {
            Layout.fillHeight: true // right content background
            Layout.fillWidth: true
            color: colorScheme.background_norm

            StackLayout {
                id: rightContent
                function showAccount(index) {
                    if (index !== undefined && index >= 0) {
                        accounts.currentIndex = index;
                    }
                    rightContent.currentIndex = 0;
                }
                function showBugReport() {
                    rightContent.currentIndex = 8;
                }
                function showConnectionModeSettings() {
                    rightContent.currentIndex = 5;
                }
                function showGeneralSettings() {
                    rightContent.currentIndex = 2;
                }
                function showHelpView() {
                    rightContent.currentIndex = 7;
                }
                function showKeychainSettings() {
                    rightContent.currentIndex = 3;
                }
                function showLocalCacheSettings() {
                    rightContent.currentIndex = 6;
                }
                function showPortSettings() {
                    rightContent.currentIndex = 4;
                }
                function showSignIn() {
                    rightContent.currentIndex = 1;
                    signIn.focus = true;
                }

                anchors.fill: parent

                AccountView {
                    // 0
                    colorScheme: root.colorScheme
                    notifications: root.notifications
                    user: {
                        if (accounts.currentIndex < 0)
                            return undefined;
                        if (Backend.users.count === 0)
                            return undefined;
                        return Backend.users.get(accounts.currentIndex);
                    }

                    onShowSetupGuide: function (user, address) {
                        root.showSetupGuide(user, address);
                    }
                    onShowSignIn: {
                        const user = this.user;
                        signIn.username = user ? user.primaryEmailOrUsername() : "";
                        rightContent.showSignIn();
                    }
                }
                GridLayout {
                    // 1 Sign In
                    columns: 2

                    Button {
                        id: backButton
                        Layout.alignment: Qt.AlignTop
                        Layout.leftMargin: 18
                        Layout.topMargin: 10
                        colorScheme: root.colorScheme
                        horizontalPadding: 8
                        icon.source: "/qml/icons/ic-arrow-left.svg"
                        secondary: true

                        onClicked: {
                            signIn.abort();
                            rightContent.showAccount();
                        }
                    }
                    SignIn {
                        id: signIn
                        Layout.bottomMargin: 68
                        Layout.fillHeight: true
                        Layout.fillWidth: true
                        Layout.leftMargin: 80 - backButton.width - 18
                        Layout.preferredWidth: 320
                        Layout.rightMargin: 80
                        Layout.topMargin: 68
                        colorScheme: root.colorScheme
                    }
                }
                GeneralSettings {
                    // 2
                    colorScheme: root.colorScheme
                    notifications: root.notifications

                    onBack: {
                        rightContent.showAccount();
                    }
                }
                KeychainSettings {
                    // 3
                    colorScheme: root.colorScheme

                    onBack: {
                        rightContent.showGeneralSettings();
                    }
                }
                PortSettings {
                    // 4
                    colorScheme: root.colorScheme
                    notifications: root.notifications

                    onBack: {
                        rightContent.showGeneralSettings();
                    }
                }
                ConnectionModeSettings {
                    // 5
                    colorScheme: root.colorScheme

                    onBack: {
                        rightContent.showGeneralSettings();
                    }
                }
                LocalCacheSettings {
                    // 6
                    colorScheme: root.colorScheme
                    notifications: root.notifications

                    onBack: {
                        rightContent.showGeneralSettings();
                    }
                }
                HelpView {
                    // 7
                    colorScheme: root.colorScheme

                    onBack: {
                        rightContent.showAccount();
                    }
                }
                BugReportFlow {
                    // 8
                    id: bugReport
                    colorScheme: root.colorScheme
                    selectedAddress: {
                        if (accounts.currentIndex < 0)
                            return "";
                        if (Backend.users.count === 0)
                            return "";
                        const user = Backend.users.get(accounts.currentIndex);
                        if (!user)
                            return "";
                        return user.addresses[0];
                    }

                    onBack: {
                        rightContent.showHelpView();
                    }
                    onBugReportWasSent: {
                        rightContent.showAccount();
                    }
                }
                Connections {
                    function onLoginAlreadyLoggedIn(index) {
                        rightContent.showAccount(index);
                    }
                    function onLoginFinished(index) {
                        rightContent.showAccount(index);
                    }

                    target: Backend
                }
            }
        }
    }
}