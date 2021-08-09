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
    title: "ProtonMail Bridge"

    width: 960
    height: 576

    minimumHeight: contentLayout.implicitHeight
    minimumWidth: contentLayout.implicitWidth

    colorScheme: ProtonStyle.currentStyle

    property var backend
    property var notifications

    signal login(string username, string password)
    signal login2FA(string username, string code)
    signal login2Password(string username, string password)
    signal loginAbort(string username)

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

    StackLayout {
        id: contentLayout

        anchors.fill: parent

        property bool _showSetup: false
        currentIndex: {
            // show welcome when there are no users or only one non-logged-in user is present
            if (backend.users.count === 0) {
                return 1
            }

            if (backend.users.count === 1 && backend.users.get(0).loggedIn === false) {
                return 1
            }

            if (contentLayout._showSetup) {
                return 2
            }

            return 0
        }

        ContentWrapper {
            id: contentWrapper
            colorScheme: root.colorScheme
            backend: root.backend
            notifications: root.notifications

            Layout.fillHeight: true
            Layout.fillWidth: true

            onShowSetupGuide: {
                root.showSetup(user,address)
            }

            onLogin: {
                root.login(username, password)
            }
            onLogin2FA: {
                root.login2FA(username, code)
            }
            onLogin2Password: {
                root.login2Password(username, password)
            }
            onLoginAbort: {
                root.loginAbort(username)
            }
        }

        WelcomeGuide {
            colorScheme: root.colorScheme
            backend: root.backend

            Layout.fillHeight: true
            Layout.fillWidth: true

            onLogin: {
                root.login(username, password)
            }
            onLogin2FA: {
                root.login2FA(username, code)
            }
            onLogin2Password: {
                root.login2Password(username, password)
            }
            onLoginAbort: {
                root.loginAbort(username)
            }
        }

        SetupGuide {
            id: setupGuide
            colorScheme: root.colorScheme
            backend: root.backend

            Layout.fillHeight: true
            Layout.fillWidth: true

            onDismissed: {
                root.showSetup(null,"")
            }
        }
    }

    NotificationPopups {
        colorScheme: root.colorScheme
        notifications: root.notifications
        mainWindow: root
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
        if (setupGuide.user) {
            contentLayout._showSetup = true
        } else {
            contentLayout._showSetup = false
        }
    }
}
