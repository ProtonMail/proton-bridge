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

import QtQuick 2.8
import ProtonUI 1.0

Rectangle {
    id: root
    property string text: "undef"
    width: copyIcon.width + valueText.width
    height: Math.max(copyIcon.height, valueText.contentHeight)
    color: "transparent"

    Rectangle {
        id: copyIcon
        width: Style.info.leftMarginIcon*2 + Style.info.iconSize
        height : Style.info.iconSize
        color: "transparent"
        anchors {
            top: root.top
            left: root.left
        }
        Text {
            anchors.centerIn: parent
            font {
                pointSize : Style.info.iconSize * Style.pt
                family    : Style.fontawesome.name
            }
            color : Style.main.textInactive
            text: Style.fa.copy
        }
        MouseArea {
            anchors.fill: parent
            onClicked : {
                valueText.select(0, valueText.length)
                valueText.copy()
                valueText.deselect()
            }
            onPressed: copyIcon.scale = 0.90
            onReleased: copyIcon.scale = 1
        }

        Accessible.role: Accessible.Button
        Accessible.name: qsTr("Copy %1 to clipboard", "Click to copy the value to system clipboard.").arg(root.text)
        Accessible.description: Accessible.name
    }

    TextEdit {
        id: valueText
        width: Style.info.widthValue
        height:  Style.main.fontSize
        anchors {
            top: root.top
            left: copyIcon.right
        }
        font {
            pointSize: Style.main.fontSize * Style.pt
        }
        color: Style.main.text
        readOnly: true
        selectByMouse: true
        selectByKeyboard: true
        wrapMode: TextEdit.Wrap
        text: root.text
        selectionColor: Style.dialog.textBlue

        Accessible.role: Accessible.StaticText
        Accessible.name: root.text
        Accessible.description: Accessible.name
    }
}
