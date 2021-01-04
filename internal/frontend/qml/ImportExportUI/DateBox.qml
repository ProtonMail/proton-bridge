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

// input for year / month / day
import QtQuick 2.8
import QtQuick.Controls 2.2
import QtQml.Models 2.2
import ProtonUI 1.0
import ImportExportUI 1.0

ComboBox {
    id: root

    property string placeholderText : "none"
    property var    dropDownStyle   : Style.dropDownLight
    property real   radius          : Style.dialog.radiusButton
    property bool   below           : true

    onDownChanged : {
        root.below = popup.y>0
    }


    font.pointSize : Style.main.fontSize * Style.pt

    spacing : Style.dialog.spacing
    height  : Style.dialog.heightInput
    width   : 10*Style.px

    function updateWidth() {
        // make the width according to localization ( especially for Months)
        var max = 10*Style.px
        if (root.model === undefined) return
        for (var i=-1; i<root.model.length; ++i){
            metrics.text = i<0 ? root.placeholderText : root.model[i]+"MM" // "M" for extra space
            max = Math.max(max, metrics.width)
        }
        root.width = root.spacing + max + root.spacing + indicatorIcon.width + root.spacing
        //console.log("width updated", root.placeholderText, root.width)
    }

    TextMetrics {
        id: metrics
        font: root.font
        text: placeholderText
    }


    indicator: Text {
        id: indicatorIcon
        color: root.enabled  ? dropDownStyle.highlight : dropDownStyle.inactive
        text: root.down ? Style.fa.chevron_up : Style.fa.chevron_down
        font.family: Style.fontawesome.name
        anchors {
            right: parent.right
            rightMargin: root.spacing
            verticalCenter: parent.verticalCenter
        }
    }

    contentItem: Text {
        id: boxItem
        leftPadding: root.spacing
        rightPadding: root.spacing

        text              : enabled && root.currentIndex>=0 ? root.displayText : placeholderText
        font              : root.font
        color             : root.enabled ? dropDownStyle.text : dropDownStyle.inactive
        verticalAlignment : Text.AlignVCenter
        elide             : Text.ElideRight
    }

    background: Rectangle {
        color: Style.transparent

        MouseArea {
            anchors.fill: parent
            onClicked: root.down ?  root.popup.close() : root.popup.open()
        }
    }


    DelegateModel { // FIXME QML DelegateModel: Error creating delegate
        id: filteredData
        model: root.model
        filterOnGroup: "filtered"
        groups: DelegateModelGroup {
            id: filtered
            name: "filtered"
            includeByDefault: true
        }
        delegate: root.delegate
    }

    function filterItems(minIndex,maxIndex) {
        // filter
        var rowCount = filteredData.items.count
        if (rowCount<=0) return
        //console.log(" filter", root.placeholderText, rowCount, minIndex, maxIndex)
        for (var iItem = 0; iItem < rowCount; iItem++) {
            var entry = filteredData.items.get(iItem);
            entry.inFiltered = ( iItem >= minIndex && iItem <= maxIndex )
            //console.log("     inserted ", iItem, rowCount, entry.model.modelData, entry.inFiltered )
        }
    }

    delegate: ItemDelegate {
        id: thisItem
        width         : view.width
        height        : Style.dialog.heightInput
        leftPadding   : root.spacing
        rightPadding  : root.spacing
        topPadding    : 0
        bottomPadding : 0

        property int index : {
            //console.log( "index: ", thisItem.DelegateModel.itemsIndex )
            return thisItem.DelegateModel.itemsIndex
        }

        onClicked : {
            //console.log("thisItem click", thisItem.index)
            root.currentIndex = thisItem.index
            root.activated(thisItem.index)
            root.popup.close()
        }


        contentItem: Text {
            text: modelData
            color: dropDownStyle.text
            font: root.font
            elide: Text.ElideRight
            verticalAlignment: Text.AlignVCenter
        }

        background: Rectangle {
            color: thisItem.hovered ? dropDownStyle.highlight : dropDownStyle.background
            Text {
                anchors{
                    right: parent.right
                    rightMargin: root.spacing
                    verticalCenter: parent.verticalCenter
                }
                font {
                    family: Style.fontawesome.name
                }
                text: root.currentIndex == thisItem.index ? Style.fa.check : ""
                color: thisItem.hovered ? dropDownStyle.text : dropDownStyle.highlight
            }

            Rectangle {
                anchors {
                    left: parent.left
                    right: parent.right
                    bottom: parent.bottom
                }
                height: Style.dialog.borderInput
                color: dropDownStyle.separator
            }
        }
    }

    popup: Popup {
        y: root.height
        x: -background.strokeWidth
        width: root.width + 2*background.strokeWidth
        modal: true
        closePolicy: Popup.CloseOnPressOutside | Popup.CloseOnEscape
        topPadding: background.radiusTopLeft + 2*background.strokeWidth
        bottomPadding: background.radiusBottomLeft + 2*background.strokeWidth
        leftPadding: 2*background.strokeWidth
        rightPadding: 2*background.strokeWidth

        contentItem: ListView {
            id: view
            clip: true
            implicitHeight: winMain.height/3
            model: filteredData // if you want to slide down to position: popup.visible ? root.delegateModel : null
            currentIndex: root.currentIndex

            ScrollIndicator.vertical: ScrollIndicator { }
        }

        background: RoundedRectangle {
            radiusTopLeft     : root.below  ?  0 : root.radius
            radiusBottomLeft  : !root.below ?  0 : root.radius
            radiusTopRight    : radiusTopLeft
            radiusBottomRight : radiusBottomLeft
            fillColor         : dropDownStyle.background
        }
    }

    Component.onCompleted: {
        //console.log(" box ", label)
        root.updateWidth()
        root.filterItems(0,model.length-1)
    }

    onModelChanged :{
        //console.log("model changed", root.placeholderText)
        root.updateWidth()
        root.filterItems(0,model.length-1)
    }
}

