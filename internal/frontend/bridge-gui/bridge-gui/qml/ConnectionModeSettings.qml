// Copyright (c) 2023 Proton AG
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
        text: qsTr("Connection mode")
        type: Label.Heading
        Layout.fillWidth: true
    }

    Label {
        colorScheme: root.colorScheme
        text: qsTr("Change the protocol Bridge and the email client use to connect for IMAP and SMTP.")
        type: Label.Body
        color: root.colorScheme.text_weak
        Layout.fillWidth: true
        wrapMode: Text.WordWrap
    }

    ColumnLayout {
        spacing: 16

        ButtonGroup{ id: imapProtocolSelection }

        Label {
            colorScheme: root.colorScheme
            text: qsTr("IMAP connection")
        }

        RadioButton {
            id: imapSSLButton
            colorScheme: root.colorScheme
            ButtonGroup.group: imapProtocolSelection
            text: qsTr("SSL")
        }

        RadioButton {
            id: imapSTARTTLSButton
            colorScheme: root.colorScheme
            ButtonGroup.group: imapProtocolSelection
            text: qsTr("STARTTLS")
        }
    }

    Rectangle {
        Layout.fillWidth: true
        height: 1
        color: root.colorScheme.border_weak
    }

    ColumnLayout {
        spacing: 16

        ButtonGroup{ id: smtpProtocolSelection }

        Label {
            colorScheme: root.colorScheme
            text: qsTr("SMTP connection")
        }

        RadioButton {
            id: smtpSSLButton
            colorScheme: root.colorScheme
            ButtonGroup.group: smtpProtocolSelection
            text: qsTr("SSL")
        }

        RadioButton {
            id: smtpSTARTTLSButton
            colorScheme: root.colorScheme
            ButtonGroup.group: smtpProtocolSelection
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
            text: qsTr("Save")
            onClicked: {
                submitButton.loading = true
                root.submit()
            }

            enabled: (!loading) && ((imapSSLButton.checked !== Backend.useSSLForIMAP) || (smtpSSLButton.checked !== Backend.useSSLForSMTP))
        }

        Button {
            colorScheme: root.colorScheme
            text: qsTr("Cancel")
            onClicked: root.back()
            secondary: true
        }

        Connections {
            target: Backend

            function onChangeMailServerSettingsFinished() {
                submitButton.loading = false
                root.back()
            }
        }
    }

    function submit(){
        submitButton.loading = true
        Backend.setMailServerSettings(Backend.imapPort, Backend.smtpPort, imapSSLButton.checked, smtpSSLButton.checked)
    }

    function setDefaultValues(){
        imapSSLButton.checked = Backend.useSSLForIMAP
        imapSTARTTLSButton.checked = !Backend.useSSLForIMAP
        smtpSSLButton.checked = Backend.useSSLForSMTP
        smtpSTARTTLSButton.checked = !Backend.useSSLForSMTP
    }

    onVisibleChanged: {
        root.setDefaultValues()
    }
}
