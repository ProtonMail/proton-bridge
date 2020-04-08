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

// Popup message
import QtQuick 2.8
import QtQuick.Controls 2.2
import ProtonUI 1.0

Rectangle {
    id: root
    color: Style.transparent
    property alias text: message.text
    visible: false

    MouseArea { // prevent action below
        anchors.fill: parent
        hoverEnabled: true
    }

    Rectangle {
        id: backgroundInp
        anchors.centerIn : root
        color  : Style.errorDialog.background
        radius : Style.errorDialog.radius
        width  : parent.width/3.
        height : contentInp.height

        Column {
            id: contentInp
            anchors.horizontalCenter: backgroundInp.horizontalCenter
            spacing: Style.dialog.heightSeparator
            topPadding: Style.dialog.heightSeparator
            bottomPadding: Style.dialog.heightSeparator

            AccessibleText {
                id: message
                font {
                    pointSize : Style.errorDialog.fontSize * Style.pt
                    bold      : true
                }
                color: Style.errorDialog.text
                horizontalAlignment: Text.AlignHCenter
                width : backgroundInp.width - 2*Style.main.rightMargin
                wrapMode: Text.Wrap
            }

            ButtonRounded {
                text        : qsTr("Okay", "todo")
                isOpaque    : true
                color_main  : Style.dialog.background
                color_minor : Style.dialog.textBlue
                onClicked   : root.hide()
                anchors.horizontalCenter : parent.horizontalCenter
            }
        }
    }

    function show(text) {
        root.text = text
        root.visible = true
    }

    function hide() {
        root.state = "Okay"
        root.visible=false
    }
}
