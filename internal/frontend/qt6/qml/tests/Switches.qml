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

    ColumnLayout {
        Layout.fillWidth: true

        spacing: parent.spacing

        Switch {
            text: "Toggle"
            colorScheme: root.colorScheme
        }

        Switch {
            text: "Toggle"
            enabled: false
            colorScheme: root.colorScheme
        }

        Switch {
            text: "Toggle"
            loading: true
            colorScheme: root.colorScheme
        }

        Switch {
            text: ""
            colorScheme: root.colorScheme
        }

        Switch {
            text: ""
            enabled: false
            colorScheme: root.colorScheme
        }

        Switch {
            text: ""
            loading: true
            colorScheme: root.colorScheme
        }
    }

    ColumnLayout {
        Layout.fillWidth: true

        spacing: parent.spacing

        Switch {
            text: "Toggle"
            checked: true
            colorScheme: root.colorScheme
        }

        Switch {
            text: "Toggle"
            checked: true
            enabled: false
            colorScheme: root.colorScheme
        }

        Switch {
            text: "Toggle"
            checked: true
            loading: true
            colorScheme: root.colorScheme
        }

        Switch {
            text: ""
            checked: true
            colorScheme: root.colorScheme
        }

        Switch {
            text: ""
            checked: true
            enabled: false
            colorScheme: root.colorScheme
        }

        Switch {
            text: ""
            checked: true
            loading: true
            colorScheme: root.colorScheme
        }
    }
}
