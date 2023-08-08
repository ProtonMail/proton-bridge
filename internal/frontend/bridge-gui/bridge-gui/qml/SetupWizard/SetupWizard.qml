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
import "." as Proton
import ".."

Item {
    id: root
    enum Client {
        AppleMail,
        MicrosoftOutlook,
        MozillaThunderbird,
        Generic
    }

    property string address
    property int client
    property string clientVersion
    property ColorScheme colorScheme
    property string userID
    property bool wasSignedOut

    function clientIconSource() {
        switch (client) {
            case SetupWizard.Client.AppleMail:
                return "/qml/icons/ic-apple-mail.svg";
            case SetupWizard.Client.MicrosoftOutlook:
                return "/qml/icons/ic-microsoft-outlook.svg";
            case SetupWizard.Client.MozillaThunderbird:
                return "/qml/icons/ic-mozilla-thunderbird.svg";
            case SetupWizard.Client.Generic:
                return "/qml/icons/ic-other-mail-clients.svg";
            default:
                console.error("Unknown mail client " + client)
                return "/qml/icons/ic-other-mail-clients.svg";
        }
    }

    function clientName() {
        switch (client) {
            case SetupWizard.Client.AppleMail:
                return "Apple Mail";
            case SetupWizard.Client.MicrosoftOutlook:
                return "Outlook";
            case SetupWizard.Client.MozillaThunderbird:
                return "Thunderbird";
            case SetupWizard.Client.Generic:
                return "your email client";
            default:
                console.error("Unknown mail client " + client)
                return "your email client";
        }
    }

    function closeWizard() {
        root.visible = false;
    }
    function showOutlookSelector() {
        root.visible = true;
        leftContent.showOutlookSelector();
        rightContent.currentIndex = 3;
    }
    function start() {
        root.visible = true;
        leftContent.showOnboarding();
        rightContent.currentIndex = 0;
    }
    function startClientConfig() {
        root.visible = true;
        leftContent.showClientSelector();
        rightContent.currentIndex = 2;
    }
    function startLogin(wasSignedOut = false) {
        root.visible = true;
        root.userID = "";
        root.address = "";
        root.wasSignedOut = wasSignedOut;
        leftContent.showLogin();
        rightContent.currentIndex = 1;
        login.reset(true);
    }

    function showClientWarning() {
        root.visible = true;
        leftContent.showClientConfigWarning();
        rightContent.currentIndex = 4
    }

    Connections {
        function onLoginFinished() {
            startClientConfig();
        }

        target: Backend
    }
    RowLayout {
        anchors.fill: parent
        spacing: 0

        Rectangle {
            id: leftHalf
            Layout.fillHeight: true
            Layout.fillWidth: true
            color: root.colorScheme.background_norm

            LeftPane {
                id: leftContent

                anchors.bottom: parent.bottom
                anchors.bottomMargin: 96
                anchors.horizontalCenter: parent.horizontalCenter
                anchors.top: parent.top
                anchors.topMargin: 96
                clip: true
                colorScheme: root.colorScheme
                width: 444
                wizard: root
            }
            Image {
                id: mailLogoWithWordmark
                anchors.bottom: parent.bottom
                anchors.bottomMargin: 48
                anchors.horizontalCenter: parent.horizontalCenter
                fillMode: Image.PreserveAspectFit
                height: 24
                mipmap: true
                source: root.colorScheme.mail_logo_with_wordmark
            }
        }
        Rectangle {
            id: rightHalf
            Layout.fillHeight: true
            Layout.fillWidth: true
            color: root.colorScheme.background_weak

            StackLayout {
                id: rightContent
                anchors.bottom: parent.bottom
                anchors.bottomMargin: 96
                anchors.horizontalCenter: parent.horizontalCenter
                anchors.top: parent.top
                anchors.topMargin: 96
                clip: true
                currentIndex: 0
                width: 444

                // stack index 0
                Onboarding {
                    Layout.fillHeight: true
                    Layout.fillWidth: true
                    colorScheme: root.colorScheme

                    onOnboardingAccepted: root.startLogin()
                }

                // stack index 1
                Login {
                    id: login
                    Layout.fillHeight: true
                    Layout.fillWidth: true
                    colorScheme: root.colorScheme

                    onLoginAbort: {
                        root.closeWizard();
                    }
                }

                // stack index 2
                ClientConfigSelector {
                    id: clientConfigSelector
                    Layout.fillHeight: true
                    Layout.fillWidth: true
                    wizard: root
                }
                // stack index 3
                ClientConfigOutlookSelector {
                    id: clientConfigOutlookSelector
                    Layout.fillHeight: true
                    Layout.fillWidth: true
                    wizard: root
                }
                // stack index 4
                ClientConfigWarning {
                    id: clientConfigWarning
                    Layout.fillHeight: true
                    Layout.fillWidth: true
                    wizard: root
                }
            }
            LinkLabel {
                id: reportProblemLink
                anchors.bottom: parent.bottom
                anchors.bottomMargin: 48
                anchors.horizontalCenter: parent.horizontalCenter
                colorScheme: root.colorScheme
                horizontalAlignment: Text.AlignRight
                text: link("#", qsTr("Report problem"))
                width: 444

                onLinkActivated: {
                    root.visible = false;
                }
            }
        }
    }
}

