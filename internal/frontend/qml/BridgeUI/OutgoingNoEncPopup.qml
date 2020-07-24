// Copyright (c) 2020 Proton Technologies AG
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

// Popup

import QtQuick 2.8
import QtQuick.Window 2.2
import BridgeUI 1.0
import ProtonUI 1.0

Window {
    id:root
    width  : Style.info.width
    height : Style.info.width/1.5
    minimumWidth  : Style.info.width
    minimumHeight : Style.info.width/1.5
    maximumWidth  : Style.info.width
    maximumHeight : Style.info.width/1.5
    color   : Style.main.background
    flags   : Qt.Window | Qt.Popup | Qt.FramelessWindowHint
    visible : false
    title   : ""
    x: 10
    y: 10
    property string messageID: ""

    // Drag and move
    MouseArea {
        property point diff: "0,0"
        property QtObject window: root

        anchors {
            fill: parent
        }

        onPressed: {
            diff = Qt.point(window.x, window.y)
            var mousePos = mapToGlobal(mouse.x, mouse.y)
            diff.x -= mousePos.x
            diff.y -= mousePos.y
        }

        onPositionChanged: {
            var currPos = mapToGlobal(mouse.x, mouse.y)
            window.x = currPos.x + diff.x
            window.y = currPos.y + diff.y
        }
    }

    Column {
        topPadding: Style.main.fontSize
        spacing: (root.height - (description.height + cancel.height + countDown.height + Style.main.fontSize))/3
        width: root.width

        Text {
            id: description
            color                    : Style.main.text
            font.pointSize           : Style.main.fontSize*Style.pt/1.2
            anchors.horizontalCenter : parent.horizontalCenter
            horizontalAlignment      : Text.AlignHCenter
            width                    : root.width - 2*Style.main.leftMargin
            wrapMode                 : Text.Wrap
            textFormat               : Text.RichText

            text: qsTr("The message with subject %1 has one or more recipients with no encryption settings. If you do not want to send this email click the cancel button.").arg("<h3>"+root.title+"</h3>")
        }

        Row {
            spacing : Style.dialog.spacing
            anchors.horizontalCenter: parent.horizontalCenter

            ButtonRounded {
                id: sendAnyway
                onClicked : root.hide(true)
                height: Style.main.fontSize*2
                //width: Style.dialog.widthButton*1.3
                fa_icon: Style.fa.send
                text: qsTr("Send now", "Confirmation of sending unencrypted email.")
            }

            ButtonRounded {
                id: cancel
                onClicked : root.hide(false)
                height: Style.main.fontSize*2
                //width: Style.dialog.widthButton*1.3
                fa_icon: Style.fa.times
                text: qsTr("Cancel", "Cancel the sending of current email")
            }
        }

        Text {
            id: countDown
            color: Style.main.text
            font.pointSize           : Style.main.fontSize*Style.pt/1.2
            anchors.horizontalCenter : parent.horizontalCenter
            horizontalAlignment      : Text.AlignHCenter
            width                    : root.width - 2*Style.main.leftMargin
            wrapMode                 : Text.Wrap
            textFormat               : Text.RichText

            text: qsTr("This popup will close after %1 and email will be sent unless you click the cancel button.").arg( "<b>" + timer.secLeft + "s</b>")
        }
    }

    Timer {
        id: timer
        property var secLeft: 0
        interval: 1000 //ms
        repeat: true
        onTriggered: {
            secLeft--
            if (secLeft <= 0) {
                root.hide(true)
            }
        }
    }

    function hide(shouldSend) {
        root.visible = false
        timer.stop()
        go.saveOutgoingNoEncPopupCoord(root.x, root.y)
        go.shouldSendAnswer(root.messageID, shouldSend)
    }

    function show(messageID, subject) {
        root.messageID = messageID
        root.title = subject
        root.visible = true
        timer.secLeft = 10
        timer.start()
    }
}


