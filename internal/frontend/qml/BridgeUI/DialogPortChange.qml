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
import QtQuick.Controls 2.2 as QC

Dialog {
    id: root

    title : "Set IMAP & SMTP settings"
    subtitle : "Changes require reconfiguration of Mail client. (Bridge will automatically restart)"
    isDialogBusy: currentIndex==1

    Column {
        id: dialogMessage
        property int heightInputs : imapPort.height + middleSep.height + smtpPort.height + buttonSep.height + buttonRow.height + secSMTPSep.height + securitySMTP.height

        Rectangle { color : "transparent"; width : Style.main.dummy; height : (root.height-dialogMessage.heightInputs)/1.6 }

        InputField {
            id: imapPort
            iconText : Style.fa.hashtag
            label    : qsTr("IMAP port", "entry field to choose port used for the IMAP server")
            text     : "undef"
        }

        Rectangle { id:middleSep; color : "transparent"; width : Style.main.dummy; height : Style.dialog.heightSeparator }

        InputField {
            id: smtpPort
            iconText : Style.fa.hashtag
            label    : qsTr("SMTP port", "entry field to choose port used for the SMTP server")
            text     : "undef"
        }

        Rectangle { id:secSMTPSep; color : Style.transparent; width : Style.main.dummy; height : Style.dialog.heightSeparator }

        // SSL button group
        Rectangle {
            anchors.horizontalCenter : parent.horizontalCenter
            width : Style.dialog.widthInput
            height : securitySMTPLabel.height + securitySMTP.height
            color : "transparent"

            AccessibleText {
                id: securitySMTPLabel
                anchors.left : parent.left
                text:qsTr("SMTP connection mode")
                color: Style.dialog.text
                font {
                    pointSize : Style.dialog.fontSize * Style.pt
                    bold      : true
                }
            }
            
            QC.ButtonGroup {
                buttons: securitySMTP.children
            }
            Row {
                id: securitySMTP
                spacing: Style.dialog.spacing
                anchors.top: securitySMTPLabel.bottom
                anchors.topMargin: Style.dialog.fontSize

                CheckBoxLabel {
                    id: securitySMTPSSL
                    text: qsTr("SSL")
                }

                CheckBoxLabel {
                    checked: true
                    id: securitySMTPSTARTTLS
                    text: qsTr("STARTTLS")
                }
            }
        }

        Rectangle { id:buttonSep; color : "transparent"; width : Style.main.dummy; height : 2*Style.dialog.heightSeparator }

        Row {
            id: buttonRow
            anchors.horizontalCenter: parent.horizontalCenter
            spacing: Style.dialog.spacing
            ButtonRounded {
                id:buttonNo
                color_main: Style.dialog.text
                fa_icon: Style.fa.times
                text: qsTr("Cancel", "dismisses current action")
                onClicked : root.hide()
            }
            ButtonRounded {
                id: buttonYes
                color_main: Style.dialog.text
                color_minor: Style.main.textBlue
                isOpaque: true
                fa_icon: Style.fa.check
                text: qsTr("Okay", "confirms and dismisses a notification")
                onClicked : root.confirmed()
            }
        }
    }

    Column {
        Rectangle { color  : "transparent";  width  : Style.main.dummy; height : (root.height-answ.height)/2 }
        Text {
            id: answ
            anchors.horizontalCenter: parent.horizontalCenter
            width : parent.width/2
            color: Style.dialog.text
            font {
                pointSize : Style.dialog.fontSize * Style.pt
                bold      : true
            }
            text : "IMAP: " + imapPort.text + "\nSMTP: " + smtpPort.text + "\nSMTP Connection Mode: " + getSelectedSSLMode() + "\n\n" +
            qsTr("Settings will be applied after the next start. You will need to reconfigure your email client(s).", "after user changes their ports they will see this notification to reconfigure their setup") +
            "\n\n" +
            qsTr("Bridge will now restart.", "after user changes their ports this appears to notify the user of restart")
            wrapMode: Text.Wrap
            horizontalAlignment: Text.AlignHCenter
        }
    }

    function areInputsOK() {
        var isOK = true
        var imapUnchanged = false
        var secSMTPUnchanged = (securitySMTPSTARTTLS.checked == go.isSMTPSTARTTLS())
        root.warning.text = ""

        if (imapPort.text!=go.getIMAPPort()) {
            if (go.isPortOpen(imapPort.text)!=0) {
                imapPort.rightIcon = Style.fa.exclamation_triangle
                root.warning.text = qsTr("Port number is not available.", "if the user changes one of their ports to a port that is occupied by another application")
                isOK=false
            } else {
                imapPort.rightIcon = Style.fa.check_circle
            }
        } else {
            imapPort.rightIcon = ""
            imapUnchanged = true
        }

        if (smtpPort.text!=go.getSMTPPort()) {
            if (go.isPortOpen(smtpPort.text)!=0) {
                smtpPort.rightIcon = Style.fa.exclamation_triangle
                root.warning.text = qsTr("Port number is not available.", "if the user changes one of their ports to a port that is occupied by another application")
                isOK=false
            } else {
                smtpPort.rightIcon = Style.fa.check_circle
            }
        } else {
            smtpPort.rightIcon = ""
            if (imapUnchanged && secSMTPUnchanged) {
                root.warning.text = qsTr("Please change at least one port number or SMTP security.", "if the user tries to change IMAP/SMTP ports to the same ports as before")
                isOK=false
            }
        }

        if (imapPort.text == smtpPort.text) {
            smtpPort.rightIcon = Style.fa.exclamation_triangle
            root.warning.text = qsTr("Port numbers must be different.", "if the user sets both the IMAP and SMTP ports to the same number")
            isOK=false
        }

        root.warning.visible = !isOK
        return isOK
    }

    function confirmed() {
        if (areInputsOK()) {
            incrementCurrentIndex()
            timer.start()
        }
    }

    function getSelectedSSLMode() {
        if (securitySMTPSTARTTLS.checked == true) {
            return "STARTTLS"
        } else {
            return "SSL"
        }
    }

    onShow : {
        imapPort.text = go.getIMAPPort()
        smtpPort.text = go.getSMTPPort()
        if (go.isSMTPSTARTTLS()) {
            securitySMTPSTARTTLS.checked = true
        } else {
            securitySMTPSSL.checked = true
        }
        areInputsOK()
        root.warning.visible = false
    }

    Shortcut {
        sequence: StandardKey.Cancel
        onActivated: root.hide()
    }

    Shortcut {
        sequence: "Enter"
        onActivated: root.confirmed()
    }

    timer.interval : 3000

    Connections {
        target: timer
        onTriggered: {
            go.setPortsAndSecurity(imapPort.text, smtpPort.text, securitySMTPSTARTTLS.checked)
            go.isRestarting = true
            Qt.quit()
        }
    }
}
