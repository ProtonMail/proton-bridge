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

// input for date range

import QtQuick 2.8
import QtQuick.Controls 2.2
import ProtonUI 1.0

CheckBox {
    id: root
    spacing: Style.dialog.spacing
    padding: 0
    property color textColor        : Style.main.text
    property color checkedColor     : Style.main.textBlue
    property color uncheckedColor   : Style.main.textInactive
    property string checkedSymbol   : Style.fa.check_square_o
    property string uncheckedSymbol : Style.fa.square_o
    background: Rectangle {
        color: Style.transparent
    }
    indicator: Text {
        text  : root.checked ? root.checkedSymbol : root.uncheckedSymbol
        color : root.checked ? root.checkedColor  : root.uncheckedColor
        font {
            pointSize : Style.dialog.iconSize * Style.pt
            family    : Style.fontawesome.name
        }
    }
    contentItem: Text {
        id: label
        text  : root.text
        color : root.textColor
        font {
            pointSize: Style.dialog.fontSize * Style.pt
        }
        horizontalAlignment: Text.AlignHCenter
        verticalAlignment: Text.AlignVCenter
        leftPadding: Style.dialog.iconSize + root.spacing
    }
}
