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
import QtQml
import QtQuick
import QtQuick.Layouts
import QtQuick.Controls
import QtQuick.Controls.impl

Rectangle {
    id: root

    property ColorScheme colorScheme
    property string iconSource
    property string text

    signal clicked

    border.color: colorScheme.border_norm
    border.width: 1
    color: {
        if (mouseArea.pressed) {
            return colorScheme.interaction_default_active;
        }
        if (mouseArea.containsMouse) {
            return colorScheme.interaction_default_hover;
        }
        return colorScheme.background_norm;
    }
    height: 68
    radius: ProtonStyle.banner_radius
    Accessible.role: Accessible.Button
    Accessible.name: root.text

    RowLayout {
        anchors.fill: parent
        anchors.margins: ProtonStyle.wizard_spacing_medium

        ColorImage {
            height: sourceSize.height
            source: iconSource
            sourceSize.height: 36
        }
        Label {
            Layout.fillWidth: true
            Layout.leftMargin: 12
            colorScheme: root.colorScheme
            horizontalAlignment: Text.AlignLeft
            text: root.text
            type: Label.LabelType.Body
            verticalAlignment: Text.AlignVCenter
        }
    }
    MouseArea {
        id: mouseArea
        acceptedButtons: Qt.LeftButton
        anchors.fill: parent
        hoverEnabled: true

        onClicked: {
            root.clicked();
        }
    }
}
