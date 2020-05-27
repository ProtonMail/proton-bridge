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

// Export dialog
import QtQuick 2.8
import QtQuick.Controls 2.2
import ProtonUI 1.0
import ImportExportUI 1.0



Button {
    id: root

    width  : 200
    height : icon.height + 4*tag.height
    scale  : pressed ? 0.95 : 1.0

    property string iconText : Style.fa.ban

    background: Rectangle { color: "transparent" }

    contentItem: Rectangle {
        id: wrapper
        color: "transparent"

        Image {
            id: icon
            anchors {
                bottom           : wrapper.bottom
                bottomMargin     : tag.height*2.5
                horizontalCenter : wrapper.horizontalCenter
            }
            fillMode : Image.PreserveAspectFit
            width    : Style.main.fontSize * 7
            mipmap   : true
            source   : "images/"+iconText+".png"

        }

        Row {
            spacing: Style.dialog.spacing
            anchors {
                bottom           : wrapper.bottom
                horizontalCenter : wrapper.horizontalCenter
            }

            Text {
                id: tag

                text  : Style.fa.plus_circle
                color : Qt.lighter( Style.dialog.textBlue, root.enabled ? 1.0 : 1.5)

                font {
                    family    : Style.fontawesome.name
                    pointSize : Style.main.fontSize * Style.pt * 1.2
                }
            }

            Text {
                text  : root.text
                color: tag.color

                font {
                    family    : tag.font.family
                    pointSize : tag.font.pointSize
                    weight    : Font.DemiBold
                    underline : true
                }
            }
        }
    }
}


