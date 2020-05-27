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

// Checkbox row for folder selection
import QtQuick 2.8
import QtQuick.Controls 2.2
import ProtonUI 1.0
import ImportExportUI 1.0

AccessibleButton {
    id: root

    property bool   isSection  : false
    property bool   isSelected : false
    property string title  : "N/A"
    property string type   : ""
    property color  color  : "black"

    height    : Style.exporting.rowHeight
    padding   : 0.0
    anchors {
        horizontalCenter: parent.horizontalCenter
    }

    background: Rectangle {
        color: isSection ? Style.exporting.background : Style.exporting.rowBackground
        Rectangle { // line
            anchors.bottom : parent.bottom
            height         : Style.dialog.borderInput
            width          : parent.width
            color          : Style.exporting.background
        }
    }

    contentItem: Rectangle {
        color: "transparent"
        id: content
        Text {
            id: checkbox
            anchors {
                verticalCenter : parent.verticalCenter
                left           : content.left
                leftMargin     : Style.exporting.leftMargin * (root.type == gui.enums.folderTypeSystem   ? 1.0 : 2.0)
            }
            font {
                family : Style.fontawesome.name
                pointSize : Style.dialog.fontSize * Style.pt
            }
            color : isSelected ? Style.main.text : Style.main.textInactive
            text  : (isSelected ? Style.fa.check_square_o : Style.fa.square_o )
        }

        Text { // icon
            id: folderIcon
            visible: !isSection
            anchors {
                verticalCenter : parent.verticalCenter
                left           : checkbox.left
                leftMargin     : Style.dialog.fontSize + Style.exporting.leftMargin
            }
            color : root.type==gui.enums.folderTypeSystem ? Style.main.textBlue : root.color
            font {
                family : Style.fontawesome.name
                pointSize : Style.dialog.fontSize * Style.pt
            }
            text  : {
                return gui.folderIcon(root.title.toLowerCase(), root.type)
            }
        }

        Text {
            text: root.title
            anchors {
                verticalCenter : parent.verticalCenter
                left           : isSection ? checkbox.left : folderIcon.left
                leftMargin     : Style.dialog.fontSize + Style.exporting.leftMargin
            }
            font {
                pointSize : Style.dialog.fontSize * Style.pt
                bold: isSection
            }
            color: Style.exporting.text
        }
    }
}
