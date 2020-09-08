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

// List of import folder and their target
import QtQuick 2.8
import QtQuick.Controls 2.2
import ProtonUI 1.0
import ImportExportUI 1.0

Rectangle {
    id: root
    color: Style.importing.rowBackground
    height: 40
    width: 300
    property real leftMargin1 : folderIcon.x - root.x
    property real leftMargin2 : selectFolder.x - root.x
    property real nameWidth : {
        var available = root.width
        available -= rowPlacement.children.length * rowPlacement.spacing  // spacing between places
        available -= 3*rowPlacement.leftPadding // left, and 2x right
        available -= folderIcon.width
        available -= arrowIcon.width
        available -= dateRangeMenu.width
        return available/3.3 // source folder label, target folder menu, target labels menu, and 0.3x label list
    }
    property real iconWidth : nameWidth*0.3

    property bool isSourceSelected: isActive
    property string lastTargetFolder: "6" // Archive
    property string lastTargetLabels: "" // no flag by default

    property string sourceID   : mboxID
    property string sourceName : name

    Rectangle {
        id: line
        anchors {
            left   : parent.left
            right  : parent.right
            bottom : parent.bottom
        }
        height : Style.main.border * 2
        color  : Style.importing.rowLine
    }

    Row {
        id: rowPlacement
        spacing: Style.dialog.spacing
        leftPadding: Style.dialog.spacing*2
        anchors.verticalCenter : parent.verticalCenter

        CheckBoxLabel {
            id: checkBox
            anchors.verticalCenter : parent.verticalCenter
            checked: root.isSourceSelected

            onClicked: root.toggleImport()
        }

        Text {
            id: folderIcon
            text : gui.folderIcon(root.sourceName, gui.enums.folderTypeFolder)
            anchors.verticalCenter : parent.verticalCenter
            color: root.isSourceSelected ? Style.main.text : Style.main.textDisabled
            font {
                family : Style.fontawesome.name
                pointSize : Style.dialog.fontSize * Style.pt
            }
        }

        Text {
            text : root.sourceName
            width: nameWidth
            elide: Text.ElideRight
            anchors.verticalCenter : parent.verticalCenter
            color: folderIcon.color
            font.pointSize : Style.dialog.fontSize * Style.pt
        }

        Text {
            id: arrowIcon
            text : Style.fa.arrow_right
            anchors.verticalCenter : parent.verticalCenter
            color: Style.main.text
            font {
                family : Style.fontawesome.name
                pointSize : Style.dialog.fontSize * Style.pt
            }
        }

        SelectFolderMenu {
            id: selectFolder
            sourceID: root.sourceID
            targets: transferRules.targetFolders(root.sourceID)
            width: nameWidth
            anchors.verticalCenter : parent.verticalCenter
            enabled: root.isSourceSelected
            onDoNotImport: root.toggleImport()
            onImportToFolder: root.importToFolder(newTargetID)
        }

        SelectLabelsMenu {
            sourceID: root.sourceID
            targets: transferRules.targetLabels(root.sourceID)
            width: nameWidth
            anchors.verticalCenter : parent.verticalCenter
            enabled: root.isSourceSelected
            onAddTargetLabel: { transferRules.addTargetID(sourceID, newTargetID) }
            onRemoveTargetLabel: { transferRules.removeTargetID(sourceID, newTargetID) }
        }

        LabelIconList {
            colorList: labelColors=="" ? [] : labelColors.split(";")
            width: iconWidth
            anchors.verticalCenter : parent.verticalCenter
            enabled: root.isSourceSelected
        }

        DateRangeMenu {
            id: dateRangeMenu
            sourceID: root.sourceID
            sourceFromDate: fromDate
            sourceToDate: toDate

            enabled: root.isSourceSelected
            anchors.verticalCenter : parent.verticalCenter

            Component.onCompleted : dateRangeMenu.updateRange()
        }
    }


    function importToFolder(newTargetID) {
        transferRules.addTargetID(root.sourceID,newTargetID)
    }

    function toggleImport() {
        transferRules.setIsRuleActive(root.sourceID, !root.isSourceSelected)
    }
}
