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
import Proton

SettingsView {
    id: root

    property var selectedAddress
    property var categoryId:-1
    property string category: Backend.getBugCategory(root.categoryId)

    signal bugReportWasSent

    function isValidEmail(text) {
        const reEmail = /^[^@]+@[^@]+\.[A-Za-z]+\s*$/;
        return reEmail.test(text);
    }

    function setCategoryId(catId) {
        root.categoryId = catId;
    }

    function setDefaultValue() {
        description.text = Backend.collectAnswers(root.categoryId);
        address.text = root.selectedAddress;
        emailClient.text = Backend.currentEmailClient;
        includeLogs.checked = true;
    }

    function submit() {
        sendButton.loading = true;
        Backend.reportBug(root.category, description.text, address.text, emailClient.text, includeLogs.checked);
    }

    fillHeight: true

    onVisibleChanged: {
        root.setDefaultValue();
    }

    Label {
        colorScheme: root.colorScheme
        text: qsTr("Send report")
        type: Label.Heading
    }
    TextArea {
        id: description

        KeyNavigation.priority: KeyNavigation.BeforeItem
        KeyNavigation.tab: address
        Layout.fillHeight: true
        Layout.fillWidth: true
        Layout.minimumHeight: heightForLinesVisible(4)
        colorScheme: root.colorScheme
        textFormat: Text.MarkdownText

        // set implicitHeight to explicit height because se don't
        // want TextArea implicitHeight (which is height of all text)
        // to be considered in SettingsView internal scroll view
        implicitHeight: height
        label: "Your answers to: " + qsTr(root.category);
        readOnly : true
    }
    TextField {
        id: address
        Layout.fillWidth: true
        colorScheme: root.colorScheme
        label: qsTr("Your contact email")
        placeholderText: qsTr("e.g. jane.doe@protonmail.com")
        validator: function (str) {
            if (!isValidEmail(str)) {
                return qsTr("Enter valid email address");
            }
            return;
        }
    }
    TextField {
        id: emailClient
        Layout.fillWidth: true
        colorScheme: root.colorScheme
        label: qsTr("Your email client (including version)")
        placeholderText: qsTr("e.g. Apple Mail 14.0")
        validator: function (str) {
            if (str.length === 0) {
                return qsTr("Enter an email client name and version");
            }
            return;
        }
    }
    RowLayout {
        CheckBox {
            id: includeLogs
            checked: true
            colorScheme: root.colorScheme
            text: qsTr("Include my recent logs")
        }
        Button {
            Layout.leftMargin: 12
            colorScheme: root.colorScheme
            secondary: true
            text: qsTr("View logs")

            onClicked: Backend.openExternalLink(Backend.logsPath)
        }
    }
    TextEdit {
        Layout.fillWidth: true
        color: root.colorScheme.text_weak
        font.family: ProtonStyle.font_family
        font.letterSpacing: ProtonStyle.caption_letter_spacing
        font.pixelSize: ProtonStyle.caption_font_size
        font.weight: ProtonStyle.fontWeight_400
        readOnly: true
        selectByMouse: true
        selectedTextColor: root.colorScheme.text_invert
        // No way to set lineHeight: ProtonStyle.caption_line_height
        selectionColor: root.colorScheme.interaction_norm
        text: qsTr("Reports are not end-to-end encrypted, please do not send any sensitive information.")
        wrapMode: Text.WordWrap
    }
    Button {
        id: sendButton
        colorScheme: root.colorScheme
        enabled: !loading
        text: qsTr("Send")

        onClicked: {
            description.validate();
            address.validate();
            emailClient.validate();
            if (description.error || address.error || emailClient.error) {
                return;
            }
            submit();
        }

        Connections {
            function onBugReportSendSuccess() {
                root.bugReportWasSent();
            }
            function onReportBugFinished() {
                sendButton.loading = false;
            }

            target: Backend
        }
    }
}