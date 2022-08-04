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
import QtQuick.Layouts
import QtQuick.Controls
import QtQuick.Controls.impl

import Proton

Rectangle {
    id: root

    property ColorScheme colorScheme
    property string title
    property string hostname
    property string port
    property string username
    property string password
    property string security

    implicitWidth: 304
    implicitHeight: content.height + 2*root._margin

    color: root.colorScheme.background_norm
    radius: ProtonStyle.card_radius

    property int _margin: 24

    ColumnLayout {
        id: content
        width: root.width - 2*root._margin
        anchors{
            top: root.top
            left: root.left
            leftMargin   : root._margin
            rightMargin  : root._margin
            topMargin    : root._margin
            bottomMargin : root._margin
        }

        spacing: 12

        Label {
            colorScheme: root.colorScheme
            text: root.title
            type: Label.Body_semibold
        }

        Item{}

        ConfigurationItem{ colorScheme: root.colorScheme; label: qsTr("Hostname") ; value: root.hostname }
        ConfigurationItem{ colorScheme: root.colorScheme; label: qsTr("Port")     ; value: root.port     }
        ConfigurationItem{ colorScheme: root.colorScheme; label: qsTr("Username") ; value: root.username }
        ConfigurationItem{ colorScheme: root.colorScheme; label: qsTr("Password") ; value: root.password }
        ConfigurationItem{ colorScheme: root.colorScheme; label: qsTr("Security") ; value: root.security }
    }
}

