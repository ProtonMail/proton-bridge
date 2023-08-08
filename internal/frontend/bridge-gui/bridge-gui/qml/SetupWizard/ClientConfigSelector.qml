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

Item {
    id: root

    property ColorScheme colorScheme: wizard.colorScheme
    readonly property bool onMacOS: (Backend.goos === "darwin")
    readonly property bool onWindows: (Backend.goos === "windows")
    property var wizard

    ColumnLayout {
        anchors.left: parent.left
        anchors.right: parent.right
        anchors.top: parent.top
        spacing: 0

        Label {
            Layout.alignment: Qt.AlignHCenter
            Layout.fillWidth: true
            colorScheme: root.colorScheme
            text: qsTr("Select your email application")
            type: Label.LabelType.Heading
        }
        Item {
            Layout.preferredHeight: 72
        }
        ClientListItem {
            Layout.fillWidth: true
            colorScheme: root.colorScheme
            iconSource: "/qml/icons/ic-apple-mail.svg"
            text: "Apple Mail"
            visible: root.onMacOS

            onClicked: {
                wizard.client = SetupWizard.Client.AppleMail;
            }
        }
        ClientListItem {
            Layout.fillWidth: true
            colorScheme: root.colorScheme
            iconSource: "/qml/icons/ic-microsoft-outlook.svg"
            text: "Microsoft Outlook"
            visible: root.onMacOS || root.onWindows

            onClicked: {
                wizard.client = SetupWizard.Client.MicrosoftOutlook;
                wizard.showOutlookSelector();
            }
        }
        ClientListItem {
            Layout.fillWidth: true
            colorScheme: root.colorScheme
            iconSource: "/qml/icons/ic-mozilla-thunderbird.svg"
            text: "Mozilla Thunderbird"

            onClicked: {
                wizard.client = SetupWizard.Client.MozillaThunderbird;
                wizard.showClientWarning();
            }
        }
        ClientListItem {
            Layout.fillWidth: true
            colorScheme: root.colorScheme
            iconSource: "/qml/icons/ic-other-mail-clients.svg"
            text: "Other"

            onClicked: {
                wizard.client = SetupWizard.Client.Generic;
                wizard.showClientWarning();
            }
        }
        Item {
            Layout.preferredHeight: 72
        }
        Button {
            Layout.fillWidth: true
            colorScheme: root.colorScheme
            secondary: true
            text: qsTr("Cancel")

            onClicked: {
                root.wizard.closeWizard();
            }
        }
    }
}

