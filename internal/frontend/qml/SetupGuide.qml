// Copyright (c) 2021 Proton Technologies AG
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


import QtQuick 2.13
import QtQuick.Layouts 1.12
import QtQuick.Controls 2.12
import QtQuick.Controls.impl 2.12

import Proton 4.0

Rectangle {
    id:root

    property ColorScheme colorScheme
    property var backend
    property var user
    property string address

    signal dismissed()

    implicitHeight: children[0].implicitHeight
    implicitWidth: children[0].implicitWidth

    color: root.colorScheme.background_norm

    RowLayout {
        anchors.fill: parent
        spacing: 0

        ColumnLayout {
            Layout.fillHeight: true
            Layout.leftMargin: 80
            Layout.rightMargin: 80
            Layout.topMargin: 30
            Layout.bottomMargin: 70
            spacing: 0

            Label {
                colorScheme: root.colorScheme
                text: qsTr("Set up email client")
                type: Label.LabelType.Heading
            }

            Label {
                colorScheme: root.colorScheme
                text: address
                color: root.colorScheme.text_weak
                type: Label.LabelType.Lead
            }

            Label {
                colorScheme: root.colorScheme
                Layout.topMargin: 32
                text: qsTr("Choose an email client")
                type: Label.LabelType.Body_semibold
            }

            ListModel {
                id: clients
                ListElement{name : "Apple Mail"          ; iconSource : "./icons/ic-apple-mail.svg"          }
                ListElement{name : "Microsoft Outlook"   ; iconSource : "./icons/ic-microsoft-outlook.svg"   }
                ListElement{name : "Mozilla Thunderbird" ; iconSource : "./icons/ic-mozilla-thunderbird.svg" }
                ListElement{name : "Other"               ; iconSource : "./icons/ic-other-mail-clients.svg"  }
            }


            Repeater {
                model: clients

                Rectangle {
                    implicitWidth: clientRow.width
                    implicitHeight: clientRow.height

                    ColumnLayout {
                        id: clientRow

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
                        onClicked: {
                            if (model.name != "Apple Mail") {
                                console.log(" TODO configure ", model.name)
                                return
                            }
                            if (user) {
                                root.user.configureAppleMail(root.address)
                            }
                            root.dismissed()
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
                    if (user) {
                        user.setupGuideSeen = true
                    }
                    root.dismissed()
                }
            }
        }
    }
}
