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

// No border button with icon

import QtQuick 2.8
import QtQuick.Controls 2.1
import ProtonUI 1.0

AccessibleButton {
    id: root

    property string iconText      : Style.fa.hashtag
    property color  textColor     : Style.main.text
    property int    fontSize      : Style.main.fontSize
    property int    iconSize      : Style.main.iconSize
    property int    margin        : iconText!="" ? Style.main.leftMarginButton : 0.0
    property bool   iconOnRight   : false
    property bool   textBold      : false
    property bool   textUnderline : false


    TextMetrics {
        id: metrics
        text: root.text
        font: showText.font
    }

    TextMetrics {
        id: metricsIcon
        text : root.iconText
        font : showIcon.font
    }

    scale   : root.pressed ? 0.96 : root.activeFocus ? 1.05 : 1.0
    height  : Math.max(metrics.height, metricsIcon.height)
    width   : metricsIcon.width*1.5 + margin + metrics.width + 4.0
    padding : 0.0

    background : Rectangle {
        color: Style.transparent
        MouseArea {
            anchors.fill : parent
            cursorShape  : Qt.PointingHandCursor
            acceptedButtons: Qt.NoButton
        }
    }

    contentItem : Rectangle {
        color: Style.transparent
        Text {
            id: showIcon
            anchors {
                left           : iconOnRight ? showText.right : parent.left
                leftMargin     : iconOnRight ? margin : 0
                verticalCenter : parent.verticalCenter
            }
            font {
                pointSize : iconSize * Style.pt
                family    : Style.fontawesome.name
            }
            color : textColor
            text  : root.iconText
        }

        Text {
            id: showText
            anchors {
                verticalCenter : parent.verticalCenter
                left           : iconOnRight ? parent.left : showIcon.right
                leftMargin     : iconOnRight ? 0 : margin
            }
            color : textColor
            font {
                pointSize : root.fontSize * Style.pt
                bold: root.textBold
                underline: root.textUnderline
            }
            text  : root.text
        }
    }
}


