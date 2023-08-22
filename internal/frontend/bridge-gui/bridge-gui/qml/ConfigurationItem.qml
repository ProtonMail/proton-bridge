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
import QtQuick
import QtQuick.Layouts
import QtQuick.Controls
import QtQuick.Controls.impl
import Proton

Item {
    id: root

    property var colorScheme
    property string label
    property string labelColor: root.colorScheme.text_norm
    property string value

    Layout.fillWidth: true
    implicitHeight: children[0].implicitHeight
    implicitWidth: children[0].implicitWidth

    ColumnLayout {
        width: root.width

        RowLayout {
            Layout.fillWidth: true

            ColumnLayout {
                Label {
                    color: labelColor
                    colorScheme: root.colorScheme
                    text: root.label
                    type: Label.Body_semibold
                }
                TextEdit {
                    id: valueText
                    Layout.fillWidth: true
                    color: root.colorScheme.text_weak
                    readOnly: true
                    selectByKeyboard: true
                    selectByMouse: true
                    selectionColor: root.colorScheme.text_weak
                    text: root.value
                    wrapMode: Text.WrapAnywhere
                }
            }
            Item {
                Layout.fillWidth: true
            }
            ColorImage {
                color: root.colorScheme.text_norm
                height: root.colorScheme.body_font_size
                source: "/qml/icons/ic-copy.svg"
                sourceSize.height: root.colorScheme.body_font_size

                MouseArea {
                    anchors.fill: parent

                    onClicked: {
                        valueText.select(0, valueText.length);
                        valueText.copy();
                        valueText.deselect();
                    }
                    onPressed: parent.scale = 0.90
                    onReleased: parent.scale = 1
                }
            }
        }
        Rectangle {
            Layout.fillWidth: true
            color: root.colorScheme.border_norm
            height: 1
        }
    }
}
