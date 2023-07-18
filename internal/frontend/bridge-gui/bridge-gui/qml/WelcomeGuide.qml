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
import Proton

Item {
    id: root

    property ColorScheme colorScheme

    implicitHeight: children[0].implicitHeight
    implicitWidth: children[0].implicitWidth

    RowLayout {
        anchors.fill: parent
        spacing: 0

        states: [
            State {
                name: "Page 1"

                PropertyChanges {
                    currentIndex: 0
                    target: signInItem
                }
            },
            State {
                name: "Page 2"

                PropertyChanges {
                    currentIndex: 1
                    target: signInItem
                }
            },
            State {
                name: "Page 3"

                PropertyChanges {
                    currentIndex: 2
                    target: signInItem
                }
            }
        ]

        Rectangle {
            Layout.fillHeight: true
            Layout.fillWidth: true
            color: root.colorScheme.background_norm
            implicitHeight: children[0].implicitHeight
            implicitWidth: children[0].implicitWidth
            visible: signInItem.currentIndex === 0

            GridLayout {
                anchors.fill: parent
                columnSpacing: 0
                columns: 3
                rowSpacing: 0

                // top margin
                Item {
                    Layout.columnSpan: 3
                    Layout.fillWidth: true

                    // Using binding component here instead of direct binding to avoid binding loop during construction of element
                    Binding on Layout.preferredHeight  {
                        value: (parent.height - welcomeContentItem.height) / 4
                    }
                }

                // left margin
                Item {
                    Layout.fillWidth: true
                    Layout.maximumWidth: 80
                    Layout.minimumWidth: 48
                    Layout.preferredHeight: welcomeContentItem.height
                }
                ColumnLayout {
                    id: welcomeContentItem
                    Layout.fillWidth: true
                    spacing: 0

                    Image {
                        Layout.alignment: Qt.AlignHCenter
                        Layout.topMargin: 16
                        source: colorScheme.welcome_img
                        sourceSize.height: 148
                        sourceSize.width: 264
                    }
                    Label {
                        Layout.alignment: Qt.AlignHCenter
                        Layout.fillWidth: true
                        Layout.topMargin: 16
                        colorScheme: root.colorScheme
                        horizontalAlignment: Text.AlignHCenter
                        text: qsTr("Welcome to\nProton Mail Bridge")
                        type: Label.LabelType.Heading
                    }
                    Label {
                        id: longTextLabel
                        Layout.alignment: Qt.AlignHCenter
                        Layout.fillWidth: true
                        Layout.preferredWidth: 320
                        Layout.topMargin: 16
                        colorScheme: root.colorScheme
                        horizontalAlignment: Text.AlignHCenter
                        text: qsTr("Add your Proton Mail account to securely access and manage your messages in your favorite email client. Bridge runs in the background and encrypts and decrypts your messages seamlessly.")
                        type: Label.LabelType.Body
                        wrapMode: Text.WordWrap
                    }
                }

                // Right margin
                Item {
                    Layout.fillWidth: true
                    Layout.maximumWidth: 80
                    Layout.minimumWidth: 48
                    Layout.preferredHeight: welcomeContentItem.height
                }

                // bottom margin
                Item {
                    Layout.columnSpan: 3
                    Layout.fillHeight: true
                    Layout.fillWidth: true
                    implicitHeight: children[0].implicitHeight + children[0].anchors.bottomMargin + children[0].anchors.topMargin
                    implicitWidth: children[0].implicitWidth

                    Image {
                        id: logoImage
                        anchors.bottom: parent.bottom
                        anchors.bottomMargin: 48
                        anchors.horizontalCenter: parent.horizontalCenter
                        anchors.topMargin: 48
                        source: colorScheme.logo_img
                        sourceSize.height: 25
                        sourceSize.width: 200
                    }
                }
            }
        }
        Rectangle {
            Layout.fillHeight: true
            Layout.fillWidth: true
            color: (signInItem.currentIndex == 0) ? root.colorScheme.background_weak : root.colorScheme.background_norm
            implicitHeight: children[0].implicitHeight
            implicitWidth: children[0].implicitWidth

            RowLayout {
                anchors.fill: parent
                spacing: 0

                Item {
                    Layout.fillHeight: true
                    Layout.fillWidth: true
                    Layout.preferredWidth: signInItem.currentIndex == 0 ? 0 : parent.width / 4
                    implicitHeight: children[0].implicitHeight + children[0].anchors.topMargin + children[0].anchors.bottomMargin
                    implicitWidth: children[0].implicitWidth + children[0].anchors.leftMargin + children[0].anchors.rightMargin

                    Button {
                        anchors.bottom: parent.bottom
                        anchors.bottomMargin: 80
                        anchors.left: parent.left
                        anchors.leftMargin: 80
                        anchors.rightMargin: 80
                        anchors.topMargin: 80
                        colorScheme: root.colorScheme
                        secondary: true
                        text: qsTr("Back")
                        visible: signInItem.currentIndex != 0

                        onClicked: {
                            signInItem.abort();
                        }
                    }
                }
                GridLayout {
                    Layout.fillHeight: true
                    Layout.fillWidth: true
                    columnSpacing: 0
                    columns: 3
                    rowSpacing: 0

                    // top margin
                    Item {
                        Layout.columnSpan: 3
                        Layout.fillWidth: true

                        // Using binding component here instead of direct binding to avoid binding loop during construction of element
                        Binding on Layout.preferredHeight  {
                            value: (parent.height - signInItem.height) / 4
                        }
                    }

                    // left margin
                    Item {
                        Layout.fillWidth: true
                        Layout.maximumWidth: 80
                        Layout.minimumWidth: 48
                        Layout.preferredHeight: signInItem.height
                    }
                    SignIn {
                        id: signInItem
                        Layout.fillWidth: true
                        Layout.preferredWidth: 320
                        colorScheme: root.colorScheme
                        focus: true
                        username: Backend.users.count === 1 && Backend.users.get(0) && (Backend.users.get(0).state === EUserState.SignedOut) ? Backend.users.get(0).username : ""
                    }

                    // Right margin
                    Item {
                        Layout.fillWidth: true
                        Layout.maximumWidth: 80
                        Layout.minimumWidth: 48
                        Layout.preferredHeight: signInItem.height
                    }

                    // bottom margin
                    Item {
                        Layout.columnSpan: 3
                        Layout.fillHeight: true
                        Layout.fillWidth: true
                    }
                }
                Item {
                    Layout.fillHeight: true
                    Layout.preferredWidth: signInItem.currentIndex === 0 ? 0 : parent.width / 4
                }
            }
        }
    }
}
