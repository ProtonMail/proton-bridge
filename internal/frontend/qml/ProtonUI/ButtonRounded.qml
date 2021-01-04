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

// Classic button with icon and text

import QtQuick 2.8
import QtQuick.Controls 2.1
import QtGraphicalEffects 1.0
import ProtonUI 1.0

AccessibleButton {
    id: root
    property string fa_icon     : ""
    property color  color_main  : Style.dialog.text
    property color  color_minor : "transparent"
    property bool   isOpaque    : false

    text      : "undef"
    state     : root.hovered || root.activeFocus ? "hover" : "normal"
    width     : Style.dialog.widthButton
    height    : Style.dialog.heightButton
    scale     : root.pressed ? 0.96 : 1.00

    background: Rectangle {
        border {
            color : root.color_main
            width : root.isOpaque ? 0 : Style.dialog.borderButton
        }
        radius : Style.dialog.radiusButton
        color  : root.isOpaque ? root.color_minor : "transparent"

        MouseArea {
            anchors.fill : parent
            cursorShape  : Qt.PointingHandCursor
            acceptedButtons: Qt.NoButton
        }
    }

    contentItem: Rectangle {
        color: "transparent"

        Row {
            id: mainText
            anchors.centerIn: parent
            spacing: 0

            Text {
                font {
                    pointSize : Style.dialog.fontSize * Style.pt
                    family    : Style.fontawesome.name
                }
                color : color_main
                text  : root.fa_icon=="" ? "" : root.fa_icon + " "
            }

            Text {
                font {
                    pointSize : Style.dialog.fontSize * Style.pt
                }
                color : color_main
                text  : root.text
            }
        }

        Glow {
            id: mainTextEffect
            anchors.fill : mainText
            source: mainText
            color: color_main
            opacity: 0.33
        }
    }

    states :[
        State {name: "normal"; PropertyChanges{ target: mainTextEffect; radius: 0  ; visible: false } },
        State {name: "hover" ; PropertyChanges{ target: mainTextEffect; radius: 3*Style.px ; visible: true  } }
    ]
}
