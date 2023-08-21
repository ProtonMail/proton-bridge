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

    property var wizard

    function showClientConfigCommon() {
        const clientName = wizard.clientName();
        titleLabel.text = qsTr("Configure %1").arg(clientName);
        descriptionLabel.text = qsTr("We will now guide you through the process of setting up your Proton account in %1.").arg(clientName);
        icon.source = wizard.clientIconSource();
        icon.sourceSize.height = 128;
        icon.sourceSize.width = 128;
        Layout.preferredHeight = 72;
        Layout.preferredWidth = 72;
    }
    function showClientConfigWarning() {
        showClientConfigCommon();
        linkLabel1.setLink("https://proton.me/support/bridge", qsTr("Why can't I use my Proton password in my email client?"));
    }
    function showClientSelector() {
        titleLabel.text = qsTr("Configure your email client");
        descriptionLabel.text = qsTr("Bridge is now connected to Proton, and has already started downloading your messages. Let’s now connect your email client to Bridge.");
        linkLabel1.clear();
        linkLabel2.clear();
        icon.source = "/qml/icons/img-mail-clients.svg";
    }
    function showLogin() {
        showOnboarding()
     }
    function showLogin2FA() {
        showOnboarding()
    }
    function showLoginMailboxPassword() {
        showOnboarding()
    }
    function showOnboarding() {
        titleLabel.text = (Backend.users.count === 0) ? qsTr("Welcome to\nProton Mail Bridge") : qsTr("Add a Proton Mail account");
        descriptionLabel.text = qsTr("Bridge is the gateway between your Proton account and your email client. It runs in the background and encrypts and decrypts your messages seamlessly. ");
        linkLabel1.setLink("https://proton.me/support/bridge", qsTr("Why do I need Bridge?"));
        linkLabel2.clear();
        icon.Layout.preferredHeight = 148;
        icon.Layout.preferredWidth = 265;
        icon.source = "/qml/icons/img-welcome.svg";
        icon.sourceSize.height = 148;
        icon.sourceSize.width = 265;
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
        spacing: 16

        Image {
            id: icon
            Layout.alignment: Qt.AlignHCenter | Qt.AlignTop
            Layout.preferredHeight: 72
            Layout.preferredWidth: 72
            fillMode: Image.PreserveAspectFit
            source: ""
            sourceSize.height: 72
            sourceSize.width: 72
        }
        Label {
            id: titleLabel
            Layout.alignment: Qt.AlignHCenter
            Layout.fillWidth: true
            colorScheme: wizard.colorScheme
            horizontalAlignment: Text.AlignHCenter
            text: ""
            type: Label.LabelType.Heading
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
