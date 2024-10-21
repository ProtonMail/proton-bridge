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

Button {
    id: root

    property var wizard
    readonly property int _iconPadding: 8 // The SVG image we use has internal padding that we need to compensate for alignment.
    readonly property int _iconSize: 24

    anchors.bottom: parent.bottom
    anchors.bottomMargin: ProtonStyle.wizard_window_margin - _iconPadding
    anchors.right: parent.right
    anchors.rightMargin: ProtonStyle.wizard_window_margin - _iconPadding
    colorScheme: wizard.colorScheme
    horizontalPadding: 0
    icon.color: wizard.colorScheme.text_weak
    icon.height: _iconSize
    icon.source: "/qml/icons/ic-question-circle.svg"
    icon.width: _iconSize
    verticalPadding: 0
    Accessible.name: qsTr("Help")

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
                Backend.openExternalLink();
            }
        }
        MenuItem {
            id: reportAProblemItem
            colorScheme: root.colorScheme
            text: qsTr("Report a problem")

            onClicked: {
                wizard.showBugReport();
            }
        }
    }
}