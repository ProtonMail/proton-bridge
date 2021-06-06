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

import "tests"

Window {
    id: root
    title: "ProtonMail Bridge"

    width: 960
    height: 576

    minimumHeight: contentLayout.implicitHeight
    minimumWidth: contentLayout.implicitWidth

    property var colorScheme: ProtonStyle.currentStyle

    property var backend
    property var users


    property bool isNoUser: backend.users.count === 0
    property bool isNoLoggedUser: backend.users.count === 1 && backend.users.get(0).loggedIn === false
    property bool showSetup: true

    signal login(string username, string password)
    signal login2FA(string username, string code)
    signal login2Password(string username, string password)
    signal loginAbort(string username)

    StackLayout {
        id: contentLayout

        anchors.fill: parent

        currentIndex: (root.isNoUser || root.isNoLoggedUser) ? 0 : ( root.showSetup ? 1 : 2)

        WelcomeWindow {
            colorScheme: root.colorScheme
            backend: root.backend
            window: root
            enabled: !banners.blocking

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
            colorScheme: root.colorScheme
            window: root
            enabled: !banners.blocking

            Layout.fillHeight: true
            Layout.fillWidth: true
        }

        ContentWrapper {
            colorScheme: root.colorScheme
            window: root
            enabled: !banners.blocking

            Layout.fillHeight: true
            Layout.fillWidth: true
        }
    }

    Banners {
        id: banners
        anchors.fill: parent
        window: root
        onTop: contentLayout.currentIndex == 0
    }

    function notifyOnlyPaidUsers()            { banners.notifyOnlyPaidUsers()            }
    function notifyConnectionLostWhileLogin() { banners.notifyConnectionLostWhileLogin() }
    function notifyUpdateManually()           { banners.notifyUpdateManually()           }
    function notifyUserAdded()                { banners.notifyUserAdded()                }

    function showSetupGuide(user)   {
        setupGuide.user = user
        root.showSetup = true
    }
}
