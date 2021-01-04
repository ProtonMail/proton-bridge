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

// Dialog with main menu

import QtQuick 2.8
import BridgeUI 1.0
import ProtonUI 1.0

Rectangle {
    id: root
    color: "#aaff5577"
    anchors {
        left   : tabbar.left
        right  : tabbar.right
        top    : tabbar.bottom
        bottom : parent.bottom
    }
    visible: false

    MouseArea {
        anchors.fill: parent
        onClicked: toggle()
    }

    Rectangle {
        color  : Style.menu.background
        radius : Style.menu.radius
        width  : Style.menu.width
        height : Style.menu.height
        anchors {
            top         : parent.top
            right       : parent.right
            topMargin   : Style.menu.topMargin
            rightMargin : Style.menu.rightMargin
        }

        MouseArea {
            anchors.fill: parent
        }

        Text {
            anchors.centerIn: parent
            text: qsTr("About")
            color: Style.menu.text
        }
    }

    function toggle(){
        if (root.visible == false) {
            root.visible = true
        } else {
            root.visible = false
        }
    }
}


