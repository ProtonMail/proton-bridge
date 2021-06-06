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
import QtQuick.Controls.impl 2.12

import Proton 4.0

RowLayout {
    id:root

    property var colorScheme
    property var window

    property var user: { "username": "janedoe@protonmail.com" }

    ColumnLayout {
        Layout.fillHeight: true
        Layout.leftMargin: 80
        Layout.rightMargin: 80
        Layout.topMargin: 30
        Layout.bottomMargin: 70

        ProtonLabel {
            text: qsTr("Set up email client")
            font.weight: ProtonStyle.fontWidth_700
            state: "heading"
        }

        ProtonLabel {
            text: user.username
            color: root.colorScheme.text_weak
            state: "lead"
        }

        ProtonLabel {
            Layout.topMargin: 32
            text: qsTr("Choose an email client")
            font.weight: ProtonStyle.fontWidth_600
            state: "body"
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

            ColumnLayout {
                RowLayout {
                    Layout.topMargin: 12
                    Layout.bottomMargin: 12
                    Layout.leftMargin: 16
                    Layout.rightMargin: 16

                    IconLabel {
                        icon.source: model.iconSource
                        icon.height: 36
                    }

                    ProtonLabel {
                        Layout.leftMargin: 12
                        text: model.name
                        state: "body"
                    }
                }

                Rectangle {
                    Layout.fillWidth: true
                    Layout.preferredHeight: 1
                    color: root.colorScheme.border_weak
                }
            }
        }

        Item { Layout.fillHeight: true }

        Button {
            text: qsTr("Set up later")
            flat: true

            onClicked: {
                root.window.showSetup = false
                root.reset()
            }
        }
    }
}
