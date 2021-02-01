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

// Dialog with adding new user

import QtQuick 2.8
import QtQuick.Layouts 1.3
import ProtonUI 1.0


StackLayout {
    id: root
    property string title      : "title"
    property string subtitle   : ""
    property alias timer       : timer
    property alias warning     : warningText
    property bool isDialogBusy : false
    property real titleHeight  : 2*titleText.anchors.topMargin + titleText.height + (warningText.visible ?  warningText.anchors.topMargin + warningText.height : 0)
    property Item background   : Rectangle {
        parent: root
        width: root.width
        height: root.height
        color        : Style.dialog.background
        visible: root.visible
        z: -1

        // Looks like StackLayout explicatly sets visible=false to all viasual children except selected.
        // We want this background to be also visible.
        onVisibleChanged: {
            if (visible != parent.visible) {
                visible = parent.visible
            }
        }

        AccessibleText {
            id: titleText
            anchors {
                top: parent.top
                horizontalCenter: parent.horizontalCenter
                topMargin: Style.dialog.titleSize
            }
            font.pointSize : Style.dialog.titleSize * Style.pt
            color : Style.dialog.text
            text  : root.title
        }

        AccessibleText {
            id: subtitleText
            anchors {
                top: titleText.bottom
                horizontalCenter: parent.horizontalCenter
            }
            font.pointSize : Style.dialog.fontSize * Style.pt
            color : Style.dialog.text
            text  : root.subtitle
            visible : root.subtitle != ""
        }

        AccessibleText {
            id:warningText
            anchors {
                top: subtitleText.bottom
                horizontalCenter: parent.horizontalCenter
            }
            font {
                bold: true
                pointSize: Style.dialog.fontSize * Style.pt
            }
            text : ""
            color: Style.main.textBlue
            visible: false
        }

        // prevent any action below
        MouseArea {
            anchors.fill: parent
            hoverEnabled: true
        }

        ClickIconText {
            anchors {
                top: parent.top
                right: parent.right
                topMargin: Style.dialog.titleSize
                rightMargin: Style.dialog.titleSize
            }
            visible   : !isDialogBusy
            iconText  : Style.fa.times
            text      : ""
            onClicked : root.hide()
            Accessible.description : qsTr("Close dialog %1", "Click to exit modal.").arg(root.title)
        }
    }

    Accessible.role: Accessible.Grouping
    Accessible.name: title
    Accessible.description: title
    Accessible.focusable: true


    visible      : false
    anchors {
        left   : parent.left
        right  : parent.right
        top    : titleBar.bottom
        bottom : parent.bottom
    }
    currentIndex : 0


    signal show()
    signal hide()

    function incrementCurrentIndex() {
        root.currentIndex++
    }

    function decrementCurrentIndex() {
        root.currentIndex--
    }

    onShow: {
        root.visible = true
        root.forceActiveFocus()
    }

    onHide: {
        root.timer.stop()
        root.currentIndex=0
        root.visible = false
        root.timer.stop()
        gui.winMain.tabbar.focusButton()
    }

    // QTimer is recommeded solution for creating trheads : http://doc.qt.io/qt-5/qtquick-threading-example.html
    Timer {
        id: timer
        interval: 300 // wait for transistion
        repeat: false
    }
}
