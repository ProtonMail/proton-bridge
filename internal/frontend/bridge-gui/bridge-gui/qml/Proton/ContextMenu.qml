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

Item {
    property var parentObject: null
    property var colorScheme: null
    property bool readOnly: false
    property bool isPassword: false

    MouseArea {
        id: controlMouseArea
        width: parentObject ? parentObject.width : 0
        height: parentObject ? parentObject.height : 0
        acceptedButtons: Qt.RightButton
        onClicked: controlContextMenu.popup()

        propagateComposedEvents: true
    }

    Menu {
        id: controlContextMenu
        colorScheme: root.colorScheme
        onVisibleChanged: {
            if (controlContextMenu.visible) {
                const hasSelectedText = parentObject.selectedText.length > 0;
                const hasClipboardText = clipboard.text.length > 0;

                copyMenuItem.visible = hasSelectedText && !isPassword;
                cutMenuItem.visible = hasSelectedText && !readOnly && !isPassword;
                pasteMenuItem.visible = hasClipboardText && !readOnly;
                controlContextMenu.visible = copyMenuItem.visible || cutMenuItem.visible || pasteMenuItem.visible;
            }
        }

        MenuItem {
            id: cutMenuItem
            colorScheme: root.colorScheme
            height: visible ? implicitHeight : 0
            text: qsTr("Cut")

            onClicked: {
                parentObject.cut()
            }
        }
        MenuItem {
            id: copyMenuItem
            colorScheme: root.colorScheme
            height: visible ? implicitHeight : 0
            text: qsTr("Copy")

            onTriggered: {
                parentObject.copy()
            }
        }
        MenuItem {
            id: pasteMenuItem
            colorScheme: root.colorScheme
            height: visible ? implicitHeight : 0

            text: qsTr("Paste")
            onTriggered: {
                parentObject.paste()
            }
        }
    }

}
