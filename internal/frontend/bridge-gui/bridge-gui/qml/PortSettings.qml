// Copyright (c) 2024 Proton AG
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

    property bool _valuesChanged: (imapField.text * 1 !== Backend.imapPort || smtpField.text * 1 !== Backend.smtpPort)
    property var notifications

    function isPortFree(field) {
        const num = field.text * 1;
        if (num === Backend.imapPort)
            return true;
        if (num === Backend.smtpPort)
            return true;
        if (!Backend.isPortFree(num)) {
            field.error = true;
            field.errorString = qsTr("Port occupied");
            return false;
        }
        return true;
    }
    function setDefaultValues() {
        imapField.text = Backend.imapPort;
        smtpField.text = Backend.smtpPort;
        imapField.error = false;
        smtpField.error = false;
    }
    function validate(port) {
        const num = port * 1;
        if (!(num > 1 && num < 65536)) {
            return qsTr("Invalid port number");
        }
        if (imapField.text === smtpField.text) {
            return qsTr("Port numbers must be different");
        }
    }

    fillHeight: false

    Component.onCompleted: root.setDefaultValues()
    onBack: {
        root.setDefaultValues();
    }

    Label {
        Layout.fillWidth: true
        colorScheme: root.colorScheme
        text: qsTr("Default ports")
        type: Label.Heading
    }
    Label {
        Layout.fillWidth: true
        color: root.colorScheme.text_weak
        colorScheme: root.colorScheme
        text: qsTr("Changes require reconfiguration of your email client.")
        type: Label.Body
        wrapMode: Text.WordWrap
    }
    RowLayout {
        spacing: 16

        TextField {
            id: imapField
            Layout.alignment: Qt.AlignTop | Qt.AlignLeft
            Layout.preferredWidth: 160
            colorScheme: root.colorScheme
            label: qsTr("IMAP port")
            validator: root.validate
        }
        TextField {
            id: smtpField
            Layout.alignment: Qt.AlignTop | Qt.AlignLeft
            Layout.preferredWidth: 160
            colorScheme: root.colorScheme
            label: qsTr("SMTP port")
            validator: root.validate
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
            enabled: (!loading) && root._valuesChanged
            text: qsTr("Save")

            onClicked: {
                // removing error here because we may have set it manually (port occupied)
                imapField.error = false;
                smtpField.error = false;

                // checking errors separately because we want to display "same port" error only once
                imapField.validate();
                if (imapField.error) {
                    return;
                }
                smtpField.validate();
                if (smtpField.error) {
                    return;
                }
                submitButton.loading = true;

                // check both ports before returning an error
                let err = false;
                err |= !isPortFree(imapField);
                err |= !isPortFree(smtpField);
                if (err) {
                    submitButton.loading = false;
                    return;
                }

                // We turn off all port error notification. They will be restored if problems persist
                root.notifications.imapPortStartupError.active = false;
                root.notifications.smtpPortStartupError.active = false;
                root.notifications.imapPortChangeError.active = false;
                root.notifications.smtpPortChangeError.active = false;
                Backend.setMailServerSettings(imapField.text, smtpField.text, Backend.useSSLForIMAP, Backend.useSSLForSMTP);
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
            }

            target: Backend
        }
    }
}
