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
import QtQuick.Controls 2.13

import Proton 4.0

ColumnLayout {
    id: root

    property var colorScheme
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
                text: modelData.username
                anchors.margins: 10
                anchors.fill: parent
                color: root.colorScheme.text_norm

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
            text: "+"

            onClicked: {
                var newUserObject = backend.userComponent.createObject(backend, { username: "test@protonmail.com", loggedIn: false } )
                backend.users.append( { object: newUserObject } )
            }
        }
        Button {
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
