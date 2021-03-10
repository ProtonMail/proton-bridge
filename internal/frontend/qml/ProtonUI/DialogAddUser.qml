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

// Dialog with adding new user

import QtQuick 2.8
import ProtonUI 1.0


Dialog {
    id: root

    title : ""

    signal createAccount()

    property alias inputPassword : inputPassword
    property alias input2FAuth : input2FAuth
    property alias inputPasswMailbox : inputPasswMailbox
    //
    property alias username : inputUsername.text
    property alias usernameElided : usernameMetrics.elidedText

    isDialogBusy : currentIndex==waitingAuthIndex || currentIndex==addingAccIndex

    property bool isFirstAccount: false

    property color buttonOpaqueMain : "white"


    property int origin: 0
    property int nameAndPasswordIndex : 0
    property int waitingAuthIndex : 2
    property int twoFAIndex : 1
    property int mailboxIndex : 3
    property int addingAccIndex : 4
    property int newAccountIndex : 5


    signal cancel()
    signal okay()

    TextMetrics {
        id: usernameMetrics
        font: dialogWaitingAuthText.font
        elideWidth : Style.dialog.widthInput
        elide      : Qt.ElideMiddle
        text       : root.username
    }

    Column { // 0
        id: dialogNameAndPassword
        property int heightInputs : inputUsername.height + buttonRow.height + middleSep.height + inputPassword.height + middleSepPassw.height

        Rectangle {
            id: topSep
            color  : "transparent"
            width  : Style.main.dummy
            // Hacky hack: +10 is to make title of Dialog bigger so longer error can fit just fine.
            height : root.height/2 + 10 - (dialogNameAndPassword.heightInputs)/2
        }

        InputField {
            id: inputUsername
            iconText   : Style.fa.user_circle
            label      : qsTr("Username", "enter username to add account")
            onAccepted : inputPassword.focusInput = true
        }

        Rectangle { id: middleSepPassw; color : "transparent"; width : Style.main.dummy; height : Style.dialog.heightSeparator}

        InputField {
            id: inputPassword
            label      : qsTr("Password", "password entry field")
            iconText   : Style.fa.lock
            isPassword : true
            onAccepted : root.okay()
        }

        Rectangle { id: middleSep; color : "transparent"; width : Style.main.dummy; height : 2*Style.dialog.heightSeparator }

        Row {
            id: buttonRow
            anchors.horizontalCenter: parent.horizontalCenter
            spacing: Style.dialog.fontSize
            ButtonRounded {
                id:buttonCancel
                fa_icon    : Style.fa.times
                text       : qsTr("Cancel", "dismisses current action")
                color_main : Style.dialog.text
                onClicked  : root.cancel()
            }
            ButtonRounded {
                id: buttonNext
                fa_icon     : Style.fa.check
                text        : qsTr("Next", "navigate to next page in add account flow")
                color_main  : buttonOpaqueMain
                color_minor : Style.dialog.textBlue
                isOpaque    : true
                onClicked   : root.okay()
            }
        }

        Rectangle {
            color : "transparent"
            width : Style.main.dummy
            height : root.height - (topSep.height + dialogNameAndPassword.heightInputs + Style.main.bottomMargin + signUpForAccount.height)
        }

        ClickIconText {
            id: signUpForAccount
            anchors.horizontalCenter: parent.horizontalCenter
            fontSize      : Style.dialog.fontSize
            iconSize      : Style.dialog.fontSize
            iconText      : "+"
            text          : qsTr ("Sign Up for an Account", "takes user to web page where they can create a ProtonMail account")
            textBold      : true
            textUnderline : true
            textColor     : Style.dialog.text
            onClicked     : root.createAccount()
        }
    }

    Column { // 1
        id: dialog2FA
        property int heightInputs : buttonRowPassw.height  + middleSep2FA.height + input2FAuth.height

        Rectangle {
            color : "transparent"
            width : Style.main.dummy
            height : (root.height - dialog2FA.heightInputs)/2
        }

        InputField {
            id: input2FAuth
            label      : qsTr("Two Factor Code", "two factor code entry field")
            iconText   : Style.fa.lock
            onAccepted : root.okay()
        }

        Rectangle { id: middleSep2FA; color : "transparent"; width : Style.main.dummy; height : 2*Style.dialog.heightSeparator }

        Row {
            id: buttonRowPassw
            anchors.horizontalCenter: parent.horizontalCenter
            spacing: Style.dialog.fontSize
            ButtonRounded {
                id: buttonBack
                fa_icon: Style.fa.times
                text: qsTr("Back", "navigate back in add account flow")
                color_main: Style.dialog.text
                onClicked : root.cancel()
            }
            ButtonRounded {
                id: buttonNextTwo
                fa_icon: Style.fa.check
                text: qsTr("Next", "navigate to next page in add account flow")
                color_main: buttonOpaqueMain
                color_minor: Style.dialog.textBlue
                isOpaque: true
                onClicked : root.okay()
            }
        }
    }

    Column { // 2
        id: dialogWaitingAuth
        Rectangle { color : "transparent"; width : Style.main.dummy; height : (root.height-dialogWaitingAuthText.height) /2 }
        Text {
            id: dialogWaitingAuthText
            anchors.horizontalCenter: parent.horizontalCenter
            color: Style.dialog.text
            font.pointSize: Style.dialog.fontSize * Style.pt
            text : qsTr("Logging in") +"\n" + root.usernameElided
            horizontalAlignment: Text.AlignHCenter
        }
    }

    Column { // 3
        id: dialogMailboxPassword
        property int heightInputs : buttonRowMailbox.height + inputPasswMailbox.height + middleSepMailbox.height

        Rectangle { color : "transparent"; width : Style.main.dummy; height : (root.height - dialogMailboxPassword.heightInputs)/2}

        InputField {
            id: inputPasswMailbox
            label      : qsTr("Mailbox password for %1", "mailbox password entry field").arg(root.usernameElided)
            iconText   : Style.fa.lock
            isPassword : true
            onAccepted : root.okay()
        }

        Rectangle { id: middleSepMailbox; color : "transparent"; width : Style.main.dummy; height : 2*Style.dialog.heightSeparator }

        Row {
            id: buttonRowMailbox
            anchors.horizontalCenter: parent.horizontalCenter
            spacing: Style.dialog.fontSize
            ButtonRounded {
                id: buttonBackBack
                fa_icon: Style.fa.times
                text: qsTr("Back", "navigate back in add account flow")
                color_main: Style.dialog.text
                onClicked : root.cancel()
            }
            ButtonRounded {
                id: buttonLogin
                fa_icon: Style.fa.check
                text: qsTr("Next", "navigate to next page in add account flow")
                color_main: buttonOpaqueMain
                color_minor: Style.dialog.textBlue
                isOpaque: true
                onClicked : root.okay()
            }
        }
    }

    Column { // 4
        id: dialogWaitingAccount

        Rectangle { color : "transparent"; width : Style.main.dummy; height : (root.height - dialogWaitingAccountText.height )/2 }

        Text {
            id: dialogWaitingAccountText
            anchors.horizontalCenter: parent.horizontalCenter
            color: Style.dialog.text
            font {
                bold : true
                pointSize: Style.dialog.fontSize * Style.pt
            }
            text : qsTr("Adding account, please wait ...", "displayed after user has logged in, before new account is displayed")
            wrapMode: Text.Wrap
        }
    }

    Column { // 5
        id: dialogFirstUserAdded

        Rectangle { color : "transparent"; width : Style.main.dummy; height : (root.height - dialogWaitingAccountText.height - okButton.height*2 )/2 }

        Text {
            id: textFirstUser
            anchors.horizontalCenter: parent.horizontalCenter
            color: Style.dialog.text
            font {
                bold : false
                pointSize: Style.dialog.fontSize * Style.pt
            }
            width: 2*root.width/3
            horizontalAlignment: Text.AlignHCenter
            textFormat: Text.RichText
            text: "<html><style>a { font-weight: bold; text-decoration: none; color: white;}</style>"+
            qsTr("Now you need to configure your client(s) to use the Bridge. Instructions for configuring your client can be found at", "") +
            "<br/><a href=\"https://protonmail.com/bridge/clients\">https://protonmail.com/bridge/clients</a>.<html>"
            wrapMode: Text.Wrap
            onLinkActivated: {
                Qt.openUrlExternally(link)
            }
            MouseArea {
                anchors.fill: parent
                cursorShape: parent.hoveredLink=="" ? Qt.PointingHandCursor : Qt.WaitCursor
                acceptedButtons: Qt.NoButton
            }
        }

        Rectangle { color : "transparent"; width : Style.main.dummy; height : okButton.height}

        ButtonRounded{
            id: okButton
            anchors.horizontalCenter: parent.horizontalCenter
            color_main: buttonOpaqueMain
            color_minor: Style.main.textBlue
            isOpaque: true
            text: qsTr("Okay", "confirms and dismisses a notification")
            onClicked: root.hide()
        }
    }

    function clear_user() {
        inputUsername.text      = ""
        inputUsername.rightIcon = ""
    }

    function clear_passwd() {
        inputPassword.text    = ""
        inputPassword.rightIcon    = ""
        inputPassword.hidePasswordText()
    }


    function clear_2fa() {
        input2FAuth.text      = ""
        input2FAuth.rightIcon = ""
    }

    function clear_passwd_mailbox() {
        inputPasswMailbox.text  = ""
        inputPasswMailbox.rightIcon  = ""
        inputPasswMailbox.hidePasswordText()
    }

    onCancel : {
        root.warning.visible=false
        if (currentIndex==0) {
            root.hide()
        } else {
            clear_passwd()
            clear_passwd_mailbox()
            currentIndex=0
        }
    }


    function check_inputs() {
        var isOK = true
        switch (currentIndex) {
            case nameAndPasswordIndex :
            isOK &= inputUsername.checkNonEmpty()
            isOK &= inputPassword.checkNonEmpty()
            break
            case twoFAIndex :
            isOK &= input2FAuth.checkNonEmpty()
            break
            case mailboxIndex :
            isOK &= inputPasswMailbox.checkNonEmpty()
            break
        }
        if (isOK) {
            warning.visible = false
            warning.text= ""
        } else {
            setWarning(qsTr("Field required", "a field that must be filled in to submit form"),0)
        }
        return isOK
    }

    function setWarning(msg, changeIndex) {
        // show message
        root.warning.text = msg
        root.warning.visible = true
    }


    onOkay : {
        var isOK = check_inputs()
        if (isOK) {
            root.origin = root.currentIndex
            switch (root.currentIndex) {
                case nameAndPasswordIndex:
                case twoFAIndex:
                root.currentIndex = waitingAuthIndex
                break;
                case mailboxIndex:
                root.currentIndex = addingAccIndex
            }
            timer.start()
        }
    }

    onShow: {
        root.title = qsTr ("Log in to your ProtonMail account", "displayed on screen when user enters username to begin adding account")
        root.warning.visible = false
        inputUsername.forceFocus()
        root.isFirstAccount =  go.isFirstStart && accountsModel.count==0
    }

    function startAgain() {
        clear_passwd()
        clear_2fa()
        clear_passwd_mailbox()
        root.currentIndex = nameAndPasswordIndex
        root.inputPassword.focusInput = true
    }

    function finishLogin(){
        root.currentIndex = addingAccIndex
        var auth = go.addAccount(inputPasswMailbox.text)
        if (auth<0) {
            startAgain()
            return
        }
    }

    Connections {
        target: timer

        onTriggered : {
            timer.repeat = false
            switch (root.origin) {
                case nameAndPasswordIndex:
                var auth = go.login(inputUsername.text, inputPassword.text)
                if (auth < 0) {
                    startAgain()
                    break
                }
                if (auth == 1) {
                    root.currentIndex = twoFAIndex
                    root.input2FAuth.focusInput = true
                    break
                }
                if (auth == 2) {
                    root.currentIndex = mailboxIndex
                    root.inputPasswMailbox.focusInput = true
                    break
                }
                root.inputPasswMailbox.text = inputPassword.text
                root.finishLogin()
                break;
                case twoFAIndex:
                var auth = go.auth2FA(input2FAuth.text)
                if (auth < 0) {
                    startAgain()
                    break
                }
                if (auth == 1) {
                    root.currentIndex = mailboxIndex
                    root.inputPasswMailbox.focusInput = true
                    break
                }
                root.inputPasswMailbox.text = inputPassword.text
                root.finishLogin()
                break;
                case mailboxIndex:
                root.finishLogin()
                break;
            }
        }
    }

    onHide: {
        // because hide slot is conneceted to processFinished it will update
        // the list evertyime `go` obejcet is finished
        clear_passwd()
        clear_passwd_mailbox()
        clear_2fa()
        clear_user()
        go.loadAccounts()
        if (root.isFirstAccount && accountsModel.count==1) {
            root.isFirstAccount=false
            root.currentIndex=5
            root.show()
            root.title=qsTr("Success, Account Added!", "shown after successful account addition")
        }
    }

    Keys.onPressed: {
        if (event.key == Qt.Key_Enter) {
            root.okay()
        }
    }
}
