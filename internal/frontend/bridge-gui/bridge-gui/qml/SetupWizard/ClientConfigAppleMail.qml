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
import QtQml
import QtQuick
import QtQuick.Layouts
import QtQuick.Controls
import QtQuick.Controls.impl
import Proton

Item {
    id: root
    enum Screen {
        CertificateInstall = 0,
        ProfileInstall = 1
    }

    property var wizard

    function showAutoConfig() {
        certInstallButton.loading = false;
        if (Backend.isTLSCertificateInstalled()) {
            stack.currentIndex = ClientConfigAppleMail.Screen.ProfileInstall;
        } else {
            stack.currentIndex = ClientConfigAppleMail.Screen.CertificateInstall;
        }
    }

    StackLayout {
        id: stack
        anchors.fill: parent

        // stack index 0
        ColumnLayout {
            id: certificateInstall
            Layout.fillHeight: true
            Layout.fillWidth: true

            Connections {
                function onCertificateInstallCanceled() {
                    // Note: this will lead to an error message in the final version.
                    certInstallButton.loading = false;
                    console.error("Certificate installation was canceled");
                }
                function onCertificateInstallFailed() {
                    // Note: this will lead to an error page later.
                    certInstallButton.loading = false;
                    console.error("Certificate installation failed");
                }
                function onCertificateInstallSuccess() {
                    certInstallButton.loading = false;
                    console.error("Certificate installed successfully");
                    stack.currentIndex = ClientConfigAppleMail.Screen.ProfileInstall;
                }

                target: Backend
            }
            Label {
                Layout.alignment: Qt.AlignHCenter
                Layout.fillWidth: true
                colorScheme: wizard.colorScheme
                horizontalAlignment: Text.AlignHCenter
                text: "Certificate install placeholder"
                type: Label.LabelType.Heading
                wrapMode: Text.WordWrap
            }
            Button {
                id: certInstallButton
                Layout.fillWidth: true
                Layout.topMargin: 48
                colorScheme: wizard.colorScheme
                enabled: !loading
                loading: false
                text: "Install Certificate Placeholder"

                onClicked: {
                    certInstallButton.loading = true;
                    Backend.installTLSCertificate();
                }
            }
            Button {
                Layout.fillWidth: true
                Layout.topMargin: 32
                colorScheme: wizard.colorScheme
                enabled: !certInstallButton.loading
                secondary: true
                text: qsTr("Cancel")

                onClicked: {
                    wizard.closeWizard();
                }
            }
        }

        // stack index 1
        ColumnLayout {
            id: profileInstall
            Layout.fillHeight: true
            Layout.fillWidth: true

            Label {
                Layout.alignment: Qt.AlignHCenter
                Layout.fillWidth: true
                colorScheme: wizard.colorScheme
                horizontalAlignment: Text.AlignHCenter
                text: "Profile install placeholder"
                type: Label.LabelType.Heading
                wrapMode: Text.WordWrap
            }
            Button {
                Layout.fillWidth: true
                Layout.topMargin: 48
                colorScheme: wizard.colorScheme
                text: "Install Profile Placeholder"

                onClicked: {
                    wizard.user.configureAppleMail(wizard.address);
                    wizard.closeWizard();
                }
            }
            Button {
                Layout.fillWidth: true
                Layout.topMargin: 32
                colorScheme: wizard.colorScheme
                secondary: true
                text: qsTr("Cancel")

                onClicked: {
                    wizard.closeWizard();
                }
            }
        }
    }
}