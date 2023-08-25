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
import QtQml
import QtQuick
import QtQuick.Window
import QtQuick.Layouts
import QtQuick.Controls
import Proton
import Notifications
import "SetupWizard"

ApplicationWindow {
    id: root

    property int _defaultHeight: 780
    property int _defaultWidth: 1080
    property var notifications

    function layoutForUserCount(userCount) {
        if (userCount === 0) {
            contentLayout.currentIndex = 1;
            setupWizard.showOnboarding();
            return;
        }
        const u = Backend.users.get(0);
        if (!u) {
            console.trace();
            return;
        }
        if ((userCount === 1) && (u.state === EUserState.SignedOut)) {
            contentLayout.currentIndex = 1;
            setupWizard.showLogin(u.primaryEmailOrUsername());
        }
    }
    function selectUser(userID) {
        contentWrapper.selectUser(userID);
    }
    function showAndRise() {
        root.show();
        root.raise();
        if (!root.active) {
            root.requestActivate();
        }
    }
    function showClientConfigurator(user, address) {
        contentLayout.currentIndex = 1;
        setupWizard.showClientConfig(user, address);
    }
    function showHelp() {
        Backend.showWebFrameWindow("https://proton.me/support/bridge");
    }
    function showLocalCacheSettings() {
        contentWrapper.showLocalCacheSettings();
    }
    function showLogin(username = "") {
        contentLayout.currentIndex = 1;
        setupWizard.showLogin(username);
    }
    function showSettings() {
        contentWrapper.showSettings();
    }
    function showWebFrameOverlay(url) {
        webFrameOverlay.visible = true;
        webFrameOverlay.url = url;
    }

    colorScheme: ProtonStyle.currentStyle
    height: _defaultHeight
    minimumWidth: _defaultWidth
    visible: true
    width: _defaultWidth

    Component.onCompleted: {
        layoutForUserCount(Backend.users.count);
    }

    // show Setup Guide on every new user
    Connections {
        function onRowsAboutToBeRemoved(parent, first, last) {
            for (let i = first; i <= last; i++) {
                const user = Backend.users.get(i);
                if (setupWizard.user === user) {
                    setupWizard.closeWizard();
                }
            }
        }
        function onRowsInserted(parent, first, _) {
            // considering that users are added one-by-one
            const user = Backend.users.get(first);
            if (user.state === EUserState.SignedOut) {
                return;
            }
            if (user.setupGuideSeen) {
                return;
            }
            root.showClientConfigurator(user, user.addresses[0]);
        }

        target: Backend.users
    }
    Connections {
        function onSelectUser(userID, forceShowWindow) {
            contentWrapper.selectUser(userID);
            if (forceShowWindow) {
                root.showAndRise();
            }
        }
        function onShowHelp() {
            root.showHelp();
            root.showAndRise();
        }
        function onShowMainWindow() {
            root.showAndRise();
        }
        function onShowSettings() {
            root.showSettings();
            root.showAndRise();
        }

        target: Backend
    }
    Connections {
        function onCountChanged(count) {
            layoutForUserCount(count);
        }

        target: Backend.users
    }
    StackLayout {
        id: contentLayout
        anchors.fill: parent
        currentIndex: 0

        ContentWrapper {
            // 0
            id: contentWrapper
            Layout.fillHeight: true
            Layout.fillWidth: true
            colorScheme: root.colorScheme
            notifications: root.notifications

            onCloseWindow: {
                root.close();
            }
            onQuitBridge: {
                // If we ever want to add a confirmation dialog before quitting:
                //root.notifications.askQuestion("Quit Bridge", "Insert warning message here.", "Quit", "Cancel", Backend.quit, null)
                root.close();
                Backend.quit();
            }
            onShowClientConfigurator: function (user, address) {
                root.showClientConfigurator(user, address);
            }
            onShowLogin: function (username) {
                root.showLogin(username);
            }
        }
        SetupWizard {
            id: setupWizard
            Layout.fillHeight: true
            Layout.fillWidth: true
            colorScheme: root.colorScheme

            onBugReportRequested: {
                contentWrapper.showBugReport();
            }
            onWizardEnded: {
                contentLayout.currentIndex = 0;
            }
        }
    }
    WebFrame {
        id: webFrameOverlay
        anchors.fill: parent
        colorScheme: root.colorScheme
        overlay: true
        url: ""
        visible: false
    }
    NotificationPopups {
        colorScheme: root.colorScheme
        mainWindow: root
        notifications: root.notifications
    }
    SplashScreen {
        id: splashScreen
        colorScheme: root.colorScheme
    }
}
