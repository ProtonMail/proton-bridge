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

RowLayout {
    id: root

    property ColorScheme colorScheme
    property string description
    property string icon
    property int iconSize: 64
    property string title

    spacing: ProtonStyle.wizard_spacing_large

    Image {
        Layout.alignment: Qt.AlignHCenter | Qt.AlignVCenter
        Layout.preferredHeight: iconSize
        Layout.preferredWidth: iconSize
        mipmap: true
        source: root.icon
    }
    ColumnLayout {
        Layout.alignment: Qt.AlignLeft | Qt.AlignVCenter
        Layout.fillWidth: true
        spacing: ProtonStyle.wizard_spacing_small

        Label {
            Layout.alignment: Qt.AlignLeft | Qt.AlignTop
            Layout.fillHeight: false
            Layout.fillWidth: true
            colorScheme: root.colorScheme
            text: root.title
            type: Label.LabelType.Body_bold
        }
        Label {
            Layout.alignment: Qt.AlignLeft | Qt.AlignTop
            Layout.fillHeight: true
            Layout.fillWidth: true
            colorScheme: root.colorScheme
            text: root.description
            type: Label.LabelType.Body
            verticalAlignment: Text.AlignTop
        }
    }
}
