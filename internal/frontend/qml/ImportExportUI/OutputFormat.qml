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

import QtQuick 2.8
import QtQuick.Controls 2.2
import ProtonUI 1.0
import ImportExportUI 1.0

Column {
    spacing: Style.dialog.spacing
    property string checkedText : group.checkedButton.text

    Text {
        id: formatLabel
        font {
            pointSize: Style.dialog.fontSize * Style.pt
            bold: true
        }
        color: Style.dialog.text
        text: qsTr("Select format of exported email:")

        InfoToolTip {
            info: qsTr("MBOX exports one file for each folder", "todo") + "\n" + qsTr("EML exports one file for each email", "todo")
            anchors {
                left: parent.right
                leftMargin: Style.dialog.spacing
                verticalCenter: parent.verticalCenter
            }
        }
    }

    Row {
        spacing : Style.main.leftMargin
        ButtonGroup {
            id: group
        }

        Repeater {
            model: [ "MBOX", "EML" ]
            delegate : RadioButton {
                id: radioDelegate
                checked: modelData=="MBOX"
                width: 5*Style.dialog.fontSize // hack due to bold
                text: modelData
                ButtonGroup.group: group
                spacing: Style.main.spacing
                indicator: Text {
                    text  : radioDelegate.checked ? Style.fa.check_circle : Style.fa.circle_o
                    color : radioDelegate.checked ? Style.main.textBlue   : Style.main.textInactive
                    font {
                        pointSize: Style.dialog.iconSize * Style.pt
                        family: Style.fontawesome.name
                    }
                    anchors.verticalCenter: parent.verticalCenter
                }
                contentItem: Text {
                    text: radioDelegate.text
                    color: Style.main.text
                    font {
                        pointSize: Style.dialog.fontSize * Style.pt
                        bold: checked
                    }
                    horizontalAlignment : Text.AlignHCenter
                    verticalAlignment   : Text.AlignVCenter
                    leftPadding: Style.dialog.iconSize
                }
            }
        }
    }
}
