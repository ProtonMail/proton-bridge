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
import Proton

Item {
    id: root
    enum InputType {
        TextInput = 1,
        Radio,
        Checkbox
    }

    property var colorScheme
    property var _bottomMargin: 20
    property var _lineHeight: 1
    property bool showSeparator: true

    property string text: ""
    property string tips: ""
    property string label: ""
    property var type: QuestionItem.InputType.TextInput
    property var answerList: ListModel{}

    property string answer:{
        if (type === QuestionItem.InputType.TextInput) {
            return textInput.text
        } else if (type === QuestionItem.InputType.Radio) {
            return selectionRadio.text
        } else if (type === QuestionItem.InputType.Checkbox) {
            return selectionCheckBox.text
        }
        return ""
    }

    implicitHeight: children[0].implicitHeight + children[0].anchors.topMargin + children[0].anchors.bottomMargin

    ColumnLayout {
        anchors.fill: parent
        spacing: 16

        Label {
            id: mainLabel
            colorScheme: root.colorScheme
            text: qsTr(root.text)
            type: Label.Body
        }
        ColumnLayout {
            spacing: 16
            Layout.bottomMargin: root._bottomMargin
            TextArea {
                id: textInput
                Layout.fillWidth: true
                Layout.fillHeight: true
                Layout.minimumHeight: root.type === QuestionItem.InputType.TextInput ? heightForLinesVisible(2) : 0
                colorScheme: root.colorScheme

                label: qsTr(root.label)
                placeholderText: qsTr(root.tips)

                visible: root.type === QuestionItem.InputType.TextInput
            }

            ButtonGroup {
                id: selectionRadio
                property string text: {
                    return checkedButton ? checkedButton.text : "";
                }
            }
            Repeater {
                model: root.answerList

                RadioButton {
                    ButtonGroup.group: selectionRadio
                    colorScheme: root.colorScheme
                    text: modelData
                    visible: root.type === QuestionItem.InputType.Radio
                }
            }
            ButtonGroup {
                id: selectionCheckBox
                exclusive: false
                property string text: {
                    var str = "";
                    for (var i = 0; i < buttons.length; ++i) {
                        if (buttons[i].checked) {
                            str += buttons[i].text + " ";
                        }
                    }
                    return str;
                }
            }
            Repeater {
                model: root.answerList

                CheckBox {
                    ButtonGroup.group: selectionCheckBox
                    colorScheme: root.colorScheme
                    text: modelData
                    visible: root.type === QuestionItem.InputType.Checkbox
                }
            }
        }
    }
    Rectangle {
        anchors.bottom: root.bottom
        anchors.left: root.left
        anchors.right: root.right
        color: colorScheme.border_weak
        height: root._lineHeight
        visible: root.showSeparator
    }
    // fill height so the footer label will always be attached to the bottom
    Item {
        Layout.fillHeight: true
    }
}