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

    function closeWizard() {
        root.visible = false;
    }
    function start() {
        root.visible = true;
        leftContent.currentIndex = 0;
        rightContent.currentIndex = 0;
    }
    function startLogin() {
        root.visible = true;
        leftContent.currentIndex = 1;
        rightContent.currentIndex = 1;
        loginLeftPane.showSignIn();
        loginRightPane.reset(true);
    }

    Connections {
        function onLoginFinished() {
            root.closeWizard();
        }

        target: Backend
    }
    RowLayout {
        anchors.fill: parent
        spacing: 0

        Rectangle {
            id: leftHalf
            Layout.fillHeight: true
            Layout.fillWidth: true
            color: root.colorScheme.background_norm

            StackLayout {
                id: leftContent
                anchors.bottom: parent.bottom
                anchors.bottomMargin: 96
                anchors.horizontalCenter: parent.horizontalCenter
                anchors.top: parent.top
                anchors.topMargin: 96
                clip: true
                currentIndex: 0
                width: 444

                // stack index 0
                OnboardingLeftPane {
                    Layout.fillHeight: true
                    Layout.fillWidth: true
                    colorScheme: root.colorScheme
                }

                // stack index 1
                LoginLeftPane {
                    id: loginLeftPane
                    Layout.fillHeight: true
                    Layout.fillWidth: true
                    colorScheme: root.colorScheme
                }
            }
            Image {
                id: mailLogoWithWordmark
                anchors.bottom: parent.bottom
                anchors.bottomMargin: 48
                anchors.horizontalCenter: parent.horizontalCenter
                fillMode: Image.PreserveAspectFit
                height: 24
                mipmap: true
                source: root.colorScheme.mail_logo_with_wordmark
            }
        }
        Rectangle {
            id: rightHalf
            Layout.fillHeight: true
            Layout.fillWidth: true
            color: root.colorScheme.background_weak

            StackLayout {
                id: rightContent
                anchors.bottom: parent.bottom
                anchors.bottomMargin: 96
                anchors.horizontalCenter: parent.horizontalCenter
                anchors.top: parent.top
                anchors.topMargin: 96
                currentIndex: 0
                clip: true
                width: 444

                // stack index 0
                OnboardingRightPane {
                    Layout.fillHeight: true
                    Layout.fillWidth: true
                    colorScheme: root.colorScheme

                    onOnboardingAccepted: root.startLogin()
                }

                // stack index 1
                LoginRightPane {
                    id: loginRightPane
                    Layout.fillHeight: true
                    Layout.fillWidth: true
                    colorScheme: root.colorScheme

                    onLoginAbort: {
                        root.closeWizard();
                    }
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

