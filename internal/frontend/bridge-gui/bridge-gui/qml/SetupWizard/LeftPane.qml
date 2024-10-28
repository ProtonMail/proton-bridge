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

Item {
    id: root

    readonly property string addAccountTitle: qsTr("Add a Proton Mail account")
    readonly property string welcomeDescription: qsTr("Bridge is the gateway between your Proton account and your email client. It runs in the background and encrypts and decrypts your messages seamlessly. ");
    readonly property string welcomeTitle: qsTr("Welcome to\nProton Mail Bridge")
    readonly property string welcomeImage: "/qml/icons/img-welcome.svg"
    readonly property int welcomeImageHeight: 148;
    readonly property int welcomeImageWidth: 265;

    property int iconHeight
    property string iconSource
    property int iconWidth
    property var wizard
    property ColorScheme colorScheme
    property var _colorScheme: wizard ? wizard.colorScheme : colorScheme

    signal startSetup()

    function showCertificateInstall() {
        showClientConfigCommon();
        if (wizard.client === SetupWizard.Client.AppleMail) {
            descriptionLabel.text = qsTr("Apple Mail configuration is mostly automated, but in order to work, Bridge needs to install a certificate in your keychain.");
            linkLabel1.setCallback(function () {
                Backend.openExternalLink("https://proton.me/support/apple-mail-certificate");
            }, qsTr("Why is this certificate needed?"), true);
        } else {
            descriptionLabel.text = qsTr("In order for Outlook to work, Bridge needs to install a certificate in your keychain.");
            linkLabel1.setCallback(function () {
                Backend.openExternalLink("https://proton.me/support/apple-mail-certificate");
            }, qsTr("Why is this certificate needed?"), true);
        }
        linkLabel2.clear();
    }

    function showClientConfigCommon() {
        titleLabel.text = "";
        linkLabel1.clear();
        linkLabel2.clear();
        iconSource = wizard.clientIconSource();
        iconHeight = 80;
        iconWidth = 80;
    }
    function showAppleMailAutoconfigProfileInstall() {
        showClientConfigCommon();
        descriptionLabel.text = qsTr("The final step before you can start using Apple Mail is to install the Bridge server profile in the system preferences.\n\nAdding a server profile is necessary to ensure that your Mac can receive and send Proton Mail messages.");
        linkLabel1.setCallback(function() { Backend.openExternalLink("https://proton.me/support/macos-certificate-warning"); }, qsTr("Why is there a yellow warning sign?"), true);
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

    function showNoAccount() {
        titleLabel.text = welcomeTitle;
        descriptionLabel.text = welcomeDescription;
        linkLabel1.setCallback(startSetup, "Start setup", false);
        linkLabel2.clear();
        root.iconSource = welcomeImage;
        root.iconHeight = welcomeImageHeight;
        root.iconWidth = welcomeImageWidth;
    }

    function showOnboarding() {
        titleLabel.text = (Backend.users.count === 0) ? welcomeTitle : addAccountTitle;
        descriptionLabel.text = welcomeDescription
        linkLabel1.setCallback(function() { Backend.openExternalLink("https://proton.me/support/why-you-need-bridge"); }, qsTr("Why do I need Bridge?"), true);
        linkLabel2.clear();
        root.iconSource = welcomeImage;
        root.iconHeight = welcomeImageHeight;
        root.iconWidth = welcomeImageWidth;
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
            colorScheme: _colorScheme
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
            colorScheme: _colorScheme
            horizontalAlignment: Text.AlignHCenter
            text: ""
            type: Label.LabelType.Body
            wrapMode: Text.WordWrap
        }
        LinkLabel {
            id: linkLabel1
            Layout.alignment: Qt.AlignHCenter
            colorScheme: _colorScheme
            visible: (text !== "")
        }
        LinkLabel {
            id: linkLabel2
            Layout.alignment: Qt.AlignHCenter
            colorScheme: _colorScheme
            visible: (text !== "")
        }
    }
}
