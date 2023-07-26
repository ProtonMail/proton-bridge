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
import QtQml
import QtQuick
import QtQuick.Layouts
import QtQuick.Controls
import QtQuick.Controls.impl
import "." as Proton

Item {
    id: root

    property ColorScheme colorScheme

    ColumnLayout {
        anchors.left: parent.left
        anchors.right: parent.right
        anchors.top: parent.top
        spacing: 96

        Label {
            Layout.alignment: Qt.AlignHCenter
            Layout.fillWidth: true
            colorScheme: root.colorScheme
            horizontalAlignment: Text.AlignHCenter
            text: qsTr("Two-step process")
            type: Label.LabelType.Heading
        }
        StepDescriptionBox {
            colorScheme: root.colorScheme
            description: "Connect Bridge to your Proton account"
            icon: "/qml/icons/ic-bridge.svg"
            title: "Step 1"
            iconSize: 48
        }
        StepDescriptionBox {
            colorScheme: root.colorScheme
            description: "Connect your email client to Bridge"
            icon: "/qml/icons/img-mail-clients.svg"
            title: "Step 2"
            iconSize: 64

        }
        Button {
            Layout.alignment: Qt.AlignHCenter
            Layout.preferredWidth: 320
            colorScheme: root.colorScheme
            text: "Let's start"
        }
    }
}