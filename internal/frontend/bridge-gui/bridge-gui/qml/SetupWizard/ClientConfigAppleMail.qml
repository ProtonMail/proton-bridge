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
import QtQml
import QtQuick
import QtQuick.Layouts
import QtQuick.Controls

Item {
    id: root

    property var wizard

    property bool profilePaneLaunched: false

    function reset() {
        profilePaneLaunched = false;
    }

    ColumnLayout {
        anchors.left: parent.left
        anchors.right: parent.right
        anchors.verticalCenter: parent.verticalCenter
        spacing: ProtonStyle.wizard_spacing_large

        ColumnLayout {
            Layout.fillWidth: true
            spacing: ProtonStyle.wizard_spacing_medium

            Label {
                Layout.alignment: Qt.AlignHCenter
                Layout.fillWidth: true
                colorScheme: wizard.colorScheme
                horizontalAlignment: Text.AlignHCenter
                text: qsTr("Install the profile")
                type: Label.LabelType.Title
                wrapMode: Text.WordWrap
            }
            Label {
                Layout.alignment: Qt.AlignHCenter
                Layout.fillWidth: true
                color: colorScheme.text_weak
                colorScheme: wizard.colorScheme
                horizontalAlignment: Text.AlignHCenter
                text: qsTr("A series of pop-ups will appear. Follow the instructions to install the profile.")
                type: Label.LabelType.Body
                wrapMode: Text.WordWrap
            }
        }
        Image {
            Layout.alignment: Qt.AlignHCenter
            height: 102
            source: "/qml/icons/img-macos-profile-screenshot.png"
            width: 364
        }
        ColumnLayout {
            Layout.fillWidth: true
            spacing: ProtonStyle.wizard_spacing_medium

            Button {
                Layout.fillWidth: true
                colorScheme: wizard.colorScheme
                text: profilePaneLaunched ? qsTr("I have installed the profile") : qsTr("Install the profile")

                onClicked: {
                    if (profilePaneLaunched) {
                        wizard.showClientConfigEnd();
                    } else {
                        wizard.user.configureAppleMail(wizard.address);
                        profilePaneLaunched = true;
                    }
                }
            }
            Button {
                Layout.fillWidth: true
                colorScheme: wizard.colorScheme
                secondary: true
                text: qsTr("Cancel")

                onClicked: {
                    wizard.closeWizard();
                }
            }
        }
    }
}
