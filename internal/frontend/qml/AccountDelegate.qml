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

import Proton 4.0

RowLayout {
    id: root
    property var colorScheme: parent.colorScheme

    property var text: "janedoe@protonmail.com"
    property var avatarText: "jd"
    property var captionText: "50.5 MB / 20 GB"

    spacing: 16

    Rectangle {
        id: avatar
        Layout.preferredHeight: account.height
        Layout.preferredWidth: account.height
        radius: 4

        color: root.colorScheme.background_avatar

        ProtonLabel {
            anchors.centerIn: avatar
            color: root.colorScheme.text_norm
            text: root.avatarText.toUpperCase()
            state: "body"
            horizontalAlignment: Qt.AlignHCenter
            verticalAlignment: Qt.AlignVCenter
        }
    }

    ColumnLayout {
        id: account
        Layout.fillHeight: true
        Layout.fillWidth: true

        ProtonLabel {
            text: root.text
            color: root.colorScheme.text_norm
            state: "body"
        }

        ProtonLabel {
            text: root.captionText
            color: root.colorScheme.text_weak
            state: "caption"
        }
    }
}
