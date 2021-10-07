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
import QtQuick.Layouts 1.12
import QtQuick.Controls 2.13
import QtQuick.Controls.impl 2.13

import Proton 4.0

SettingsView {
    id: root

    property bool _valuesOK: !imapField.error && !smtpField.error
    property bool _valuesChanged: (
        imapField.text*1 != root.backend.portIMAP ||
        smtpField.text*1 != root.backend.portSMTP
    )

    Label {
        colorScheme: root.colorScheme
        text: qsTr("Default ports")
        type: Label.Heading
        Layout.fillWidth: true
    }

    Label {
        colorScheme: root.colorScheme
        text: qsTr("Changes require reconfiguration of your email client. Bridge will automatically restart.")
        type: Label.Body
        color: root.colorScheme.text_weak
        Layout.fillWidth: true
        wrapMode: Text.WordWrap
    }

    RowLayout {
        spacing: 16

        TextField {
            id: imapField
            colorScheme: root.colorScheme
            label: qsTr("IMAP port")
            Layout.preferredWidth: 160
            onEditingFinished: root.validate(imapField)
        }
        TextField {
            id: smtpField
            colorScheme: root.colorScheme
            label: qsTr("SMTP port")
            Layout.preferredWidth: 160
            onEditingFinished: root.validate(smtpField)
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
            enabled: root._valuesOK && root._valuesChanged
            onClicked: {
                submitButton.loading = true
                root.submit()
            }
        }

        Button {
            colorScheme: root.colorScheme
            text: qsTr("Cancel")
            onClicked: root.back()
            secondary: true
        }

        Connections {
            target: root.backend

            onChangePortFinished: submitButton.loading = false
        }
    }

    onBack: {
        root.setDefaultValues()
    }

    function validate(field) {
        var num = field.text*1
        if (! (num > 1 && num < 65536) )  {
            field.error = true
            field.assistiveText = qsTr("Invalid port number.")
            return
        }

        if (imapField.text == smtpField.text) {
            field.error = true
            field.assistiveText = qsTr("Port numbers must be different.")
            return
        }

        field.error = false
        field.assistiveText = ""
    }

    function isPortFree(field) {
        field.error = false
        field.assistiveText = ""

        var num = field.text*1
        if (num == root.backend.portIMAP) return true
        if (num == root.backend.portSMTP) return true
        if (!root.backend.isPortFree(num)) {
            field.error = true
            field.assistiveText = qsTr("Port occupied.")
            submitButton.loading = false
            return false
        }
    }

    function submit(){
        submitButton.loading = true
        if (!isPortFree(imapField)) return
        if (!isPortFree(smtpField)) return
        root.backend.changePorts(imapField.text, smtpField.text)
    }

    function setDefaultValues(){
        imapField.text = backend.portIMAP
        smtpField.text = backend.portSMTP
    }

    Component.onCompleted: root.setDefaultValues()
}
