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

// Notify user

import QtQuick 2.8
import ProtonUI 1.0

Rectangle {
    id: root
    property int   posx // x-coordinate of triangle
    property bool  isTriangleBelow
    property string text
    property alias bubbleColor: bubble.color
    anchors {
        top  : tabbar.bottom
        left : tabbar.left
        leftMargin : {
            // position of bubble calculated from posx
            return Math.max(
                Style.main.leftMargin,  // keep minimal left margin
                Math.min(
                    root.posx - root.width/2, // fit triangle in the middle if possible
                    tabbar.width - root.width - Style.main.rightMargin // keep minimal right margin
                )
            )
        }
        topMargin: 0
    }
    height  : triangle.height + bubble.height
    width   : bubble.width
    color   : "transparent"
    visible : false


    Rectangle {
        id : triangle
        anchors {
            top          : root.isTriangleBelow  ? undefined   : root.top
            bottom       : root.isTriangleBelow  ? root.bottom : undefined
            bottomMargin : 1*Style.px
            left         : root.left
            leftMargin   : root.posx - triangle.width/2 - root.anchors.leftMargin
        }
        width: 2*Style.tabbar.heightTriangle+2
        height: Style.tabbar.heightTriangle
        color: "transparent"
        Canvas {
            anchors.fill: parent
            rotation: root.isTriangleBelow ? 180 : 0
            onPaint: {
                var ctx = getContext("2d")
                ctx.fillStyle = bubble.color
                ctx.moveTo(0 , height)
                ctx.lineTo(width/2, 0)
                ctx.lineTo(width , height)
                ctx.closePath()
                ctx.fill()
            }
        }
    }

    Rectangle {
        id: bubble
        anchors {
            top: root.top
            left: root.left
            topMargin: (root.isTriangleBelow ? 0 : triangle.height)
        }
        width  : mainText.contentWidth + Style.main.leftMargin + Style.main.rightMargin
        height : 2*Style.main.fontSize
        radius : Style.bubble.radius
        color  : Style.bubble.background

        AccessibleText {
            id: mainText
            anchors {
                horizontalCenter : parent.horizontalCenter
                top: parent.top
                topMargin    : Style.main.fontSize
            }

            text: "<html><style>a { color: "+Style.main.textBlue+";}</style>"+root.text+"<html>"
            width : Style.bubble.width - ( Style.main.leftMargin + Style.main.rightMargin )
            font.pointSize: Style.main.fontSize * Style.pt
            horizontalAlignment: Text.AlignHCenter
            textFormat: Text.RichText
            wrapMode: Text.WordWrap
            color: Style.bubble.text
            onLinkActivated: {
                Qt.openUrlExternally(link)
            }
            MouseArea {
                anchors.fill: mainText
                cursorShape: mainText.hoveredLink ? Qt.PointingHandCursor : Qt.ArrowCursor
                acceptedButtons: Qt.NoButton
            }

            Accessible.name: qsTr("Message")
            Accessible.description: root.text
        }

        ButtonRounded {
            id: okButton
            visible: !root.isTriangleBelow
            anchors {
                bottom           : parent.bottom
                horizontalCenter : parent.horizontalCenter
                bottomMargin     : Style.main.fontSize
            }
            text: qsTr("Okay", "confirms and dismisses a notification")
            height: Style.main.fontSize*2
            color_main: Style.main.text
            color_minor: Style.main.textBlue
            isOpaque: true
            onClicked: hide()
        }
    }

    function place(index) {
        if (index < 0) {
            // add accounts
            root.isTriangleBelow = true
            bubble.height = 3.25*Style.main.fontSize
            root.posx = 2*Style.main.leftMargin
            bubble.width = mainText.contentWidth - Style.main.leftMargin
        } else {
            root.isTriangleBelow = false
            bubble.height = (
                bubble.anchors.topMargin +  // from top
                mainText.contentHeight + // the text content
                Style.main.fontSize + // gap between button
                okButton.height + okButton.anchors.bottomMargin // from bottom and button
            )
            if (index < 3) {
                // possition accordig to top tab
                var margin  = Style.main.leftMargin    + Style.tabbar.widthButton/2
                root.posx = margin + index*tabbar.spacing
            } else {
                // quit button
                root.posx = tabbar.width - 2*Style.main.rightMargin
            }
        }
    }

    function show() {
        root.visible=true
        gui.winMain.activeContent = false
    }

    function hide() {
        root.visible=false
        go.bubbleClosed()
        gui.winMain.activeContent = true
        gui.winMain.tabbar.focusButton()
    }
}
