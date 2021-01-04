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

// one line input text field with label

import QtQuick 2.8
import QtQuick.Controls 2.1
import ProtonUI 1.0

Column {
    id: root
    property alias focusInput      : inputField.focus
    property alias label           : textlabel.text
    property alias iconText        : iconInput.text
    property alias placeholderText : inputField.placeholderText
    property alias text            : inputField.text
    property bool isPassword       : false
    property string rightIcon      : ""

    signal accepted()
    signal editingFinished()

    spacing: Style.dialog.heightSeparator
    anchors.horizontalCenter : parent.horizontalCenter

    AccessibleText {
        id: textlabel
        anchors.left : parent.left
        font {
            pointSize : Style.dialog.fontSize * Style.pt
            bold      : true
        }
        horizontalAlignment: Text.AlignHCenter
        color : Style.dialog.text
    }

    Rectangle {
        id: inputWrap
        anchors.horizontalCenter : parent.horizontalCenter
        width  : Style.dialog.widthInput
        height : Style.dialog.heightInput
        color : "transparent"

        Text {
            id: iconInput
            anchors {
                top  : parent.top
                left : parent.left
            }
            color          : Style.dialog.text
            font {
                pointSize : Style.dialog.iconSize * Style.pt
                family : Style.fontawesome.name
            }
            text: "o"
        }

        TextField {
            id: inputField
            anchors {
                fill: inputWrap
                leftMargin   : Style.dialog.iconSize+Style.dialog.fontSize
                bottomMargin : inputWrap.height - Style.dialog.iconSize
            }
            verticalAlignment   : TextInput.AlignTop
            horizontalAlignment : TextInput.AlignLeft
            selectByMouse       : true
            color               : Style.dialog.text
            selectionColor      : Style.main.textBlue
            font {
                pointSize : Style.dialog.fontSize * Style.pt
            }
            padding: 0
            background: Rectangle {
                anchors.fill: parent
                color : "transparent"
            }
            Component.onCompleted : {
                if (isPassword) {
                    echoMode = TextInput.Password
                } else  {
                    echoMode = TextInput.Normal
                }
            }

            Accessible.name: textlabel.text
            Accessible.description: textlabel.text
        }

        Text {
            id: iconRight
            anchors {
                top  : parent.top
                right : parent.right
            }
            color          : Style.dialog.text
            font {
                pointSize : Style.dialog.iconSize * Style.pt
                family    : Style.fontawesome.name
            }
            text: ( !isPassword ? "" : ( 
                inputField.echoMode == TextInput.Password ? Style.fa.eye : Style.fa.eye_slash
            )) + " " + rightIcon
            MouseArea {
                anchors.fill: parent
                onClicked: {
                    if (isPassword) {
                        if (inputField.echoMode == TextInput.Password) inputField.echoMode = TextInput.Normal
                        else inputField.echoMode = TextInput.Password
                    }
                }
            }
        }

        Rectangle {
            anchors {
                left   : parent.left
                right  : parent.right
                bottom : parent.bottom
            }
            height: Math.max(Style.main.border,1)
            color: Style.dialog.text
        }
    }

    function clear() {
        inputField.text = ""
        rightIcon = ""
    }

    function checkNonEmpty() {
        if (inputField.text == "") {
            rightIcon = Style.fa.exclamation_triangle
            root.placeholderText = ""
            inputField.focus = true
            return false
        } else {
            rightIcon = Style.fa.check_circle
        }
        return true
    }

    function hidePasswordText() {
        if (root.isPassword) inputField.echoMode = TextInput.Password
    }

    function checkIsANumber(){
        if (/^\d+$/.test(inputField.text)) {
            rightIcon = Style.fa.check_circle
            return true
        }
        rightIcon = Style.fa.exclamation_triangle
        root.placeholderText = ""
        inputField.focus = true
        return false
    }

    function forceFocus() {
        inputField.forceActiveFocus()
    }

    Connections {
        target: inputField
        onAccepted: root.accepted()
        onEditingFinished: root.editingFinished()
    }

    Keys.onPressed: {
        if (event.key == Qt.Key_Enter) {
            root.accepted()
        }
    }
}
