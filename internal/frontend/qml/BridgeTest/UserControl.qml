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
import QtQuick.Layouts 1.12
import QtQuick.Controls 2.13

ColumnLayout {
    property var user
    property var backend

    spacing : 5

    Layout.fillHeight: true
    //Layout.fillWidth: true

    property var colorScheme

    TextField {
        Layout.fillWidth: true

        text: user !== undefined ? user.username : ""

        onEditingFinished: {
            user.username = text
        }
    }

    RowLayout {
        Layout.fillWidth: true

        Button {
            //Layout.fillWidth: true

            text: "Login"
            enabled: user !== undefined && !user.loggedIn && user.username.length > 0

            onClicked: {
                if (user === backend.loginUser) {
                    var newUserObject = backend.userComponent.createObject(backend, {username: user.username, loggedIn: true})
                    backend.users.append( { object: newUserObject } )

                    user.username = ""
                    user.resetLoginRequests()
                    return
                }

                user.loggedIn = true
                user.resetLoginRequests()
            }
        }

        Button {
            //Layout.fillWidth: true

            text: "Logout"
            enabled: user !== undefined && user.loggedIn && user.username.length > 0

            onClicked: {
                user.loggedIn = false
                user.resetLoginRequests()
            }
        }
    }


    RowLayout {
        Layout.fillWidth: true

        Label {
            id: loginLabel
            text: "Login:"

            Layout.preferredWidth: Math.max(loginLabel.implicitWidth, faLabel.implicitWidth, passLabel.implicitWidth)
        }

        Button {
            text: "name/pass error"
            enabled: user !== undefined && user.isLoginRequested && !user.isLogin2FARequested && !user.isLogin2PasswordProvided

            onClicked: {
                user.loginUsernamePasswordError()
                user.resetLoginRequests()
            }
        }

        Button {
            text: "free user error"
            enabled: user !== undefined && user.isLoginRequested
            onClicked: {
                user.loginFreeUserError()
                user.resetLoginRequests()
            }
        }

        Button {
            text: "connection error"
            enabled: user !== undefined && user.isLoginRequested
            onClicked: {
                user.loginConnectionError()
                user.resetLoginRequests()
            }
        }
    }

    RowLayout {
        Layout.fillWidth: true

        Label {
            id: faLabel
            text: "2FA:"

            Layout.preferredWidth: Math.max(loginLabel.implicitWidth, faLabel.implicitWidth, passLabel.implicitWidth)
        }

        Button {
            text: "request"

            enabled: user !== undefined && user.isLoginRequested && !user.isLogin2FARequested && !user.isLogin2PasswordRequested
            onClicked: {
                user.login2FARequested()
                user.isLogin2FARequested = true
            }
        }

        Button {
            text: "error"

            enabled: user !== undefined && user.isLogin2FAProvided && !(user.isLogin2PasswordRequested && !user.isLogin2PasswordProvided)
            onClicked: {
                user.login2FAError()
                user.isLogin2FAProvided = false
            }
        }

        Button {
            text: "Abort"

            enabled: user !== undefined && user.isLogin2FAProvided && !(user.isLogin2PasswordRequested && !user.isLogin2PasswordProvided)
            onClicked: {
                user.login2FAErrorAbort()
                user.resetLoginRequests()
            }
        }
    }

    RowLayout {
        Layout.fillWidth: true

        Label {
            id: passLabel
            text: "2 Password:"

            Layout.preferredWidth: Math.max(loginLabel.implicitWidth, faLabel.implicitWidth, passLabel.implicitWidth)
        }

        Button {
            text: "request"

            enabled: user !== undefined && user.isLoginRequested && !user.isLogin2PasswordRequested && !(user.isLogin2FARequested && !user.isLogin2FAProvided)
            onClicked: {
                user.login2PasswordRequested()
                user.isLogin2PasswordRequested = true
            }
        }

        Button {
            text: "error"

            enabled: user !== undefined && user.isLogin2PasswordProvided && !(user.isLogin2FARequested && !user.isLogin2FAProvided)
            onClicked: {
                user.login2PasswordError()

                user.isLogin2PasswordProvided = false
            }
        }

        Button {
            text: "Abort"

            enabled: user !== undefined && user.isLogin2PasswordProvided && !(user.isLogin2FARequested && !user.isLogin2FAProvided)
            onClicked: {
                user.login2PasswordErrorAbort()
                user.resetLoginRequests()
            }
        }
    }


    Item {
        Layout.fillHeight: true
    }
}
