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

    // Primary buttons
    ButtonsColumn {
        colorScheme: root.colorScheme
        Layout.fillWidth: true
        Layout.fillHeight: true

        iconLoading: "/qml/icons/Loader_16.svg"
    }

    // Secondary buttons
    ButtonsColumn {
        colorScheme: root.colorScheme
        Layout.fillWidth: true
        Layout.fillHeight: true

        secondary: true
        iconLoading: "/qml/icons/Loader_16.svg"
    }

    // Secondary icons
    ButtonsColumn {
        colorScheme: root.colorScheme
        Layout.fillWidth: true
        Layout.fillHeight: true

        secondary: true
        textNormal: ""
        iconNormal: "/qml/icons/ic-cross-close.svg"
        textDisabled: ""
        iconDisabled: "/qml/icons/ic-cross-close.svg"
        textLoading: ""
        iconLoading: "/qml/icons/Loader_16.svg"
    }

    // Icons
    ButtonsColumn {
        colorScheme: root.colorScheme
        Layout.fillWidth: true
        Layout.fillHeight: true

        textNormal: ""
        iconNormal: "/qml/icons/ic-cross-close.svg"
        textDisabled: ""
        iconDisabled: "/qml/icons/ic-cross-close.svg"
        textLoading: ""
        iconLoading: "/qml/icons/Loader_16.svg"
    }
}
