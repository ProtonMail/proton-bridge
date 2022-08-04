// Copyright (c) 2022 Proton AG
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
import QtQuick.Layouts
import QtQuick.Controls

import "../Proton"

Rectangle {
    id: root
    property ColorScheme colorScheme
    color: colorScheme.background_norm
    clip: true

    implicitHeight: children[0].implicitHeight + children[0].anchors.topMargin + children[0].anchors.bottomMargin
    implicitWidth: children[0].implicitWidth + children[0].anchors.leftMargin + children[0].anchors.rightMargin

    ScrollView {
        anchors.fill: parent

        ColumnLayout {
            anchors.margins: 20

            width: root.width

            spacing: 5

            Buttons {
                colorScheme: root.colorScheme
                Layout.fillWidth: true
                Layout.margins: 20
            }

            CheckBoxes {
                colorScheme: root.colorScheme
                Layout.fillWidth: true
                Layout.margins: 20
            }

            ComboBoxes {
                colorScheme: root.colorScheme
                Layout.fillWidth: true
                Layout.margins: 20
            }

            RadioButtons {
                colorScheme: root.colorScheme
                Layout.fillWidth: true
                Layout.margins: 20
            }

            Switches {
                colorScheme: root.colorScheme
                Layout.fillWidth: true
                Layout.margins: 20
            }

            TextAreas {
                colorScheme: root.colorScheme
                Layout.fillWidth: true
                Layout.margins: 20
            }

            TextFields {
                colorScheme: root.colorScheme
                Layout.fillWidth: true
                Layout.margins: 20
            }
        }
    }
}
