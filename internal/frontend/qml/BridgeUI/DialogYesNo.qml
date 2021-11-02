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

// Dialog with Yes/No buttons

import QtQuick 2.8
import BridgeUI 1.0
import ProtonUI 1.0

Dialog {
    id: root

    title : ""

    property string input

    property alias question  : msg.text
    property alias note      : noteText.text
    property alias answer    : answ.text
    property alias buttonYes : buttonYes
    property alias buttonNo  : buttonNo

    isDialogBusy: currentIndex==1

    signal confirmed()

    Column {
        id: dialogMessage
        property int heightInputs : msg.height+
        middleSep.height+
        buttonRow.height +
        (checkboxSep.visible ? checkboxSep.height : 0 ) +
        (noteSep.visible ? noteSep.height : 0 ) +
        (checkBoxWrapper.visible ? checkBoxWrapper.height : 0 ) +
        (root.note!="" ? noteText.height : 0 )

        Rectangle { color : "transparent"; width : Style.main.dummy; height : (root.height-dialogMessage.heightInputs)/2 }

        AccessibleText {
            id:noteText
            anchors.horizontalCenter: parent.horizontalCenter
            color: Style.dialog.text
            font {
                pointSize: Style.dialog.fontSize * Style.pt
                bold: false
            }
            width: 2*root.width/3
            horizontalAlignment: Text.AlignHCenter
            wrapMode: Text.Wrap
        }
        Rectangle { id: noteSep; visible: note!=""; color : "transparent"; width : Style.main.dummy; height : Style.dialog.heightSeparator}

        AccessibleText {
            id: msg
            anchors.horizontalCenter: parent.horizontalCenter
            color: Style.dialog.text
            font {
                pointSize: Style.dialog.fontSize * Style.pt
                bold: true
            }
            width: 2*parent.width/3
            text : ""
            horizontalAlignment: Text.AlignHCenter
            wrapMode: Text.Wrap
        }

        Rectangle { id: checkboxSep; visible: checkBoxWrapper.visible; color : "transparent"; width : Style.main.dummy; height : Style.dialog.heightSeparator}
        Row {
            id: checkBoxWrapper
            property bool isChecked : false
            visible: root.state=="deleteUser"
            anchors.horizontalCenter: parent.horizontalCenter
            spacing: Style.dialog.spacing

            function toggle() {
                checkBoxWrapper.isChecked = !checkBoxWrapper.isChecked
            }

            Text {
                id: checkbox
                font {
                    pointSize : Style.dialog.iconSize * Style.pt
                    family    : Style.fontawesome.name
                }
                anchors.verticalCenter : parent.verticalCenter
                text: checkBoxWrapper.isChecked ? Style.fa.check_square_o : Style.fa.square_o
                color: checkBoxWrapper.isChecked ? Style.main.textBlue : Style.main.text

                MouseArea {
                    anchors.fill: parent
                    onPressed: checkBoxWrapper.toggle()
                    cursorShape: Qt.PointingHandCursor
                }
            }
            Text {
                id: checkBoxNote
                anchors.verticalCenter : parent.verticalCenter
                text: qsTr("Additionally delete all stored preferences and data", "when removing an account, this extra preference additionally deletes all cached data")
                color: Style.main.text
                font.pointSize: Style.dialog.fontSize * Style.pt

                MouseArea {
                    anchors.fill: parent
                    onPressed: checkBoxWrapper.toggle()
                    cursorShape: Qt.PointingHandCursor

                    Accessible.role: Accessible.CheckBox
                    Accessible.checked: checkBoxWrapper.isChecked
                    Accessible.name: checkBoxNote.text
                    Accessible.description: checkBoxNote.text
                    Accessible.ignored: checkBoxNote.text == ""
                    Accessible.onToggleAction: checkBoxWrapper.toggle()
                    Accessible.onPressAction: checkBoxWrapper.toggle()
                }
            }
        }

        Rectangle { id: middleSep; color : "transparent"; width : Style.main.dummy; height : 2*Style.dialog.heightSeparator }

        Row {
            id: buttonRow
            anchors.horizontalCenter: parent.horizontalCenter
            spacing: Style.dialog.spacing
            ButtonRounded {
                id:buttonNo
                visible: root.state != "toggleEarlyAccess"
                color_main: Style.dialog.text
                fa_icon: Style.fa.times
                text: qsTr("No")
                onClicked : root.hide()
            }
            ButtonRounded {
                id: buttonYes
                color_main: Style.dialog.text
                color_minor: Style.main.textBlue
                isOpaque: true
                fa_icon: Style.fa.check
                text: root.state == "toggleEarlyAccess" ?  qsTr("Ok") : qsTr("Yes")
                onClicked : {
                    currentIndex=1
                    root.confirmed()
                }
            }
        }
    }


    Column {
        Rectangle { color  : "transparent";  width  : Style.main.dummy; height : (root.height-answ.height)/2 }
        AccessibleText {
            id: answ
            anchors.horizontalCenter: parent.horizontalCenter
            color: Style.old.pm_white
            font {
                pointSize : Style.dialog.fontSize * Style.pt
                bold      : true
            }
            width: 3*parent.width/4
            horizontalAlignment: Text.AlignHCenter
            text : qsTr("Waiting...", "in general this displays between screens when processing data takes a long time")
            wrapMode: Text.Wrap
        }
    }


    states : [
        State {
            name: "quit"
            PropertyChanges {
                target: root
                currentIndex : 0
                title        : qsTr("Close Bridge", "quits the application")
                question     : qsTr("Are you sure you want to close the Bridge?", "asked when user tries to quit the application")
                note         : ""
                answer       : qsTr("Closing Bridge...", "displayed when user is quitting application")
            }
        },
        State {
            name: "logout"
            PropertyChanges {
                target: root
                currentIndex : 1
                title        : qsTr("Logout", "title of page that displays during account logout")
                question     : ""
                note         : ""
                answer       : qsTr("Logging out...", "displays during account logout")
            }
        },
        State {
            name: "deleteUser"
            PropertyChanges {
                target: root
                currentIndex : 0
                title        : qsTr("Delete account", "title of page that displays during account deletion")
                question     : qsTr("Are you sure you want to remove this account?", "displays during account deletion")
                note         : ""
                answer       : qsTr("Deleting ...", "displays during account deletion")
            }
        },
        State {
            name: "clearChain"
            PropertyChanges {
                target       : root
                currentIndex : 0
                title        : qsTr("Clear keychain", "title of page that displays during keychain clearing")
                question     : qsTr("Are you sure you want to clear your keychain?", "displays during keychain clearing")
                note         : qsTr("This will remove all accounts that you have added to the Bridge and disconnect you from your email client(s).", "displays during keychain clearing")
                answer       : qsTr("Clearing the keychain ...", "displays during keychain clearing")
            }
        },
        State {
            name: "clearCache"
            PropertyChanges {
                target: root
                currentIndex : 0
                title        : qsTr("Clear cache", "title of page that displays during cache clearing")
                question     : qsTr("Are you sure you want to clear your local cache?", "displays during cache clearing")
                note         : qsTr("This will delete all of your stored preferences as well as cached email data for all accounts, and requires you to reconfigure your client.", "displays during cache clearing")
                answer       : qsTr("Clearing the cache ...", "displays during cache clearing")
            }
        },
        State {
            name: "checkUpdates"
            PropertyChanges {
                target: root
                currentIndex : 1
                title        : ""
                question     : ""
                note         : ""
                answer       : qsTr("Checking for updates ...", "displays if user clicks the Check for Updates button in the Help tab")
            }
        },
        State {
            name: "addressmode"
            PropertyChanges {
                target: root
                currentIndex : 0
                title        : ""
                question     : qsTr("Do you want to continue?", "asked when the user changes between split and combined address mode")
                note         : qsTr("Changing between split and combined address mode will require you to delete your account(s) from your email client and begin the setup process from scratch.", "displayed when the user changes between split and combined address mode")
                answer       : qsTr("Configuring address mode...", "displayed when the user changes between split and combined address mode")
            }
        },
        State {
            name: "toggleAutoStart"
            PropertyChanges {
                target: root
                currentIndex : 1
                question     : ""
                note         : ""
                title        : ""
                answer       : {
                    var msgTurnOn = qsTr("Turning on automatic start of Bridge...", "when the automatic start feature is selected")
                    var msgTurnOff = qsTr("Turning off automatic start of Bridge...", "when the automatic start feature is deselected")
                    return go.isAutoStart==false ? msgTurnOff  :  msgTurnOn
                }
            }
        },
        State {
            name: "toggleAllowProxy"
            PropertyChanges {
                target: root
                currentIndex : 0
                question     : {
                    var questionTurnOn = qsTr("Do you want to allow alternative routing?")
                    var questionTurnOff = qsTr("Do you want to disallow alternative routing?")
                    return go.isProxyAllowed==false ? questionTurnOn  :  questionTurnOff
                }
                note         : qsTr("In case Proton sites are blocked, this setting allows Bridge to try alternative network routing to reach Proton, which can be useful for bypassing firewalls or network issues. We recommend keeping this setting on for greater reliability.")
                title        : {
                    var titleTurnOn = qsTr("Allow alternative routing")
                    var titleTurnOff = qsTr("Disallow alternative routing")
                    return go.isProxyAllowed==false ? titleTurnOn  :  titleTurnOff
                }
                answer       : {
                    var msgTurnOn = qsTr("Allowing Bridge to use alternative routing to connect to Proton...", "when the allow proxy feature is selected")
                    var msgTurnOff = qsTr("Disallowing Bridge to use alternative routing to connect to Proton...", "when the allow proxy feature is deselected")
                    return go.isProxyAllowed==false ? msgTurnOn  :  msgTurnOff
                }
            }
        },
        State {
            name: "toggleEarlyAccessOn"
            PropertyChanges {
                target: root
                currentIndex : 0
                question     : qsTr("Do you want to be the first to get the latest updates? Please keep in mind that early versions may be less stable.")
                note         : ""
                title        : qsTr("Enable early access")
                answer       : qsTr("Enabling early access...")
            }
        },
        State {
            name: "toggleEarlyAccessOff"
            PropertyChanges {
                target: root
                currentIndex : 0
                question     : qsTr("Are you sure you want to leave early access? Please keep in mind this operation clears the cache and restarts Bridge.")
                note         : qsTr("This will delete all of your stored preferences as well as cached email data for all accounts, and requires you to reconfigure your client.")
                title        : qsTr("Disable early access")
                answer       : qsTr("Disabling early access...")
            }
        },
        State {
            name: "noKeychain"
            PropertyChanges {
                target: root
                currentIndex : 0
                note     : qsTr(
                    "%1 is not able to detected a supported password manager (pass, gnome-keyring or other provider of FreeDesktop Secret Service API). Please install and setup supported password manager and restart the application.",
                    "Error message when no keychain is detected"
                ).arg(go.programTitle)
                question         : qsTr("Do you want to close application now?", "when no password manager found." )
                title        : "No system password manager detected"
                answer       : qsTr("Closing Bridge...", "displayed when user is quitting application")
            }
        },
        State {
            name: "undef";
            PropertyChanges {
                target: root
                currentIndex : 1
                question     : ""
                note         : ""
                title        : ""
                answer       : ""
            }
        }
    ]


    Shortcut {
        sequence: StandardKey.Cancel
        onActivated: root.hide()
    }

    Shortcut {
        sequence: "Enter"
        onActivated: root.confirmed()
    }

    onHide: {
        checkBoxWrapper.isChecked = false
        state = "undef"
    }

    onShow: {
        // hide all other dialogs
        winMain.dialogAddUser     .visible = false
        winMain.dialogChangePort  .visible = false
        winMain.dialogCredits     .visible = false
        root.visible = true
    }

    onConfirmed : {
        if (state == "quit" || state == "instance exists" ) {
            timer.interval = 1000
        } else {
            timer.interval = 300
        }
        answ.forceActiveFocus()
        timer.start()
    }

    Connections {
        target: timer
        onTriggered: {
            if ( state == "addressmode"       ) { go.switchAddressMode   (input) }
            if ( state == "clearChain"        ) { go.clearKeychain       ()      }
            if ( state == "clearCache"        ) { go.clearCache          ()      }
            if ( state == "deleteUser"        ) { go.deleteAccount       (input, checkBoxWrapper.isChecked) }
            if ( state == "logout"            ) { go.logoutAccount       (input) }
            if ( state == "toggleAutoStart"   ) { go.toggleAutoStart     ()      }
            if ( state == "toggleAllowProxy"  ) { go.toggleAllowProxy    ()      }
            if ( state == "toggleEarlyAccessOn"  ) { go.toggleEarlyAccess   ()      }
            if ( state == "toggleEarlyAccessOff" ) { go.toggleEarlyAccess   ()      }
            if ( state == "quit"              ) { Qt.quit                ()      }
            if ( state == "instance exists"   ) { Qt.quit                ()      }
            if ( state == "noKeychain"        ) { Qt.quit                ()      }
            if ( state == "checkUpdates"      ) { }
        }
    }

    Keys.onPressed: {
        if (event.key == Qt.Key_Enter) {
            root.confirmed()
        }
    }
}
