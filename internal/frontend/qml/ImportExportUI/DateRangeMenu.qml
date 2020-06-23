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
    id:root

    width  : icon.width + indicator.width + 3*padding
    height : icon.height + 3*padding

    property real padding : Style.dialog.spacing
    property bool down    : popup.visible

    property var  structure   : transferRules
    property string  sourceID : ""
    property int  sourceFromDate : 0
    property int  sourceToDate : 0

    color: Style.transparent

    RoundedRectangle {
        anchors.fill: parent
        radiusTopLeft: root.down ? 0 : Style.dialog.radiusButton
        fillColor: root.down ? Style.main.textBlue : Style.transparent
    }

    Text {
        id: icon
        text: Style.fa.calendar_o
        anchors {
            left           : parent.left
            leftMargin     : root.padding
            verticalCenter : parent.verticalCenter
        }

        color: root.enabled ? (
            root.down ? Style.main.background : Style.main.text
        ) : Style.main.textDisabled

        font.family : Style.fontawesome.name

        Text {
            anchors {
                verticalCenter: parent.bottom
                horizontalCenter: parent.right
            }

            color          : !root.down && root.enabled ? Style.main.textRed : icon.color
            text           : Style.fa.exclamation_circle
            visible        : !dateRangeInput.allDates
            font.pointSize : root.padding * Style.pt * 1.5
            font.family    : Style.fontawesome.name
        }
    }


    Text {
        id: indicator
        anchors {
            right          : parent.right
            rightMargin    : root.padding
            verticalCenter : parent.verticalCenter
        }

        text  : root.down ? Style.fa.chevron_up : Style.fa.chevron_down
        color : !root.down && root.enabled ? Style.main.textBlue : icon.color
        font.family : Style.fontawesome.name
    }

    MouseArea {
        anchors.fill: root
        onClicked: {
            popup.open()
        }
    }

    Popup {
        id: popup

        x      : -width
        modal  : true
        clip   : true

        topPadding : 0

        background: RoundedRectangle {
            fillColor   : Style.bubble.paneBackground
            strokeColor : fillColor
            radiusTopRight: 0

            RoundedRectangle {
                anchors {
                    left: parent.left
                    right: parent.right
                    top: parent.top
                }
                height: Style.dialog.heightInput
                fillColor: Style.dropDownDark.highlight
                strokeColor: fillColor
                radiusTopRight: 0
                radiusBottomLeft: 0
                radiusBottomRight: 0
            }
        }

        contentItem : Column {
            spacing: Style.dialog.spacing

            Text {
                anchors {
                    left: parent.left
                }

                text              : qsTr("Import date range")
                font.bold         : Style.dropDownDark.labelBold
                color             : Style.dropDownDark.text
                height            : Style.dialog.heightInput
                verticalAlignment : Text.AlignVCenter
            }

            DateRange {
                id: dateRangeInput
                allDates: true
                structure: root.structure
                sourceID: root.sourceID
                dropDownStyle: Style.dropDownDark
            }
        }

        onAboutToShow : updateRange()
        onAboutToHide : dateRangeInput.applyRange()
    }

    function updateRange() {
        dateRangeInput.setRangeFromTo(root.sourceFromDate, root.sourceToDate)
    }

    Connections {
        target:root
        onSourceFromDateChanged: root.updateRange()
        onSourceToDateChanged: root.updateRange()
    }
}
