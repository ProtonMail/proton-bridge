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

// Header of window with logo and buttons

import QtQuick 2.8
import ProtonUI 1.0
import QtQuick.Window 2.2


Rectangle {
    id: root
    // dimensions
    property Window parentWin
    property string title: "ProtonMail Bridge"
    property bool hasIcon : true
    anchors.top   : parent.top
    anchors.right : parent.right
    width         : Style.main.width
    height        : Style.title.height
    // style
    color        : Style.title.background

    signal hideClicked()

    // Drag to move : https://stackoverflow.com/a/18927884
    MouseArea {
        property variant clickPos: "1,1"
        anchors.fill: parent
        onPressed: {
            clickPos  = Qt.point(mouse.x,mouse.y)
        }
        onPositionChanged: {
            var delta = Qt.point(mouse.x-clickPos.x, mouse.y-clickPos.y)
            parentWin.x += delta.x;
            parentWin.y += delta.y;
        }
    }

    // logo
    Image {
        id: imgLogo
        height             : Style.title.imgHeight
        fillMode           : Image.PreserveAspectFit
        visible: root.hasIcon
        anchors {
            left           : root.left
            leftMargin     : Style.title.leftMargin
            verticalCenter : root.verticalCenter
        }
        //source             : "qrc://logo.svg"
        source             : "logo.svg"
        smooth             : true
    }

    TextMetrics {
        id: titleMetrics
        elideWidth: 2*root.width/3
        elide: Qt.ElideMiddle
        font: titleText.font
        text: root.title
    }

    // Title
    Text {
        id: titleText
        anchors {
            left           : hasIcon ? imgLogo.right : parent.left
            leftMargin     : hasIcon ? Style.title.leftMargin : Style.main.leftMargin
            verticalCenter : root.verticalCenter
        }
        text           : titleMetrics.elidedText
        color          : Style.title.text
        font.pointSize : Style.title.fontSize * Style.pt
    }

    // Underline Button
    Rectangle {
        id: buttonUndrLine
        anchors {
            verticalCenter : root.verticalCenter
            right          : buttonCross.left
            rightMargin    : 2*Style.title.fontSize
        }
        width  : Style.title.fontSize
        height : Style.title.fontSize
        color  : "transparent"
        Canvas {
            anchors.fill: parent
            onPaint: {
                var val = Style.title.fontSize
                var ctx = getContext("2d")
                ctx.strokeStyle = 'white'
                ctx.strokeWidth = 4
                ctx.moveTo(0  , val-1)
                ctx.lineTo(val, val-1)
                ctx.stroke()
            }
        }
        MouseArea {
            anchors.fill: parent
            onClicked: root.hideClicked()
        }
    }

    // Cross Button
    Rectangle {
        id: buttonCross
        anchors {
            verticalCenter : root.verticalCenter
            right          : root.right
            rightMargin    : Style.main.rightMargin
        }
        width  : Style.title.fontSize
        height : Style.title.fontSize
        color  : "transparent"
        Canvas {
            anchors.fill: parent
            onPaint: {
                var val = Style.title.fontSize
                var ctx = getContext("2d")
                ctx.strokeStyle = 'white'
                ctx.strokeWidth = 4
                ctx.moveTo(0,0)
                ctx.lineTo(val,val)
                ctx.moveTo(val,0)
                ctx.lineTo(0,val)
                ctx.stroke()
            }
        }
        MouseArea {
            anchors.fill: parent
            onClicked: root.hideClicked()
        }
    }
}
