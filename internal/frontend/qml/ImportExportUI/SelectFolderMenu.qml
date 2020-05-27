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

// This is global combo box which can be adjusted to choose folder target, folder label or global label
import QtQuick 2.8
import QtQuick.Controls 2.2
import ProtonUI 1.0
import ImportExportUI 1.0

ComboBox {
    id: root
    //fixme rounded
    height: Style.main.fontSize*2 //fixme
    property string folderType: gui.enums.folderTypeFolder
    property string selectedIDs
    property string sourceID
    property bool isFolderType: root.folderType == gui.enums.folderTypeFolder
    property bool hasTarget: root.selectedIDs != ""
    property bool below: true

    signal doNotImport()
    signal importToFolder(string newTargetID)

    leftPadding: Style.dialog.spacing

    onDownChanged : {
        if (root.down) view.model.updateFilter()
        root.below = popup.y>0
    }

    contentItem : Text {
        id: boxText
        verticalAlignment: Text.AlignVCenter
        font {
            family: Style.fontawesome.name
            pointSize : Style.dialog.fontSize * Style.pt
            bold: root.down
        }
        elide: Text.ElideRight
        textFormat: Text.StyledText

        text : root.displayText
        color: !root.enabled ? Style.main.textDisabled : ( root.down ? Style.main.background : Style.main.text )
    }

    displayText: {
        //console.trace()
        //console.log("updatebox", view.currentIndex, root.hasTarget, root.selectedIDs, root.sourceID, root.folderType)
        if (!root.hasTarget) {
            if (root.isFolderType) return qsTr("Do not import")
            return qsTr("No labels selected")
        }
        if (!root.isFolderType) return Style.fa.tags + " " + qsTr("Add/Remove labels")

        // We know here that it has a target and this is folder dropdown so we must find the first folder
        var selSplit = root.selectedIDs.split(";")
        for (var selIndex in selSplit) {
            var selectedID = selSplit[selIndex]
            var selectedType = structurePM.getType(selectedID)
            if (selectedType == gui.enums.folderTypeLabel) continue; // skip type::labele
            var selectedName = structurePM.getName(selectedID)
            if (selectedName == "") continue; // empty name seems like wrong ID
            var icon = gui.folderIcon(selectedName, selectedType)
            if (selectedType == gui.enums.folderTypeSystem) {
                return icon + " " + selectedName
            }
            var iconColor = structurePM.getColor(selectedID)
            return '<font color="'+iconColor+'">'+ icon + "</font> " + selectedName
        }
        return ""
    }


    background : RoundedRectangle {
        fillColor         : root.down ? Style.main.textBlue : Style.transparent
        strokeColor       : root.down ? fillColor : Style.main.line
        radiusTopLeft     : root.down && !root.below ? 0 : Style.dialog.radiusButton
        radiusBottomLeft  : root.down && root.below ? 0 : Style.dialog.radiusButton
        radiusTopRight    : radiusTopLeft
        radiusBottomRight : radiusBottomLeft

        MouseArea {
            anchors.fill: parent
            onClicked : {
                if (root.down) root.popup.close()
                else root.popup.open()
            }
        }
    }

    indicator : Text {
        text: (root.down && root.below) || (!root.down && !root.below) ? Style.fa.chevron_up : Style.fa.chevron_down
        anchors {
            right: parent.right
            verticalCenter: parent.verticalCenter
            rightMargin: Style.dialog.spacing
        }
        font {
            family : Style.fontawesome.name
            pointSize : Style.dialog.fontSize * Style.pt
        }
        color: root.enabled && !root.down ? Style.main.textBlue : root.contentItem.color
    }

    // Popup objects
    delegate: Rectangle {
        id: thisDelegate

        height : Style.main.fontSize * 2
        width  : selectNone.width

        property bool isHovered: area.containsMouse

        color: isHovered ? root.popup.hoverColor : root.popup.backColor


        property bool isSelected : {
            var selected = root.selectedIDs.split(";")
            for (var iSel in selected) {
                var sel = selected[iSel]
                if (folderId == sel){
                    return true
                }
            }
            return false
        }

        Text {
            id: targetIcon
            text: gui.folderIcon(folderName,folderType)
            color : folderType != gui.enums.folderTypeSystem ? folderColor : root.popup.textColor
            anchors {
                verticalCenter: parent.verticalCenter
                left: parent.left
                leftMargin: root.leftPadding
            }
            font {
                family : Style.fontawesome.name
                pointSize : Style.dialog.fontSize * Style.pt
            }
        }

        Text {
            id: targetName

            anchors {
                verticalCenter: parent.verticalCenter
                left: targetIcon.right
                right: parent.right
                leftMargin: Style.dialog.spacing
                rightMargin: Style.dialog.spacing
            }

            text: folderName
            color : root.popup.textColor
            elide: Text.ElideRight

            font {
                family : Style.fontawesome.name
                pointSize : Style.dialog.fontSize * Style.pt
            }
        }

        Text {
            id: targetIndicator
            anchors {
                right: parent.right
                verticalCenter: parent.verticalCenter
            }

            text    : thisDelegate.isSelected ? Style.fa.check_square : Style.fa.square_o
            visible : thisDelegate.isSelected || !root.isFolderType
            color   : root.popup.textColor
            font {
                family : Style.fontawesome.name
                pointSize : Style.dialog.fontSize * Style.pt
            }
        }

        Rectangle {
            id: line
            anchors {
                bottom : parent.bottom
                left   : parent.left
                right  : parent.right
            }
            height : Style.main.lineWidth
            color  : Style.main.line
        }

        MouseArea {
            id: area
            anchors.fill: parent

            onClicked: {
                //console.log(" click delegate")
                if (root.isFolderType) { // don't update if selected
                    if (!thisDelegate.isSelected) {
                        root.importToFolder(folderId)
                    }
                    root.popup.close()
                }
                if (root.folderType==gui.enums.folderTypeLabel) {
                    if (thisDelegate.isSelected) {
                        structureExternal.removeTargetLabelID(sourceID,folderId)
                    } else  {
                        structureExternal.addTargetLabelID(sourceID,folderId)
                    }
                }
            }
            hoverEnabled: true
        }
    }

    popup : Popup {
        y: root.height
        width: root.width
        modal: true
        closePolicy: Popup.CloseOnPressOutside | Popup.CloseOnEscape
        padding: Style.dialog.spacing

        property var textColor  : Style.main.background
        property var backColor  : Style.main.text
        property var hoverColor : Style.main.textBlue

        contentItem : Column {
            // header
            Rectangle {
                id: selectNone
                width: root.popup.width - 2*root.popup.padding
                //height: root.isFolderType ? 2* Style.main.fontSize : 0
                height: 2*Style.main.fontSize
                color: area.containsMouse ? root.popup.hoverColor : root.popup.backColor
                visible : root.isFolderType

                Text {
                    anchors {
                        left           : parent.left
                        leftMargin     : Style.dialog.spacing
                        verticalCenter : parent.verticalCenter
                    }
                    text: root.isFolderType ? qsTr("Do not import") : ""
                    color: root.popup.textColor
                    font {
                        pointSize: Style.dialog.fontSize * Style.pt
                        bold: true
                    }
                }

                Rectangle {
                    id: line
                    anchors {
                        bottom : parent.bottom
                        left   : parent.left
                        right  : parent.right
                    }
                    height : Style.dialog.borderInput
                    color  : Style.main.line
                }

                MouseArea {
                    id: area
                    anchors.fill: parent
                    onClicked: {
                        //console.log(" click no set")
                        root.doNotImport()
                        root.popup.close()
                    }
                    hoverEnabled: true
                }
            }

            // scroll area
            Rectangle {
                width: selectNone.width
                height: winMain.height/4
                color: root.popup.backColor

                ListView {
                    id: view

                    clip         : true
                    anchors.fill : parent

                    section.property : "sectionName"
                    section.delegate : Text{text: sectionName}

                    model : FilterStructure {
                        filterOnGroup : root.folderType
                        delegate      : root.delegate
                    }
                }
            }

            // footer
            Rectangle {
                id: addFolderOrLabel
                width: selectNone.width
                height: addButton.height + 3*Style.dialog.spacing
                color: root.popup.backColor

                Rectangle {
                    anchors {
                        top   : parent.top
                        left  : parent.left
                        right : parent.right
                    }
                    height : Style.dialog.borderInput
                    color  : Style.main.line
                }

                ButtonRounded {
                    id: addButton
                    anchors.centerIn: addFolderOrLabel
                    width: parent.width * 0.681

                    fa_icon    : Style.fa.plus_circle
                    text       : root.isFolderType ? qsTr("Create new folder") : qsTr("Create new label")
                    color_main : root.popup.textColor
                }

                MouseArea {
                    anchors.fill : parent

                    onClicked  : {
                        //console.log("click", addButton.text)
                        var newName = ""
                        if ( typeof folderName !== 'undefined' && !structurePM.hasFolderWithName (folderName) ) {
                            newName = folderName
                        }
                        winMain.popupFolderEdit.show(newName, "", "", root.folderType, sourceID)
                        root.popup.close()
                    }
                }
            }
        }

        background : RoundedRectangle {
            strokeColor       : root.popup.backColor
            fillColor         : root.popup.backColor
            radiusTopLeft     : root.below  ? 0 : Style.dialog.radiusButton
            radiusBottomLeft  : !root.below ? 0 : Style.dialog.radiusButton
            radiusTopRight    : radiusTopLeft
            radiusBottomRight : radiusBottomLeft
        }
    }
}

