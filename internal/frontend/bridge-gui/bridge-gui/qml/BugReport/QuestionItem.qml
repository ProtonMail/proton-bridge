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
import QtQuick.Controls.impl
import Proton

Item {
    id: root
    enum InputType {
        TextInput = 1,
        Radio,
        Checkbox
    }

    property string _typeOpen: "open"
    property string _typeChoice: "choice"
    property string _typeMutlichoice: "multichoice"
    property var colorScheme
    property var _bottomMargin: 20
    property var _lineHeight: 1
    property bool showSeparator: true

    property string text: ""
    property string tips: ""
    property string label: ""
    property bool mandatory: false
    property var type: root._typeOpen
    property var answerList: ListModel{}
    property int maxChar: 150

    property string answer:{
        if (type === root._typeOpen) {
            return textInput.text
        } else if (type === root._typeChoice) {
            return selectionRadio.text
        } else if (type === root._typeMutlichoice) {
            return selectionCheckBox.text
        }
        return ""
    }
    property bool error: {
            if (root.type === root._typeOpen)
                return textInput.error;
            if (root.type === root._typeChoice)
                return selectionRadio.error;
            if (root.type === root._typeMutlichoice)
                return selectionCheckBox.error;
            return false
    }

    function setDefaultValue(defaultValue) {
        textInput.setDefaultValue(defaultValue)
        selectionRadio.setDefaultValue(defaultValue)
        selectionCheckBox.setDefaultValue(defaultValue)
    }

    function validate() {

    if (root.type === root._typeOpen)
        textInput.validate()
    else if (root.type === root._typeChoice)
        selectionRadio.validate()
    else if (root.type === root._typeMutlichoice)
        selectionCheckBox.validate()
    }

    implicitHeight: children[0].implicitHeight + children[0].anchors.topMargin + children[0].anchors.bottomMargin

    ColumnLayout {
        anchors.fill: parent
        spacing: 16

        Label {
            id: mainLabel
            colorScheme: root.colorScheme
            text: root.mandatory ? qsTr(root.text+" *") : qsTr(root.text)
            type: Label.Body
        }
        ColumnLayout {
            spacing: 16
            Layout.bottomMargin: root._bottomMargin
            TextArea {
                id: textInput
                Layout.fillWidth: true
                Layout.fillHeight: true
                Layout.minimumHeight: root.type === root._typeOpen ? heightForLinesVisible(2) : 0
                colorScheme: root.colorScheme

                property int _maxLength: root.maxChar
                property int _minLength: 1

                label: qsTr(root.label)
                hint: textInput.text.length + "/" + _maxLength
                placeholderText: qsTr(root.tips)

                function setDefaultValue(defaultValue) {
                    textInput.text = root.type === root._typeOpen ? defaultValue : ""
                }

                validator: function (text) {
                    if (mandatory && textInput.text.length < textInput._minLength) {
                        return qsTr("Field is mandatory");
                    }
                    if (textInput.text.length > textInput._maxLength) {
                        return qsTr("max. %1 characters").arg(_maxLength);
                    }
                    return;
                }
                onTextChanged: {
                    // Rise max length error immediately while typing if mandatory field
                    if (textInput.text.length > textInput._maxLength) {
                        validate();
                    }
                }

                visible: root.type === root._typeOpen
            }

            ButtonGroup {
                id: selectionRadio

                property string text: {
                    return checkedButton ? checkedButton.text : "";
                }
                property bool error: false

                function setDefaultValue(defaultValue) {
                    const values = root.type === root._typeChoice ? defaultValue : [];
                    for (var i = 0; i < buttons.length; ++i) {
                        buttons[i].checked  = values.includes(buttons[i].text);
                    }
                }
                function validate() {
                    if (mandatory && selectionRadio.text.length === 0) {
                        error = true;
                        return
                    }
                    error = false;
                }

                onTextChanged: {
                    validate();
                }
            }
            Repeater {
                model: root.answerList

                RadioButton {
                    ButtonGroup.group: selectionRadio
                    colorScheme: root.colorScheme
                    text: modelData
                    visible: root.type === root._typeChoice
                }
            }
            ButtonGroup {
                id: selectionCheckBox
                exclusive: false
                property string delimitor: ", "
                property string text: {
                    var str = "";
                    for (var i = 0; i < buttons.length; ++i) {
                        if (buttons[i].checked) {
                            str += buttons[i].text + delimitor;
                        }
                    }
                    return str.slice(0, -delimitor.length);
                }
                property bool error: false

                function setDefaultValue(defaultValue) {
                    const values = root.type === root._typeMutlichoice ? defaultValue.split(delimitor) : [];
                    for (var i = 0; i < buttons.length; ++i) {
                        buttons[i].checked  = values.includes(buttons[i].text);
                    }
                }

                function validate() {
                    if (mandatory && selectionCheckBox.text.length === 0) {
                        error = true;
                        return
                    }
                    error = false;
                }

                onTextChanged: {
                    validate();
                }
            }
            Repeater {
                model: root.answerList

                CheckBox {
                    ButtonGroup.group: selectionCheckBox
                    colorScheme: root.colorScheme
                    text: modelData
                    visible: root.type === root._typeMutlichoice
                }
            }

            RowLayout {
                id: footerLayout
                Layout.fillWidth: true
                spacing: 0

                visible: {
                    if (root.type === root._typeOpen)
                        return false
                    return root.error
                }

                ColorImage {
                    id: errorIcon
                    Layout.rightMargin: 4
                    color: root.colorScheme.signal_danger
                    height: errorChoice.height
                    source: "/qml/icons/ic-exclamation-circle-filled.svg"
                    sourceSize.height: errorChoice.height
                }
                Label {
                    id: errorChoice
                    Layout.fillWidth: true
                    color: root.colorScheme.signal_danger
                    colorScheme: root.colorScheme
                    text: "Field is mandatory"
                    type: Label.LabelType.Caption_semibold
                }
            }
        }
    }

    Rectangle {
        anchors.bottom: root.bottom
        anchors.left: root.left
        anchors.right: root.right
        color: root.colorScheme.border_weak
        height: root._lineHeight
        visible: root.showSeparator
    }
    // fill height so the footer label will always be attached to the bottom
    Item {
        Layout.fillHeight: true
    }
}