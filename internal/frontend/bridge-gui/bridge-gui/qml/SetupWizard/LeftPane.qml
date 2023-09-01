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

    property int iconHeight
    property string iconSource
    property int iconWidth
    property var wizard

    function showAppleMailAutoconfigCertificateInstall() {
        showAppleMailAutoconfigCommon();
        descriptionLabel.text = qsTr("Apple Mail configuration is mostly automated, but in order to work, Bridge needs to install a certificate in your keychain.");
        linkLabel1.setCallback(function() { Backend.showHelpOverlay("WhyCertificate.html"); }, qsTr("Why is this certificate needed?"), false);
    }
    function showAppleMailAutoconfigCommon() {
        titleLabel.text = "";
        linkLabel1.clear();
        linkLabel2.clear();
        iconSource = wizard.clientIconSource();
        iconHeight = 80;
        iconWidth = 80;
    }
    function showAppleMailAutoconfigProfileInstall() {
        showAppleMailAutoconfigCommon();
        descriptionLabel.text = qsTr("The final step before you can start using Apple Mail is to install the Bridge server profile in the system preferences.\n\nAdding a server profile is necessary to ensure that your Mac can receive and send Proton Mails.");
        linkLabel1.setCallback(function() { Backend.showHelpOverlay("WhyProfileWarning.html"); }, qsTr("Why is there a yellow warning sign?"), false);
        linkLabel2.setCallback(wizard.showClientParams, qsTr("Configure Apple Mail manually"), false);
    }
    function showClientSelector(newAccount = true) {
        titleLabel.text = "";
        descriptionLabel.text = newAccount ? qsTr("Bridge is now connected to Proton, and has already started downloading your messages. Let’s now connect your email client to Bridge.") : qsTr("Let’s connect your email client to Bridge.");
        linkLabel1.clear();
        linkLabel2.clear();
        iconSource = "/qml/icons/img-client-config-selector.svg";
        iconHeight = 104;
        iconWidth = 266;
    }
    function showLogin() {
        showOnboarding();
    }
    function showLogin2FA() {
        showOnboarding();
    }
    function showLoginMailboxPassword() {
        showOnboarding();
    }
    function showOnboarding() {
        titleLabel.text = (Backend.users.count === 0) ? qsTr("Welcome to\nProton Mail Bridge") : qsTr("Add a Proton Mail account");
        descriptionLabel.text = qsTr("Bridge is the gateway between your Proton account and your email client. It runs in the background and encrypts and decrypts your messages seamlessly. ");
        linkLabel1.setCallback(function() { Backend.showHelpOverlay("WhyBridge.html"); }, qsTr("Why do I need Bridge?"), false);
        linkLabel2.clear();
        root.iconSource = "/qml/icons/img-welcome.svg";
        root.iconHeight = 148;
        root.iconWidth = 265;
    }

    Connections {
        function onLogin2FARequested() {
            showLogin2FA();
        }
        function onLogin2PasswordRequested() {
            showLoginMailboxPassword();
        }

        target: Backend
    }
    ColumnLayout {
        anchors.left: parent.left
        anchors.right: parent.right
        anchors.verticalCenter: parent.verticalCenter
        spacing: ProtonStyle.wizard_spacing_medium

        Image {
            id: icon
            Layout.alignment: Qt.AlignHCenter | Qt.AlignTop
            Layout.preferredHeight: root.iconHeight
            Layout.preferredWidth: root.iconWidth
            source: root.iconSource
            sourceSize.height: root.iconHeight
            sourceSize.width: root.iconWidth
        }
        Label {
            id: titleLabel
            Layout.alignment: Qt.AlignHCenter
            Layout.fillWidth: true
            colorScheme: wizard.colorScheme
            horizontalAlignment: Text.AlignHCenter
            text: ""
            type: Label.LabelType.Heading
            visible: text.length !== 0
            wrapMode: Text.WordWrap
        }
        Label {
            id: descriptionLabel
            Layout.alignment: Qt.AlignHCenter
            Layout.fillWidth: true
            colorScheme: wizard.colorScheme
            horizontalAlignment: Text.AlignHCenter
            text: ""
            type: Label.LabelType.Body
            wrapMode: Text.WordWrap
        }
        LinkLabel {
            id: linkLabel1
            Layout.alignment: Qt.AlignHCenter
            colorScheme: wizard.colorScheme
            visible: (text !== "")
        }
        LinkLabel {
            id: linkLabel2
            Layout.alignment: Qt.AlignHCenter
            colorScheme: wizard.colorScheme
            visible: (text !== "")
        }
    }
}
