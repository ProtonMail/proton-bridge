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
import QtQuick.Dialogs
import Proton

SettingsView {
    id: root

    property url diskCachePath: pathDialog.shortcuts.home
    property var notifications

    function refresh() {
        diskCacheSetting.description = Backend.nativePath(root.diskCachePath);
        submitButton.enabled = (!submitButton.loading) && !Backend.areSameFileOrFolder(Backend.diskCachePath, root.diskCachePath);
    }
    function setDefaultValues() {
        root.diskCachePath = Backend.diskCachePath;
        root.refresh();
    }
    function submit() {
        submitButton.loading = true;
        Backend.setDiskCachePath(root.diskCachePath);
    }

    fillHeight: false

    onBack: {
        root.setDefaultValues();
    }
    onVisibleChanged: {
        root.setDefaultValues();
    }

    Label {
        Layout.fillWidth: true
        colorScheme: root.colorScheme
        text: qsTr("Local cache")
        type: Label.Heading
    }
    Label {
        Layout.fillWidth: true
        color: root.colorScheme.text_weak
        colorScheme: root.colorScheme
        text: qsTr("Bridge stores your encrypted messages locally to optimize communication with your client.")
        type: Label.Body
        wrapMode: Text.WordWrap
    }
    SettingsItem {
        id: diskCacheSetting
        Layout.fillWidth: true
        actionText: qsTr("Change location")
        colorScheme: root.colorScheme
        descriptionWrap: Text.WrapAnywhere
        text: qsTr("Current cache location")
        type: SettingsItem.Button

        onClicked: {
            pathDialog.open();
        }

        FolderDialog {
            id: pathDialog
            currentFolder: root.diskCachePath
            title: qsTr("Select cache location")

            onAccepted: {
                root.diskCachePath = pathDialog.selectedFolder;
                root.refresh();
            }
        }
    }
    RowLayout {
        spacing: 12

        Button {
            id: submitButton
            colorScheme: root.colorScheme
            text: qsTr("Save")

            onClicked: {
                root.submit();
            }
        }
        Button {
            colorScheme: root.colorScheme
            secondary: true
            text: qsTr("Cancel")

            onClicked: root.back()
        }
        Connections {
            function onDiskCachePathChangeFinished() {
                submitButton.loading = false;
                root.setDefaultValues();
            }

            target: Backend
        }
    }
}
