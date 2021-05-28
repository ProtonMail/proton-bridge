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

import QtQuick 2.13
import QtQuick.Window 2.13
import QtQuick.Controls 2.13

Window {
    id: testroot
    width   : 250
    height  : 600
    flags   : Qt.Window | Qt.Dialog | Qt.FramelessWindowHint
    visible : true
    title   : "GUI test Window"
    color   : "#10101010"

    Column {
        anchors.horizontalCenter: parent.horizontalCenter
        spacing : 5
        Button {
            text: "Show window"
            onClicked: {
                bridge._mainWindow.visible = true
            }
        }
        Button {
            text: "Hide window"
            onClicked: {
                bridge._mainWindow.visible = false
            }
        }
    }

    Component.onCompleted : {
        testroot.x= 10
        testroot.y= 100
        bridge._mainWindow.visible = true
    }


    Bridge {id:bridge}
}
