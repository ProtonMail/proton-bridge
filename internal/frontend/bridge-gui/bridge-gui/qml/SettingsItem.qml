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

import QtQuick
import QtQuick.Layouts
import QtQuick.Controls

import Proton

Item {
    id: root
    property var colorScheme

    property string text: "Text"
    property string actionText: "Action"
    property string actionIcon: ""
    property string description: "Lorem ipsum dolor sit amet"
    property alias descriptionWrap: descriptionLabel.wrapMode
    property var    type: SettingsItem.ActionType.Toggle

    property bool checked: true
    property bool loading: false
    property bool showSeparator: true

    property var _bottomMargin: 20
    property var _lineWidth: 1
    property var _toggleTopMargin: 6

    signal clicked

    enum ActionType {
        Toggle = 1, Button = 2, PrimaryButton = 3
    }

    implicitHeight: children[0].implicitHeight + children[0].anchors.topMargin + children[0].anchors.bottomMargin
    implicitWidth: children[0].implicitWidth + children[0].anchors.leftMargin + children[0].anchors.rightMargin

    RowLayout {
        anchors.fill: parent
        spacing: 16

        ColumnLayout {
            Layout.fillHeight: true
            Layout.fillWidth: true
            Layout.bottomMargin: root._bottomMargin

            spacing: 4

            Label {
                id: mainLabel
                colorScheme: root.colorScheme
                text: root.text
                type: Label.Body_semibold
            }

            Label {
                id: descriptionLabel
                Layout.fillHeight: true
                Layout.fillWidth: true

                Layout.preferredWidth: parent.width

                wrapMode: Text.WordWrap
                colorScheme: root.colorScheme
                text: root.description
                color: root.colorScheme.text_weak
            }
        }

        Toggle {
            Layout.alignment: Qt.AlignTop
            Layout.topMargin: root._toggleTopMargin
            id: toggle
            colorScheme: root.colorScheme
            visible: root.type === SettingsItem.ActionType.Toggle

            checked: root.checked
            loading: root.loading
            onClicked: { if (!root.loading) root.clicked() }
        }

        Button {
            Layout.alignment: Qt.AlignTop

            id: button
            colorScheme: root.colorScheme
            visible: root.type === SettingsItem.Button || root.type === SettingsItem.PrimaryButton
            text: root.actionText + (root.actionIcon != "" ? "  " : "")
            loading: root.loading
            icon.source: root.actionIcon
            onClicked: { if (!root.loading) root.clicked() }
            secondary: root.type !== SettingsItem.PrimaryButton
        }
    }

    Rectangle {
        anchors.left: root.left
        anchors.right: root.right
        anchors.bottom: root.bottom
        color: colorScheme.border_weak
        height: root._lineWidth
        visible: root.showSeparator
    }
}
