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

    property bool _valuesChanged: (
        imapField.text*1 !== Backend.portIMAP ||
        smtpField.text*1 !== Backend.portSMTP
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
            validator: root.validate
        }
        TextField {
            id: smtpField
            colorScheme: root.colorScheme
            label: qsTr("SMTP port")
            Layout.preferredWidth: 160
            validator: root.validate
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
            enabled: root._valuesChanged
            onClicked: {
                // removing error here because we may have set it manually (port occupied)
                imapField.error = false
                smtpField.error = false

                // checking errors seperatly because we want to display "same port" error only once
                imapField.validate()
                if (imapField.error) {
                    return
                }
                smtpField.validate()
                if (smtpField.error) {
                    return
                }

                submitButton.loading = true

                // check both ports before returning an error
                var err = false
                err |= !isPortFree(imapField)
                err |= !isPortFree(smtpField)
                if (err) {
                    submitButton.loading = false
                    return
                }

                Backend.changePorts(imapField.text, smtpField.text)
            }
        }

        Button {
            colorScheme: root.colorScheme
            text: qsTr("Cancel")
            onClicked: root.back()
            secondary: true
        }

        Connections {
            target: Backend

            function onChangePortFinished() {
                submitButton.loading = false
            }
        }
    }

    onBack: {
        root.setDefaultValues()
    }

    function validate(port) {
        var num = port*1
        if (! (num > 1 && num < 65536) )  {
            return qsTr("Invalid port number")
        }

        if (imapField.text == smtpField.text) {
            return qsTr("Port numbers must be different")
        }

        return
    }

    function isPortFree(field) {
        var num = field.text*1
        if (num === Backend.portIMAP) return true
        if (num === Backend.portSMTP) return true
        if (!Backend.isPortFree(num)) {
            field.error = true
            field.errorString = qsTr("Port occupied")
            return false
        }

        return true
    }

    function setDefaultValues(){
        imapField.text = Backend.portIMAP
        smtpField.text = Backend.portSMTP
    }

    Component.onCompleted: root.setDefaultValues()
}
