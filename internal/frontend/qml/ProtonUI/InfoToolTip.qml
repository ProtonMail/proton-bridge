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

// on hover information

import QtQuick 2.8
import QtQuick.Controls 2.2
import ProtonUI 1.0

Text { // info icon
    id:root
    property alias info : tip.text
    font {
        family: Style.fontawesome.name
        pointSize : Style.dialog.iconSize * Style.pt
    }
    text: Style.fa.info_circle
    color: Style.main.textDisabled

    MouseArea {
        anchors.fill: parent
        hoverEnabled: true

        onEntered : tip.visible=true
        onExited  : tip.visible=false
    }

    ToolTip {
        id: tip
        width: Style.bubble.width
        x: - 0.2*tip.width
        y: - tip.height

        topPadding : Style.main.fontSize/2
        bottomPadding : Style.main.fontSize/2
        leftPadding : Style.bubble.widthPane + Style.dialog.spacing
        rightPadding: Style.dialog.spacing
        delay: 800

        background : Rectangle {
            id: bck
            color: Style.bubble.paneBackground
            radius : Style.bubble.radius


            Text {
                id: icon
                color: Style.bubble.background
                text: Style.fa.info_circle
                font {
                    family    : Style.fontawesome.name
                    pointSize : Style.dialog.iconSize * Style.pt
                }
                anchors {
                    verticalCenter : bck.verticalCenter
                    left           : bck.left
                    leftMargin     : (Style.bubble.widthPane - icon.width) / 2
                }
            }

            Rectangle { // right edge
                anchors {
                    fill       : bck
                    leftMargin  : Style.bubble.widthPane
                }
                radius: parent.radius
                color: Style.bubble.background
            }

            Rectangle { // center background
                anchors {
                    fill        : parent
                    leftMargin  : Style.bubble.widthPane
                    rightMargin : Style.bubble.widthPane
                }
                color: Style.bubble.background
            }
        }

        contentItem : Text {
            text: tip.text
            color: Style.bubble.text
            wrapMode: Text.Wrap
            font.pointSize: Style.main.fontSize * Style.pt
        }
    }
}
