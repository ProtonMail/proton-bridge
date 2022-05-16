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
import QtQuick.Window 2.13
import QtQuick.Layouts 1.12
import QtQuick.Controls 2.12

import "../Proton"

RowLayout {
    id: root
    property ColorScheme colorScheme

    // Norm
    ColumnLayout {
        Layout.fillWidth: true

        spacing: parent.spacing

        TextField {
            colorScheme: root.colorScheme
            Layout.fillWidth: true

            placeholderText: "Placeholder"
            label: "Label"
            hint: "Hint"
            assistiveText: "Assistive text"
        }

        TextField {
            colorScheme: root.colorScheme
            Layout.fillWidth: true

            text: "Value"
            placeholderText: "Placeholder"
            label: "Label"
            hint: "Hint"
            assistiveText: "Assistive text"
        }


        TextField {
            colorScheme: root.colorScheme
            Layout.fillWidth: true
            error: true

            text: "Value"
            placeholderText: "Placeholder"
            label: "Label"
            hint: "Hint"
            errorString: "Error message"
        }


        TextField {
            colorScheme: root.colorScheme
            Layout.fillWidth: true

            text: "Value"
            placeholderText: "Placeholder"
            label: "Label"
            hint: "Hint"
            assistiveText: "Assistive text"

            enabled: false
        }
    }

    // Masked
    ColumnLayout {
        Layout.fillWidth: true

        spacing: parent.spacing

        TextField {
            colorScheme: root.colorScheme
            Layout.fillWidth: true
            echoMode: TextInput.Password
            placeholderText: "Password"
            label: "Label"
            hint: "Hint"
            assistiveText: "Assistive text"
        }

        TextField {
            colorScheme: root.colorScheme
            Layout.fillWidth: true
            text: "Password"

            echoMode: TextInput.Password
            placeholderText: "Password"
            label: "Label"
            hint: "Hint"
            assistiveText: "Assistive text"
        }

        TextField {
            colorScheme: root.colorScheme
            Layout.fillWidth: true
            text: "Password"
            error: true

            echoMode: TextInput.Password
            placeholderText: "Password"
            label: "Label"
            hint: "Hint"
            errorString: "Error message"
        }

        TextField {
            colorScheme: root.colorScheme
            Layout.fillWidth: true
            text: "Password"
            enabled: false

            echoMode: TextInput.Password
            placeholderText: "Password"
            label: "Label"
            hint: "Hint"
            assistiveText: "Assistive text"
        }
    }

    // Varia
    ColumnLayout {
        Layout.fillWidth: true

        spacing: parent.spacing

        TextField {
            colorScheme: root.colorScheme
            Layout.fillWidth: true

            placeholderText: "Type 42 here"
            label: "42 Validator"
            hint: "Accepts only \"42\""
            assistiveText: "Type sometihng here, preferably 42"

            validator: function(str) {
                if (str === "42") {
                    return
                }

                return "Not 42"
            }
        }

        TextField {
            colorScheme: root.colorScheme
            Layout.fillWidth: true

            placeholderText: "Placeholder"
            label: "Label"
        }

        TextField {
            colorScheme: root.colorScheme
            Layout.fillWidth: true

            placeholderText: "Placeholder"
            hint: "Hint"
        }

        TextField {
            colorScheme: root.colorScheme
            Layout.fillWidth: true

            placeholderText: "Placeholder"
            assistiveText: "Assistive text"
        }
    }
}
