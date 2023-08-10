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
import ".."

Rectangle {
    id: root
    property var wizard
    property ColorScheme colorScheme: wizard.colorScheme
    color: colorScheme.background_weak
    readonly property bool genericClient: SetupWizard.Client.Generic === wizard.client

    Item {
        id: centeredContainer
        anchors.horizontalCenter: parent.horizontalCenter
        anchors.top: parent.top
        anchors.bottom: parent.bottom
        width: 800

        ColumnLayout {
            anchors.left: parent.left
            anchors.right: parent.right
            anchors.top: parent.top
            anchors.bottomMargin: 96
            anchors.topMargin: 32
            spacing: 0
            Label {
                Layout.alignment: Qt.AlignHCenter
                Layout.fillWidth: true
                colorScheme: root.colorScheme
                horizontalAlignment: Text.AlignHCenter
                text: qsTr("Configure %1").arg(wizard.clientName())
                type: Label.LabelType.Heading
                wrapMode: Text.WordWrap
            }
            Label {
                id: descriptionLabel
                Layout.alignment: Qt.AlignHCenter
                Layout.fillWidth: true
                Layout.topMargin: 8
                color: colorScheme.text_weak
                colorScheme: root.colorScheme
                horizontalAlignment: Text.AlignHCenter
                text: genericClient ? qsTr("Here are the IMAP and SMTP configuration parameters for your email client") :
                    qsTr("Here are your email configuration parameters for %1. \nWe have prepared an easy to follow configuration guide to help you setup your account in %1.").arg(wizard.clientName())
                type: Label.LabelType.Body
                wrapMode: Text.WordWrap
            }
            RowLayout {
                id: configuration

                Layout.fillHeight: true
                Layout.fillWidth: true
                Layout.topMargin: 32
                spacing: 64
                Configuration {
                    Layout.fillWidth: true
                    colorScheme: root.colorScheme
                    hostname: Backend.hostname
                    password: wizard.user ? wizard.user.password : ""
                    port: Backend.imapPort.toString()
                    security: Backend.useSSLForIMAP ? "SSL" : "STARTTLS"
                    title: qsTr("IMAP")
                    username: wizard.address
                }
                Configuration {
                    Layout.fillWidth: true
                    colorScheme: root.colorScheme
                    hostname: Backend.hostname
                    password: wizard.user ? wizard.user.password : ""
                    port: Backend.smtpPort.toString()
                    security: Backend.useSSLForSMTP ? "SSL" : "STARTTLS"
                    title: qsTr("SMTP")
                    username: wizard.address
                }
            }

            Button {
                Layout.alignment: Qt.AlignHCenter
                Layout.preferredWidth: 444
                Layout.topMargin: 32
                colorScheme: root.colorScheme
                text: qsTr("Open configuration guide")
                visible: !genericClient
            }

            Button {
                Layout.alignment: Qt.AlignHCenter
                Layout.preferredWidth: 444
                Layout.topMargin: 32
                colorScheme: root.colorScheme
                text: qsTr("Done")
                onClicked: wizard.closeWizard()
            }
        }

        LinkLabel {
            id: reportProblemLink
            anchors.bottom: parent.bottom
            anchors.bottomMargin: 48
            anchors.right: parent.right
            colorScheme: root.colorScheme
            horizontalAlignment: Text.AlignRight
            text: link("#", qsTr("Report problem"))

            onLinkActivated: {
                wizard.closeWizard();
            }
        }
    }
}
