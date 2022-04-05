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

import QtQuick 2.13
import QtQuick.Layouts 1.12
import QtQuick.Controls 2.12
import QtQuick.Controls.impl 2.12

import Proton 4.0

Item {
    id: root
    Layout.fillWidth: true

    property var colorScheme
    property string label
    property string value

    implicitHeight: children[0].implicitHeight
    implicitWidth: children[0].implicitWidth

    ColumnLayout {
        width: root.width

        RowLayout {
            Layout.fillWidth: true

            ColumnLayout {
                Label {
                    colorScheme: root.colorScheme
                    text: root.label
                    type: Label.Body
                }
                TextEdit {
                    id: valueText
                    text: root.value
                    color: root.colorScheme.text_weak
                    readOnly: true
                    selectByMouse: true
                    selectByKeyboard: true
                    selectionColor: root.colorScheme.text_weak
                }
            }

            Item {
                Layout.fillWidth: true
            }

            ColorImage {
                source: "icons/ic-copy.svg"
                color: root.colorScheme.text_norm
                height: root.colorScheme.body_font_size
                sourceSize.height: root.colorScheme.body_font_size

                MouseArea {
                    anchors.fill: parent
                    onClicked : {
                        valueText.select(0, valueText.length)
                        valueText.copy()
                        valueText.deselect()
                    }
                    onPressed: parent.scale = 0.90
                    onReleased: parent.scale = 1
                }

            }
        }

        Rectangle {
            Layout.fillWidth: true
            height: 1
            color: root.colorScheme.border_norm
        }
    }
}
