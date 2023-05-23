// Copyright (c) 2023 Proton AG
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

import QtQml
import QtQuick
import QtQuick.Layouts
import QtQuick.Controls
import QtQuick.Controls.impl

import Proton

FocusScope {
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
        Backend.loginAbort(usernameTextField.text)
    }

    implicitHeight: children[0].implicitHeight
    implicitWidth: children[0].implicitWidth

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
            target: Backend

            function onLoginUsernamePasswordError(errorMsg) {
                console.assert(stackLayout.currentIndex == 0, "Unexpected loginUsernamePasswordError")
                console.assert(signInButton.loading == true, "Unexpected loginUsernamePasswordError")

                stackLayout.loginFailed()
                if (errorMsg!="") errorLabel.text = errorMsg
                else errorLabel.text = qsTr("Incorrect login credentials")
            }

            function onLoginFreeUserError() {
                console.assert(stackLayout.currentIndex == 0, "Unexpected loginFreeUserError")
                stackLayout.loginFailed()
            }

            function onLoginConnectionError(errorMsg) {
                if (stackLayout.currentIndex == 0 ) {
                    stackLayout.loginFailed()
                }
            }

            function onLogin2FARequested(username) {
                console.assert(stackLayout.currentIndex == 0, "Unexpected login2FARequested")
                twoFactorUsernameLabel.text = username
                stackLayout.currentIndex = 1
                twoFactorPasswordTextField.focus = true
            }

            function onLogin2FAError(errorMsg) {
                console.assert(stackLayout.currentIndex == 1, "Unexpected login2FAError")

                twoFAButton.loading = false

                twoFactorPasswordTextField.enabled = true
                twoFactorPasswordTextField.error = true
                twoFactorPasswordTextField.errorString = qsTr("Your code is incorrect")
                twoFactorPasswordTextField.focus = true
            }

            function onLogin2FAErrorAbort(errorMsg) {
                console.assert(stackLayout.currentIndex == 1, "Unexpected login2FAErrorAbort")
                root.reset()
                errorLabel.text = qsTr("Incorrect login credentials. Please try again.")
            }

            function onLogin2PasswordRequested() {
                console.assert(stackLayout.currentIndex == 0 || stackLayout.currentIndex == 1, "Unexpected login2PasswordRequested")
                stackLayout.currentIndex = 2
                secondPasswordTextField.focus = true
            }
            function onLogin2PasswordError(errorMsg) {
                console.assert(stackLayout.currentIndex == 2, "Unexpected login2PasswordError")

                secondPasswordButton.loading = false

                secondPasswordTextField.enabled = true
                secondPasswordTextField.error = true
                secondPasswordTextField.errorString = qsTr("Your mailbox password is incorrect")
                secondPasswordTextField.focus = true
            }
            function onLogin2PasswordErrorAbort(errorMsg) {
                console.assert(stackLayout.currentIndex == 2, "Unexpected login2PasswordErrorAbort")
                root.reset()
                errorLabel.text = qsTr("Incorrect login credentials. Please try again.")
            }

            function onLoginFinished(index) {
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
                usernameTextField.errorString = ""
                usernameTextField.focus = true

                passwordTextField.enabled = true
                passwordTextField.error = false
                passwordTextField.errorString = ""
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
                    source: "/qml/icons/ic-exclamation-circle-filled.svg"
                    height: errorLabel.lineHeight
                    sourceSize.height: errorLabel.lineHeight
                }

                Label {
                    colorScheme: root.colorScheme
                    id: errorLabel
                    wrapMode: Text.WordWrap
                    Layout.fillWidth: true;
                    Layout.leftMargin: 4
                    color: root.colorScheme.signal_danger

                    type: root.error ? Label.LabelType.Caption_semibold : Label.LabelType.Caption
                }
            }

            TextField {
                colorScheme: root.colorScheme
                id: usernameTextField
                label: qsTr("Email or username")
                focus: true
                Layout.fillWidth: true
                Layout.topMargin: 24
                validateOnEditingFinished: false

                onTextChanged: {
                    // remove "invalid username / password error"
                    if (error || errorLabel.text.length > 0) {
                        errorLabel.text = ""
                        usernameTextField.error = false
                        passwordTextField.error = false
                    }
                }

                validator: function(str) {
                    if (str.length === 0) {
                        return qsTr("Enter email or username")
                    }
                    return
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
                validateOnEditingFinished: false

                onTextChanged: {
                    // remove "invalid username / password error"
                    if (error || errorLabel.text.length > 0) {
                        errorLabel.text = ""
                        usernameTextField.error = false
                        passwordTextField.error = false
                    }
                }

                validator: function(str) {
                    if (str.length === 0) {
                        return qsTr("Enter password")
                    }
                    return
                }

                onAccepted: signInButton.checkAndSignIn()
            }

            Button {
                colorScheme: root.colorScheme
                id: signInButton
                text: loading ? qsTr("Signing in") : qsTr("Sign in")
                enabled: !loading
                Layout.fillWidth: true
                Layout.topMargin: 24


                onClicked: checkAndSignIn()

                function checkAndSignIn() {
                    usernameTextField.validate()
                    passwordTextField.validate()

                    if (usernameTextField.error || passwordTextField.error) {
                        return
                    }

                    usernameTextField.enabled = false
                    passwordTextField.enabled = false

                    loading = true

                    Backend.login(usernameTextField.text, Qt.btoa(passwordTextField.text))
                }
            }

            Label {
                colorScheme: root.colorScheme
                textFormat: Text.StyledText
                text: link("https://proton.me/mail/pricing", qsTr("Create or upgrade your account"))
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
                twoFactorPasswordTextField.errorString = ""
                twoFactorPasswordTextField.text = ""
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
                validateOnEditingFinished: false
                Layout.fillWidth: true
                Layout.topMargin: 32

                validator: function(str) {
                    if (str.length === 0) {
                        return qsTr("Enter the 6-digit code")
                    }
                }

                onTextChanged: {
                    if (text.length >= 6) {
                        twoFAButton.onClicked()
                    }
                }

                onAccepted: {
                    twoFAButton.onClicked()
                }

            }

            Button {
                colorScheme: root.colorScheme
                id: twoFAButton
                text: loading ? qsTr("Authenticating") : qsTr("Authenticate")
                enabled: !loading
                Layout.fillWidth: true
                Layout.topMargin: 24

                onClicked: {
                    twoFactorPasswordTextField.validate()

                    if (twoFactorPasswordTextField.error) {
                        return
                    }

                    twoFactorPasswordTextField.enabled = false
                    loading = true
                    Backend.login2FA(usernameTextField.text, Qt.btoa(twoFactorPasswordTextField.text))
                }
            }
        }

        ColumnLayout {
            id: login2PasswordLayout

            function reset() {
                secondPasswordButton.loading = false

                secondPasswordTextField.enabled = true
                secondPasswordTextField.error = false
                secondPasswordTextField.errorString = ""
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
                validateOnEditingFinished: false

                validator: function(str) {
                    if (str.length === 0) {
                        return qsTr("Enter password")
                    }
                    return
                }

                onAccepted: {
                    secondPasswordButton.onClicked()
                }
            }

            Button {
                colorScheme: root.colorScheme
                id: secondPasswordButton
                text: loading ? qsTr("Unlocking") : qsTr("Unlock")
                enabled: !loading

                Layout.fillWidth: true
                Layout.topMargin: 24

                onClicked: {
                    secondPasswordTextField.validate()

                    if (secondPasswordTextField.error) {
                        return
                    }

                    secondPasswordTextField.enabled = false
                    loading = true
                    Backend.login2Password(usernameTextField.text, Qt.btoa(secondPasswordTextField.text))
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
