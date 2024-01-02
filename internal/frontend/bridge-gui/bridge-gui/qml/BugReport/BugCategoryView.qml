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
import QtQuick.Layouts
import QtQuick.Controls
import Proton
import ".."

SettingsView {
    id: root

    signal categorySelected(int categoryId)

    fillHeight: true

    property var categories: Backend.bugCategories

    Label {
        Layout.fillWidth: true
        colorScheme: root.colorScheme
        text: qsTr("What do you want to report?")
        type: Label.Heading
    }

    Repeater {
        model: root.categories

        CategoryItem {
            Layout.fillWidth: true
            actionIcon: "/qml/icons/ic-chevron-right.svg"
            colorScheme: root.colorScheme
            text: modelData.name
            hint: modelData.hint ? modelData.hint: ""

            onClicked: root.categorySelected(index)
        }
    }

    // fill height so the footer label will always be attached to the bottom
    Item {
        Layout.fillHeight: true
    }
}