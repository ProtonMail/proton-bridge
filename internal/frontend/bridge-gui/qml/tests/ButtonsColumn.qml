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

import QtQuick.Layouts
import QtQuick
import QtQuick.Controls

import "../Proton"

ColumnLayout {
    id: root
    property ColorScheme colorScheme

    property string textNormal: "Button"
    property string iconNormal: ""
    property string textDisabled: "Disabled"
    property string iconDisabled: ""
    property string textLoading: "Loading"
    property string iconLoading: ""
    property bool secondary: false

    Button {
        colorScheme: root.colorScheme
        Layout.fillWidth: true

        Layout.minimumHeight: implicitHeight
        Layout.minimumWidth: implicitWidth

        text: root.textNormal
        icon.source: iconNormal
        secondary: root.secondary
    }


    Button {
        colorScheme: root.colorScheme
        Layout.fillWidth: true

        Layout.minimumHeight: implicitHeight
        Layout.minimumWidth: implicitWidth

        text: root.textDisabled
        icon.source: iconDisabled
        secondary: root.secondary

        enabled: false
    }

    Button {
        colorScheme: root.colorScheme
        Layout.fillWidth: true

        Layout.minimumHeight: implicitHeight
        Layout.minimumWidth: implicitWidth

        text: root.textLoading
        icon.source: iconLoading
        secondary: root.secondary

        loading: true
    }
}
