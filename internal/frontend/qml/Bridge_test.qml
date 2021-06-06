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
import QtQuick.Controls 2.13

import QtQml.Models 2.12

import Proton 4.0

import "./BridgeTest"

Window {
    id: root

    width: 640
    height: 480
    x: 100
    y: 100

    property var colorScheme: ProtonStyle.darkStyle

    flags   : Qt.Window | Qt.Dialog
    visible : true
    title   : "Bridge Test GUI"
    color   : colorScheme.background_norm

    function _log(msg, color) {
        logTextArea.text += "<p style='color: " + color + ";'>" + msg + "</p>"
        logTextArea.text += "\n"
    }

    function log(msg) {
        console.log(msg)
        _log(msg, root.colorScheme.signal_info)
    }

    function error(msg) {
        console.error(msg)
        _log(msg, root.colorScheme.signal_danger)
    }

    // No user object should be put in this list until a successful login
    property var users: UserModel {
        id: _users

        onRowsInserted: {
            for (var i = first; i <= last; i++) {
                _usersTest.insert(i + 1, { object: get(i) } )
            }
        }

        onRowsRemoved: {
            _usersTest.remove(first + 1, first - last + 1)
        }

        onRowsMoved: {
            _usersTest.move(start + 1, row + 1, end - start + 1)
        }

        onDataChanged: {
            for (var i = topLeft.row; i <= bottomRight.row; i++) {
                _usersTest.set(i + 1, { object: get(i) } )
            }
        }
    }

    // this list is used on test gui: it contains same users list as users above + fake user to represent login request of new user on pos 0
    property var usersTest: UserModel {
        id: _usersTest
    }

    property var userComponent: Component {
        id: _userComponent

        QtObject {
            property string username: ""
            property bool loggedIn: false

            signal loginUsernamePasswordError()
            signal loginFreeUserError()
            signal loginConnectionError()
            signal login2FARequested()
            signal login2FAError()
            signal login2FAErrorAbort()
            signal login2PasswordRequested()
            signal login2PasswordError()
            signal login2PasswordErrorAbort()

            // Test purpose only:
            property bool isFakeUser: this === root.loginUser

            function userSignal(msg) {
                if (isFakeUser) {
                    return
                }

                root.log("<- User (" + username + "): " + msg)
            }

            onLoginUsernamePasswordError: {
                userSignal("loginUsernamePasswordError")
            }
            onLoginFreeUserError: {
                userSignal("loginFreeUserError")
            }
            onLoginConnectionError: {
                userSignal("loginConnectionError")
            }
            onLogin2FARequested: {
                userSignal("login2FARequested")
            }
            onLogin2FAError: {
                userSignal("login2FAError")
            }
            onLogin2FAErrorAbort: {
                userSignal("login2FAErrorAbort")
            }
            onLogin2PasswordRequested: {
                userSignal("login2PasswordRequested")
            }
            onLogin2PasswordError: {
                userSignal("login2PasswordError")
            }
            onLogin2PasswordErrorAbort: {
                userSignal("login2PasswordErrorAbort")
            }

            function resetLoginRequests() {
                isLoginRequested = false
                isLogin2FARequested = false
                isLogin2FAProvided = false
                isLogin2PasswordRequested = false
                isLogin2PasswordProvided = false
            }

            property bool isLoginRequested: false

            property bool isLogin2FARequested: false
            property bool isLogin2FAProvided: false

            property bool isLogin2PasswordRequested: false
            property bool isLogin2PasswordProvided: false
        }
    }

    // this it fake user used only for representing first login request
    property var loginUser
    Component.onCompleted: {
        var newLoginUser = _userComponent.createObject()
        root.loginUser = newLoginUser
        _usersTest.append({object: newLoginUser})

        newLoginUser.loginUsernamePasswordError.connect(root.loginUsernamePasswordError)
        newLoginUser.loginFreeUserError.connect(root.loginFreeUserError)
        newLoginUser.loginConnectionError.connect(root.loginConnectionError)
        newLoginUser.login2FARequested.connect(root.login2FARequested)
        newLoginUser.login2FAError.connect(root.login2FAError)
        newLoginUser.login2FAErrorAbort.connect(root.login2FAErrorAbort)
        newLoginUser.login2PasswordRequested.connect(root.login2PasswordRequested)
        newLoginUser.login2PasswordError.connect(root.login2PasswordError)
        newLoginUser.login2PasswordErrorAbort.connect(root.login2PasswordErrorAbort)
    }


    TabBar {
        id: tabBar
        anchors.left: parent.left
        anchors.right: parent.right

        TabButton {
            text: "Global settings"
        }

        TabButton {
            text: "User control"
        }

        TabButton {
            text: "Playground"
        }

        TabButton {
            text: "Log"
        }
    }

    StackLayout {
        anchors.top: tabBar.bottom
        anchors.left: parent.left
        anchors.right: parent.right
        anchors.bottom: parent.bottom

        currentIndex: tabBar.currentIndex

        anchors.margins: 10

        RowLayout {
            id: globalTab
            spacing : 5
            property alias colorScheme: root.colorScheme

            ColumnLayout {
                spacing : 5

                property alias colorScheme: globalTab.colorScheme

                ProtonLabel {
                    text: "Global settings"
                    color: globalTab.colorScheme.text_norm
                }

                ButtonGroup {
                    id: styleRadioGroup
                }

                RadioButton {
                    Layout.fillWidth: true

                    text: "Light UI"
                    checked: ProtonStyle.currentStyle === ProtonStyle.lightStyle
                    ButtonGroup.group: styleRadioGroup

                    onCheckedChanged: {
                        if (checked && ProtonStyle.currentStyle !== ProtonStyle.lightStyle) {
                            ProtonStyle.currentStyle = ProtonStyle.lightStyle
                        }
                    }
                }

                RadioButton {
                    Layout.fillWidth: true

                    text: "Dark UI"
                    checked: ProtonStyle.currentStyle === ProtonStyle.darkStyle
                    ButtonGroup.group: styleRadioGroup

                    onCheckedChanged: {
                        if (checked && ProtonStyle.currentStyle !== ProtonStyle.darkStyle) {
                            ProtonStyle.currentStyle = ProtonStyle.darkStyle
                        }
                    }


                }

                Button {
                    //Layout.fillWidth: true

                    text: "Open Bridge"
                    enabled: bridge === undefined || bridge === null
                    onClicked: {
                        bridge = bridgeComponent.createObject()
                    }
                }

                Button {
                    //Layout.fillWidth: true

                    text: "Close Bridge"
                    enabled: bridge !== undefined && bridge !== null
                    onClicked: {
                        bridge.destroy()
                    }
                }

                Item {
                    Layout.fillHeight: true
                }
            }

            ColumnLayout {
                spacing : 5

                property alias colorScheme: globalTab.colorScheme

                ProtonLabel {
                    text: "Notifications"
                    color: globalTab.colorScheme.text_norm
                }

                Button {
                    text: "Notify: danger"
                    enabled: bridge !== undefined && bridge !== null
                    onClicked: {
                        bridge.mainWindow.notifyOnlyPaidUsers()
                    }
                }

                Button {
                    text: "Notify: warning"
                    enabled: bridge !== undefined && bridge !== null
                    onClicked: {
                        bridge.mainWindow.notifyUpdateManually()
                    }
                }

                Button {
                    text: "Notify: success"
                    enabled: bridge !== undefined && bridge !== null
                    onClicked: {
                        bridge.mainWindow.notifyUserAdded()
                    }
                }



                Item {
                    Layout.fillHeight: true
                }
            }
        }

        RowLayout {
            id: usersTab
            UserList {
                id: usersListView
                Layout.fillHeight: true
                colorScheme: root.colorScheme
                backend: root
            }

            UserControl {
                colorScheme: root.colorScheme
                backend: root
                user: ((root.usersTest.count > usersListView.currentIndex) && usersListView.currentIndex != -1) ? root.usersTest.get(usersListView.currentIndex) : undefined
            }
        }

        RowLayout {
            id: playgroundTab

            property var colorScheme: root.colorScheme

            AccountDelegate{}
        }

        TextArea {
            id: logTextArea
            Layout.fillHeight: true
            Layout.fillWidth: true

            Layout.preferredWidth: 400
            Layout.preferredHeight: 200

            textFormat: TextEdit.RichText
            //readOnly: true
        }
    }

    property var bridge

    // this signals are used only when trying to login with new user (i.e. not in users model)
    signal loginUsernamePasswordError()
    signal loginFreeUserError()
    signal loginConnectionError()
    signal login2FARequested()
    signal login2FAError()
    signal login2FAErrorAbort()
    signal login2PasswordRequested()
    signal login2PasswordError()
    signal login2PasswordErrorAbort()

    onLoginUsernamePasswordError: {
        console.debug("<- loginUsernamePasswordError")
    }
    onLoginFreeUserError: {
        console.debug("<- loginFreeUserError")
    }
    onLoginConnectionError: {
        console.debug("<- loginConnectionError")
    }
    onLogin2FARequested: {
        console.debug("<- login2FARequested")
    }
    onLogin2FAError: {
        console.debug("<- login2FAError")
    }
    onLogin2FAErrorAbort: {
        console.debug("<- login2FAErrorAbort")
    }
    onLogin2PasswordRequested: {
        console.debug("<- login2PasswordRequested")
    }
    onLogin2PasswordError: {
        console.debug("<- login2PasswordError")
    }
    onLogin2PasswordErrorAbort: {
        console.debug("<- login2PasswordErrorAbort")
    }

    Component {
        id: bridgeComponent

        Bridge {
            backend: root

            onLogin: {
                root.log("-> login(" + username + ", " + password + ")")

                loginUser.username = username
                loginUser.isLoginRequested = true
            }

            onLogin2FA: {
                root.log("-> login2FA(" + username + ", " + code + ")")

                loginUser.isLogin2FAProvided = true
            }

            onLogin2Password: {
                root.log("-> login2FA(" + username + ", " + password + ")")

                loginUser.isLogin2PasswordProvided = true
            }

            onLoginAbort: {
                root.log("-> loginAbort(" + username + ")")

                loginUser.resetLoginRequests()
            }
        }
    }

    onClosing: {
        Qt.quit()
    }
}
