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
import QtQuick.Controls 2.13

import Proton 4.0

ColumnLayout {
    id: root

    property ColorScheme colorScheme
    property var backend

    property alias currentIndex: usersListView.currentIndex
    ListView {
        id: usersListView
        Layout.fillHeight: true
        Layout.preferredWidth: 200

        model: backend.usersTest
        highlightFollowsCurrentItem: true

        delegate: Item {

            implicitHeight: children[0].implicitHeight + anchors.topMargin + anchors.bottomMargin
            implicitWidth: children[0].implicitWidth + anchors.leftMargin + anchors.rightMargin

            width: usersListView.width

            anchors.margins: 10

            Label {
                colorScheme: root.colorScheme
                text: modelData.username
                anchors.margins: 10
                anchors.fill: parent

                MouseArea {
                    anchors.fill: parent
                    onClicked: {
                        usersListView.currentIndex = index
                    }
                }
            }
        }

        highlight: Rectangle {
            color: root.colorScheme.interaction_default_active
        }
    }

    RowLayout {
        Layout.fillWidth: true

        Button {
            colorScheme: root.colorScheme

            text: "+"

            onClicked: {
                var newUserObject = backend.userComponent.createObject(backend)
                newUserObject.username = backend.loginUser.username.length > 0 ? backend.loginUser.username : "test@protonmail.com"
                newUserObject.loggedIn = true
                newUserObject.setupGuideSeen = true // backend.loginUser.setupGuideSeen

                backend.loginUser.username = ""
                backend.loginUser.loggedIn = false
                backend.loginUser.setupGuideSeen = false

                backend.users.append( { object: newUserObject } )
            }
        }
        Button {
            colorScheme: root.colorScheme
            text: "-"

            enabled: usersListView.currentIndex != 0

            onClicked: {
                // var userObject = backend.users.get(usersListView.currentIndex - 1)
                backend.users.remove(usersListView.currentIndex - 1)
                // userObject.deleteLater()
            }
        }
    }
}
