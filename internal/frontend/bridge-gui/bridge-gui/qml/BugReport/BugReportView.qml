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
import Proton
import ".."

SettingsView {
    id: root
    property var selectedAddress
    property var categoryId: -1
    property string category: Backend.getBugCategory(root.categoryId)
    property var suggestions: null

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

    ColumnLayout {
        spacing: 32

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
            readOnly: true
        }

        ColumnLayout {
            id: suggestionBox
            visible: suggestions && suggestions.length > 0
            spacing: 8
            RowLayout {
                Label {
                    colorScheme: root.colorScheme
                    text: qsTr("We believe these links may help you solve your problem")
                    type: Label.Body_semibold
                }
                InfoTooltip {
                    colorScheme: root.colorScheme
                    text: qsTr("The links will open in an external browser. If you cancel the report, your input will be preserved until you restart Bridge.")
                    Layout.bottomMargin: -4
                }
            }
            Repeater {
                model: suggestions
                LinkLabel {
                    required property var modelData
                    colorScheme: root.colorScheme
                    text: modelData.title
                    link: modelData.url
                    external: true
                }
            }
        }

        RowLayout {
            spacing: 12

            TextField {
                id: address
                Layout.preferredWidth: 1
                Layout.fillWidth: true
                colorScheme: root.colorScheme
                label: qsTr("Your contact email")
                placeholderText: qsTr("e.g. jane.doe@protonmail.com")
                validator: function (str) {
                    if (!isValidEmail(str)) {
                        return qsTr("Enter valid email address");
                    }
                }
            }
            TextField {
                id: emailClient
                Layout.preferredWidth: 1
                Layout.fillWidth: true
                colorScheme: root.colorScheme
                label: qsTr("Your email client (including version)")
                placeholderText: qsTr("e.g. Apple Mail 14.0")
                validator: function (str) {
                    if (str.length === 0) {
                        return qsTr("Enter an email client name and version");
                    }
                }
            }
        }
        RowLayout {
            spacing: 12

            CheckBox {
                id: includeLogs
                checked: true
                colorScheme: root.colorScheme
                text: qsTr("Include my recent logs")
            }

            Button {
                colorScheme: root.colorScheme
                secondary: true
                text: qsTr("View logs")

                onClicked: Backend.openExternalLink(Backend.logsPath)
            }

            Label {
                Layout.fillWidth: true
                verticalAlignment: Qt.AlignVCenter
                colorScheme: root.colorScheme
                type: Label.Caption
                color: root.colorScheme.text_weak
                text: qsTr("Reports are not end-to-end encrypted, please do not send any sensitive information.")
                wrapMode: Text.WordWrap
            }
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

                function onReceivedKnowledgeBaseSuggestions(suggestions) {
                    root.suggestions = suggestions
                }

                target: Backend
            }
        }
    }
}