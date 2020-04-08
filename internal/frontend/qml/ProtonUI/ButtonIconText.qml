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

//  Button with full window width containing two icons (left and right) and text

import QtQuick 2.8
import QtQuick.Controls 2.1
import ProtonUI 1.0

AccessibleButton {
    id: root
    property alias leftIcon  : leftIcon
    property alias rightIcon : rightIcon
    property alias main      : mainText

    // dimensions
    width  : viewContent.width
    height : Style.main.heightRow
    topPadding: 0
    bottomPadding: 0
    leftPadding: Style.main.leftMargin
    rightPadding: Style.main.rightMargin

    background : Rectangle{
        color: Qt.lighter(Style.main.background, root.hovered || root.activeFocus ? ( root.pressed ? 1.2: 1.1) :1.0)
        // line
        Rectangle {
            anchors.bottom : parent.bottom
            width          : parent.width
            height         : Style.main.heightLine
            color          : Style.main.line
        }
        // pointing cursor
        MouseArea {
            anchors.fill : parent
            cursorShape  : Qt.PointingHandCursor
            acceptedButtons: Qt.NoButton
        }
    }

    contentItem : Rectangle {
        color: "transparent"
        // Icon left
        Text {
            id: leftIcon
            anchors {
                verticalCenter : parent.verticalCenter
                left           : parent.left
            }
            font {
                family    : Style.fontawesome.name
                pointSize : Style.settings.iconSize * Style.pt
            }
            color       : Style.main.textBlue
            text        : Style.fa.hashtag
        }

        // Icon/Text right
        Text {
            id: rightIcon
            anchors {
                verticalCenter : parent.verticalCenter
                right           : parent.right
            }
            font {
                family    : Style.fontawesome.name
                pointSize : Style.settings.iconSize * Style.pt
            }
            color       : Style.main.textBlue
            text        : Style.fa.hashtag
        }

        // Label
        Text {
            id: mainText
            anchors {
                verticalCenter : parent.verticalCenter
                left           : leftIcon.right
                leftMargin     : leftIcon.text!="" ? Style.main.leftMargin : 0
            }
            font.pointSize  : Style.settings.fontSize * Style.pt
            color           : Style.main.text
            text            : root.text
        }
    }
}
