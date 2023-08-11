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

    enum RootStack {
        TwoPanesView = 0,
        ClientConfigParameters = 1
    }

    enum ContentStack {
        Onboarding = 0,
        Login = 1,
        ClientConfigSelector = 2,
        ClientConfigOutlookSelector = 3,
        ClientConfigWarning = 4
    }

    property int client
    property string clientVersion
    property ColorScheme colorScheme
    property var user
    property string address

    signal wizardEnded()
    signal showBugReport()

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
        wizardEnded()
    }

    function showOutlookSelector() {
        rootStackLayout.currentIndex = SetupWizard.RootStack.TwoPanesView;
        leftContent.showOutlookSelector();
        rightContent.currentIndex = SetupWizard.ContentStack.ClientConfigOutlookSelector;
    }

    function start() {
        rootStackLayout.currentIndex = SetupWizard.RootStack.TwoPanesView;
        leftContent.showOnboarding();
        rightContent.currentIndex = SetupWizard.ContentStack.Onboarding;
    }

    function startClientConfig(user, address) {
        root.user = user
        root.address = address
        rootStackLayout.currentIndex = SetupWizard.RootStack.TwoPanesView;
        leftContent.showClientSelector();
        rightContent.currentIndex = SetupWizard.ContentStack.ClientConfigSelector;
    }

    function startLogin(username = "") {
        rootStackLayout.currentIndex = SetupWizard.RootStack.TwoPanesView;
        root.address = "";
        leftContent.showLogin();
        rightContent.currentIndex = SetupWizard.ContentStack.Login;
        login.reset(true);
        login.username = username;
    }

    function showClientWarning() {
        rootStackLayout.currentIndex = SetupWizard.RootStack.TwoPanesView;
        leftContent.showClientConfigWarning();
        rightContent.currentIndex = SetupWizard.ContentStack.ClientConfigWarning
    }

    function showClientParams() {
        rootStackLayout.currentIndex = SetupWizard.RootStack.ClientConfigParameters;

    }

    Connections {
        function onLoginFinished(userIndex, wasSignedOut) {
            if (wasSignedOut) {
                closeWizard();
                return;
            }
            let user = Backend.users.get(userIndex)
            let address = user ? user.addresses[0] : ""
            startClientConfig(user, address);
        }

        target: Backend
    }

    StackLayout {
        id: rootStackLayout
        anchors.fill: parent

        // rootStackLayout index 0 SetupWizard.RootStack.TwoPanesView
        RowLayout {
            Layout.fillHeight: true
            Layout.fillWidth: true
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
                        closeWizard();
                        showBugReport();
                    }
                }
            }
        }

        // rootStackLayout index 1 SetupWizard.RootStack.ClientConfigParameters
        ClientConfigParameters {
            id: clientConfigParameters
            Layout.fillHeight: true
            Layout.fillWidth: true
            wizard: root
        }
    }
}

