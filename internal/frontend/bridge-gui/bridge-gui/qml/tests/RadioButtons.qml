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

import QtQuick
import QtQuick.Window
import QtQuick.Layouts
import QtQuick.Controls

import "../Proton"

RowLayout {
    id: root
    property ColorScheme colorScheme

    ColumnLayout {
        Layout.fillWidth: true

        spacing: parent.spacing

        RadioButton {
            colorScheme: root.colorScheme
            text: "Radio"
        }

        RadioButton {
            colorScheme: root.colorScheme
            text: "Radio"
            error: true
        }

        RadioButton {
            colorScheme: root.colorScheme
            text: "Radio"
            enabled: false
        }
        RadioButton {
            colorScheme: root.colorScheme
            text: ""
        }

        RadioButton {
            colorScheme: root.colorScheme
            text: ""
            error: true
        }

        RadioButton {
            colorScheme: root.colorScheme
            text: ""
            enabled: false
        }
    }

    ColumnLayout {
        Layout.fillWidth: true

        spacing: parent.spacing

        RadioButton {
            colorScheme: root.colorScheme
            text: "Radio"
            checked: true
        }

        RadioButton {
            colorScheme: root.colorScheme
            text: "Radio"
            checked: true
            error: true
        }

        RadioButton {
            colorScheme: root.colorScheme
            text: "Radio"
            checked: true
            enabled: false
        }
        RadioButton {
            colorScheme: root.colorScheme
            text: ""
            checked: true
        }

        RadioButton {
            colorScheme: root.colorScheme
            text: ""
            checked: true
            error: true
        }

        RadioButton {
            colorScheme: root.colorScheme
            text: ""
            checked: true
            enabled: false
        }
    }
}
