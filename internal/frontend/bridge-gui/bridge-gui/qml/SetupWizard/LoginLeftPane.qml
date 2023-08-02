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

Item {
    id: root

    property ColorScheme colorScheme
    property string description: ""
    property string helpLink: ""

    function show2FA() {
        root.description = qsTr("You have enabled two-factor authentication. Please enter the 6-digit code provided by your authenticator application.");
        root.helpLink = "";
    }
    function showMailboxPassword() {
        root.description = qsTr("You have secured your account with a separate mailbox password. ");
        root.helpLink = "";
    }
    function showSignIn() {
        root.description = qsTr("Let's start by signing in to your Proton account.");
        root.helpLink = linkLabel.link("https://proton.me/mail/pricing", qsTr("Create or upgrade your account"));
    }

    Connections {
        function onLogin2FARequested() {
            show2FA();
        }
        function onLogin2PasswordRequested() {
            showMailboxPassword();
        }

        target: Backend
    }
    ColumnLayout {
        anchors.left: parent.left
        anchors.right: parent.right
        anchors.top: parent.top
        spacing: 0

        Image {
            Layout.alignment: Qt.AlignHCenter | Qt.AlignTop
            Layout.preferredHeight: 72
            Layout.preferredWidth: 72
            fillMode: Image.PreserveAspectFit
            source: "/qml/icons/ic-bridge.svg"
            sourceSize.height: 128
            sourceSize.width: 128
        }
        Label {
            Layout.alignment: Qt.AlignHCenter
            Layout.fillWidth: true
            Layout.topMargin: 16
            colorScheme: root.colorScheme
            horizontalAlignment: Text.AlignHCenter
            text: qsTr("Sign in to your Proton Account")
            type: Label.LabelType.Heading
            wrapMode: Text.WordWrap
        }
        Label {
            Layout.alignment: Qt.AlignHCenter
            Layout.fillWidth: true
            Layout.topMargin: 96
            colorScheme: root.colorScheme
            horizontalAlignment: Text.AlignHCenter
            text: description
            type: Label.LabelType.Body
            wrapMode: Text.WordWrap
        }
        Label {
            id: linkLabel
            Layout.alignment: Qt.AlignHCenter
            Layout.fillWidth: false
            Layout.topMargin: 96
            colorScheme: root.colorScheme
            horizontalAlignment: Text.AlignHCenter
            text: root.helpLink
            type: Label.LabelType.Body
            visible: {
                root.helpLink !== "";
            }

            onLinkActivated: function (link) {
                Qt.openUrlExternally(link);
            }

            HoverHandler {
                acceptedDevices: PointerDevice.Mouse
                cursorShape: Qt.PointingHandCursor
            }
        }
    }
}