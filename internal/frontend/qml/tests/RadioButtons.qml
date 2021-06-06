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

    ColumnLayout {
        Layout.fillWidth: true
        property var colorScheme: parent.colorScheme

        spacing: parent.spacing

        RadioButton {
            text: "Radio"
        }

        RadioButton {
            text: "Radio"
            error: true
        }

        RadioButton {
            text: "Radio"
            enabled: false
        }
        RadioButton {
            text: ""
        }

        RadioButton {
            text: ""
            error: true
        }

        RadioButton {
            text: ""
            enabled: false
        }
    }

    ColumnLayout {
        Layout.fillWidth: true
        property var colorScheme: parent.colorScheme

        spacing: parent.spacing

        RadioButton {
            text: "Radio"
            checked: true
        }

        RadioButton {
            text: "Radio"
            checked: true
            error: true
        }

        RadioButton {
            text: "Radio"
            checked: true
            enabled: false
        }
        RadioButton {
            text: ""
            checked: true
        }

        RadioButton {
            text: ""
            checked: true
            error: true
        }

        RadioButton {
            text: ""
            checked: true
            enabled: false
        }
    }
}
