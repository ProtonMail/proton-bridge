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
import QtQuick
import QtQuick.Layouts
import QtQuick.Controls
import QtQuick.Controls.impl
import Proton

SettingsView {
    id: root

    property bool _isAdvancedShown: false
    property var notifications
    property var allUsersLoaded: false
    property var hasInternetConnection: true

    fillHeight: false

    Label {
        Layout.fillWidth: true
        colorScheme: root.colorScheme
        text: qsTr("Settings")
        type: Label.Heading
    }
    SettingsItem {
        id: autoUpdate
        Layout.fillWidth: true
        checked: Backend.isAutomaticUpdateOn
        colorScheme: root.colorScheme
        description: qsTr("Bridge will automatically update in the background.")
        text: qsTr("Automatic updates")
        type: SettingsItem.Toggle

        onClicked: Backend.toggleAutomaticUpdate(!autoUpdate.checked)
    }
    SettingsItem {
        id: autostart
        Layout.fillWidth: true
        checked: Backend.isAutostartOn
        colorScheme: root.colorScheme
        description: qsTr("Bridge will open upon startup.")
        text: qsTr("Open on startup")
        type: SettingsItem.Toggle

        onClicked: {
            autostart.loading = true;
            Backend.toggleAutostart(!autostart.checked);
        }

        Connections {
            function onToggleAutostartFinished() {
                autostart.loading = false;
            }

            target: Backend
        }
    }
    SettingsItem {
        id: beta
        Layout.fillWidth: true
        checked: Backend.isBetaEnabled
        colorScheme: root.colorScheme
        description: qsTr("Be among the first to try new features.")
        text: qsTr("Beta access")
        type: SettingsItem.Toggle

        onClicked: {
            if (!beta.checked) {
                root.notifications.askEnableBeta();
            } else {
                Backend.toggleBeta(false);
            }
        }
    }
    RowLayout {
        ColorImage {
            Layout.alignment: Qt.AlignCenter
            color: root.colorScheme.interaction_norm
            height: root.colorScheme.body_font_size
            source: root._isAdvancedShown ? "/qml/icons/ic-chevron-down.svg" : "/qml/icons/ic-chevron-right.svg"
            sourceSize.height: root.colorScheme.body_font_size

            MouseArea {
                anchors.fill: parent

                onClicked: root._isAdvancedShown = !root._isAdvancedShown
            }
        }
        Label {
            id: advSettLabel
            color: root.colorScheme.interaction_norm
            colorScheme: root.colorScheme
            text: qsTr("Advanced settings")
            type: Label.Body

            MouseArea {
                anchors.fill: parent

                onClicked: root._isAdvancedShown = !root._isAdvancedShown
            }
        }
    }
    SettingsItem {
        id: keychains
        Layout.fillWidth: true
        actionText: qsTr("Change")
        checked: Backend.isDoHEnabled
        colorScheme: root.colorScheme
        description: qsTr("Change which keychain Bridge uses as default")
        text: qsTr("Change keychain")
        type: SettingsItem.Button
        visible: root._isAdvancedShown && Backend.availableKeychain.length > 1

        onClicked: root.parent.showKeychainSettings()
    }
    SettingsItem {
        id: doh
        Layout.fillWidth: true
        checked: Backend.isDoHEnabled
        colorScheme: root.colorScheme
        description: qsTr("If Proton’s servers are blocked in your location, alternative network routing will be used to reach Proton.")
        text: qsTr("Alternative routing")
        type: SettingsItem.Toggle
        visible: root._isAdvancedShown

        onClicked: Backend.toggleDoH(!doh.checked)
    }
    SettingsItem {
        id: darkMode
        Layout.fillWidth: true
        checked: Backend.colorSchemeName === "dark"
        colorScheme: root.colorScheme
        description: qsTr("Choose dark color theme.")
        text: qsTr("Dark mode")
        type: SettingsItem.Toggle
        visible: root._isAdvancedShown

        onClicked: Backend.changeColorScheme(darkMode.checked ? "light" : "dark")
    }
    SettingsItem {
        id: allMail
        Layout.fillWidth: true
        checked: Backend.isAllMailVisible
        colorScheme: root.colorScheme
        description: qsTr("Choose to list the All Mail folder in your local client.")
        text: qsTr("Show All Mail")
        type: SettingsItem.Toggle
        visible: root._isAdvancedShown

        onClicked: root.notifications.askChangeAllMailVisibility(Backend.isAllMailVisible)
    }
    SettingsItem {
        id: telemetry
        Layout.fillWidth: true
        checked: !Backend.isTelemetryDisabled
        colorScheme: root.colorScheme
        description: qsTr("Help us improve Proton services by sending anonymous usage statistics.")
        text: qsTr("Collect usage diagnostics")
        type: SettingsItem.Toggle
        visible: root._isAdvancedShown

        onClicked: Backend.toggleIsTelemetryDisabled(telemetry.checked)
    }
    SettingsItem {
        id: ports
        Layout.fillWidth: true
        actionText: qsTr("Change")
        colorScheme: root.colorScheme
        description: qsTr("Choose which ports are used by default.")
        text: qsTr("Default ports")
        type: SettingsItem.Button
        visible: root._isAdvancedShown

        onClicked: root.parent.showPortSettings()
    }
    SettingsItem {
        id: imap
        Layout.fillWidth: true
        actionText: qsTr("Change")
        colorScheme: root.colorScheme
        description: qsTr("Change the protocol Bridge and the email client use to connect for IMAP and SMTP.")
        text: qsTr("Connection mode")
        type: SettingsItem.Button
        visible: root._isAdvancedShown

        onClicked: root.parent.showConnectionModeSettings()
    }
    SettingsItem {
        id: cache
        Layout.fillWidth: true
        actionText: qsTr("Configure")
        colorScheme: root.colorScheme
        description: qsTr("Configure Bridge's local cache.")
        text: qsTr("Local cache")
        type: SettingsItem.Button
        visible: root._isAdvancedShown

        onClicked: root.parent.showLocalCacheSettings()
    }
    SettingsItem {
        id: exportTLSCertificates
        Layout.fillWidth: true
        actionText: qsTr("Export")
        colorScheme: root.colorScheme
        description: qsTr("Export the TLS private key and certificate used by the IMAP and SMTP servers.")
        text: qsTr("Export TLS certificates")
        type: SettingsItem.Button
        visible: root._isAdvancedShown

        onClicked: {
            Backend.exportTLSCertificates();
        }
    }
    SettingsItem {
        id: repair
        Layout.fillWidth: true
        actionText: qsTr("Repair")
        colorScheme: root.colorScheme
        description: qsTr("Reload all accounts, cached data, and download all emails again. Email clients stay connected to Bridge.")
        text: qsTr("Repair Bridge")
        type: SettingsItem.Button
        visible: root._isAdvancedShown
        enabled: root.allUsersLoaded && Backend.users.count && root.hasInternetConnection

        onClicked: {
            root.notifications.askRepairBridge();
        }

        Connections {
            function onInternetOff() {
                root.hasInternetConnection = false;
                repair.description = qsTr("This feature requires internet access to the Proton servers.")

            }
            function onInternetOn() {
                root.hasInternetConnection = true;
                repair.description = qsTr("Reload all accounts, cached data, and download all emails again. Email clients stay connected to Bridge.")
            }
            function onAllUsersLoaded() {
                root.allUsersLoaded = true;
            }
            target: Backend
        }
    }
    SettingsItem {
        id: reset
        Layout.fillWidth: true
        actionText: qsTr("Reset")
        colorScheme: root.colorScheme
        description: qsTr("Remove all accounts, clear cached data, and restore the original settings.")
        text: qsTr("Reset Bridge")
        type: SettingsItem.Button
        visible: root._isAdvancedShown

        onClicked: {
            root.notifications.askResetBridge();
        }
    }
}
