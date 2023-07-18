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

SettingsView {
    id: root

    property bool _valuesChanged: keychainSelection.checkedButton && keychainSelection.checkedButton.text !== Backend.currentKeychain

    function setDefaultValues() {
        for (const bi in keychainSelection.buttons) {
            const button = keychainSelection.buttons[bi];
            if (button.text === Backend.currentKeychain) {
                button.checked = true;
                break;
            }
        }
    }

    fillHeight: false

    Component.onCompleted: root.setDefaultValues()
    onBack: {
        root.setDefaultValues();
    }

    Label {
        Layout.fillWidth: true
        colorScheme: root.colorScheme
        text: qsTr("Default keychain")
        type: Label.Heading
    }
    Label {
        Layout.fillWidth: true
        color: root.colorScheme.text_weak
        colorScheme: root.colorScheme
        text: qsTr("Change which keychain Bridge uses as default")
        type: Label.Body
        wrapMode: Text.WordWrap
    }
    ColumnLayout {
        spacing: 16

        ButtonGroup {
            id: keychainSelection
        }
        Repeater {
            model: Backend.availableKeychain

            RadioButton {
                ButtonGroup.group: keychainSelection
                colorScheme: root.colorScheme
                text: modelData
            }
        }
    }
    Rectangle {
        Layout.fillWidth: true
        color: root.colorScheme.border_weak
        height: 1
    }
    RowLayout {
        spacing: 12

        Button {
            id: submitButton
            colorScheme: root.colorScheme
            enabled: root._valuesChanged
            text: qsTr("Save and restart")

            onClicked: {
                Backend.changeKeychain(keychainSelection.checkedButton.text);
            }
        }
        Button {
            colorScheme: root.colorScheme
            secondary: true
            text: qsTr("Cancel")

            onClicked: root.back()
        }
        Connections {
            function onChangeKeychainFinished() {
                submitButton.loading = false;
                root.back();
            }

            target: Backend
        }
    }
}
