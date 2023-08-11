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
import QtQml
import QtQuick
import QtQuick.Layouts
import QtQuick.Controls
import QtQuick.Controls.impl
import Proton

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
import QtQml
import QtQuick
import QtQuick.Layouts
import QtQuick.Controls
import QtQuick.Controls.impl
import Proton

Item {
    id: root

    property var wizard

    ColumnLayout {
        anchors.left: parent.left
        anchors.right: parent.right
        anchors.top: parent.top
        spacing: 0

        Label {
            Layout.fillWidth: true
            colorScheme: wizard.colorScheme
            horizontalAlignment: Text.AlignHCenter
            text: qsTr("A word of warning")
            type: Label.LabelType.Heading
            wrapMode: Text.WordWrap
        }
        Item {
            Layout.preferredHeight: 96
        }
        Label {
            Layout.alignment: Qt.AlignHCenter
            Layout.fillWidth: true
            horizontalAlignment: Text.AlignHCenter
            colorScheme: wizard.colorScheme
            text: qsTr("Do not enter your Proton account password in you email application.")
            type: Label.LabelType.Body
            wrapMode: Text.WordWrap
        }
        Item {
            Layout.preferredHeight: 96
        }
        Label {
            Layout.alignment: Qt.AlignHCenter
            Layout.fillWidth: true
            colorScheme: wizard.colorScheme
            horizontalAlignment: Text.AlignHCenter
            text: qsTr("We have generated a new password for you. It will work only on this computer, and can safely be entered in your email client.")
            type: Label.LabelType.Body
            wrapMode: Text.WordWrap
        }
        Item {
            Layout.preferredHeight: 96
        }
        Button {
            Layout.fillWidth: true
            colorScheme: wizard.colorScheme
            text: qsTr("I understand")

            onClicked: {
                root.wizard.showClientParams();
            }
        }
        Item {
            Layout.preferredHeight: 32
        }
        Button {
            Layout.fillWidth: true
            colorScheme: wizard.colorScheme
            secondary: true
            text: qsTr("Cancel")

            onClicked: {
                root.wizard.closeWizard();
            }
        }
    }
}

