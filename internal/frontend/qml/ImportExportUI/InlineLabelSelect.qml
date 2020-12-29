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
import ImportExportUI 1.0


Row {
    id: root
    spacing: Style.dialog.spacing

    property alias labelWidth : label.width

    property string labelName     : ""
    property string labelColor    : ""
    property alias  labelSelected : masterLabelCheckbox.checked

    Text {
        id: label
        text : qsTr("Add import label")
        font {
            bold: true
            pointSize: Style.main.fontSize * Style.pt
        }
        color: Style.main.text
        anchors.verticalCenter: parent.verticalCenter
    }

    InfoToolTip {
        info: qsTr( "When master import label is selected then all imported emails will have this label.", "Tooltip text for master import label")
        anchors.verticalCenter: parent.verticalCenter
    }

    CheckBoxLabel {
        id: masterLabelCheckbox
        text                   : ""
        anchors.verticalCenter : parent.verticalCenter
        checkedSymbol          : Style.fa.toggle_on
        uncheckedSymbol        : Style.fa.toggle_off
        uncheckedColor         : Style.main.textDisabled
        symbolPointSize        : Style.dialog.iconSize * Style.pt * 1.1
        spacing                : Style.dialog.spacing*2

        TextMetrics {
            id: metrics
            text: masterLabelCheckbox.checkedSymbol
            font {
                family: Style.fontawesome.name
                pointSize: masterLabelCheckbox.symbolPointSize
            }
        }


        Rectangle {
            color: parent.checked ? dotBackground.color : Style.exporting.sliderBackground
            width: metrics.width*0.9
            height: metrics.height*0.6
            radius: height/2
            z: -1

            anchors {
                left: masterLabelCheckbox.left
                verticalCenter: masterLabelCheckbox.verticalCenter
                leftMargin: 0.05 * metrics.width
            }

            Rectangle {
                id: dotBackground
                color  : Style.exporting.background
                height : parent.height
                width  : height
                radius : height/2
                anchors {
                    left           : parent.left
                    verticalCenter : parent.verticalCenter
                }

            }
        }
    }

    Rectangle {
        // label
        color  : Style.transparent
        radius : Style.dialog.radiusButton
        border {
            color : Style.dialog.line
            width : Style.dialog.borderInput
        }
        anchors.verticalCenter : parent.verticalCenter

        scale: area.pressed ? 0.95 : 1

        width: content.width
        height: content.height


        Row {
            id: content

            spacing   : Style.dialog.spacing
            padding   : Style.dialog.spacing

            anchors.verticalCenter: parent.verticalCenter

            // label icon color
            Text {
                text: Style.fa.tag
                color: root.labelSelected ? root.labelColor : Style.dialog.line
                anchors.verticalCenter: parent.verticalCenter
                font {
                    family: Style.fontawesome.name
                    pointSize: Style.main.fontSize * Style.pt
                }
            }

            TextMetrics {
                id:labelMetrics
                text: root.labelName
                elide: Text.ElideRight
                elideWidth:gui.winMain.width*0.303

                font {
                    pointSize: Style.main.fontSize * Style.pt
                }
            }

            // label text
            Text {
                text: labelMetrics.elidedText
                color: root.labelSelected ? Style.dialog.text : Style.dialog.line
                font: labelMetrics.font
                anchors.verticalCenter: parent.verticalCenter
            }

            // edit icon
            Text {
                text: Style.fa.edit
                color: root.labelSelected ? Style.main.textBlue : Style.dialog.line
                anchors.verticalCenter: parent.verticalCenter
                font {
                    family: Style.fontawesome.name
                    pointSize: Style.main.fontSize * Style.pt
                }
            }
        }

        MouseArea {
            id: area
            anchors.fill: parent
            enabled: root.labelSelected
            onClicked : {
                if (!root.labelSelected) return
                // NOTE: "createLater" is hack
                winMain.popupFolderEdit.show(root.labelName, "createLater", root.labelColor, gui.enums.folderTypeLabel, "")
            }
        }
    }

    function reset(){
        labelColor  = go.leastUsedColor()
        labelName = qsTr("Imported", "default name of global label followed by date") + " " + gui.niceDateTime()
        labelSelected=true
    }

    Connections {
        target: winMain.popupFolderEdit

        onEdited : {
            if (newName!="") root.labelName = newName
            if (newColor!="") root.labelColor = newColor
        }
    }


    /*
     SelectLabelsMenu {
         id: labelMenu
         width       : winMain.width/5
         sourceID    : root.sourceID
         selectedIDs : root.structure.getTargetLabelIDs(root.sourceID)
         anchors.verticalCenter: parent.verticalCenter
     }

     LabelIconList {
         id: iconList
         selectedIDs : root.structure.getTargetLabelIDs(root.sourceID)
         anchors.verticalCenter: parent.verticalCenter
     }


     Connections {
         target: structureExternal
         onDataChanged: {
             iconList.selectedIDs = root.structure.getTargetLabelIDs(root.sourceID)
             labelMenu.selectedIDs = root.structure.getTargetLabelIDs(root.sourceID)
         }
     }

     Connections {
         target: structurePM
         onDataChanged:{
             iconList.selectedIDs = root.structure.getTargetLabelIDs(root.sourceID)
             labelMenu.selectedIDs = root.structure.getTargetLabelIDs(root.sourceID)
         }
     }
     */
}
