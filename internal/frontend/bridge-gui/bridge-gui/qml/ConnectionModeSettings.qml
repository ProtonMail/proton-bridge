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
import QtQuick
import QtQuick.Layouts
import QtQuick.Controls
import QtQuick.Controls.impl
import Proton

SettingsView {
    id: root
    function setDefaultValues() {
        imapSSLButton.checked = Backend.useSSLForIMAP;
        imapSTARTTLSButton.checked = !Backend.useSSLForIMAP;
        smtpSSLButton.checked = Backend.useSSLForSMTP;
        smtpSTARTTLSButton.checked = !Backend.useSSLForSMTP;
    }
    function submit() {
        submitButton.loading = true;
        Backend.setMailServerSettings(Backend.imapPort, Backend.smtpPort, imapSSLButton.checked, smtpSSLButton.checked);
    }

    fillHeight: false

    onVisibleChanged: {
        root.setDefaultValues();
    }

    Label {
        Layout.fillWidth: true
        colorScheme: root.colorScheme
        text: qsTr("Connection mode")
        type: Label.Heading
    }
    Label {
        Layout.fillWidth: true
        color: root.colorScheme.text_weak
        colorScheme: root.colorScheme
        text: qsTr("Change the protocol Bridge and the email client use to connect for IMAP and SMTP.")
        type: Label.Body
        wrapMode: Text.WordWrap
    }
    ColumnLayout {
        spacing: 16

        ButtonGroup {
            id: imapProtocolSelection
        }
        Label {
            colorScheme: root.colorScheme
            text: qsTr("IMAP connection")
        }
        RadioButton {
            id: imapSSLButton
            ButtonGroup.group: imapProtocolSelection
            colorScheme: root.colorScheme
            text: qsTr("SSL")
        }
        RadioButton {
            id: imapSTARTTLSButton
            ButtonGroup.group: imapProtocolSelection
            colorScheme: root.colorScheme
            text: qsTr("STARTTLS")
        }
    }
    Rectangle {
        Layout.fillWidth: true
        color: root.colorScheme.border_weak
        height: 1
    }
    ColumnLayout {
        spacing: 16

        ButtonGroup {
            id: smtpProtocolSelection
        }
        Label {
            colorScheme: root.colorScheme
            text: qsTr("SMTP connection")
        }
        RadioButton {
            id: smtpSSLButton
            ButtonGroup.group: smtpProtocolSelection
            colorScheme: root.colorScheme
            text: qsTr("SSL")
        }
        RadioButton {
            id: smtpSTARTTLSButton
            ButtonGroup.group: smtpProtocolSelection
            colorScheme: root.colorScheme
            text: qsTr("STARTTLS")
        }
    }
    Rectangle {
        Layout.fillWidth: true
        color: root.colorScheme.border_weak
        height: 1
    }
    RowLayout {
        spacing: 12

        Button {
            id: submitButton
            colorScheme: root.colorScheme
            enabled: (!loading) && ((imapSSLButton.checked !== Backend.useSSLForIMAP) || (smtpSSLButton.checked !== Backend.useSSLForSMTP))
            text: qsTr("Save")

            onClicked: {
                submitButton.loading = true;
                root.submit();
            }
        }
        Button {
            colorScheme: root.colorScheme
            secondary: true
            text: qsTr("Cancel")

            onClicked: root.back()
        }
        Connections {
            function onChangeMailServerSettingsFinished() {
                submitButton.loading = false;
                root.back();
            }

            target: Backend
        }
    }
}
