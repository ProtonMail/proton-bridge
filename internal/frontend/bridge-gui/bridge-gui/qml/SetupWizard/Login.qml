// Copyright (c) 2024 Proton AG
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

FocusScope {
    id: root
    enum RootStack {
        Login,
        TOTP,
        MailboxPassword,
        HV
    }

    property alias currentIndex: stackLayout.currentIndex
    property alias username: usernameTextField.text
    property var wizard
    property string hvLinkUrl: ""

    signal loginAbort(string username, bool wasSignedOut)

    function abort() {
        root.reset();
        loginAbort(usernameTextField.text, false);
        Backend.loginAbort(usernameTextField.text);
    }
    function reset(clearUsername = false) {
        stackLayout.currentIndex = Login.RootStack.Login;
        loginLayout.reset(clearUsername);
        totpLayout.reset();
        mailboxPasswordLayout.reset();
        if (username.length === 0) {
            usernameTextField.forceActiveFocus();
        } else {
            passwordTextField.forceActiveFocus();
        }
        passwordTextField.hidePassword();
        secondPasswordTextField.hidePassword();
    }
    function resetViaHv() {
        usernameTextField.enabled = false;
        passwordTextField.enabled = false;
        signInButton.loading = true;
        secondPasswordButton.loading = false;
        secondPasswordTextField.enabled = true;
        totpLayout.reset();
    }

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
                console.assert(stackLayout.currentIndex === Login.RootStack.TOTP, "Unexpected login2FAError");
                twoFAButton.loading = false;
                twoFactorPasswordTextField.enabled = true;
                twoFactorPasswordTextField.error = true;
                twoFactorPasswordTextField.errorString = qsTr("Your code is incorrect");
                twoFactorPasswordTextField.focus = true;
            }
            function onLogin2FAErrorAbort(_) {
                console.assert(stackLayout.currentIndex === Login.RootStack.TOTP, "Unexpected login2FAErrorAbort");
                root.reset();
                errorLabel.text = qsTr("Incorrect login credentials. Please try again.");
            }
            function onLogin2FARequested(username) {
                console.assert(stackLayout.currentIndex === Login.RootStack.Login, "Unexpected login2FARequested");
                twoFactorUsernameLabel.text = username;
                stackLayout.currentIndex = Login.RootStack.TOTP;
                twoFactorPasswordTextField.focus = true;
            }
            function onLogin2PasswordError(_) {
                console.assert(stackLayout.currentIndex === Login.RootStack.MailboxPassword, "Unexpected login2PasswordError");
                secondPasswordButton.loading = false;
                secondPasswordTextField.enabled = true;
                secondPasswordTextField.error = true;
                secondPasswordTextField.errorString = qsTr("Your mailbox password is incorrect");
                secondPasswordTextField.focus = true;
            }
            function onLogin2PasswordErrorAbort(_) {
                console.assert(stackLayout.currentIndex === Login.RootStack.MailboxPassword, "Unexpected login2PasswordErrorAbort");
                root.reset();
                errorLabel.text = qsTr("Incorrect login credentials. Please try again.");
            }
            function onLogin2PasswordRequested(username) {
                console.assert(stackLayout.currentIndex === Login.RootStack.Login || stackLayout.currentIndex === Login.RootStack.TOTP, "Unexpected login2PasswordRequested");
                stackLayout.currentIndex = Login.RootStack.MailboxPassword;
                mailboxPasswordUsernameLabel.text = username;
                secondPasswordTextField.focus = true;
            }
            function onLoginAlreadyLoggedIn(_) {
                stackLayout.currentIndex = Login.RootStack.Login;
                root.reset();
            }
            function onLoginConnectionError(_) {
                if (stackLayout.currentIndex === Login.RootStack.Login) {
                    stackLayout.loginFailed();
                }
            }
            function onLoginFinished(_) {
                stackLayout.currentIndex = Login.RootStack.Login;
                root.reset();
            }
            function onLoginFreeUserError() {
                console.assert(stackLayout.currentIndex === Login.RootStack.Login, "Unexpected loginFreeUserError");
                stackLayout.loginFailed();
            }
            function onLoginUsernamePasswordError(errorMsg) {
                console.assert(stackLayout.currentIndex === Login.RootStack.Login, "Unexpected loginUsernamePasswordError");
                stackLayout.loginFailed();
                if (errorMsg !== "")
                    errorLabel.text = errorMsg;
                else
                    errorLabel.text = qsTr("Incorrect login credentials");
            }
            function onLoginHvRequested(hvUrl) {
                console.assert(stackLayout.currentIndex === Login.RootStack.Login || stackLayout.currentIndex === Login.RootStack.MailboxPassword, "Unexpected loginHvRequested");
                stackLayout.currentIndex = Login.RootStack.HV;
                hvUsernameLabel.text = usernameTextField.text;
                hvLinkUrl = hvUrl;
            }
            function onLoginHvError(_) {
                console.assert(stackLayout.currentIndex === Login.RootStack.Login || stackLayout.currentIndex === Login.RootStack.MailboxPassword, "Unexpected onLoginHvInvalidTokenError");
                stackLayout.currentIndex = Login.RootStack.Login;
                root.resetViaHv();
                root.reset()
            }

            target: Backend
        }
        Item {
            ColumnLayout {
                id: loginLayout
                function clearErrors() {
                    usernameTextField.error = false;
                    usernameTextField.errorString = "";
                    passwordTextField.error = false;
                    passwordTextField.errorString = "";
                    errorLabel.text = "";
                }
                function reset(clearUsername = false) {
                    signInButton.loading = false;
                    errorLabel.text = "";
                    usernameTextField.enabled = true;
                    usernameTextField.focus = true;
                    if (clearUsername) {
                        usernameTextField.text = "";
                    }
                    passwordTextField.enabled = true;
                    passwordTextField.text = "";
                    clearErrors();
                }

                anchors.left: parent.left
                anchors.right: parent.right
                anchors.verticalCenter: parent.verticalCenter
                spacing: ProtonStyle.wizard_spacing_medium

                ColumnLayout {
                    Layout.fillWidth: true
                    spacing: ProtonStyle.wizard_spacing_small

                    Label {
                        Layout.alignment: Qt.AlignHCenter
                        Layout.fillWidth: true
                        colorScheme: wizard.colorScheme
                        horizontalAlignment: Text.AlignHCenter
                        text: qsTr("Sign in")
                        type: Label.LabelType.Title
                    }
                    Label {
                        Layout.alignment: Qt.AlignHCenter
                        Layout.fillWidth: true
                        color: wizard.colorScheme.text_weak
                        colorScheme: wizard.colorScheme
                        horizontalAlignment: Text.AlignHCenter
                        text: qsTr("Enter your Proton Account details.")
                        type: Label.LabelType.Body
                    }
                }
                RowLayout {
                    Layout.fillWidth: true
                    spacing: 0

                    ColorImage {
                        color: wizard.colorScheme.signal_danger
                        height: errorLabel.lineHeight
                        source: "/qml/icons/ic-exclamation-circle-filled.svg"
                        sourceSize.height: errorLabel.lineHeight
                        visible: errorLabel.text.length > 0
                    }
                    Label {
                        id: errorLabel
                        Layout.fillWidth: true
                        Layout.leftMargin: 4
                        color: wizard.colorScheme.signal_danger
                        colorScheme: wizard.colorScheme
                        type: Label.LabelType.Caption_semibold
                        wrapMode: Text.WordWrap
                    }
                }
                TextField {
                    id: usernameTextField
                    Layout.fillWidth: true
                    colorScheme: wizard.colorScheme
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
                        loginLayout.clearErrors();
                    }
                }
                TextField {
                    id: passwordTextField
                    Layout.fillWidth: true
                    colorScheme: wizard.colorScheme
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
                        loginLayout.clearErrors();
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
                    colorScheme: wizard.colorScheme
                    enabled: !loading
                    text: loading ? qsTr("Signing in") : qsTr("Sign in")

                    onClicked: {
                        checkAndSignIn();
                    }
                }
                Button {
                    Layout.fillWidth: true
                    colorScheme: wizard.colorScheme
                    enabled: !signInButton.loading
                    secondary: true
                    secondaryIsOpaque: true
                    text: qsTr("Cancel")

                    onClicked: {
                        root.abort();
                    }
                }
                LinkLabel {
                    Layout.alignment: Qt.AlignHCenter
                    colorScheme: wizard.colorScheme
                    external: true
                    link: "https://proton.me/mail/pricing"
                    text: qsTr("Create or upgrade your account")
                }
            }
        }
        Item {
            ColumnLayout {
                id: totpLayout
                function reset() {
                    twoFAButton.loading = false;
                    twoFactorPasswordTextField.enabled = true;
                    twoFactorPasswordTextField.error = false;
                    twoFactorPasswordTextField.errorString = "";
                    twoFactorPasswordTextField.text = "";
                }

                anchors.left: parent.left
                anchors.right: parent.right
                anchors.verticalCenter: parent.verticalCenter
                spacing: ProtonStyle.wizard_spacing_medium

                ColumnLayout {
                    Layout.fillWidth: true
                    spacing: ProtonStyle.wizard_spacing_small

                    Label {
                        Layout.alignment: Qt.AlignHCenter
                        Layout.fillWidth: true
                        colorScheme: wizard.colorScheme
                        horizontalAlignment: Text.AlignHCenter
                        text: qsTr("Two-factor authentication")
                        type: Label.LabelType.Title
                    }
                    Label {
                        id: twoFactorUsernameLabel
                        Layout.alignment: Qt.AlignHCenter
                        Layout.fillWidth: true
                        color: wizard.colorScheme.text_weak
                        colorScheme: wizard.colorScheme
                        horizontalAlignment: Text.AlignHCenter
                        text: ""
                        type: Label.LabelType.Body
                    }
                }
                Label {
                    id: descriptionLabel
                    Layout.alignment: Qt.AlignHCenter
                    Layout.fillWidth: true
                    colorScheme: wizard.colorScheme
                    horizontalAlignment: Text.AlignHCenter
                    text: qsTr("You have enabled two-factor authentication. Please enter the 6-digit code provided by your authenticator application.")
                    type: Label.LabelType.Body
                    wrapMode: Text.WordWrap
                }
                TextField {
                    id: twoFactorPasswordTextField
                    Layout.fillWidth: true
                    colorScheme: wizard.colorScheme
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
                    colorScheme: wizard.colorScheme
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
                Button {
                    Layout.fillWidth: true
                    colorScheme: wizard.colorScheme
                    enabled: !twoFAButton.loading
                    secondary: true
                    secondaryIsOpaque: true
                    text: qsTr("Cancel")

                    onClicked: {
                        root.abort();
                    }
                }
            }
        }
        Item {
            ColumnLayout {
                id: mailboxPasswordLayout
                function reset() {
                    secondPasswordButton.loading = false;
                    secondPasswordTextField.enabled = true;
                    secondPasswordTextField.error = false;
                    secondPasswordTextField.errorString = "";
                    secondPasswordTextField.text = "";
                }

                anchors.left: parent.left
                anchors.right: parent.right
                anchors.verticalCenter: parent.verticalCenter
                spacing: ProtonStyle.wizard_spacing_medium

                ColumnLayout {
                    Layout.fillWidth: true
                    spacing: ProtonStyle.wizard_spacing_small

                    Label {
                        Layout.alignment: Qt.AlignHCenter
                        Layout.fillWidth: true
                        colorScheme: wizard.colorScheme
                        horizontalAlignment: Text.AlignHCenter
                        text: qsTr("Unlock your mailbox")
                        type: Label.LabelType.Title
                    }
                    Label {
                        id: mailboxPasswordUsernameLabel
                        Layout.alignment: Qt.AlignHCenter
                        Layout.fillWidth: true
                        color: wizard.colorScheme.text_weak
                        colorScheme: wizard.colorScheme
                        horizontalAlignment: Text.AlignHCenter
                        text: ""
                        type: Label.LabelType.Body
                    }
                }
                Label {
                    Layout.alignment: Qt.AlignHCenter
                    Layout.fillWidth: true
                    colorScheme: wizard.colorScheme
                    horizontalAlignment: Text.AlignHCenter
                    text: qsTr("You have secured your account with a separate mailbox password.")
                    type: Label.LabelType.Body
                    wrapMode: Text.WordWrap
                }
                TextField {
                    id: secondPasswordTextField
                    Layout.fillWidth: true
                    colorScheme: wizard.colorScheme
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
                    colorScheme: wizard.colorScheme
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
                Button {
                    Layout.fillWidth: true
                    colorScheme: wizard.colorScheme
                    enabled: !secondPasswordButton.loading
                    secondary: true
                    secondaryIsOpaque: true
                    text: qsTr("Cancel")

                    onClicked: {
                        root.abort();
                    }
                }
            }
        }
        Item {
            id: hvLayout

            ColumnLayout {
                Layout.fillWidth: true
                anchors.left: parent.left
                anchors.right: parent.right
                anchors.verticalCenter: parent.verticalCenter
                spacing: ProtonStyle.wizard_spacing_extra_large

                ColumnLayout {
                    spacing: ProtonStyle.wizard_spacing_medium

                    ColumnLayout {
                        spacing: ProtonStyle.wizard_spacing_small

                        Label {
                            Layout.alignment: Qt.AlignHCenter
                            Layout.fillWidth: true
                            colorScheme: wizard.colorScheme
                            horizontalAlignment: Text.AlignHCenter
                            text: qsTr("Human verification")
                            type: Label.LabelType.Title
                            wrapMode: Text.WordWrap
                        }

                        Label {
                            id: hvUsernameLabel
                            Layout.alignment: Qt.AlignHCenter
                            Layout.fillWidth: true
                            color: wizard.colorScheme.text_weak
                            colorScheme: wizard.colorScheme
                            horizontalAlignment: Text.AlignHCenter
                            type: Label.LabelType.Body
                        }

                    }

                    Label {
                        Layout.alignment: Qt.AlignHCenter
                        Layout.fillWidth: true
                        colorScheme: wizard.colorScheme
                        horizontalAlignment: Text.AlignHCenter
                        text: qsTr("Please open the following link in your favourite web browser to verify you are human.")
                        type: Label.LabelType.Body
                        wrapMode: Text.WordWrap
                    }

                }


                Label {
                    id: hvRequestedUrlText
                    type: Label.LabelType.Lead
                    Layout.alignment: Qt.AlignHCenter
                    Layout.fillWidth: true
                    colorScheme: wizard.colorScheme
                    horizontalAlignment: Text.AlignLeft
                    text: "<a href='" + hvLinkUrl + "'>" + hvLinkUrl.replace("&", "&amp;")+ "</a>"
                    MouseArea {
                        anchors.fill: parent
                        cursorShape: Qt.PointingHandCursor
                        onClicked: {
                            Qt.openUrlExternally(hvLinkUrl);
                        }
                    }
                }


                ColumnLayout {
                    spacing: ProtonStyle.wizard_spacing_medium

                    Button {
                        id: hVContinueButton
                        Layout.fillWidth: true
                        colorScheme: wizard.colorScheme
                        text: qsTr("Continue")

                        function checkAndSignInHv() {
                            console.assert(stackLayout.currentIndex === Login.RootStack.HV ||  stackLayout.currentIndex === Login.RootStack.MailboxPassword, "Unexpected checkInAndSignInHv")
                            stackLayout.currentIndex = Login.RootStack.Login
                            usernameTextField.validate();
                            passwordTextField.validate();
                            if (usernameTextField.error || passwordTextField.error) {
                                return;
                            }
                            root.resetViaHv();
                            Backend.loginHv(usernameTextField.text, Qt.btoa(passwordTextField.text));
                        }

                        onClicked: {
                            checkAndSignInHv()
                        }
                    }
                    Button {
                        Layout.fillWidth: true
                        colorScheme: wizard.colorScheme
                        secondary: true
                        secondaryIsOpaque: true
                        text: qsTr("Cancel")
                        onClicked: {
                            root.abort();
                        }
                    }
                }
            }
        }
    }
}
