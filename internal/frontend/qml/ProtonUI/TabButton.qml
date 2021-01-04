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

// Button with text and icon for tabbar

import QtQuick 2.8
import ProtonUI 1.0
import QtQuick.Controls 2.1

AccessibleButton {
    id: root
    property alias iconText  : icon.text
    property alias title     : titleText.text
    property color textColor : {
        if (root.state=="deactivated")  {
            return Qt.lighter(Style.tabbar.textInactive, root.hovered || root.activeFocus ? 1.25 : 1.0)
        }
        if (root.state=="activated") {
            return Style.tabbar.text
        }
    }

    text: root.title
    Accessible.description: root.title + " tab"

    width  : titleMetrics.width // Style.tabbar.widthButton
    height : Style.tabbar.heightButton
    padding: 0

    background: Rectangle {
        color  : Style.transparent

        MouseArea {
            anchors.fill: parent
            cursorShape: Qt.PointingHandCursor
            acceptedButtons: Qt.NoButton
        }
    }

    contentItem : Rectangle {
        color: "transparent"
        scale     : root.pressed ? 0.96 : 1.00

        Text {
            id: icon
            // dimenstions
            anchors {
                top : parent.top
                horizontalCenter : parent.horizontalCenter
            }
            // style
            color : root.textColor
            font  {
                family : Style.fontawesome.name
                pointSize   : Style.tabbar.iconSize * Style.pt
            }
        }

        TextMetrics {
            id: titleMetrics
            text : root.title
            font.pointSize : titleText.font.pointSize
        }

        Text {
            id: titleText
            // dimenstions
            anchors {
                bottom           : parent.bottom
                horizontalCenter : parent.horizontalCenter
            }
            // style
            color : root.textColor
            font {
                pointSize : Style.tabbar.fontSize * Style.pt
                bold      : root.state=="activated"
            }
        }
    }

    states: [
        State {
            name: "activated"
        },
        State {
            name: "deactivated"
        }
    ]
}
