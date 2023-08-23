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
        CertificateInstall,
        ProfileInstall
    }

    property var wizard

    signal appleMailAutoconfigCertificateInstallPageShown
    signal appleMailAutoconfigProfileInstallPageShow

    function showAutoconfig() {
        certificateInstall.waitingForCert = false;
        if (Backend.isTLSCertificateInstalled()) {
            stack.currentIndex = ClientConfigAppleMail.Screen.ProfileInstall;
            appleMailAutoconfigProfileInstallPageShow();
        } else {
            stack.currentIndex = ClientConfigAppleMail.Screen.CertificateInstall;
            appleMailAutoconfigCertificateInstallPageShown();
        }
    }

    StackLayout {
        id: stack
        anchors.fill: parent

        // stack index 0
        Item {
            id: certificateInstall

            property bool waitingForCert: false

            Layout.fillHeight: true
            Layout.fillWidth: true

            ColumnLayout {
                anchors.left: parent.left
                anchors.right: parent.right
                anchors.verticalCenter: parent.verticalCenter
                spacing: 24

                Connections {
                    function onCertificateInstallCanceled() {
                        // Note: this will lead to an error message in the final version.
                        certificateInstall.waitingForCert = false;
                        console.error("Certificate installation was canceled");
                    }
                    function onCertificateInstallFailed() {
                        // Note: this will lead to an error page later.
                        certificateInstall.waitingForCert = false;
                        console.error("Certificate installation failed");
                    }
                    function onCertificateInstallSuccess() {
                        certificateInstall.waitingForCert = false;
                        console.error("Certificate installed successfully");
                        root.showAutoconfig();
                    }

                    target: Backend
                }
                ColumnLayout {
                    Layout.fillWidth: true
                    spacing: 16

                    Label {
                        Layout.alignment: Qt.AlignHCenter
                        Layout.fillWidth: true
                        colorScheme: wizard.colorScheme
                        horizontalAlignment: Text.AlignHCenter
                        text: "Install the bridge certificate"
                        type: Label.LabelType.Title
                        wrapMode: Text.WordWrap
                    }
                    Label {
                        Layout.alignment: Qt.AlignHCenter
                        Layout.fillWidth: true
                        color: colorScheme.text_weak
                        colorScheme: wizard.colorScheme
                        horizontalAlignment: Text.AlignHCenter
                        text: "After clicking on the button below, a system pop-up will ask you for your credential, please enter your macOS user credentials (not your Proton account’s) and validate."
                        type: Label.LabelType.Body
                        wrapMode: Text.WordWrap
                    }
                }
                Image {
                    Layout.alignment: Qt.AlignHCenter
                    height: 182
                    opacity: certificateInstall.waitingForCert ? 0.3 : 1.0
                    source: "/qml/icons/img-macos-cert-screenshot.png"
                    width: 140
                }
                ColumnLayout {
                    Layout.fillWidth: true
                    spacing: 16

                    Button {
                        Layout.fillWidth: true
                        colorScheme: wizard.colorScheme
                        enabled: !certificateInstall.waitingForCert
                        loading: certificateInstall.waitingForCert
                        text: "Install the certificate"

                        onClicked: {
                            certificateInstall.waitingForCert = true;
                            Backend.installTLSCertificate();
                        }
                    }
                    Button {
                        Layout.fillWidth: true
                        colorScheme: wizard.colorScheme
                        enabled: !certificateInstall.waitingForCert
                        secondary: true
                        text: qsTr("Cancel")

                        onClicked: {
                            wizard.closeWizard();
                        }
                    }
                }
            }
        }
        // stack index 1
        Item {
            id: profileInstall
            Layout.fillHeight: true
            Layout.fillWidth: true

            ColumnLayout {
                anchors.left: parent.left
                anchors.right: parent.right
                anchors.verticalCenter: parent.verticalCenter
                spacing: 24

                ColumnLayout {
                    Layout.fillWidth: true
                    spacing: 16

                    Label {
                        Layout.alignment: Qt.AlignHCenter
                        Layout.fillWidth: true
                        colorScheme: wizard.colorScheme
                        horizontalAlignment: Text.AlignHCenter
                        text: "Install the profile"
                        type: Label.LabelType.Title
                        wrapMode: Text.WordWrap
                    }
                    Label {
                        Layout.alignment: Qt.AlignHCenter
                        Layout.fillWidth: true
                        color: colorScheme.text_weak
                        colorScheme: wizard.colorScheme
                        horizontalAlignment: Text.AlignHCenter
                        text: qsTr("A system pop-up will appear. Double click on the entry with your email, and click ’Install’ in the dialog that appears.")
                        type: Label.LabelType.Body
                        wrapMode: Text.WordWrap
                    }
                }
                Image {
                    Layout.alignment: Qt.AlignHCenter
                    height: 102
                    opacity: certificateInstall.waitingForCert ? 0.3 : 1.0
                    source: "/qml/icons/img-macos-profile-screenshot.png"
                    width: 364
                }
                ColumnLayout {
                    Layout.fillWidth: true
                    spacing: 16

                    Button {
                        Layout.fillWidth: true
                        colorScheme: wizard.colorScheme
                        text: "Install the profile"

                        onClicked: {
                            wizard.user.configureAppleMail(wizard.address);
                            wizard.showClientConfigEnd();
                        }
                    }
                    Button {
                        Layout.fillWidth: true
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
    }
}

// Label {
//     Layout.alignment: Qt.AlignHCenter
//     Layout.fillWidth: true
//     colorScheme: wizard.colorScheme
//     horizontalAlignment: Text.AlignHCenter
//     text: "Profile install placeholder"
//     type: Label.LabelType.Heading
//     wrapMode: Text.WordWrap
// }
// Button {
//     Layout.fillWidth: true
//     Layout.topMargin: 48
//     colorScheme: wizard.colorScheme
//     text: "Install Profile Placeholder"
//     onClicked: {
//         wizard.user.configureAppleMail(wizard.address);
//         wizard.closeWizard();
//     }
// }
// Button {
//     Layout.fillWidth: true
//     Layout.topMargin: 32
//     colorScheme: wizard.colorScheme
//     secondary: true
//     text: qsTr("Cancel")
//     onClicked: {
//         wizard.closeWizard();
//     }
// }
