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

import QtQml 2.12
import QtQuick 2.13
import QtQuick.Window 2.13
import QtQuick.Layouts 1.12
import QtQuick.Controls 2.12

import Proton 4.0
import Notifications 1.0

import "tests"

ApplicationWindow {
    id: root

    width: 960
    height: 576

    minimumHeight: contentLayout.implicitHeight
    minimumWidth: contentLayout.implicitWidth

    colorScheme: ProtonStyle.currentStyle

    property var backend
    property var notifications

    // This is needed because on MacOS if first window shown is not transparent -
    // all other windows of application will not have transparent background (black
    // instead of transparency). In our case that mean that if MainWindow will be
    // shown before StatusWindow - StatusWindow will not have transparent corners.
    color: "transparent"

    // show Setup Guide on every new user
    Connections {
        target: root.backend.users

        onRowsInserted: {
            // considerring that users are added one-by-one
            var user = root.backend.users.get(first)

            if (!user.loggedIn) {
                return
            }

            if (user.setupGuideSeen) {
                return
            }

            root.showSetup(user,user.addresses[0])
        }

        onRowsAboutToBeRemoved: {
            for (var i = first; i <= last; i++ ) {
                var user = root.backend.users.get(i)

                if (setupGuide.user === user) {
                    setupGuide.user = null
                    contentLayout._showSetup = false
                    return
                }
            }
        }
    }

    Connections {
        target: root.backend

        onShowMainWindow: {
            root.showAndRise()
        }

        onLoginFinished: {
            console.debug("Login finished", index)
        }
    }

    StackLayout {
        id: contentLayout

        anchors.fill: parent

        property bool _showSetup: false
        currentIndex: {
            // show welcome when there are no users or only one non-logged-in user is present
            if (backend.users.count === 0) {
                return 1
            }

            var u = backend.users.get(0)

            if (!u) {
                console.trace()
                console.log("empty user")
                return 1
            }

            if (backend.users.count === 1 && u.loggedIn === false) {
                return 1
            }

            if (contentLayout._showSetup) {
                return 2
            }

            return 0
        }

        ContentWrapper { // 0
            id: contentWrapper
            colorScheme: root.colorScheme
            backend: root.backend
            notifications: root.notifications

            Layout.fillHeight: true
            Layout.fillWidth: true

            onShowSetupGuide: {
                root.showSetup(user,address)
            }
        }

        WelcomeGuide { // 1
            colorScheme: root.colorScheme
            backend: root.backend

            Layout.fillHeight: true
            Layout.fillWidth: true
        }

        SetupGuide { // 2
            id: setupGuide
            colorScheme: root.colorScheme
            backend: root.backend

            Layout.fillHeight: true
            Layout.fillWidth: true

            onDismissed: {
                root.showSetup(null,"")
            }

            onFinished: {
                // TODO: Do not close window. Trigger backend to check that
                // there is a successfully connected client. Then backend
                // should send another signal to close the setup guide.
                root.showSetup(null,"")
            }
        }

    }

    NotificationPopups {
        colorScheme: root.colorScheme
        notifications: root.notifications
        mainWindow: root
        backend: root.backend
    }

    SplashScreen {
        id: splashScreen
        colorScheme: root.colorScheme
        backend: root.backend
    }

    function showLocalCacheSettings() { contentWrapper.showLocalCacheSettings() }
    function showSettings() { contentWrapper.showSettings() }
    function showHelp() { contentWrapper.showHelp() }

    function showSignIn(username) {
        if (contentLayout.currentIndex == 1) return
        contentWrapper.showSignIn(username)
    }

    function showSetup(user, address) {
        setupGuide.user = user
        setupGuide.address = address
        setupGuide.reset()
        if (setupGuide.user) {
            contentLayout._showSetup = true
        } else {
            contentLayout._showSetup = false
        }
    }

    function showAndRise() {
        root.show()
        root.raise()
        if (!root.active) {
            root.requestActivate()
        }
    }
}
