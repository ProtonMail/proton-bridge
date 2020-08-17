// Copyright (c) 2020 Proton Technologies AG
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

// credits

import QtQuick 2.8
import ProtonUI 1.0
import ImportExportUI 1.0

Item {
    id: root
    Rectangle {
        anchors.centerIn: parent
        width: Style.main.width
        height: root.parent.height - 6*Style.dialog.titleSize
        color: "transparent"

        ListView {
            anchors.fill: parent
            clip: true
            model: go.credits.split(";")

            delegate: AccessibleText {
                anchors.horizontalCenter: parent.horizontalCenter
                text: modelData
                color: Style.main.text
                font.pointSize:  Style.main.fontSize * Style.pt
            }

            footer: ButtonRounded {
                anchors.horizontalCenter: parent.horizontalCenter
                text: qsTr("Close", "close window")
                onClicked: dialogCredits.hide()
            }
        }
    }
}
