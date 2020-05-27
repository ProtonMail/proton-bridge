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

// Import report modal
import QtQuick 2.11
import QtQuick.Controls 2.4
import ProtonUI 1.0
import ImportExportUI 1.0


Rectangle {
    id: root

    property alias text      : cellText.text
    property bool  isHeader  : false
    property bool  isHovered : false
    property bool  isWider   : cellText.contentWidth > root.width

    width  : 20*Style.px
    height : cellText.height
    z      : root.isHovered ? 3 : 1
    color  : Style.transparent

    Rectangle {
        anchors {
            fill    : cellText
            margins : -2*Style.px
        }
        color : root.isWider ? Style.main.background : Style.transparent
        border {
            color : root.isWider ? Style.main.textDisabled : Style.transparent
            width : Style.px
        }
    }

    Text {
        id: cellText
        color : root.isHeader  ? Style.main.textDisabled : Style.main.text
        elide : root.isHovered ? Text.ElideNone          : Text.ElideRight
        width : root.isHovered ? cellText.contentWidth   : root.width
        font {
            pointSize : Style.main.textSize * Style.pt
            family    : Style.fontawesome.name
        }
    }

    MouseArea {
        anchors.fill : root
        hoverEnabled : !root.isHeader
        onEntered    : { root.isHovered = true  }
        onExited     : { root.isHovered = false }
    }
}
