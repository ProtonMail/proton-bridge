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
import Proton

Item {
    id: root
    enum ActionType {
        Toggle = 1,
        Button,
        PrimaryButton
    }

    property var _bottomMargin: 20
    property var _lineWidth: 1
    property var _toggleTopMargin: 6
    property string actionIcon: ""
    property string actionText: "Action"
    property bool checked: true
    property var colorScheme
    property string description: "Lorem ipsum dolor sit amet"
    property alias descriptionWrap: descriptionLabel.wrapMode
    property bool loading: false
    property bool showSeparator: true
    property string text: "Text"
    property var type: SettingsItem.ActionType.Toggle

    signal clicked

    implicitHeight: children[0].implicitHeight + children[0].anchors.topMargin + children[0].anchors.bottomMargin
    implicitWidth: children[0].implicitWidth + children[0].anchors.leftMargin + children[0].anchors.rightMargin
    Accessible.name: text
    Accessible.role: Accessible.Grouping

    RowLayout {
        anchors.fill: parent
        spacing: 16

        ColumnLayout {
            Layout.bottomMargin: root._bottomMargin
            Layout.fillHeight: true
            Layout.fillWidth: true
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
                color: root.colorScheme.text_weak
                colorScheme: root.colorScheme
                text: root.description
                wrapMode: Text.WordWrap
            }
        }
        Toggle {
            id: toggle
            Layout.alignment: Qt.AlignTop
            Layout.topMargin: root._toggleTopMargin
            checked: root.checked
            colorScheme: root.colorScheme
            loading: root.loading
            visible: root.type === SettingsItem.ActionType.Toggle
            Accessible.role: Accessible.CheckBox
            Accessible.name: root.Accessible.name + " toggle"

            onClicked: {
                if (!root.loading)
                    root.clicked();
            }
        }
        Button {
            id: button
            Layout.alignment: Qt.AlignTop
            colorScheme: root.colorScheme
            icon.source: root.actionIcon
            loading: root.loading
            secondary: root.type !== SettingsItem.PrimaryButton
            text: root.actionText
            visible: root.type === SettingsItem.Button || root.type === SettingsItem.PrimaryButton
            Accessible.role: Accessible.Button
            Accessible.name: root.Accessible.name + " button"

            onClicked: {
                if (!root.loading)
                    root.clicked();
            }
        }
    }
    Rectangle {
        anchors.bottom: root.bottom
        anchors.left: root.left
        anchors.right: root.right
        color: colorScheme.border_weak
        height: root._lineWidth
        visible: root.showSeparator
    }
}
