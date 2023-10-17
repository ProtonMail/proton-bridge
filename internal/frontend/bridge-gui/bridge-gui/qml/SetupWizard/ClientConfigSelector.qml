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

    readonly property bool onMacOS: (Backend.goos === "darwin")
    readonly property bool onWindows: (Backend.goos === "windows")
    property var wizard

    ColumnLayout {
        anchors.left: parent.left
        anchors.right: parent.right
        anchors.verticalCenter: parent.verticalCenter
        spacing: ProtonStyle.wizard_spacing_medium

        Label {
            Layout.alignment: Qt.AlignHCenter
            Layout.fillWidth: true
            Layout.topMargin: ProtonStyle.wizard_spacing_medium
            colorScheme: wizard.colorScheme
            horizontalAlignment: Qt.AlignHCenter
            text: qsTr("Select your email client")
            type: Label.LabelType.Title
            wrapMode: Text.WordWrap
        }
        ClientListItem {
            Layout.fillWidth: true
            colorScheme: wizard.colorScheme
            iconSource: "/qml/icons/ic-apple-mail.svg"
            text: "Apple Mail"
            visible: root.onMacOS

            onClicked: {
                wizard.client = SetupWizard.Client.AppleMail;
                wizard.showAppleMailAutoConfig();
            }
        }
        ClientListItem {
            Layout.fillWidth: true
            colorScheme: wizard.colorScheme
            iconSource: "/qml/icons/ic-microsoft-outlook.svg"
            text: "Microsoft Outlook"
            visible: root.onMacOS || root.onWindows

            onClicked: {
                wizard.client = SetupWizard.Client.MicrosoftOutlook;
                wizard.showClientParams();
            }
        }
        ClientListItem {
            Layout.fillWidth: true
            colorScheme: wizard.colorScheme
            iconSource: "/qml/icons/ic-mozilla-thunderbird.svg"
            text: "Mozilla Thunderbird"

            onClicked: {
                wizard.client = SetupWizard.Client.MozillaThunderbird;
                wizard.showClientParams();
            }
        }
        ClientListItem {
            Layout.fillWidth: true
            colorScheme: wizard.colorScheme
            iconSource: "/qml/icons/ic-other-mail-clients.svg"
            text: qsTr("Other")

            onClicked: {
                wizard.client = SetupWizard.Client.Generic;
                wizard.showClientParams();
            }
        }
        Button {
            Layout.fillWidth: true
            Layout.topMargin: 20
            colorScheme: wizard.colorScheme
            secondary: true
            secondaryIsOpaque: true
            text: qsTr("Setup later")

            onClicked: {
                root.wizard.closeWizard();
            }
        }
    }
}

