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

// Important information under title bar

import QtQuick 2.8
import QtQuick.Window 2.2
import QtQuick.Controls 2.1
import ProtonUI 1.0

Rectangle {
    id: root
    height: 0
    visible: state != "ok"
    state: "ok"
    color: "black"
    property var fontSize : 1.0 * Style.main.fontSize

    Row {
        anchors.centerIn: root
        visible: root.visible
        spacing: Style.main.leftMarginButton

        AccessibleText {
            id: message
            font.pointSize: root.fontSize * Style.pt

            text: qsTr("Connection security error: Your network connection to Proton services may be insecure.", "message in bar showed when TLS Pinning fails")
        }

        ClickIconText {
            anchors.verticalCenter : message.verticalCenter
            iconText  : ""
            text      : qsTr("Learn more", "This button opens TLS Pinning issue modal with more explanation")
            visible   : root.visible
            onClicked : {
                winMain.dialogTlsCert.show()
            }
            fontSize : root.fontSize
            textUnderline: true
        }
    }


    states: [
        State {
            name: "notOK"
            PropertyChanges {
                target: root
                height: 2* Style.main.fontSize
                color: Style.main.textRed
            }
        }
    ]
}
