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
    enum ContentStack {
        Onboarding,
        Login,
        ClientConfigSelector,
        ClientConfigAppleMail
    }
    enum RootStack {
        TwoPanesView,
        ClientConfigParameters,
        ClientConfigEnd
    }

    property string address
    property int client
    property string clientVersion
    property ColorScheme colorScheme
    property var user

    signal showBugReport
    signal wizardEnded

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
            console.error("Unknown mail client " + client);
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
            console.error("Unknown mail client " + client);
            return "your email client";
        }
    }
    function closeWizard() {
        wizardEnded();
    }
    function showAppleMailAutoConfig() {
        rootStackLayout.currentIndex = SetupWizard.RootStack.TwoPanesView;
        leftContent.showClientSelector();
        rightContent.currentIndex = SetupWizard.ContentStack.ClientConfigAppleMail;
        clientConfigAppleMail.showAutoConfig();
    }
    function showClientConfig(user, address) {
        root.user = user;
        root.address = address;
        rootStackLayout.currentIndex = SetupWizard.RootStack.TwoPanesView;
        leftContent.showClientSelector();
        rightContent.currentIndex = SetupWizard.ContentStack.ClientConfigSelector;
    }
    function showClientParams() {
        rootStackLayout.currentIndex = SetupWizard.RootStack.ClientConfigParameters;
    }
    function showClientConfigEnd() {
        rootStackLayout.currentIndex = SetupWizard.RootStack.ClientConfigEnd;
    }
    function showLogin(username = "") {
        rootStackLayout.currentIndex = SetupWizard.RootStack.TwoPanesView;
        root.address = "";
        leftContent.showLogin();
        rightContent.currentIndex = SetupWizard.ContentStack.Login;
        login.reset(true);
        login.username = username;
    }
    function showOnboarding() {
        rootStackLayout.currentIndex = SetupWizard.RootStack.TwoPanesView;
        root.address = ""
        root.user = null
        leftContent.showOnboarding();
        rightContent.currentIndex = SetupWizard.ContentStack.Onboarding;
    }

    Connections {
        function onLoginFinished(userIndex, wasSignedOut) {
            if (wasSignedOut) {
                closeWizard();
                return;
            }
            let user = Backend.users.get(userIndex);
            let address = user ? user.addresses[0] : "";
            showClientConfig(user, address);
        }

        target: Backend
    }
    StackLayout {
        id: rootStackLayout
        anchors.fill: parent

        // rootStackLayout index 0
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
                    anchors.bottomMargin: 92
                    anchors.horizontalCenter: parent.horizontalCenter
                    anchors.top: parent.top
                    anchors.topMargin: 40
                    clip: true
                    width: 364
                    wizard: root
                }
                Image {
                    id: mailLogoWithWordmark
                    anchors.bottom: parent.bottom
                    anchors.bottomMargin: 40
                    anchors.horizontalCenter: parent.horizontalCenter
                    height: 36
                    width: 134
                    sourceSize.height: 36
                    sourceSize.width: 134
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
                    anchors.bottomMargin: 92
                    anchors.horizontalCenter: parent.horizontalCenter
                    anchors.top: parent.top
                    anchors.topMargin: 40
                    clip: true
                    currentIndex: 0
                    width: 364

                    // rightContent stack index 0
                    Onboarding {
                        wizard: root
                    }

                    // rightContent tack index 1
                    Login {
                        id: login
                        wizard: root

                        onLoginAbort: {
                            root.closeWizard();
                        }
                    }

                    // rightContent stack index 2
                    ClientConfigSelector {
                        id: clientConfigSelector
                        wizard: root
                    }
                    // rightContent stack index 3
                    ClientConfigAppleMail {
                        id: clientConfigAppleMail
                        wizard: root
                    }
                }
            }
        }

        // rootStackLayout index 1
        ClientConfigParameters {
            id: clientConfigParameters
            Layout.fillHeight: true
            Layout.fillWidth: true
            wizard: root
        }

        // rootStackLayout index 2
        ClientConfigEnd {
            id: clientConfigEnd
            Layout.fillHeight: true
            Layout.fillWidth: true
            wizard: root
        }
    }
    HelpButton {
        wizard: root
    }
}

