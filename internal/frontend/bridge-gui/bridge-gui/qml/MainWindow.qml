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
import QtQml
import QtQuick
import QtQuick.Window
import QtQuick.Layouts
import QtQuick.Controls
import Proton
import Notifications
import "SetupWizard"
import "BugReport"

ApplicationWindow {
    id: root

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
    function showClientConfigurator(user, address, justLoggedIn) {
        contentLayout.currentIndex = 1;
        setupWizard.showClientConfig(user, address, justLoggedIn);
    }
    function showHelp() {
        contentWrapper.showHelp();
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

    colorScheme: ProtonStyle.currentStyle
    height: screen.height < ProtonStyle.window_default_height + 100 ? ProtonStyle.window_minimum_height : ProtonStyle.window_default_height
    minimumHeight: ProtonStyle.window_minimum_height
    minimumWidth: ProtonStyle.window_minimum_width
    visible: true
    width: ProtonStyle.window_default_width

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
            root.showClientConfigurator(user, user.addresses[0], false);
        }

        target: Backend.users
    }
    Connections {
        function onSelectUser(userID, forceShowWindow) {
            contentWrapper.selectUser(userID);
            if (setupWizard.visible) {
                setupWizard.closeWizard()
            }
            if (forceShowWindow) {
                root.showAndRise();
            }
        }
        function onShowHelp() {
            root.showHelp();
            if (setupWizard.visible) {
                setupWizard.closeWizard()
            }

            root.showAndRise();
        }
        function onShowMainWindow() {
            root.showAndRise();
        }
        function onShowSettings() {
            if (setupWizard.visible) {
                setupWizard.closeWizard()
            }
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
            onShowClientConfigurator: function (user, address, justLoggedIn) {
                root.showClientConfigurator(user, address, justLoggedIn);
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
