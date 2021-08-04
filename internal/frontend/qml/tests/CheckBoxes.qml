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
    id: root
    property ColorScheme colorScheme

    ColumnLayout {
        Layout.fillWidth: true

        spacing: parent.spacing

        CheckBox {
            text: "Checkbox"
        }

        CheckBox {
            text: "Checkbox"
            error: true
        }

        CheckBox {
            text: "Checkbox"
            enabled: false
        }
        CheckBox {
            text: ""
        }

        CheckBox {
            text: ""
            error: true
        }

        CheckBox {
            text: ""
            enabled: false
        }
    }

    ColumnLayout {
        Layout.fillWidth: true

        spacing: parent.spacing

        CheckBox {
            text: "Checkbox"
            checked: true
        }

        CheckBox {
            text: "Checkbox"
            checked: true
            error: true
        }

        CheckBox {
            text: "Checkbox"
            checked: true
            enabled: false
        }
        CheckBox {
            text: ""
            checked: true
        }

        CheckBox {
            text: ""
            checked: true
            error: true
        }

        CheckBox {
            text: ""
            checked: true
            enabled: false
        }
    }
}
