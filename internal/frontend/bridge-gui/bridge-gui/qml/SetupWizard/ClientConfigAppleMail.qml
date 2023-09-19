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
        if (Backend.isTLSCertificateInstalled()) {
            showProfileInstall();
        } else {
            showCertificateInstall();
        }
    }
    function showCertificateInstall() {
        certificateInstall.reset();
        stack.currentIndex = ClientConfigAppleMail.Screen.CertificateInstall;
        appleMailAutoconfigCertificateInstallPageShown();
    }
    function showProfileInstall() {
        profileInstall.reset();
        stack.currentIndex = ClientConfigAppleMail.Screen.ProfileInstall;
        appleMailAutoconfigProfileInstallPageShow();
    }

    StackLayout {
        id: stack
        anchors.fill: parent

        // stack index 0
        Item {
            id: certificateInstall

            property string errorString: ""
            property bool showBugReportLink: false
            property bool waitingForCert: false

            function clearError() {
                errorString = "";
                showBugReportLink = false;
            }
            function reset() {
                waitingForCert = false;
                clearError();
            }

            Layout.fillHeight: true
            Layout.fillWidth: true

            ColumnLayout {
                anchors.left: parent.left
                anchors.right: parent.right
                anchors.verticalCenter: parent.verticalCenter
                spacing: ProtonStyle.wizard_spacing_large

                Connections {
                    function onCertificateInstallCanceled() {
                        certificateInstall.waitingForCert = false;
                        certificateInstall.errorString = qsTr("Apple Mail cannot be configured if you do not install the certificate. Please retry.");
                        certificateInstall.showBugReportLink = false;
                    }
                    function onCertificateInstallFailed() {
                        certificateInstall.waitingForCert = false;
                        certificateInstall.errorString = qsTr("An error occurred while installing the certificate.");
                        certificateInstall.showBugReportLink = true;
                    }
                    function onCertificateInstallSuccess() {
                        certificateInstall.reset();
                        root.showAutoconfig();
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
                    opacity: certificateInstall.waitingForCert ? 0.3 : 1.0
                    source: "/qml/icons/img-macos-cert-screenshot.png"
                    width: 140
                }
                ColumnLayout {
                    Layout.fillWidth: true
                    spacing: ProtonStyle.wizard_spacing_medium

                    Button {
                        Layout.fillWidth: true
                        colorScheme: wizard.colorScheme
                        enabled: !certificateInstall.waitingForCert
                        loading: certificateInstall.waitingForCert
                        text: qsTr("Install the certificate")

                        onClicked: {
                            certificateInstall.clearError();
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
                                visible: certificateInstall.errorString.length > 0
                            }
                            Label {
                                id: errorLabel
                                Layout.fillWidth: true
                                color: wizard.colorScheme.signal_danger
                                colorScheme: wizard.colorScheme
                                horizontalAlignment: Text.AlignHCenter
                                text: certificateInstall.errorString
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
                            visible: certificateInstall.showBugReportLink
                        }
                    }
                }
            }
        }
        // stack index 1
        Item {
            id: profileInstall

            property bool profilePaneLaunched: false

            function reset() {
                profilePaneLaunched = false;
            }

            Layout.fillHeight: true
            Layout.fillWidth: true

            ColumnLayout {
                anchors.left: parent.left
                anchors.right: parent.right
                anchors.verticalCenter: parent.verticalCenter
                spacing: ProtonStyle.wizard_spacing_large

                ColumnLayout {
                    Layout.fillWidth: true
                    spacing: ProtonStyle.wizard_spacing_medium

                    Label {
                        Layout.alignment: Qt.AlignHCenter
                        Layout.fillWidth: true
                        colorScheme: wizard.colorScheme
                        horizontalAlignment: Text.AlignHCenter
                        text: qsTr("Install the profile")
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
                    source: "/qml/icons/img-macos-profile-screenshot.png"
                    width: 364
                }
                ColumnLayout {
                    Layout.fillWidth: true
                    spacing: ProtonStyle.wizard_spacing_medium

                    Button {
                        Layout.fillWidth: true
                        colorScheme: wizard.colorScheme
                        text: profileInstall.profilePaneLaunched ? qsTr("I have installed the profile") : qsTr("Install the profile")

                        onClicked: {
                            if (profileInstall.profilePaneLaunched) {
                                wizard.showClientConfigEnd();
                            } else {
                                wizard.user.configureAppleMail(wizard.address);
                                profileInstall.profilePaneLaunched = true;
                            }
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
