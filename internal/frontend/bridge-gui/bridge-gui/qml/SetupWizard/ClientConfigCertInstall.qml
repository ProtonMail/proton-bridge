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

import QtQml
import QtQuick
import QtQuick.Layouts
import QtQuick.Controls
import QtQuick.Controls.impl

Item {
    id: root

    property string errorString: ""
    property bool showBugReportLink: false
    property bool waitingForCert: false
    property var wizard

    function clearError() {
        errorString = "";
        showBugReportLink = false;
    }
    function reset() {
        waitingForCert = false;
        clearError();
    }

    ColumnLayout {
        anchors.left: parent.left
        anchors.right: parent.right
        anchors.verticalCenter: parent.verticalCenter
        spacing: ProtonStyle.wizard_spacing_large

        Connections {
            function onCertificateInstallCanceled() {
                root.waitingForCert = false;
                root.errorString = qsTr("%1 cannot be configured if you do not install the certificate. Please retry.").arg(wizard.clientName());
                root.showBugReportLink = false;
            }
            function onCertificateInstallFailed() {
                root.waitingForCert = false;
                root.errorString = qsTr("An error occurred while installing the certificate.");
                root.showBugReportLink = true;
            }


            target: Backend
        }
        ColumnLayout {
            Layout.fillWidth: true
            spacing: ProtonStyle.wizard_spacing_medium

            Label {
                Layout.alignment: Qt.AlignHCenter
                Layout.fillWidth: true
                colorScheme: wizard.colorScheme
                horizontalAlignment: Text.AlignHCenter
                text: qsTr("Install the bridge certificate")
                type: Label.LabelType.Title
                wrapMode: Text.WordWrap
            }
            Label {
                Layout.alignment: Qt.AlignHCenter
                Layout.fillWidth: true
                color: colorScheme.text_weak
                colorScheme: wizard.colorScheme
                horizontalAlignment: Text.AlignHCenter
                text: qsTr("After clicking on the button below, a system pop-up will ask you for your credentials, please enter your macOS user credentials (not your Proton account’s) and validate.")
                type: Label.LabelType.Body
                wrapMode: Text.WordWrap
            }
        }
        Image {
            Layout.alignment: Qt.AlignHCenter
            height: 182
            opacity: root.waitingForCert ? 0.3 : 1.0
            source: "/qml/icons/img-macos-cert-screenshot.png"
            width: 140
        }
        ColumnLayout {
            Layout.fillWidth: true
            spacing: ProtonStyle.wizard_spacing_medium

            Button {
                Layout.fillWidth: true
                colorScheme: wizard.colorScheme
                enabled: !root.waitingForCert
                loading: root.waitingForCert
                text: qsTr("Install the certificate")

                onClicked: {
                    root.clearError();
                    root.waitingForCert = true;
                    Backend.installTLSCertificate();
                }
            }
            Button {
                Layout.fillWidth: true
                colorScheme: wizard.colorScheme
                enabled: !root.waitingForCert
                secondary: true
                text: qsTr("Cancel")

                onClicked: {
                    wizard.closeWizard();
                }
            }
            ColumnLayout {
                Layout.fillWidth: true
                spacing: ProtonStyle.wizard_spacing_small

                RowLayout {
                    Layout.fillWidth: true
                    spacing: ProtonStyle.wizard_spacing_extra_small

                    ColorImage {
                        color: wizard.colorScheme.signal_danger
                        height: errorLabel.lineHeight
                        source: "/qml/icons/ic-exclamation-circle-filled.svg"
                        sourceSize.height: errorLabel.lineHeight
                        visible: root.errorString.length > 0
                    }
                    Label {
                        id: errorLabel
                        Layout.fillWidth: true
                        color: wizard.colorScheme.signal_danger
                        colorScheme: wizard.colorScheme
                        horizontalAlignment: Text.AlignHCenter
                        text: root.errorString
                        type: Label.LabelType.Body_semibold
                        wrapMode: Text.WordWrap
                    }
                }
                LinkLabel {
                    Layout.alignment: Qt.AlignHCenter
                    callback: wizard.showBugReport
                    colorScheme: wizard.colorScheme
                    link: "#"
                    text: qsTr("Report the problem")
                    visible: root.showBugReportLink
                }
            }
        }
    }
}