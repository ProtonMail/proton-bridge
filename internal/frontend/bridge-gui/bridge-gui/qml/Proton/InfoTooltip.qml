// Copyright (c) 2024 Proton AG
//
// This file is part of Proton Mail Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.
import QtQuick
import QtQuick.Controls
import QtQuick.Controls.impl
import QtQuick.Layouts

ColorImage {
    id: root

    property var colorScheme
    property string text
    property int size: 16

    color: root.colorScheme.interaction_norm
    height: size
    width: size
    source: "/qml/icons/ic-info-circle.svg"
    sourceSize.height: size
    sourceSize.width: size
    visible: root.hint !== ""
    MouseArea {
        id: imageArea
        anchors.fill: parent
        hoverEnabled: true
    }
    ToolTip {
        id: toolTipinfo
        text: root.text
        visible: imageArea.containsMouse
        implicitWidth: Math.min(400, tooltipText.implicitWidth)
        background: Rectangle {
            radius: 4
            border.color: root.colorScheme.border_weak
            color: root.colorScheme.background_weak
        }
        contentItem: Text {
            id: tooltipText
            color: root.colorScheme.text_norm
            text: root.text
            wrapMode: Text.WordWrap

            horizontalAlignment: Text.AlignHCenter
            verticalAlignment: Text.AlignVCenter
        }
    }
}