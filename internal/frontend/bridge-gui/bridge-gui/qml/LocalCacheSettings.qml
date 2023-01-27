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
import QtQuick.Controls.impl
import QtQuick.Dialogs

import Proton

SettingsView {
    id: root

    fillHeight: false

    property var notifications
    property url diskCachePath: pathDialog.shortcuts.home

    function refresh() {
        diskCacheSetting.description = Backend.nativePath(root.diskCachePath)
        submitButton.enabled = (!submitButton.loading) && !Backend.areSameFileOrFolder(Backend.diskCachePath, root.diskCachePath)
    }

    Label {
        colorScheme: root.colorScheme
        text: qsTr("Local cache")
        type: Label.Heading
        Layout.fillWidth: true
    }

    Label {
        colorScheme: root.colorScheme
        text: qsTr("Bridge stores your encrypted messages locally to optimize communication with your client.")
        type: Label.Body
        color: root.colorScheme.text_weak
        Layout.fillWidth: true
        wrapMode: Text.WordWrap
    }

    SettingsItem {
        id: diskCacheSetting
        colorScheme: root.colorScheme
        text: qsTr("Current cache location")
        actionText: qsTr("Change location")
        descriptionWrap: Text.WrapAnywhere
        type: SettingsItem.Button
        onClicked: {
            pathDialog.open()
        }

        Layout.fillWidth: true

        FolderDialog {
            id: pathDialog
            title: qsTr("Select cache location")
            currentFolder: root.diskCachePath
            onAccepted: {
                root.diskCachePath = pathDialog.selectedFolder
                root.refresh()
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
                root.submit()
            }
        }

        Button {
            colorScheme: root.colorScheme
            text: qsTr("Cancel")
            onClicked: root.back()
            secondary: true
        }

        Connections {
            target: Backend

            function onDiskCachePathChangeFinished() {
                submitButton.loading = false
                root.setDefaultValues()
            }
        }
    }

    onBack: {
        root.setDefaultValues()
    }

    function submit() {
        submitButton.loading = true
        Backend.setDiskCachePath(root.diskCachePath)
    }

    function setDefaultValues(){
        root.diskCachePath = Backend.diskCachePath
        root.refresh();
    }

    onVisibleChanged: {
        root.setDefaultValues()
    }
}
