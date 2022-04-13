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

import QtQml 2.12
import QtQuick 2.13
import QtQuick.Layouts 1.12
import QtQuick.Controls 2.12

import Proton 4.0

Item {
    id: root

    property ColorScheme colorScheme

    property var backend

    implicitHeight: children[0].implicitHeight
    implicitWidth: children[0].implicitWidth

    RowLayout {
        anchors.fill: parent
        spacing: 0

        Rectangle {
            color: root.colorScheme.background_norm

            Layout.fillHeight: true
            Layout.fillWidth: true

            implicitHeight: children[0].implicitHeight
            implicitWidth: children[0].implicitWidth

            visible: signInItem.currentIndex == 0

            GridLayout {
                anchors.fill: parent

                columnSpacing: 0
                rowSpacing: 0

                columns: 3

                // top margin
                Item {
                    Layout.columnSpan: 3
                    Layout.fillWidth: true

                    // Using binding component here instead of direct binding to avoid binding loop during construction of element
                    Binding on Layout.preferredHeight {
                        value: (parent.height - welcomeContentItem.height) / 4
                    }
                }

                // left margin
                Item {
                    Layout.minimumWidth: 48
                    Layout.maximumWidth: 80
                    Layout.fillWidth: true
                    Layout.preferredHeight: welcomeContentItem.height
                }

                ColumnLayout {
                    id: welcomeContentItem
                    Layout.fillWidth: true
                    spacing: 0

                    Image {
                        source: colorScheme.welcome_img
                        Layout.alignment: Qt.AlignHCenter
                        Layout.topMargin: 16
                        sourceSize.height: 148
                        sourceSize.width: 264
                    }

                    Label {
                        colorScheme: root.colorScheme
                        text: qsTr("Welcome to\nProton Mail Bridge")
                        Layout.alignment: Qt.AlignHCenter
                        Layout.fillWidth: true
                        Layout.topMargin: 16

                        horizontalAlignment: Text.AlignHCenter

                        type: Label.LabelType.Heading
                    }

                    Label {
                        colorScheme: root.colorScheme
                        id: longTextLabel
                        text: qsTr("Add your Proton Mail account to securely access and manage your messages in your favorite email client. Bridge runs in the background and encrypts and decrypts your messages seamlessly.")
                        Layout.alignment: Qt.AlignHCenter
                        Layout.fillWidth: true
                        Layout.topMargin: 16
                        Layout.preferredWidth: 320

                        wrapMode: Text.WordWrap

                        horizontalAlignment: Text.AlignHCenter

                        type: Label.LabelType.Body
                    }
                }

                // Right margin
                Item {
                    Layout.minimumWidth: 48
                    Layout.maximumWidth: 80
                    Layout.fillWidth: true
                    Layout.preferredHeight: welcomeContentItem.height
                }

                // bottom margin
                Item {
                    Layout.columnSpan: 3
                    Layout.fillWidth: true
                    Layout.fillHeight: true

                    implicitHeight: children[0].implicitHeight + children[0].anchors.bottomMargin + children[0].anchors.topMargin
                    implicitWidth: children[0].implicitWidth

                    Image {
                        id: logoImage
                        source: colorScheme.logo_img

                        anchors.horizontalCenter: parent.horizontalCenter
                        anchors.bottom: parent.bottom
                        anchors.topMargin: 48
                        anchors.bottomMargin: 48
                        sourceSize.height: 25
                        sourceSize.width: 200
                    }
                }
            }
        }

        Rectangle {
            color: (signInItem.currentIndex == 0) ? root.colorScheme.background_weak : root.colorScheme.background_norm
            Layout.fillHeight: true
            Layout.fillWidth: true

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
                        colorScheme: root.colorScheme
                        anchors.left: parent.left
                        anchors.bottom: parent.bottom

                        anchors.leftMargin: 80
                        anchors.rightMargin: 80
                        anchors.topMargin: 80
                        anchors.bottomMargin: 80

                        visible: signInItem.currentIndex != 0

                        secondary: true
                        text: qsTr("Back")

                        onClicked: {
                            signInItem.abort()
                        }
                    }
                }

                GridLayout {
                    Layout.fillHeight: true
                    Layout.fillWidth: true

                    columnSpacing: 0
                    rowSpacing: 0

                    columns: 3

                    // top margin
                    Item {
                        Layout.columnSpan: 3
                        Layout.fillWidth: true

                        // Using binding component here instead of direct binding to avoid binding loop during construction of element
                        Binding on Layout.preferredHeight {
                            value: (parent.height - signInItem.height) / 4
                        }
                    }

                    // left margin
                    Item {
                        Layout.minimumWidth: 48
                        Layout.maximumWidth: 80
                        Layout.fillWidth: true
                        Layout.preferredHeight: signInItem.height
                    }


                    SignIn {
                        id: signInItem
                        colorScheme: root.colorScheme

                        Layout.preferredWidth: 320
                        Layout.fillWidth: true

                        username: root.backend.users.count === 1 && root.backend.users.get(0) && root.backend.users.get(0).loggedIn === false ? root.backend.users.get(0).username : ""
                        backend: root.backend
                    }

                    // Right margin
                    Item {
                        Layout.minimumWidth: 48
                        Layout.maximumWidth: 80
                        Layout.fillWidth: true
                        Layout.preferredHeight: signInItem.height
                    }

                    // bottom margin
                    Item {
                        Layout.columnSpan: 3
                        Layout.fillWidth: true
                        Layout.fillHeight: true
                    }
                }

                Item {
                    Layout.fillHeight: true
                    Layout.preferredWidth: signInItem.currentIndex == 0 ? 0 : parent.width / 4
                }
            }
        }

        states: [
            State {
                name: "Page 1"
                PropertyChanges {
                    target: signInItem
                    currentIndex: 0
                }
            },
            State {
                name: "Page 2"
                PropertyChanges {
                    target: signInItem
                    currentIndex: 1
                }
            },
            State {
                name: "Page 3"
                PropertyChanges {
                    target: signInItem
                    currentIndex: 2
                }
            }
        ]
    }
}
