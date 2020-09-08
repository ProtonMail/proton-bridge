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
    id: dateRange

    property var   structure : transferRules
    property string sourceID : "-1"

    property alias allDates      : allDatesBox.checked
    property alias inputDateFrom : inputDateFrom
    property alias inputDateTo   : inputDateTo

    property alias labelWidth: label.width

    function getRange() {common.getRange()}
    function applyRange() {common.applyRange()}

    DateRangeFunctions {id:common}

    spacing: Style.dialog.spacing*2

    Text {
        id: label
        anchors.verticalCenter: parent.verticalCenter
        text : qsTr("Date range")
        font {
            bold: true
            family: Style.fontawesome.name
            pointSize: Style.main.fontSize * Style.pt
        }
        color: Style.main.text
    }

    DateInput {
        id: inputDateFrom
        label: ""
        anchors.verticalCenter: parent.verticalCenter
        currentDate: new Date(0) // default epoch start
        maxDate: inputDateTo.currentDate
    }

    Text {
        text : Style.fa.arrows_h
        anchors.verticalCenter: parent.verticalCenter
        horizontalAlignment: Text.AlignHCenter
        verticalAlignment: Text.AlignVCenter
        color: Style.main.text
        font.family: Style.fontawesome.name
    }

    DateInput {
        id: inputDateTo
        label: ""
        anchors.verticalCenter: parent.verticalCenter
        currentDate: new Date() // default now
        minDate: inputDateFrom.currentDate
        isMaxDateToday: true
    }

    CheckBoxLabel {
        id: allDatesBox
        text                   : qsTr("All dates")
        anchors.verticalCenter : parent.verticalCenter
        checkedSymbol          : Style.fa.toggle_on
        uncheckedSymbol        : Style.fa.toggle_off
        uncheckedColor         : Style.main.textDisabled
        symbolPointSize        : Style.dialog.iconSize * Style.pt * 1.1
        spacing                : Style.dialog.spacing*2

        TextMetrics {
            id: metrics
            text: allDatesBox.checkedSymbol
            font {
                family: Style.fontawesome.name
                pointSize: allDatesBox.symbolPointSize
            }
        }


        Rectangle {
            color: allDatesBox.checked ? dotBackground.color : Style.exporting.sliderBackground
            width: metrics.width*0.9
            height: metrics.height*0.6
            radius: height/2
            z: -1

            anchors {
                left: allDatesBox.left
                verticalCenter: allDatesBox.verticalCenter
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
}
