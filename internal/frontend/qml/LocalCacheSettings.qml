// Copyright (c) 2022 Proton AG
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

import QtQuick 2.13
import QtQuick.Layouts 1.12
import QtQuick.Controls 2.13
import QtQuick.Controls.impl 2.13
import QtQuick.Dialogs 1.1

import Proton 4.0

SettingsView {
    id: root

    fillHeight: false

    property var notifications
    property bool _diskCacheEnabled: true
    property url _diskCachePath: pathDialog.shortcuts.home

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
        Layout.maximumWidth: this.parent.Layout.maximumWidth
        wrapMode: Text.WordWrap
    }

    SettingsItem {
        colorScheme: root.colorScheme
        text: qsTr("Enable local cache")
        description: qsTr("Recommended for optimal performance.")
        type: SettingsItem.Toggle
        checked: root._diskCacheEnabled
        onClicked: root._diskCacheEnabled = !root._diskCacheEnabled

        Layout.fillWidth: true
    }

    SettingsItem {
        colorScheme: root.colorScheme
        text: qsTr("Current cache location")
        actionText: qsTr("Change location")
        description: root.backend.goos === "windows" ?
                         root._diskCachePath.toString().replace("file:///", "").replace(new RegExp("/", 'g'), "\\") + "\\" :
                         root._diskCachePath.toString().replace("file://", "") + "/"
        descriptionWrap: Text.WrapAnywhere
        type: SettingsItem.Button
        enabled: root._diskCacheEnabled
        onClicked: {
            pathDialog.open()
        }

        Layout.fillWidth: true

        FileDialog {
            id: pathDialog
            title: qsTr("Select cache location")
            folder: root._diskCachePath
            onAccepted: root._diskCachePath = pathDialog.fileUrl
            selectFolder: true
        }
    }

    RowLayout {
        spacing: 12

        Button {
            id: submitButton
            colorScheme: root.colorScheme
            text: qsTr("Save and restart")
            enabled: (
                root.backend.diskCachePath != root._diskCachePath ||
                root.backend.isDiskCacheEnabled != root._diskCacheEnabled
            )
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
            target: root.backend

            onChangeLocalCacheFinished: {
                submitButton.loading = false
                root.setDefaultValues()
            }
        }
    }

    onBack: {
        root.setDefaultValues()
    }

    function submit(){
        if (!root._diskCacheEnabled && root.backend.isDiskCacheEnabled) {
            root.notifications.askDisableLocalCache()
            return
        }

        if (root._diskCacheEnabled && !root.backend.isDiskCacheEnabled) {
            root.notifications.askEnableLocalCache(root._diskCachePath)
            return
        }

        // Not asking, only changing path
        submitButton.loading = true
        root.backend.changeLocalCache(root.backend.isDiskCacheEnabled, root._diskCachePath)
    }

    function setDefaultValues(){
        root._diskCacheEnabled = root.backend.isDiskCacheEnabled
        root._diskCachePath = root.backend.diskCachePath
    }

    onVisibleChanged: {
        root.setDefaultValues()
    }
}
