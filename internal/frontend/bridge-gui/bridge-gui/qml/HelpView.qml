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
import QtQuick
import QtQuick.Layouts
import QtQuick.Controls
import Proton

SettingsView {
    id: root
    fillHeight: true

    Label {
        Layout.fillWidth: true
        colorScheme: root.colorScheme
        text: qsTr("Help")
        type: Label.Heading
    }
    SettingsItem {
        id: setupPage
        Layout.fillWidth: true
        actionIcon: "/qml/icons/ic-external-link.svg"
        actionText: qsTr("Go to help topics")
        colorScheme: root.colorScheme
        description: qsTr("Get help setting up your client with our instructions and FAQs.")
        text: qsTr("Installation and setup")
        type: SettingsItem.PrimaryButton

        onClicked: {
            Backend.openExternalLink();
        }
    }
    SettingsItem {
        id: checkUpdates
        Layout.fillWidth: true
        actionText: qsTr("Check now")
        colorScheme: root.colorScheme
        description: qsTr("Check that you're using the latest version of Bridge.\nTo stay up to date, enable auto-updates in settings.")
        text: qsTr("Updates")
        type: SettingsItem.Button

        onClicked: {
            checkUpdates.loading = true;
            Backend.checkUpdates();
        }

        Connections {
            function onCheckUpdatesFinished() {
                checkUpdates.loading = false;
            }

            target: Backend
        }
    }
    SettingsItem {
        id: logs
        Layout.fillWidth: true
        actionText: qsTr("View logs")
        colorScheme: root.colorScheme
        description: qsTr("Open and review logs to troubleshoot.")
        text: qsTr("Logs")
        type: SettingsItem.Button

        onClicked: Backend.openExternalLink(Backend.logsPath)
    }
    SettingsItem {
        id: reportBug
        Layout.fillWidth: true
        actionText: qsTr("Report problem")
        colorScheme: root.colorScheme
        description: qsTr("Something not working as expected? Let us know.")
        text: qsTr("Report a problem")
        type: SettingsItem.Button

        onClicked: {
            Backend.updateCurrentMailClient();
            Backend.notifyReportBugClicked();
            root.parent.showBugReport();
        }
    }

    // fill height so the footer label will always be attached to the bottom
    Item {
        Layout.fillHeight: true
        Layout.fillWidth: true
    }
    Label {
        Layout.alignment: Qt.AlignHCenter
        color: root.colorScheme.text_weak
        colorScheme: root.colorScheme
        horizontalAlignment: Text.AlignHCenter
        text: qsTr("%1 v%2 (%3)<br>© 2017-%4 %5<br>%6 %7<br>%8").arg(Backend.appname).arg(Backend.version).arg(Backend.tag).arg(Backend.buildYear()).arg(Backend.vendor).arg(link(Backend.licensePath, qsTr("License"))).arg(link(Backend.dependencyLicensesLink, qsTr("Dependencies"))).arg(link(Backend.releaseNotesLink, qsTr("Release notes")))
        textFormat: Text.StyledText
        type: Label.Caption

        onLinkActivated: function (link) {
            Backend.openExternalLink(link)
        }
    }
}
