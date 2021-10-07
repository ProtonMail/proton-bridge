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
import QtQuick.Controls 2.12

import Proton 4.0

SettingsView {
    id: root

    property var selectedAddress

    Label {
        text: qsTr("Report a problem")
        colorScheme: root.colorScheme
        type: Label.Heading
    }


    TextArea {
        id: description
        property int _minLength: 150
        property int _maxLength: 800
        property bool _inputOK: description.text.length>=description._minLength && description.text.length<=description._maxLength

        label: qsTr("Description")
        colorScheme: root.colorScheme
        Layout.fillWidth: true
        Layout.minimumHeight: 100
        hint: description.text.length + "/" + _maxLength
        placeholderText: qsTr("Tell us what went wrong or isn't working (min. %1 characters).").arg(_minLength)

        onEditingFinished: {
            if (!description._inputOK) {
                description.error = true
                if (description.text.length <= description._minLength) {
                    description.assistiveText = qsTr("Enter a problem description (min. %1 characters).").arg(_minLength)
                } else {
                    description.assistiveText = qsTr("Enter a problem description (max. %1 characters).").arg(_maxLength)
                }
            } else {
                description.error = false
                description.assistiveText = ""
            }
        }
        onTextChanged: {
            description.error = false
            description.assistiveText = ""
        }
    }


    TextField {
        id: address
        property bool _inputOK: root.isValidEmail(address.text)

        label: qsTr("Your contact email")
        colorScheme: root.colorScheme
        Layout.fillWidth: true
        placeholderText: qsTr("e.g. jane.doe@protonmail.com")

        onEditingFinished: {
            if (!address._inputOK) {
                address.error = true
                address.assistiveText = qsTr("Enter valid email address")
            } else {
                address.assistiveText = ""
                address.error = false
            }
        }
        onTextChanged: {
            address.error = false
            address.assistiveText = ""
        }
    }

    TextField {
        id: emailClient
        property bool _inputOK: emailClient.text.length > 0

        label: qsTr("Your email client (including version)")
        colorScheme: root.colorScheme
        Layout.fillWidth: true
        placeholderText: qsTr("e.g. Apple Mail 14.0")

        onEditingFinished: {
            if (!emailClient._inputOK) {
                emailClient.assistiveText = qsTr("Enter an email client name and version")
                emailClient.error = true
            } else {
                emailClient.assistiveText = ""
                emailClient.error = false
            }
        }
        onTextChanged: {
            emailClient.error = false
            emailClient.assistiveText = ""
        }
    }


    RowLayout {
        CheckBox {
            id: includeLogs
            text: qsTr("Include my recent logs")
            colorScheme: root.colorScheme
            checked: true
        }
        Button {
            Layout.leftMargin: 12
            text: qsTr("View logs")
            secondary: true
            colorScheme: root.colorScheme
            onClicked: Qt.openUrlExternally("file://"+root.backend.logsPath)
        }
    }

    Label {
        text: {
            var address = "bridge@protonmail.com"
            var mailTo = `<a href="mailto://${address}">${address}</a>`
            return qsTr("These reports are not end-to-end encrypted. In case of sensitive information, contact us at %1.").arg(mailTo)
        }
        colorScheme: root.colorScheme
        Layout.fillWidth: true
        wrapMode: Text.WordWrap
        type: Label.Caption
        color: root.colorScheme.text_weak
    }

    Button {
        id: sendButton
        text: qsTr("Send")
        colorScheme: root.colorScheme
        onClicked: root.submit()
        enabled: description._inputOK && address._inputOK && emailClient._inputOK

        Connections {target: root.backend; onReportBugFinished: sendButton.loading = false }
    }

    function setDefaultValue() {
        description.text = ""
        description.error = false
        description.assistiveText = ""

        address.text = root.selectedAddress
        address.error = false
        address.assistiveText = ""

        emailClient.text = root.backend.currentEmailClient
        emailClient.error = false
        emailClient.assistiveText = ""

        includeLogs.checked = true
    }

    function isValidEmail(text){
        var reEmail = /\w+@\w+\.\w+/
        return reEmail.test(text)
    }

    function submit() {
        sendButton.loading = true
        root.backend.reportBug(
            description.text,
            address.text,
            emailClient.text,
            includeLogs.checked
        )
    }

    Component.onCompleted: root.setDefaultValue()

    onVisibleChanged: {
        root.setDefaultValue()
    }
}
