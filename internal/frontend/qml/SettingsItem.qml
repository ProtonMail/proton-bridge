// Copyright (c) 2021 Proton Technologies AG
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

ColumnLayout {
    id: root
    property var colorScheme

    property string text: "Text"
    property string actionText: "Action"
    property string actionIcon: ""
    property string description: "Lorem ipsum dolor sit amet"
    property var    type: SettingsItem.ActionType.Toggle

    property bool checked: true
    property bool disabled: false
    property bool loading: false

    signal clicked

    spacing: 20

    Layout.fillWidth: true
    Layout.maximumWidth: root.parent.Layout.maximumWidth

    enum ActionType {
        Toggle = 1, Button = 2, PrimaryButton = 3
    }

    RowLayout {
        Layout.fillWidth: true

        ColumnLayout {
            Label {
                id:mainLabel
                colorScheme: root.colorScheme
                text: root.text
                type: Label.Body_semibold
            }

            Label {
                Layout.minimumWidth: mainLabel.width
                Layout.maximumWidth: root.Layout.maximumWidth - root.spacing - (
                    toggle.visible ? toggle.width : button.width
                )

                wrapMode: Text.WordWrap
                colorScheme: root.colorScheme
                text: root.description
                color: root.colorScheme.text_weak
            }
        }

        Item {
            Layout.fillWidth: true
            Layout.fillHeight: true
        }

        Toggle {
            id: toggle
            colorScheme: root.colorScheme
            visible: root.type == SettingsItem.ActionType.Toggle

            checked: root.checked
            loading: root.loading
            onClicked: { if (!root.loading) root.clicked() }
        }

        Button {
            id: button
            colorScheme: root.colorScheme
            visible: root.type == SettingsItem.Button || root.type == SettingsItem.PrimaryButton
            text: root.actionText + (root.actionIcon != "" ? "  " : "")
            loading: root.loading
            icon.source: root.actionIcon
            onClicked: { if (!root.loading) root.clicked() }
            secondary: root.type != SettingsItem.PrimaryButton
        }
    }

    Rectangle {
        Layout.fillWidth: true
        color: colorScheme.border_weak
        height: 1
    }
}
