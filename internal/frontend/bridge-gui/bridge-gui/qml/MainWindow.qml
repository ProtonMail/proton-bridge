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
    function showHelp() {
        showWebViewOverlay("https://proton.me/support/bridge");
    }
    function showLocalCacheSettings() {
        contentWrapper.showLocalCacheSettings();
    }
    function showSettings() {
        contentWrapper.showSettings();
    }
    function showSetup(user, address) {
        setupGuide.user = user;
        setupGuide.address = address;
        setupGuide.reset();
        contentLayout._showSetup = !!setupGuide.user;
    }
    function showSignIn(username) {
        if (contentLayout.currentIndex === 1)
            return;
        contentWrapper.showSignIn(username);
    }
    function showWebViewOverlay(url) {
        webViewOverlay.visible = true;
        webViewOverlay.url = url;
    }

    colorScheme: ProtonStyle.currentStyle
    height: _defaultHeight
    minimumWidth: _defaultWidth
    visible: true
    width: _defaultWidth

    // show Setup Guide on every new user
    Connections {
        function onRowsAboutToBeRemoved(parent, first, last) {
            for (let i = first; i <= last; i++) {
                const user = Backend.users.get(i);
                if (setupGuide.user === user) {
                    setupGuide.user = null;
                    contentLayout._showSetup = false;
                    return;
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
            root.showSetup(user, user.addresses[0]);
        }

        target: Backend.users
    }
    Connections {
        function onLoginFinished(index, wasSignedOut) {
            // const user = Backend.users.get(index);
            // if (user && !wasSignedOut) {
            //    root.showSetup(user, user.addresses[0]);
            // }
            // console.debug("Login finished", index);
        }
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
    StackLayout {
        id: contentLayout

        property bool _showSetup: false

        anchors.fill: parent
        currentIndex: {
            // show welcome when there are no users
            if (Backend.users.count === 0) {
                setupWizard.start();
                return 0;
            }
            const u = Backend.users.get(0);
            if (!u) {
                console.trace();
                console.log("empty user");
                return 1;
            }
            if ((Backend.users.count === 1) && (u.state === EUserState.SignedOut)) {
                showSignIn(u.primaryEmailOrUsername());
                return 0;
            }
            if (contentLayout._showSetup) {
                return 2;
            }
            return 0;
        }

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
            onShowSetupGuide: function (user, address) {
                setupWizard.startClientConfig();
            }
            onShowSetupWizard: {
                setupWizard.start();
            }
        }
        WelcomeGuide {
            Layout.fillHeight: true
            Layout.fillWidth: true // 1
            colorScheme: root.colorScheme
        }
        SetupGuide {
            // 2
            id: setupGuide
            Layout.fillHeight: true
            Layout.fillWidth: true
            colorScheme: root.colorScheme

            onDismissed: {
                root.showSetup(null, "");
            }
            onFinished: {
                // TODO: Do not close window. Trigger Backend to check that
                // there is a successfully connected client. Then Backend
                // should send another signal to close the setup guide.
                root.showSetup(null, "");
            }
        }
    }
    WebView {
        id: webViewOverlay
        anchors.fill: parent
        colorScheme: root.colorScheme
        overlay: true
        url: ""
        visible: false
    }
    SetupWizard {
        id: setupWizard
        anchors.fill: parent
        colorScheme: root.colorScheme
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
