// Copyright (c) 2023 Proton AG
//
// This file is part of Proton Mail Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

import QtQuick
import QtQuick.Layouts
import QtQuick.Controls

import Proton

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
        actionIcon: "/qml/icons/ic-external-link.svg"
        description: qsTr("Get help setting up your client with our instructions and FAQs.")
        type: SettingsItem.PrimaryButton
        onClicked: {Qt.openUrlExternally("https://proton.me/support/bridge")}

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
            Backend.checkUpdates()
        }

        Connections {
            target: Backend
            function onCheckUpdatesFinished() { checkUpdates.loading = false }
        }

        Layout.fillWidth: true
    }

    SettingsItem {
        id: logs
        colorScheme: root.colorScheme
        text: qsTr("Logs")
        actionText: qsTr("View logs")
        description: qsTr("Open and review logs to troubleshoot.")
        type: SettingsItem.Button
        onClicked: Qt.openUrlExternally(Backend.logsPath)

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
            Backend.updateCurrentMailClient()
            root.parent.showBugReport()
        }

        Layout.fillWidth: true
    }

    // fill height so the footer label will be always attached to the bottom
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

        text: qsTr("%1 v%2 (%3)<br>© 2017-%4 %5<br>%6 %7<br>%8").
        arg(Backend.appname).
        arg(Backend.version).
        arg(Backend.tag).
        arg(Backend.buildYear()).
        arg(Backend.vendor).
        arg(link(Backend.licensePath, qsTr("License"))).
        arg(link(Backend.dependencyLicensesLink, qsTr("Dependencies"))).
        arg(link(Backend.releaseNotesLink, qsTr("Release notes")))

        onLinkActivated: function(link) { Qt.openUrlExternally(link) }
    }
}
