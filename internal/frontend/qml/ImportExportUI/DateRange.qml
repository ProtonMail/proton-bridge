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


Column {
    id: dateRange

    property var   structure : transferRules
    property string sourceID : "-1"

    property alias allDates      : allDatesBox.checked
    property alias inputDateFrom : inputDateFrom
    property alias inputDateTo   : inputDateTo

    function getRange() {common.getRange()}
    function setRangeFromTo(from, to) {common.setRangeFromTo(from, to)}
    function applyRange() {common.applyRange()}

    property var dropDownStyle : Style.dropDownLight
    property var isDark : dropDownStyle.background == Style.dropDownDark.background

    spacing: Style.dialog.spacing

    DateRangeFunctions {id:common}

    DateInput {
        id: inputDateFrom
        label: qsTr("From:")
        currentDate: gui.netBday
        maxDate: inputDateTo.currentDate
        dropDownStyle: dateRange.dropDownStyle
    }

    Rectangle {
        width: inputDateTo.width
        height: Style.dialog.borderInput / 2
        color: isDark ? dropDownStyle.separator : Style.transparent
    }

    DateInput {
        id: inputDateTo
        label: qsTr("To:")
        metricsLabel: inputDateFrom.label
        currentDate: new Date() // now
        minDate: inputDateFrom.currentDate
        dropDownStyle: dateRange.dropDownStyle
    }

    Rectangle {
        width: inputDateTo.width
        height: Style.dialog.borderInput
        color: isDark ? dropDownStyle.separator : Style.transparent
    }

    CheckBoxLabel {
        id: allDatesBox
        text            : qsTr("All dates")
        anchors.right   : inputDateTo.right
        checkedSymbol   : Style.fa.toggle_on
        uncheckedSymbol : Style.fa.toggle_off
        uncheckedColor  : Style.main.textDisabled
        textColor       : dropDownStyle.text
        symbolPointSize : Style.dialog.iconSize * Style.pt * 1.1
        spacing         : Style.dialog.spacing*2

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
