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
import QtQuick.Layouts
import QtQuick.Controls
import QtQuick.Controls.impl
import Proton

FocusScope {
    id: root

    property ColorScheme colorScheme
    property alias currentIndex: stackLayout.currentIndex
    property alias username: usernameTextField.text

    signal loginAbort(string username, bool wasSignedOut)

    function abort() {
        root.reset();
        loginAbort(usernameTextField.text, false);
        Backend.loginAbort(usernameTextField.text);
    }
    function reset(clearUsername = false) {
        stackLayout.currentIndex = 0;
        loginNormalLayout.reset(clearUsername);
        login2FALayout.reset();
        login2PasswordLayout.reset();
    }

    implicitHeight: children[0].implicitHeight
    implicitWidth: children[0].implicitWidth
    state: "Page 1"

    states: [
        State {
            name: "Page 1"

            PropertyChanges {
                currentIndex: 0
                target: stackLayout
            }
        },
        State {
            name: "Page 2"

            PropertyChanges {
                currentIndex: 1
                target: stackLayout
            }
        },
        State {
            name: "Page 3"

            PropertyChanges {
                currentIndex: 2
                target: stackLayout
            }
        }
    ]

    StackLayout {
        id: stackLayout
        function loginFailed() {
            signInButton.loading = false;
            usernameTextField.enabled = true;
            usernameTextField.error = true;
            passwordTextField.enabled = true;
            passwordTextField.error = true;
        }

        anchors.fill: parent

        Connections {
            function onLogin2FAError(_) {
                console.assert(stackLayout.currentIndex === 1, "Unexpected login2FAError");
                twoFAButton.loading = false;
                twoFactorPasswordTextField.enabled = true;
                twoFactorPasswordTextField.error = true;
                twoFactorPasswordTextField.errorString = qsTr("Your code is incorrect");
                twoFactorPasswordTextField.focus = true;
            }
            function onLogin2FAErrorAbort(_) {
                console.assert(stackLayout.currentIndex === 1, "Unexpected login2FAErrorAbort");
                root.reset();
                errorLabel.text = qsTr("Incorrect login credentials. Please try again.");
            }
            function onLogin2FARequested(username) {
                console.assert(stackLayout.currentIndex === 0, "Unexpected login2FARequested");
                twoFactorUsernameLabel.text = username;
                stackLayout.currentIndex = 1;
                twoFactorPasswordTextField.focus = true;
            }
            function onLogin2PasswordError(_) {
                console.assert(stackLayout.currentIndex === 2, "Unexpected login2PasswordError");
                secondPasswordButton.loading = false;
                secondPasswordTextField.enabled = true;
                secondPasswordTextField.error = true;
                secondPasswordTextField.errorString = qsTr("Your mailbox password is incorrect");
                secondPasswordTextField.focus = true;
            }
            function onLogin2PasswordErrorAbort(_) {
                console.assert(stackLayout.currentIndex === 2, "Unexpected login2PasswordErrorAbort");
                root.reset();
                errorLabel.text = qsTr("Incorrect login credentials. Please try again.");
            }
            function onLogin2PasswordRequested() {
                console.assert(stackLayout.currentIndex === 0 || stackLayout.currentIndex === 1, "Unexpected login2PasswordRequested");
                stackLayout.currentIndex = 2;
                secondPasswordTextField.focus = true;
            }
            function onLoginAlreadyLoggedIn(_) {
                stackLayout.currentIndex = 0;
                root.reset();
            }
            function onLoginConnectionError(_) {
                if (stackLayout.currentIndex === 0) {
                    stackLayout.loginFailed();
                }
            }
            function onLoginFinished(_) {
                stackLayout.currentIndex = 0;
                root.reset();
            }
            function onLoginFreeUserError() {
                console.assert(stackLayout.currentIndex === 0, "Unexpected loginFreeUserError");
                stackLayout.loginFailed();
            }
            function onLoginUsernamePasswordError(errorMsg) {
                console.assert(stackLayout.currentIndex === 0, "Unexpected loginUsernamePasswordError");
                stackLayout.loginFailed();
                if (errorMsg !== "")
                    errorLabel.text = errorMsg;
                else
                    errorLabel.text = qsTr("Incorrect login credentials");
            }

            target: Backend
        }
        ColumnLayout {
            id: loginNormalLayout
            function reset(clearUsername = false) {
                signInButton.loading = false;
                errorLabel.text = "";
                usernameTextField.enabled = true;
                usernameTextField.error = false;
                usernameTextField.errorString = "";
                usernameTextField.focus = true;
                if (clearUsername) {
                    usernameTextField.text = "";
                }
                passwordTextField.enabled = true;
                passwordTextField.error = false;
                passwordTextField.errorString = "";
                passwordTextField.text = "";
            }

            spacing: 0

            Label {
                Layout.alignment: Qt.AlignHCenter
                Layout.topMargin: 16
                colorScheme: root.colorScheme
                text: qsTr("Sign in")
                type: Label.LabelType.Title
            }
            Label {
                id: subTitle
                Layout.alignment: Qt.AlignHCenter
                Layout.topMargin: 8
                color: root.colorScheme.text_weak
                colorScheme: root.colorScheme
                text: qsTr("Enter your Proton Account details.")
                type: Label.LabelType.Body
            }
            RowLayout {
                Layout.fillWidth: true
                Layout.topMargin: 36
                spacing: 0
                visible: errorLabel.text.length > 0

                ColorImage {
                    color: root.colorScheme.signal_danger
                    height: errorLabel.lineHeight
                    source: "/qml/icons/ic-exclamation-circle-filled.svg"
                    sourceSize.height: errorLabel.lineHeight
                }
                Label {
                    id: errorLabel
                    Layout.fillWidth: true
                    Layout.leftMargin: 4
                    color: root.colorScheme.signal_danger
                    colorScheme: root.colorScheme
                    type: root.error ? Label.LabelType.Caption_semibold : Label.LabelType.Caption
                    wrapMode: Text.WordWrap
                }
            }
            TextField {
                id: usernameTextField
                Layout.fillWidth: true
                Layout.topMargin: 24
                colorScheme: root.colorScheme
                focus: true
                label: qsTr("Email or username")
                validateOnEditingFinished: false
                validator: function (str) {
                    if (str.length === 0) {
                        return qsTr("Enter email or username");
                    }
                }

                onAccepted: passwordTextField.forceActiveFocus()
                onTextChanged: {
                    // remove "invalid username / password error"
                    if (error || errorLabel.text.length > 0) {
                        errorLabel.text = "";
                        usernameTextField.error = false;
                        passwordTextField.error = false;
                    }
                }
            }
            TextField {
                id: passwordTextField
                Layout.fillWidth: true
                Layout.topMargin: 8
                colorScheme: root.colorScheme
                echoMode: TextInput.Password
                label: qsTr("Password")
                validateOnEditingFinished: false
                validator: function (str) {
                    if (str.length === 0) {
                        return qsTr("Enter password");
                    }
                }

                onAccepted: signInButton.checkAndSignIn()
                onTextChanged: {
                    // remove "invalid username / password error"
                    if (error || errorLabel.text.length > 0) {
                        errorLabel.text = "";
                        usernameTextField.error = false;
                        passwordTextField.error = false;
                    }
                }
            }
            Button {
                id: signInButton
                function checkAndSignIn() {
                    usernameTextField.validate();
                    passwordTextField.validate();
                    if (usernameTextField.error || passwordTextField.error) {
                        return;
                    }
                    usernameTextField.enabled = false;
                    passwordTextField.enabled = false;
                    loading = true;
                    Backend.login(usernameTextField.text, Qt.btoa(passwordTextField.text));
                }

                Layout.fillWidth: true
                Layout.topMargin: 24
                colorScheme: root.colorScheme
                enabled: !loading
                text: loading ? qsTr("Signing in") : qsTr("Sign in")

                onClicked: {
                    checkAndSignIn();
                }
            }
            Button {
                id: cancelButton
                Layout.fillWidth: true
                Layout.topMargin: 24
                colorScheme: root.colorScheme
                enabled: !loading
                secondary: true
                text: qsTr("Cancel")

                onClicked: {
                    root.abort();
                }
            }
        }
        ColumnLayout {
            id: login2FALayout
            function reset() {
                twoFAButton.loading = false;
                twoFactorPasswordTextField.enabled = true;
                twoFactorPasswordTextField.error = false;
                twoFactorPasswordTextField.errorString = "";
                twoFactorPasswordTextField.text = "";
            }

            spacing: 0

            Label {
                Layout.alignment: Qt.AlignCenter
                Layout.topMargin: 16
                colorScheme: root.colorScheme
                text: qsTr("Two-factor authentication")
                type: Label.LabelType.Heading
            }
            Label {
                id: twoFactorUsernameLabel
                Layout.alignment: Qt.AlignCenter
                Layout.topMargin: 8
                color: root.colorScheme.text_weak
                colorScheme: root.colorScheme
                type: Label.LabelType.Lead
            }
            TextField {
                id: twoFactorPasswordTextField
                Layout.fillWidth: true
                Layout.topMargin: 32
                assistiveText: qsTr("Enter the 6-digit code")
                colorScheme: root.colorScheme
                label: qsTr("Two-factor code")
                validateOnEditingFinished: false
                validator: function (str) {
                    if (str.length === 0) {
                        return qsTr("Enter the 6-digit code");
                    }
                }

                onAccepted: {
                    twoFAButton.onClicked();
                }
                onTextChanged: {
                    if (text.length >= 6) {
                        twoFAButton.onClicked();
                    }
                }
            }
            Button {
                id: twoFAButton
                Layout.fillWidth: true
                Layout.topMargin: 24
                colorScheme: root.colorScheme
                enabled: !loading
                text: loading ? qsTr("Authenticating") : qsTr("Authenticate")

                onClicked: {
                    twoFactorPasswordTextField.validate();
                    if (twoFactorPasswordTextField.error) {
                        return;
                    }
                    twoFactorPasswordTextField.enabled = false;
                    loading = true;
                    Backend.login2FA(usernameTextField.text, Qt.btoa(twoFactorPasswordTextField.text));
                }
            }
        }
        ColumnLayout {
            id: login2PasswordLayout
            function reset() {
                secondPasswordButton.loading = false;
                secondPasswordTextField.enabled = true;
                secondPasswordTextField.error = false;
                secondPasswordTextField.errorString = "";
                secondPasswordTextField.text = "";
            }

            spacing: 0

            Label {
                Layout.alignment: Qt.AlignCenter
                Layout.topMargin: 16
                colorScheme: root.colorScheme
                text: qsTr("Unlock your mailbox")
                type: Label.LabelType.Heading
            }
            TextField {
                id: secondPasswordTextField
                Layout.fillWidth: true
                Layout.topMargin: 8 + implicitHeight + 24 + subTitle.implicitHeight
                colorScheme: root.colorScheme
                echoMode: TextInput.Password
                label: qsTr("Mailbox password")
                validateOnEditingFinished: false
                validator: function (str) {
                    if (str.length === 0) {
                        return qsTr("Enter password");
                    }
                }

                onAccepted: {
                    secondPasswordButton.onClicked();
                }
            }
            Button {
                id: secondPasswordButton
                Layout.fillWidth: true
                Layout.topMargin: 24
                colorScheme: root.colorScheme
                enabled: !loading
                text: loading ? qsTr("Unlocking") : qsTr("Unlock")

                onClicked: {
                    secondPasswordTextField.validate();
                    if (secondPasswordTextField.error) {
                        return;
                    }
                    secondPasswordTextField.enabled = false;
                    loading = true;
                    Backend.login2Password(usernameTextField.text, Qt.btoa(secondPasswordTextField.text));
                }
            }
        }
    }
}
