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
import "." as Proton

Item {
    id: root

    property ColorScheme colorScheme

    function start() {
        root.visible = true;
    }

    RowLayout {
        anchors.fill: parent
        spacing: 0

        Rectangle {
            id: leftHalf
            Layout.fillHeight: true
            Layout.fillWidth: true
            color: root.colorScheme.background_norm

            Item {
                id: leftContent
                anchors.bottom: parent.bottom
                anchors.bottomMargin: 96
                anchors.horizontalCenter: parent.horizontalCenter
                anchors.top: parent.top
                anchors.topMargin: 96
                width: 444

                OnboardingLeftPane {
                    anchors.fill: parent
                    colorScheme: root.colorScheme
                }
            }
            Image {
                id: mailLogoWithWordmark
                anchors.bottom: parent.bottom
                anchors.bottomMargin: 48
                anchors.horizontalCenter: parent.horizontalCenter
                antialiasing: true
                fillMode: Image.PreserveAspectFit
                height: 24
                smooth: true
                source: root.colorScheme.mail_logo_with_wordmark
                sourceSize.height: 24
                sourceSize.width: 142
            }
        }
        Rectangle {
            id: rightHalf
            Layout.fillHeight: true
            Layout.fillWidth: true
            color: root.colorScheme.background_weak

            Item {
                id: rightContent
                anchors.bottom: parent.bottom
                anchors.bottomMargin: 96
                anchors.horizontalCenter: parent.horizontalCenter
                anchors.top: parent.top
                anchors.topMargin: 96
                width: 444

                OnboardingRightPane {
                    anchors.fill: parent
                    colorScheme: root.colorScheme
                }
            }
            Label {
                id: reportProblemLink
                anchors.bottom: parent.bottom
                anchors.bottomMargin: 48
                anchors.horizontalCenter: parent.horizontalCenter
                colorScheme: root.colorScheme
                horizontalAlignment: Text.AlignRight
                text: link("#", "Report problem")
                width: 444

                onLinkActivated: {
                    root.visible = false;
                }

                HoverHandler {
                    id: mouse
                    acceptedDevices: PointerDevice.Mouse
                    cursorShape: Qt.PointingHandCursor
                }
            }
        }
    }
}

