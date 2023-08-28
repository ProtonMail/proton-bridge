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

Rectangle {
    id: root

    property ColorScheme colorScheme: wizard.colorScheme
    property var wizard

    clip: true
    color: colorScheme.background_norm

    Item {
        id: centeredContainer
        anchors.bottom: parent.bottom
        anchors.bottomMargin: 84
        anchors.horizontalCenter: parent.horizontalCenter
        anchors.top: parent.top
        anchors.topMargin: 32
        clip: true
        width: ProtonStyle.wizard_pane_width

        ColumnLayout {
            anchors.left: parent.left
            anchors.right: parent.right
            anchors.verticalCenter: parent.verticalCenter
            spacing: ProtonStyle.wizard_spacing_medium

            Image {
                Layout.alignment: Qt.AlignHCenter
                Layout.preferredHeight: sourceSize.height
                Layout.preferredWidth: sourceSize.width
                source: "/qml/icons/img-client-config-success.svg"
                sourceSize.height: 104
                sourceSize.width: 190
            }
            Label {
                Layout.alignment: Qt.AlignHCenter
                Layout.fillWidth: true
                colorScheme: wizard.colorScheme
                horizontalAlignment: Text.AlignHCenter
                text: qsTr("Congratulations! You're all setup")
                type: Label.LabelType.Title
                wrapMode: Text.WordWrap
            }
            Label {
                Layout.alignment: Qt.AlignHCenter
                Layout.fillWidth: true
                color: colorScheme.text_weak
                colorScheme: wizard.colorScheme
                horizontalAlignment: Text.AlignHCenter
                text: wizard.address
                type: Label.LabelType.Body
                wrapMode: Text.WordWrap
            }
            Label {
                Layout.alignment: Qt.AlignHCenter
                Layout.fillWidth: true
                colorScheme: wizard.colorScheme
                horizontalAlignment: Text.AlignHCenter
                text: qsTr("Your client has been configured. While complete synchronization might take some time, you can already send encrypted emails.")
                type: Label.LabelType.Body
                wrapMode: Text.WordWrap
            }
            Button {
                Layout.fillWidth: true
                colorScheme: root.colorScheme
                text: qsTr("Done")

                onClicked: wizard.closeWizard()
            }
        }
    }
    Image {
        id: mailLogoWithWordmark
        anchors.bottom: parent.bottom
        anchors.bottomMargin: 32
        anchors.horizontalCenter: parent.horizontalCenter
        height: 36
        source: root.colorScheme.mail_logo_with_wordmark
        sourceSize.height: height
        sourceSize.width: width
        width: 134
    }
}
