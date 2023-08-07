// Copyright (c) 2023 Proton AG
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
import Proton

Item {
    id: root

    property ColorScheme colorScheme
    property string iconSource
    property string text

    signal clicked

    implicitHeight: clientRow.height

    ColumnLayout {
        id: clientRow
        anchors.left: parent.left
        anchors.right: parent.right
        spacing: 0

        RowLayout {
            Layout.bottomMargin: 12
            Layout.leftMargin: 16
            Layout.rightMargin: 16
            Layout.topMargin: 12

            ColorImage {
                height: 36
                source: iconSource
                sourceSize.height: 36
            }
            Label {
                Layout.leftMargin: 12
                colorScheme: root.colorScheme
                text: root.text
                type: Label.LabelType.Body
            }
        }
        Rectangle {
            Layout.fillWidth: true
            Layout.preferredHeight: 1
            color: root.colorScheme.border_weak
        }
    }
    MouseArea {
        anchors.fill: parent
        cursorShape: Qt.PointingHandCursor

        onClicked: {
            root.clicked();
        }
    }
}
