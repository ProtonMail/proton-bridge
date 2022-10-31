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

SettingsView {
    id: root

    fillHeight: false

    Label {
        colorScheme: root.colorScheme
        text: qsTr("SMTP connection mode")
        type: Label.Heading
        Layout.fillWidth: true
    }

    Label {
        colorScheme: root.colorScheme
        text: qsTr("Changes require reconfiguration of email client. Bridge will automatically restart.")
        type: Label.Body
        color: root.colorScheme.text_weak
        Layout.fillWidth: true
        wrapMode: Text.WordWrap
    }

    ColumnLayout {
        spacing: 16

        ButtonGroup{ id: protocolSelection }

        Label {
            colorScheme: root.colorScheme
            text: qsTr("SMTP connection security")
        }

        RadioButton {
            id: sslButton
            colorScheme: root.colorScheme
            ButtonGroup.group: protocolSelection
            text: qsTr("SSL")
        }

        RadioButton {
            id: starttlsButton
            colorScheme: root.colorScheme
            ButtonGroup.group: protocolSelection
            text: qsTr("STARTTLS")
        }
    }

    Rectangle {
        Layout.fillWidth: true
        height: 1
        color: root.colorScheme.border_weak
    }

    RowLayout {
        spacing: 12

        Button {
            id: submitButton
            colorScheme: root.colorScheme
            text: qsTr("Save and restart")
            onClicked: {
                submitButton.loading = true
                root.submit()
            }

            enabled: sslButton.checked !== Backend.useSSLforSMTP
        }

        Button {
            colorScheme: root.colorScheme
            text: qsTr("Cancel")
            onClicked: root.back()
            secondary: true
        }

        Connections {
            target: Backend

            function onToggleUseSSLFinished() {
                submitButton.loading = false
            }
        }
    }

    function submit(){
        submitButton.loading = true
        Backend.toggleUseSSLforSMTP(sslButton.checked)
    }

    function setDefaultValues(){
        sslButton.checked = Backend.useSSLforSMTP
        starttlsButton.checked = !Backend.useSSLforSMTP
    }

    onVisibleChanged: {
        root.setDefaultValues()
    }
}
