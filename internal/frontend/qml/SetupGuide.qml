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


import QtQuick 2.13
import QtQuick.Layouts 1.12
import QtQuick.Controls 2.12
import QtQuick.Controls.impl 2.12

import Proton 4.0

Item {
    id:root

    property ColorScheme colorScheme
    property var backend
    property var user
    property string address

    signal dismissed()
    signal finished()

    implicitHeight: children[0].implicitHeight
    implicitWidth: children[0].implicitWidth


    ListModel {
        id: clients
        property string name : "Apple Mail"
        property string iconSource : "./icons/ic-apple-mail.svg"
        property bool haveAutoSetup: true
        property string link: "https://protonmail.com/bridge/applemail"

        Component.onCompleted : {
            if (root.backend.goos == "darwin") {
                append({
                    "name"          : "Apple Mail",
                    "iconSource"    : "./icons/ic-apple-mail.svg",
                    "haveAutoSetup" : true,
                    "link"          : "https://protonmail.com/bridge/applemail"
                })
                append({
                    "name"          : "Microsoft Outlook",
                    "iconSource"    : "./icons/ic-microsoft-outlook.svg",
                    "haveAutoSetup" : false,
                    "link"          : "https://protonmail.com/bridge/outlook2019-mac"
                })
            }
            if (root.backend.goos == "windows") {
                append({
                    "name"          : "Microsoft Outlook",
                    "iconSource"    : "./icons/ic-microsoft-outlook.svg",
                    "haveAutoSetup" : false,
                    "link"          : "https://protonmail.com/bridge/outlook2019"
                })
            }

            append({
                "name"          : "Mozilla Thunderbird",
                "iconSource"    : "./icons/ic-mozilla-thunderbird.svg",
                "haveAutoSetup" : false,
                "link"          : "https://protonmail.com/bridge/thunderbird"
            })

            append({
                "name"          : "Other",
                "iconSource"    : "./icons/ic-other-mail-clients.svg",
                "haveAutoSetup" : false,
                "link"          : "https://protonmail.com/bridge/clients"
            })

        }
    }

    Rectangle {
        anchors.fill: root
        color: root.colorScheme.background_norm
    }

    StackLayout {
        id: guidePages
        anchors.fill: parent
        anchors.leftMargin: 80
        anchors.rightMargin: 80
        anchors.topMargin: 30
        anchors.bottomMargin: 70


        ColumnLayout { // 0: Client selection
            id: clientView
            Layout.fillHeight: true

            property int columnWidth: 268

            spacing: 8

            Label {
                colorScheme: root.colorScheme
                text: qsTr("Setting up email client")
                type: Label.LabelType.Heading
            }

            Label {
                colorScheme: root.colorScheme
                text: address
                color: root.colorScheme.text_weak
                type: Label.LabelType.Lead
            }

            RowLayout {
                Layout.topMargin: 32-clientView.spacing
                spacing: 24

                ColumnLayout {
                    id: clientColumn
                    Layout.alignment: Qt.AlignTop

                    Label {
                        id: labelA
                        colorScheme: root.colorScheme
                        text: qsTr("Choose an email client")
                        type: Label.LabelType.Body_semibold
                    }

                    ListView {
                        id: clientList
                        Layout.fillHeight: true
                        width: clientView.columnWidth

                        model: clients

                        highlight: Rectangle {
                            color: root.colorScheme.interaction_default_active
                            radius: ProtonStyle.context_item_radius
                        }

                        delegate: Item {
                            implicitWidth: clientRow.width
                            implicitHeight: clientRow.height

                            ColumnLayout {
                                id: clientRow
                                width: clientList.width

                                RowLayout {
                                    Layout.topMargin: 12
                                    Layout.bottomMargin: 12
                                    Layout.leftMargin: 16
                                    Layout.rightMargin: 16

                                    ColorImage {
                                        source: model.iconSource
                                        height: 36
                                        sourceSize.height: 36
                                    }

                                    Label {
                                        colorScheme: root.colorScheme
                                        Layout.leftMargin: 12
                                        text: model.name
                                        type: Label.LabelType.Body
                                    }
                                }

                                Rectangle {
                                    Layout.fillWidth: true
                                    Layout.preferredHeight: 1
                                    color: root.colorScheme.border_weak
                                }
                            }

                            MouseArea {
                                anchors.fill: parent
                                cursorShape: Qt.PointingHandCursor
                                onClicked: {
                                    clientList.currentIndex = index
                                    if (!model.haveAutoSetup) {
                                        root.setupAction(1,index)
                                    }
                                }
                            }
                        }
                    }
                }

                ColumnLayout {
                    id: actionColumn
                    visible: clientList.currentIndex >= 0 && clients.get(clientList.currentIndex).haveAutoSetup
                    Layout.alignment: Qt.AlignTop

                    Label {
                        colorScheme: root.colorScheme
                        text: qsTr("Choose configuration mode")
                        type: Label.LabelType.Body_semibold
                    }

                    ListView {
                        id: actionList
                        Layout.fillHeight: true
                        width: clientView.columnWidth

                        model: [
                            qsTr("Configure automatically"),
                            qsTr("Configure manually"),
                        ]

                        highlight: Rectangle {
                            color: root.colorScheme.interaction_default_active
                            radius: ProtonStyle.context_item_radius
                        }

                        delegate: Item {
                            implicitWidth: children[0].width
                            implicitHeight: children[0].height

                            ColumnLayout {
                                width: actionList.width

                                Label {
                                    Layout.topMargin: 20
                                    Layout.bottomMargin: 20
                                    Layout.leftMargin: 16
                                    Layout.rightMargin: 16
                                    colorScheme: root.colorScheme
                                    text: modelData
                                    type: Label.LabelType.Body
                                }

                                Rectangle {
                                    Layout.fillWidth: true
                                    Layout.preferredHeight: 1
                                    color: root.colorScheme.border_weak
                                }
                            }

                            MouseArea {
                                anchors.fill: parent
                                cursorShape: Qt.PointingHandCursor
                                onClicked: {
                                    actionList.currentIndex = index
                                    root.setupAction(index,clientList.currentIndex)
                                }
                            }
                        }
                    }
                }
            }

            Item { Layout.fillHeight: true }

            Button {
                colorScheme: root.colorScheme
                text: qsTr("Set up later")
                flat: true

                onClicked: {
                    root.setupAction(-1,-1)
                    if (user) {
                        user.setupGuideSeen = true
                    }
                    root.dismissed()
                }
            }
        }
    }

    function setupAction(actionID,clientID){
        if (user) {
            user.setupGuideSeen = true
        }

        switch (actionID) {
            case -1: root.dismissed(); break; // dismiss
            case 0: // automatic
            if (user) {
                switch (clientID) {
                    case 0:
                    root.user.configureAppleMail(root.address)
                    break;
                }
            }
            root.finished()
            break;
            case 1: // manual
            var clientObj = clients.get(clientID)
            if (clientObj != undefined && clientObj.link != "" ) {
                Qt.openUrlExternally(clientObj.link)
            } else {
                console.log("unexpected client index", actionID, clientID)
            }
            root.finished();
            break;
            default:
            console.log("unexpected client setup action", actionID, clientID)
        }
    }

    function reset(){
        guidePages.currentIndex = 0
        clientList.currentIndex = -1
        actionList.currentIndex = -1
    }
}
