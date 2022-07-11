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
import QtQuick.Layouts 1.12
import QtQuick.Controls 2.13

import Proton 4.0

ColumnLayout {
    id: root

    property var user
    property var userIndex
    property var backend

    spacing : 5

    Layout.fillHeight: true
    //Layout.fillWidth: true

    property ColorScheme colorScheme

    TextField {
        colorScheme: root.colorScheme
        Layout.fillWidth: true

        text: user !== undefined ? user.username : ""

        onEditingFinished: {
            user.username = text
        }
    }

    ColumnLayout {
        Layout.fillWidth: true

        Switch {
            id: userLoginSwitch
            colorScheme: root.colorScheme

            text: "LoggedIn"
            enabled: user !== undefined && user.username.length > 0

            checked: user ? user.loggedIn : false

            onCheckedChanged: {
                if (!user) {
                    return
                }

                if (checked) {
                    if (user === backend.loginUser) {
                        var newUserObject = backend.userComponent.createObject(backend, {username: user.username, loggedIn: true, setupGuideSeen: user.setupGuideSeen})
                        backend.users.append( { object: newUserObject } )

                        user.username = ""
                        user.resetLoginRequests()
                        return
                    }

                    user.loggedIn = true
                    user.resetLoginRequests()
                    return
                } else {
                    user.loggedIn = false
                    user.resetLoginRequests()
                }
            }
        }

        Switch {
            colorScheme: root.colorScheme

            text: "Setup guide seen"
            enabled: user !== undefined && user.username.length > 0

            checked: user ? user.setupGuideSeen : false

            onCheckedChanged: {
                if (!user) {
                    return
                }

                user.setupGuideSeen = checked
            }
        }
    }


    RowLayout {
        Layout.fillWidth: true

        Label {
            colorScheme: root.colorScheme
            id: loginLabel
            text: "Login:"

            Layout.preferredWidth: Math.max(loginLabel.implicitWidth, faLabel.implicitWidth, passLabel.implicitWidth)
        }

        Button {
            colorScheme: root.colorScheme
            text: "name/pass error"
            enabled: user !== undefined //&& user.isLoginRequested && !user.isLogin2FARequested && !user.isLogin2PasswordProvided

            onClicked: {
                root.backend.loginUsernamePasswordError("")
                user.resetLoginRequests()
            }
        }

        Button {
            colorScheme: root.colorScheme
            text: "free user error"
            enabled: user !== undefined //&& user.isLoginRequested
            onClicked: {
                root.backend.loginFreeUserError()
                user.resetLoginRequests()
            }
        }

        Button {
            colorScheme: root.colorScheme
            text: "connection error"
            enabled: user !== undefined //&& user.isLoginRequested
            onClicked: {
                root.backend.loginConnectionError("")
                user.resetLoginRequests()
            }
        }
    }

    RowLayout {
        Layout.fillWidth: true

        Label {
            colorScheme: root.colorScheme
            id: faLabel
            text: "2FA:"

            Layout.preferredWidth: Math.max(loginLabel.implicitWidth, faLabel.implicitWidth, passLabel.implicitWidth)
        }

        Button {
            colorScheme: root.colorScheme
            text: "request"

            enabled: user !== undefined //&& user.isLoginRequested && !user.isLogin2FARequested && !user.isLogin2PasswordRequested
            onClicked: {
                root.backend.login2FARequested(user.username)
                user.isLogin2FARequested = true
            }
        }

        Button {
            colorScheme: root.colorScheme
            text: "error"

            enabled: user !== undefined //&& user.isLogin2FAProvided && !(user.isLogin2PasswordRequested && !user.isLogin2PasswordProvided)
            onClicked: {
                root.backend.login2FAError("")
                user.isLogin2FAProvided = false
            }
        }

        Button {
            colorScheme: root.colorScheme
            text: "Abort"

            enabled: user !== undefined //&& user.isLogin2FAProvided && !(user.isLogin2PasswordRequested && !user.isLogin2PasswordProvided)
            onClicked: {
                root.backend.login2FAErrorAbort("")
                user.resetLoginRequests()
            }
        }
    }

    RowLayout {
        Layout.fillWidth: true

        Label {
            colorScheme: root.colorScheme
            id: passLabel
            text: "2 Password:"

            Layout.preferredWidth: Math.max(loginLabel.implicitWidth, faLabel.implicitWidth, passLabel.implicitWidth)
        }

        Button {
            colorScheme: root.colorScheme
            text: "request"

            enabled: user !== undefined //&& user.isLoginRequested && !user.isLogin2PasswordRequested && !(user.isLogin2FARequested && !user.isLogin2FAProvided)
            onClicked: {
                root.backend.login2PasswordRequested("")
                user.isLogin2PasswordRequested = true
            }
        }

        Button {
            colorScheme: root.colorScheme
            text: "error"

            enabled: user !== undefined //&& user.isLogin2PasswordProvided && !(user.isLogin2FARequested && !user.isLogin2FAProvided)
            onClicked: {
                root.backend.login2PasswordError("")

                user.isLogin2PasswordProvided = false
            }
        }

        Button {
            colorScheme: root.colorScheme
            text: "Abort"

            enabled: user !== undefined //&& user.isLogin2PasswordProvided && !(user.isLogin2FARequested && !user.isLogin2FAProvided)
            onClicked: {
                root.backend.login2PasswordErrorAbort("")
                user.resetLoginRequests()
            }
        }
    }

    RowLayout {
        Button {
            colorScheme: root.colorScheme
            text: "Login Finished"

            onClicked: {
                root.backend.loginFinished(0+loginFinishedIndex.text)
                user.resetLoginRequests()
            }
        }
        TextField {
            id: loginFinishedIndex
            colorScheme: root.colorScheme
            label: "Index:"
            text: root.userIndex
        }
    }

    RowLayout {
        Button {
            colorScheme: root.colorScheme
            text: "Already logged in"

            onClicked: {
                root.backend.loginAlreadyLoggedIn(0+loginAlreadyLoggedInIndex.text)
                user.resetLoginRequests()
            }
        }
        TextField {
            id: loginAlreadyLoggedInIndex
            colorScheme: root.colorScheme
            label: "Index:"
            text: root.userIndex
        }
    }

    RowLayout {
        TextField {
            colorScheme: root.colorScheme
            label: "used:"
            text: user && user.usedBytes ? user.usedBytes : 0
            onEditingFinished: {
                user.usedBytes = parseFloat(text)
            }
            implicitWidth: 200
        }
        TextField {
            colorScheme: root.colorScheme
            label: "total:"
            text: user && user.totalBytes ? user.totalBytes : 0
            onEditingFinished: {
                user.totalBytes = parseFloat(text)
            }
            implicitWidth: 200
        }
    }

    RowLayout {
        Label {colorScheme: root.colorScheme; text: "Split mode"}
        Toggle { colorScheme: root.colorScheme; checked: user ? user.splitMode : false; onClicked: {user.splitMode = !user.splitMode}}
        Button { colorScheme: root.colorScheme; text: "Toggle Finished"; onClicked: {user.toggleSplitModeFinished()}}
    }

    TextArea { // TODO: this is causing binding loop on imlicitWidth
        colorScheme: root.colorScheme
        text: user && user.addresses ? user.addresses.join("\n") : "user@protonmail.com"
        Layout.fillWidth: true

        onEditingFinished: {
            user.addresses = text.split("\n")
        }
    }

    Item {
        Layout.fillHeight: true
    }
}
