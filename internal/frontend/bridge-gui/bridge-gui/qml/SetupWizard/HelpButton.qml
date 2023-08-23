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

Button {
    id: root

    property var wizard

    anchors.bottom: parent.bottom
    anchors.bottomMargin: 32
    anchors.right: parent.right
    anchors.rightMargin: 32
    colorScheme: wizard.colorScheme
    horizontalPadding: 0
    icon.color: wizard.colorScheme.text_weak
    icon.height: 24
    icon.source: "/qml/icons/ic-question-circle.svg"
    icon.width: 24
    verticalPadding: 0

    onClicked: {
        menu.popup(-menu.width + root.width, -menu.height);
    }

    Menu {
        id: menu
        colorScheme: root.colorScheme
        modal: true

        MenuItem {
            id: getHelpItem
            colorScheme: root.colorScheme
            text: qsTr("Get help")

            onClicked: {
                console.error("Get help");
            }
        }
        MenuItem {
            id: reportAProblemItem
            colorScheme: root.colorScheme
            text: qsTr("Report a problem")

            onClicked: {
                console.error("Report a problem");
            }
        }
    }
}