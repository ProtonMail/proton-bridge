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
import QtQuick.Controls 2.12
import QtQuick.Controls.impl 2.12

import Proton 4.0

Item {
    id: root
    property ColorScheme colorScheme

    function reset() {
        stackLayout.currentIndex = 0
        loginNormalLayout.reset()
        login2FALayout.reset()
        login2PasswordLayout.reset()

    }

    function abort() {
        root.reset()
        root.backend.loginAbort(usernameTextField.text)
    }

    implicitHeight: children[0].implicitHeight
    implicitWidth: children[0].implicitWidth

    property var backend

    property alias username: usernameTextField.text
    state: "Page 1"

    property alias currentIndex: stackLayout.currentIndex

    StackLayout {
        id: stackLayout
        anchors.fill: parent

        function loginFailed() {
            signInButton.loading = false

            usernameTextField.enabled = true
            usernameTextField.error = true

            passwordTextField.enabled = true
            passwordTextField.error = true
        }

        Connections {
            target: root.backend

            onLoginUsernamePasswordError: {
                console.assert(stackLayout.currentIndex == 0, "Unexpected loginUsernamePasswordError")
                console.assert(signInButton.loading == true, "Unexpected loginUsernamePasswordError")

                stackLayout.loginFailed()
                if (errorMsg!="") errorLabel.text = errorMsg
                else errorLabel.text = qsTr("Incorrect login credentials")
            }

            onLoginFreeUserError: {
                console.assert(stackLayout.currentIndex == 0, "Unexpected loginFreeUserError")
                stackLayout.loginFailed()
            }

            onLoginConnectionError: {
                if (stackLayout.currentIndex == 0 ) {
                    stackLayout.loginFailed()
                }
            }

            onLogin2FARequested: {
                console.assert(stackLayout.currentIndex == 0, "Unexpected login2FARequested")
                twoFactorUsernameLabel.text = username
                stackLayout.currentIndex = 1
            }
            onLogin2FAError: {
                console.assert(stackLayout.currentIndex == 1, "Unexpected login2FAError")

                twoFAButton.loading = false

                twoFactorPasswordTextField.enabled = true
                twoFactorPasswordTextField.error = true
                twoFactorPasswordTextField.assistiveText = qsTr("Your code is incorrect")
            }
            onLogin2FAErrorAbort: {
                console.assert(stackLayout.currentIndex == 1, "Unexpected login2FAErrorAbort")
                root.reset()
                errorLabel.text = qsTr("Incorrect login credentials. Please try again.")
            }

            onLogin2PasswordRequested: {
                console.assert(stackLayout.currentIndex == 0 || stackLayout.currentIndex == 1, "Unexpected login2PasswordRequested")
                stackLayout.currentIndex = 2
            }
            onLogin2PasswordError: {
                console.assert(stackLayout.currentIndex == 2, "Unexpected login2PasswordError")

                secondPasswordButton.loading = false

                secondPasswordTextField.enabled = true
                secondPasswordTextField.error = true
                secondPasswordTextField.assistiveText = qsTr("Your mailbox password is incorrect")
            }
            onLogin2PasswordErrorAbort: {
                console.assert(stackLayout.currentIndex == 2, "Unexpected login2PasswordErrorAbort")
                root.reset()
                errorLabel.text = qsTr("Incorrect login credentials. Please try again.")
            }

            onLoginFinished: {
                stackLayout.currentIndex = 0
                root.reset()
            }
        }

        ColumnLayout {
            id: loginNormalLayout

            function reset() {
                signInButton.loading = false

                errorLabel.text = ""

                usernameTextField.enabled = true
                usernameTextField.error = false
                usernameTextField.assistiveText = ""

                passwordTextField.enabled = true
                passwordTextField.error = false
                passwordTextField.assistiveText = ""
                passwordTextField.text = ""
            }

            spacing: 0

            Label {
                colorScheme: root.colorScheme
                text: qsTr("Sign in")
                Layout.alignment: Qt.AlignHCenter
                Layout.topMargin: 16
                type: Label.LabelType.Title
            }

            Label {
                colorScheme: root.colorScheme
                id: subTitle
                text: qsTr("Enter your Proton Account details.")
                Layout.alignment: Qt.AlignHCenter
                Layout.topMargin: 8
                color: root.colorScheme.text_weak
                type: Label.LabelType.Body
            }

            RowLayout {
                Layout.fillWidth: true
                Layout.topMargin: 36

                spacing: 0
                visible: errorLabel.text.length > 0

                ColorImage {
                    color: root.colorScheme.signal_danger
                    source: "./icons/ic-exclamation-circle-filled.svg"
                    height: errorLabel.height
                    sourceSize.height: errorLabel.height
                }

                Label {
                    colorScheme: root.colorScheme
                    id: errorLabel
                    Layout.leftMargin: 4
                    color: root.colorScheme.signal_danger

                    type: root.error ? Label.LabelType.Caption_semibold : Label.LabelType.Caption
                }
            }

            TextField {
                colorScheme: root.colorScheme
                id: usernameTextField
                label: qsTr("Username or email")

                Layout.fillWidth: true
                Layout.topMargin: 24

                onTextEdited: { // TODO: repeating?
                    if (error || errorLabel.text.length > 0) {
                        errorLabel.text = ""

                        usernameTextField.error = false
                        usernameTextField.assistiveText = ""

                        passwordTextField.error = false
                        passwordTextField.assistiveText = ""
                    }
                }

                onAccepted: passwordTextField.forceActiveFocus()
            }

            TextField {
                colorScheme: root.colorScheme
                id: passwordTextField
                label: qsTr("Password")

                Layout.fillWidth: true
                Layout.topMargin: 8
                echoMode: TextInput.Password

                onTextEdited: {
                    if (error || errorLabel.text.length > 0) {
                        errorLabel.text = ""

                        usernameTextField.error = false
                        usernameTextField.assistiveText = ""

                        passwordTextField.error = false
                        passwordTextField.assistiveText = ""
                    }
                }

                onAccepted: signInButton.checkAndSignIn()
            }

            Button {
                colorScheme: root.colorScheme
                id: signInButton
                text: qsTr("Sign in")

                Layout.fillWidth: true
                Layout.topMargin: 24


                onClicked: checkAndSignIn()

                function checkAndSignIn() {
                    var err = false

                    if (usernameTextField.text.length == 0) {
                        usernameTextField.error = true
                        usernameTextField.assistiveText = qsTr("Enter username or email")
                        err = true
                    } else {
                        usernameTextField.error = false
                        usernameTextField.assistiveText = qsTr("")
                    }

                    if (passwordTextField.text.length == 0) {
                        passwordTextField.error = true
                        passwordTextField.assistiveText = qsTr("Enter password")
                        err = true
                    } else {
                        passwordTextField.error = false
                        passwordTextField.assistiveText = qsTr("")
                    }

                    if (err) {
                        return
                    }

                    usernameTextField.enabled = false
                    passwordTextField.enabled = false

                    enabled = false
                    loading = true

                    root.backend.login(usernameTextField.text, Qt.btoa(passwordTextField.text))
                }
            }

            Label {
                colorScheme: root.colorScheme
                textFormat: Text.StyledText
                text: link("https://protonmail.com/signup", qsTr("Create or upgrade your account"))
                Layout.alignment: Qt.AlignHCenter
                Layout.topMargin: 24
                type: Label.LabelType.Body

                onLinkActivated: {
                    Qt.openUrlExternally(link)
                }

            }
        }

        ColumnLayout {
            id: login2FALayout

            function reset() {
                twoFAButton.loading = false

                twoFactorPasswordTextField.enabled = true
                twoFactorPasswordTextField.error = false
                twoFactorPasswordTextField.assistiveText = ""
                twoFactorPasswordTextField.text=""
            }

            spacing: 0

            Label {
                colorScheme: root.colorScheme
                text: qsTr("Two-factor authentication")
                Layout.topMargin: 16
                Layout.alignment: Qt.AlignCenter
                type: Label.LabelType.Heading
            }

            Label {
                colorScheme: root.colorScheme
                id: twoFactorUsernameLabel

                Layout.alignment: Qt.AlignCenter
                Layout.topMargin: 8
                type: Label.LabelType.Lead
                color: root.colorScheme.text_weak
            }

            TextField {
                colorScheme: root.colorScheme
                id: twoFactorPasswordTextField
                label: qsTr("Two-factor code")
                assistiveText: qsTr("Enter the 6-digit code")

                Layout.fillWidth: true
                Layout.topMargin: 32

                onTextEdited: {
                    if (error) {
                        twoFactorPasswordTextField.error = false
                        twoFactorPasswordTextField.assistiveText = ""
                    }
                }
            }

            Button {
                colorScheme: root.colorScheme
                id: twoFAButton
                text: loading ? qsTr("Authenticating") : qsTr("Authenticate")

                Layout.fillWidth: true
                Layout.topMargin: 24

                onClicked: {
                    var err = false

                    if (twoFactorPasswordTextField.text.length == 0) {
                        twoFactorPasswordTextField.error = true
                        twoFactorPasswordTextField.assistiveText = qsTr("Enter username or email")
                        err = true
                    } else {
                        twoFactorPasswordTextField.error = false
                        twoFactorPasswordTextField.assistiveText = qsTr("")
                    }

                    if (err) {
                        return
                    }

                    twoFactorPasswordTextField.enabled = false

                    enabled = false
                    loading = true

                    root.backend.login2FA(usernameTextField.text, Qt.btoa(twoFactorPasswordTextField.text))
                }
            }
        }

        ColumnLayout {
            id: login2PasswordLayout

            function reset() {
                secondPasswordButton.loading = false

                secondPasswordTextField.enabled = true
                secondPasswordTextField.error = false
                secondPasswordTextField.assistiveText = ""
                secondPasswordTextField.text = ""
            }

            spacing: 0

            Label {
                colorScheme: root.colorScheme
                text: qsTr("Unlock your mailbox")
                Layout.topMargin: 16
                Layout.alignment: Qt.AlignCenter
                type: Label.LabelType.Heading
            }

            TextField {
                colorScheme: root.colorScheme
                id: secondPasswordTextField
                label: qsTr("Mailbox password")

                Layout.fillWidth: true
                Layout.topMargin: 8 + implicitHeight + 24 + subTitle.implicitHeight
                echoMode: TextInput.Password

                onTextEdited: {
                    if (error) {
                        secondPasswordTextField.error = false
                        secondPasswordTextField.assistiveText = ""
                    }
                }
            }

            Button {
                colorScheme: root.colorScheme
                id: secondPasswordButton
                text: loading ? qsTr("Unlocking") : qsTr("Unlock")

                Layout.fillWidth: true
                Layout.topMargin: 24

                onClicked: {
                    var err = false

                    if (secondPasswordTextField.text.length == 0) {
                        secondPasswordTextField.error = true
                        secondPasswordTextField.assistiveText = qsTr("Enter username or email")
                        err = true
                    } else {
                        secondPasswordTextField.error = false
                        secondPasswordTextField.assistiveText = qsTr("")
                    }

                    if (err) {
                        return
                    }

                    secondPasswordTextField.enabled = false

                    enabled = false
                    loading = true

                    root.backend.login2Password(usernameTextField.text, Qt.btoa(secondPasswordTextField.text))
                }
            }
        }
    }

    states: [
        State {
            name: "Page 1"
            PropertyChanges {
                target: stackLayout
                currentIndex: 0
            }
        },
        State {
            name: "Page 2"
            PropertyChanges {
                target: stackLayout
                currentIndex: 1
            }
        },
        State {
            name: "Page 3"
            PropertyChanges {
                target: stackLayout
                currentIndex: 2
            }
        }
    ]
}
