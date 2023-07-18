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
import QtQuick
import QtQuick.Layouts
import QtQuick.Controls
import QtQuick.Controls.impl
import Proton

Item {
    id: root

    property string address
    property ColorScheme colorScheme
    property var user

    signal dismissed
    signal finished

    function reset() {
        guidePages.currentIndex = 0;
        clientList.currentIndex = -1;
        actionList.currentIndex = -1;
    }
    function setupAction(actionID, clientID) {
        if (user) {
            user.setupGuideSeen = true;
        }
        switch (actionID) {
        case -1:
            root.dismissed();
            break; // dismiss
        case 0 // automatic
        :
            if (user) {
                switch (clientID) {
                case 0:
                    root.user.configureAppleMail(root.address);
                    Backend.notifyAutoconfigClicked("AppleMail");
                    break;
                }
            }
            root.finished();
            break;
        case 1 // manual
        :
            let clientObj = clients.get(clientID);
            if (clientObj !== undefined && clientObj.link !== "") {
                Qt.openUrlExternally(clientObj.link);
                Backend.notifyKBArticleClicked(clientObj.link);
            } else {
                console.log("unexpected client index", actionID, clientID);
            }
            root.finished();
            break;
        default:
            console.log("unexpected client setup action", actionID, clientID);
        }
    }

    implicitHeight: children[0].implicitHeight
    implicitWidth: children[0].implicitWidth

    ListModel {
        id: clients

        property bool haveAutoSetup: true
        property string iconSource: "/qml/icons/ic-apple-mail.svg"
        property string link: "https://proton.me/support/protonmail-bridge-clients-apple-mail"
        property string name: "Apple Mail"

        Component.onCompleted: {
            if (Backend.goos === "darwin") {
                append({
                        "name": "Apple Mail",
                        "iconSource": "/qml/icons/ic-apple-mail.svg",
                        "haveAutoSetup": true,
                        "link": "https://proton.me/support/protonmail-bridge-clients-apple-mail"
                    });
                append({
                        "name": "Microsoft Outlook",
                        "iconSource": "/qml/icons/ic-microsoft-outlook.svg",
                        "haveAutoSetup": false,
                        "link": "https://proton.me/support/protonmail-bridge-clients-macos-outlook-2019"
                    });
            }
            if (Backend.goos === "windows") {
                append({
                        "name": "Microsoft Outlook",
                        "iconSource": "/qml/icons/ic-microsoft-outlook.svg",
                        "haveAutoSetup": false,
                        "link": "https://proton.me/support/protonmail-bridge-clients-windows-outlook-2019"
                    });
            }
            append({
                    "name": "Mozilla Thunderbird",
                    "iconSource": "/qml/icons/ic-mozilla-thunderbird.svg",
                    "haveAutoSetup": false,
                    "link": "https://proton.me/support/protonmail-bridge-clients-windows-thunderbird"
                });
            append({
                    "name": "Other",
                    "iconSource": "/qml/icons/ic-other-mail-clients.svg",
                    "haveAutoSetup": false,
                    "link": "https://proton.me/support/protonmail-bridge-configure-client"
                });
        }
    }
    Rectangle {
        anchors.fill: root
        color: root.colorScheme.background_norm
    }
    StackLayout {
        id: guidePages
        anchors.bottomMargin: 70
        anchors.fill: parent
        anchors.leftMargin: 80
        anchors.rightMargin: 80
        anchors.topMargin: 30

        ColumnLayout {
            // 0: Client selection
            id: clientView

            property int columnWidth: 268

            Layout.fillHeight: true
            spacing: 8

            Label {
                colorScheme: root.colorScheme
                text: qsTr("Setting up email client")
                type: Label.LabelType.Heading
            }
            Label {
                color: root.colorScheme.text_weak
                colorScheme: root.colorScheme
                text: address
                type: Label.LabelType.Lead
            }
            RowLayout {
                Layout.topMargin: 32 - clientView.spacing
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
                        model: clients
                        width: clientView.columnWidth

                        delegate: Item {
                            implicitHeight: clientRow.height
                            implicitWidth: clientRow.width

                            ColumnLayout {
                                id: clientRow
                                width: clientList.width

                                RowLayout {
                                    Layout.bottomMargin: 12
                                    Layout.leftMargin: 16
                                    Layout.rightMargin: 16
                                    Layout.topMargin: 12

                                    ColorImage {
                                        height: 36
                                        source: model.iconSource
                                        sourceSize.height: 36
                                    }
                                    Label {
                                        Layout.leftMargin: 12
                                        colorScheme: root.colorScheme
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
                                    clientList.currentIndex = index;
                                    if (!model.haveAutoSetup) {
                                        root.setupAction(1, index);
                                    }
                                }
                            }
                        }
                        highlight: Rectangle {
                            color: root.colorScheme.interaction_default_active
                            radius: ProtonStyle.context_item_radius
                        }
                    }
                }
                ColumnLayout {
                    id: actionColumn
                    Layout.alignment: Qt.AlignTop
                    visible: clientList.currentIndex >= 0 && clients.get(clientList.currentIndex).haveAutoSetup

                    Label {
                        colorScheme: root.colorScheme
                        text: qsTr("Choose configuration mode")
                        type: Label.LabelType.Body_semibold
                    }
                    ListView {
                        id: actionList
                        Layout.fillHeight: true
                        model: [qsTr("Configure automatically"), qsTr("Configure manually")]
                        width: clientView.columnWidth

                        delegate: Item {
                            implicitHeight: children[0].height
                            implicitWidth: children[0].width

                            ColumnLayout {
                                width: actionList.width

                                Label {
                                    Layout.bottomMargin: 20
                                    Layout.leftMargin: 16
                                    Layout.rightMargin: 16
                                    Layout.topMargin: 20
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
                                    actionList.currentIndex = index;
                                    root.setupAction(index, clientList.currentIndex);
                                }
                            }
                        }
                        highlight: Rectangle {
                            color: root.colorScheme.interaction_default_active
                            radius: ProtonStyle.context_item_radius
                        }
                    }
                }
            }
            Item {
                Layout.fillHeight: true
            }
            Button {
                colorScheme: root.colorScheme
                flat: true
                text: qsTr("Set up later")

                onClicked: {
                    root.setupAction(-1, -1);
                    if (user) {
                        user.setupGuideSeen = true;
                    }
                    root.dismissed();
                }
            }
        }
    }
}
