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

// one line input text field with label
import QtQuick 2.8
import QtQuick.Controls 2.2
import QtQuick.Controls.Styles 1.4
import ProtonUI 1.0
import QtGraphicalEffects 1.0

Column {
    id: root
    property alias label: textlabel.text
    property alias placeholderText: inputField.placeholderText
    property alias echoMode: inputField.echoMode
    property alias text: inputField.text
    property alias field: inputField

    signal accepted()

    spacing: Style.dialog.heightSeparator

    Text {
        id: textlabel
        font {
            pointSize: Style.dialog.fontSize * Style.pt
            bold: true
        }
        color: Style.dialog.text
    }


    TextField {
        id: inputField
        width: Style.dialog.widthInput
        height: Style.dialog.heightButton
        selectByMouse  : true
        selectionColor : Style.main.textBlue
        padding        : Style.dialog.radiusButton
        color          : Style.dialog.text
        font {
            pointSize : Style.dialog.fontSize * Style.pt
        }
        background: Rectangle {
            color : Style.dialog.background
            radius:  Style.dialog.radiusButton
            border {
                color : Style.dialog.line
                width : Style.dialog.borderInput
            }
            layer.enabled: true
            layer.effect: FastBlur {
                anchors.fill: parent
                radius: 8 * Style.px
            }
        }
    }

    Connections {
        target     : inputField
        onAccepted : root.accepted()
    }
}
