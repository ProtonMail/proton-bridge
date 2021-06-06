// Copyright (c) 2021 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

import QtQuick 2.13
import QtQuick.Window 2.13
import QtQuick.Layouts 1.12
import QtQuick.Controls 2.12

import Proton 4.0

RowLayout {
    property var colorScheme: parent.colorScheme

    // Norm
    ColumnLayout {
        Layout.fillWidth: true
        property var colorScheme: parent.colorScheme

        spacing: parent.spacing

        TextField {
            Layout.fillWidth: true

            placeholderText: "Placeholder"
            label: "Label"
            hint: "Hint"
            assistiveText: "Assistive text"
        }

        TextField {
            Layout.fillWidth: true

            text: "Value"
            placeholderText: "Placeholder"
            label: "Label"
            hint: "Hint"
            assistiveText: "Assistive text"
        }


        TextField {
            Layout.fillWidth: true
            error: true

            text: "Value"
            placeholderText: "Placeholder"
            label: "Label"
            hint: "Hint"
            assistiveText: "Error message"
        }


        TextField {
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
        property var colorScheme: parent.colorScheme

        spacing: parent.spacing

        TextField {
            Layout.fillWidth: true
            echoMode: TextInput.Password
            placeholderText: "Password"
            label: "Label"
            hint: "Hint"
            assistiveText: "Assistive text"
        }

        TextField {
            Layout.fillWidth: true
            text: "Password"

            echoMode: TextInput.Password
            placeholderText: "Password"
            label: "Label"
            hint: "Hint"
            assistiveText: "Assistive text"
        }

        TextField {
            Layout.fillWidth: true
            text: "Password"
            error: true

            echoMode: TextInput.Password
            placeholderText: "Password"
            label: "Label"
            hint: "Hint"
            assistiveText: "Error message"
        }

        TextField {
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
        property var colorScheme: parent.colorScheme

        spacing: parent.spacing

        TextField {
            Layout.fillWidth: true

            placeholderText: "Placeholder"
            label: "Label"
            hint: "Hint"

            Rectangle {
                anchors.fill: parent
                border.color: "red"
                border.width: 1
                z: parent.z - 1
            }
        }

        TextField {
            Layout.fillWidth: true

            placeholderText: "Placeholder"
            assistiveText: "Assistive text"

            Rectangle {
                anchors.fill: parent
                border.color: "red"
                border.width: 1
                z: parent.z - 1
            }
        }

        TextField {
            Layout.fillWidth: true

            placeholderText: "Placeholder"

            Rectangle {
                anchors.fill: parent
                border.color: "red"
                border.width: 1
                z: parent.z - 1
            }
        }
    }
}
