// Copyright (c) 2024 Proton AG
// This file is part of Proton Mail Bridge.
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.
import QtQuick
import QtQuick.Controls
import "."
import "Proton"

Rectangle {
    property var target: parent

    border.color: "red"
    border.width: 1
    color: "transparent"
    height: target.height
    width: target.width
    x: target.x
    y: target.y
    //z: parent.z - 1
    z: 10000000

    Label {
        anchors.centerIn: parent
        color: "black"
        colorScheme: ProtonStyle.currentStyle
        text: parent.width + "x" + parent.height
    }
    Rectangle {
        border.color: "green"
        border.width: 1
        color: "transparent"
        height: target.implicitHeight
        width: target.implicitWidth
        //z: parent.z - 1
        z: 10000000
    }
}
