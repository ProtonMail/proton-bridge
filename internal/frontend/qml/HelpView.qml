// Copyright (c) 2022 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

import QtQuick 2.13
import QtQuick.Layouts 1.12
import QtQuick.Controls 2.12

import Proton 4.0

SettingsView {
    id: root

    fillHeight: true

    Label {
        colorScheme: root.colorScheme
        text: qsTr("Help")
        type: Label.Heading
        Layout.fillWidth: true
    }

    SettingsItem {
        id: setupPage
        colorScheme: root.colorScheme
        text: qsTr("Installation and setup")
        actionText: qsTr("Go to help topics")
        actionIcon: "./icons/ic-external-link.svg"
        description: qsTr("Get help setting up your client with our instructions and FAQs.")
        type: SettingsItem.PrimaryButton
        onClicked: {Qt.openUrlExternally("https://protonmail.com/support/categories/bridge/")}

        Layout.fillWidth: true
    }

    SettingsItem {
        id: checkUpdates
        colorScheme: root.colorScheme
        text: qsTr("Updates")
        actionText: qsTr("Check now")
        description: qsTr("Check that you're using the latest version of Bridge. To stay up to date, enable auto-updates in settings.")
        type: SettingsItem.Button
        onClicked: {
            checkUpdates.loading = true
            root.backend.checkUpdates()
        }

        Connections {target: root.backend; onCheckUpdatesFinished: checkUpdates.loading = false}

        Layout.fillWidth: true
    }

    SettingsItem {
        id: logs
        colorScheme: root.colorScheme
        text: qsTr("Logs")
        actionText: qsTr("View logs")
        description: qsTr("Open and review logs to troubleshoot.")
        type: SettingsItem.Button
        onClicked: Qt.openUrlExternally(root.backend.logsPath)

        Layout.fillWidth: true
    }

    SettingsItem {
        id: reportBug
        colorScheme: root.colorScheme
        text: qsTr("Report a problem")
        actionText: qsTr("Report a problem")
        description: qsTr("Something not working as expected? Let us know.")
        type: SettingsItem.Button
        onClicked: {
            root.backend.updateCurrentMailClient()
            root.parent.showBugReport()
        }

        Layout.fillWidth: true
    }

    // fill height so the footer label will be allways attached to the bottom
    Item {
        Layout.fillHeight: true
        Layout.fillWidth: true
    }

    Label {
        Layout.alignment: Qt.AlignHCenter
        colorScheme: root.colorScheme
        type: Label.Caption
        color: root.colorScheme.text_weak
        textFormat: Text.StyledText

        horizontalAlignment: Text.AlignHCenter

        text: qsTr("Proton Mail Bridge v%1<br>© 2022 Proton AG<br>%2 %3").
            arg(root.backend.version).
            arg(link(root.backend.licensePath, qsTr("License"))).
            arg(link(root.backend.releaseNotesLink, qsTr("Release notes")))

        onLinkActivated: Qt.openUrlExternally(link)
    }
}
