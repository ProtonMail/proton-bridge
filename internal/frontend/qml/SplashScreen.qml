// Copyright (c) 2022 Proton Technologies AG
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

import QtQml 2.12
import QtQuick 2.13
import QtQuick.Layouts 1.12
import QtQuick.Controls 2.12
import QtQuick.Controls.impl 2.12

import Proton 4.0

Rectangle {
    id: root

    property ColorScheme colorScheme
    property var backend

    color: root.colorScheme.background_norm

    ColumnLayout {
        anchors.centerIn: root

        //width: 320

        spacing: 20

        Label {
            Layout.bottomMargin: 12;
            Layout.alignment: Qt.AlignHCenter;

            colorScheme: root.colorScheme;
            text: "What's new in Bridge"
            type: Label.Heading
            horizontalAlignment: Text.AlignCenter
        }

        Repeater {
            model: ListModel {
                ListElement { icon: "ic-illustrative-view-html-code" ; title: qsTr("New interface") ; description: qsTr("Entirely redesigned GUI with more intuitive setup.")}
                ListElement { icon: "ic-card-identity"               ; title: qsTr("Status view")   ; description: qsTr("Important notifications and available storage at a glance.")}
                ListElement { icon: "ic-drive"                       ; title: qsTr("Local cache")   ; description: qsTr("New and improved cache for major stability and performance enhancements.")}
            }

            Item {
                implicitWidth: children[0].implicitWidth
                implicitHeight: children[0].implicitHeight

                RowLayout {
                    id: row
                    spacing: 25

                    Item {
                        Layout.topMargin: itemTitle.height/2
                        Layout.alignment: Qt.AlignTop
                        Layout.preferredWidth: 24
                        Layout.preferredHeight: 24

                        ColorImage {
                            anchors.top: parent.top
                            anchors.left: parent.left

                            color: root.colorScheme.interaction_norm
                            source: "./icons/"+model.icon+".svg"
                            width: parent.width
                            sourceSize.width: parent.width
                        }
                    }

                    ColumnLayout {
                        spacing: 0

                        Label {
                            id: itemTitle
                            colorScheme: root.colorScheme
                            text: model.title
                            type: Label.Body_bold
                        }

                        Label {
                            Layout.preferredWidth: 320
                            colorScheme: root.colorScheme
                            text: model.description
                            wrapMode: Text.WordWrap
                        }
                    }
                }
            }

        }

        Item {
            Layout.alignment: Qt.AlignHCenter;
            implicitWidth: children[0].width
            implicitHeight: children[0].height

            RowLayout {
                spacing: 10

                Label {
                    colorScheme: root.colorScheme;
                    text: qsTr("Full release notes")
                    Layout.alignment: Qt.AlignHCenter
                    type: Label.LabelType.Body
                    onLinkActivated: Qt.openUrlExternally(link)
                    color: root.colorScheme.interaction_norm
                }

                ColorImage {
                    color: root.colorScheme.interaction_norm
                    source: "./icons/ic-external-link.svg"
                    width: 16
                    sourceSize.width: 16
                }

            }

            MouseArea {
                anchors.fill: parent
                cursorShape: Qt.PointingHandCursor
                onClicked: Qt.openUrlExternally(root.backend.releaseNotesLink)
            }
        }


        Button {
            Layout.topMargin: 12
            Layout.fillWidth: true
            colorScheme: root.colorScheme
            text: "Start using Bridge"
            onClicked: root.backend.showSplashScreen = false
        }
    }
}

