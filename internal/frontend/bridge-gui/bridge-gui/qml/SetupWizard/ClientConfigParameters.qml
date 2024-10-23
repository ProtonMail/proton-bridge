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
import QtQuick.Controls.impl
import ".."

Rectangle {
    id: root

    property ColorScheme colorScheme: wizard.colorScheme
    readonly property bool genericClient: SetupWizard.Client.Generic === wizard.client
    property var wizard

    clip: true
    color: colorScheme.background_weak

    Item {
        id: centeredContainer
        anchors.bottom: parent.bottom
        anchors.horizontalCenter: parent.horizontalCenter
        anchors.top: parent.top
        width: 640

        ColumnLayout {
            anchors.left: parent.left
            anchors.right: parent.right
            anchors.verticalCenter: parent.verticalCenter
            spacing: ProtonStyle.wizard_spacing_medium

            Label {
                Layout.alignment: Qt.AlignHCenter
                Layout.fillWidth: true
                colorScheme: wizard.colorScheme
                horizontalAlignment: Text.AlignHCenter
                text: qsTr("Configure %1").arg(wizard.clientName())
                type: Label.LabelType.Title
                wrapMode: Text.WordWrap
            }
            Rectangle {
                Layout.fillWidth: true
                border.color: colorScheme.border_norm
                border.width: 1
                color: "transparent"
                height: childrenRect.height + 2 * ProtonStyle.wizard_spacing_medium
                radius: 12

                RowLayout {
                    anchors.left: parent.left
                    anchors.margins: ProtonStyle.wizard_spacing_medium
                    anchors.right: parent.right
                    anchors.top: parent.top
                    spacing: ProtonStyle.wizard_spacing_small

                    Label {
                        Layout.fillHeight: true
                        Layout.fillWidth: true
                        colorScheme: wizard.colorScheme
                        horizontalAlignment: Text.AlignLeft
                        text: (SetupWizard.Client.MicrosoftOutlook === wizard.client) ? qsTr("Are you unsure about your Outlook version or do you need assistance in configuring Outlook?") : qsTr("Do you need assistance in configuring %1?".arg(wizard.clientName()))
                        type: Label.LabelType.Body
                        verticalAlignment: Text.AlignVCenter
                        wrapMode: Text.WordWrap
                    }
                    Button {
                        colorScheme: root.colorScheme
                        icon.source: "/qml/icons/ic-external-link.svg"
                        text: qsTr("Open guide")

                        onClicked: function () {
                            Backend.openExternalLink(wizard.setupGuideLink());
                        }
                    }
                }
            }
            Rectangle {
                Layout.fillWidth: true
                border.color: colorScheme.signal_warning
                border.width: 1
                color: "transparent"
                height: childrenRect.height + 2 * ProtonStyle.wizard_spacing_medium
                radius: ProtonStyle.banner_radius

                RowLayout {
                    anchors.left: parent.left
                    anchors.margins: ProtonStyle.wizard_spacing_medium
                    anchors.right: parent.right
                    anchors.top: parent.top
                    spacing: ProtonStyle.wizard_spacing_medium

                    ColorImage {
                        id: image
                        height: 36
                        source: "/qml/icons/ic-warning-orange.svg"
                        sourceSize.height: height
                        sourceSize.width: width
                        width: height
                    }
                    Label {
                        Layout.fillHeight: true
                        Layout.fillWidth: true
                        colorScheme: wizard.colorScheme
                        horizontalAlignment: Text.AlignLeft
                        text: qsTr("Copy paste the provided configuration parameters. Use the password below (not your Proton password), when adding your Proton account to %1.".arg(wizard.clientName()))
                        type: Label.LabelType.Body
                        verticalAlignment: Text.AlignVCenter
                        wrapMode: Text.WordWrap
                    }
                }
            }
            RowLayout {
                id: configuration
                Layout.fillWidth: true
                spacing: ProtonStyle.wizard_spacing_extra_large

                Configuration {
                    Layout.fillWidth: true
                    colorScheme: wizard.colorScheme
                    highlightPassword: true
                    hostname: Backend.hostname
                    password: wizard.user ? wizard.user.password : ""
                    port: Backend.imapPort.toString()
                    security: Backend.useSSLForIMAP ? "SSL" : "STARTTLS"
                    title: "IMAP"
                    username: wizard.address
                }
                Configuration {
                    Layout.fillWidth: true
                    colorScheme: wizard.colorScheme
                    highlightPassword: true
                    hostname: Backend.hostname
                    password: wizard.user ? wizard.user.password : ""
                    port: Backend.smtpPort.toString()
                    security: Backend.useSSLForSMTP ? "SSL" : "STARTTLS"
                    title: "SMTP"
                    username: wizard.address
                }
            }
            Button {
                Layout.alignment: Qt.AlignHCenter
                Layout.preferredWidth: 304
                colorScheme: root.colorScheme
                secondary: true
                secondaryIsOpaque: true
                text: qsTr("Continue")

                onClicked: wizard.showClientConfigEnd()
            }
        }
    }
}

