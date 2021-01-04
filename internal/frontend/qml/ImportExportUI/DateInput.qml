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

// input for date
import QtQuick 2.8
import QtQuick.Controls 2.2
import ProtonUI 1.0
import ImportExportUI 1.0

Rectangle {
    id: root

    width  : row.width + (root.label == "" ? 0 : textlabel.width)
    height : row.height
    color  : Style.transparent

    property alias  label        : textlabel.text
    property string metricsLabel : root.label
    property var dropDownStyle   : Style.dropDownLight

    // dates
    property date  currentDate    : new Date()  // default now
    property date  minDate        : new Date(0) // default epoch start
    property date  maxDate        : new Date()  // default now
    property bool  isMaxDateToday : false
    property int   unix           : Math.floor(currentDate.getTime()/1000)

    onMinDateChanged: {
        if (isNaN(minDate.getTime()) || minDate.getTime() > maxDate.getTime()) {
            minDate = new Date(0)
        }
        //console.log(" minDate changed:", root.label, minDate.toDateString())
        updateRange()
    }
    onMaxDateChanged: {
        if (isNaN(maxDate.getTime()) || minDate.getTime() > maxDate.getTime()) {
            maxDate = new Date()
        }
        //console.log(" maxDate changed:", root.label, maxDate.toDateString())
        updateRange()
    }

    RoundedRectangle {
        id: background
        anchors.fill      : row
        strokeColor       : dropDownStyle.line
        strokeWidth       : Style.dialog.borderInput
        fillColor         : dropDownStyle.background
        radiusTopLeft     : row.children[0].down                     && !row.children[0].below                     ? 0 : Style.dialog.radiusButton
        radiusBottomLeft  : row.children[0].down                     && row.children[0].below                      ? 0 : Style.dialog.radiusButton
        radiusTopRight    : row.children[row.children.length-1].down && !row.children[row.children.length-1].below ? 0 : Style.dialog.radiusButton
        radiusBottomRight : row.children[row.children.length-1].down && row.children[row.children.length-1].below  ? 0 : Style.dialog.radiusButton
    }

    TextMetrics {
        id: textMetrics
        text: root.metricsLabel+"M"
        font: textlabel.font
    }

    Text {
        id: textlabel
        anchors {
            left           : root.left
            verticalCenter : root.verticalCenter
        }
        font {
            pointSize: Style.dialog.fontSize * Style.pt
            bold: dropDownStyle.labelBold
        }
        color: dropDownStyle.text
        width: textMetrics.width
        verticalAlignment: Text.AlignVCenter
    }

    Row {
        id: row

        anchors {
            left   : root.label=="" ? root.left : textlabel.right
            bottom : root.bottom
        }
        padding   : Style.dialog.borderInput

        DateBox {
            id: monthInput
            placeholderText: qsTr("Month")
            enabled: !allDates
            model: gui.allMonths
            onActivated: updateRange()
            anchors.verticalCenter: parent.verticalCenter
            dropDownStyle: root.dropDownStyle
            onDownChanged: {
                if (root.isMaxDateToday){
                    root.maxDate = new Date()
                }
            }
        }

        Rectangle {
            width: Style.dialog.borderInput
            height: monthInput.height
            color: dropDownStyle.line
            anchors.verticalCenter: parent.verticalCenter
        }

        DateBox {
            id: dayInput
            placeholderText: qsTr("Day")
            enabled: !allDates
            model: gui.allDays
            onActivated: updateRange()
            anchors.verticalCenter: parent.verticalCenter
            dropDownStyle: root.dropDownStyle
            onDownChanged: {
                if (root.isMaxDateToday){
                    root.maxDate = new Date()
                }
            }
        }

        Rectangle {
            width: Style.dialog.borderInput
            height: monthInput.height
            color: dropDownStyle.line
        }

        DateBox {
            id: yearInput
            placeholderText: qsTr("Year")
            enabled: !allDates
            model: gui.allYears
            onActivated: updateRange()
            anchors.verticalCenter: parent.verticalCenter
            dropDownStyle: root.dropDownStyle
            onDownChanged: {
                if (root.isMaxDateToday){
                    root.maxDate = new Date()
                }
            }
        }
    }


    function setDate(d) {
        //console.trace()
        //console.log( "    setDate ", label, d)
        if (isNaN(d = parseInt(d))) return
        var newUnix = Math.min(maxDate.getTime(), d*1000) // seconds to ms
        newUnix = Math.max(minDate.getTime(), newUnix)
        root.updateRange(new Date(newUnix))
        //console.log( "        set ", currentDate.getTime())
    }


    function updateRange(curr) {
        if (curr === undefined || isNaN(curr.getTime())) curr = root.getCurrentDate()
        //console.log( "    update", label, curr, curr.getTime())
        //console.trace()
        if (isNaN(curr.getTime())) return // shouldn't happen
        // full system date range
        var firstYear = parseInt(gui.allYears[0])
        var firstDay = parseInt(gui.allDays[0])
        if ( isNaN(firstYear) || isNaN(firstDay) ) return
        // get minimal and maximal available year, month, day
        // NOTE: The order is important!!!
        var minYear  = minDate.getFullYear()
        var maxYear  = maxDate.getFullYear()
        var minMonth = (curr.getFullYear() == minYear  ? minDate.getMonth() : 0  )
        var maxMonth = (curr.getFullYear() == maxYear  ? maxDate.getMonth() : 11 )
        var minDay   = (
            curr.getFullYear() == minYear  &&
            curr.getMonth()    == minMonth ?
            minDate.getDate() : firstDay
        )
        var maxDay   = (
            curr.getFullYear() == maxYear  &&
            curr.getMonth()    == maxMonth ?
            maxDate.getDate() : gui.daysInMonth(curr.getFullYear(), curr.getMonth()+1)
        )

        //console.log("update ranges: ", root.label, minYear, maxYear, minMonth+1, maxMonth+1, minDay, maxDay)
        //console.log("update indexes: ", root.label, firstYear-minYear, firstYear-maxYear, minMonth, maxMonth, minDay-firstDay, maxDay-firstDay)


        yearInput.filterItems(firstYear-maxYear, firstYear-minYear)
        monthInput.filterItems(minMonth,maxMonth) // getMonth() is index not a month (i.e. Jan==0)
        dayInput.filterItems(minDay-1,maxDay-1)

        // keep ordering from model not from filter
        yearInput  .currentIndex = firstYear - curr.getFullYear()
        monthInput .currentIndex = curr.getMonth() // getMonth() is index not a month (i.e. Jan==0)
        dayInput   .currentIndex = curr.getDate()-firstDay

        /*
         console.log(
             "update current indexes: ", root.label,
             curr.getFullYear() , '->' , yearInput.currentIndex  ,
             gui.allMonths[curr.getMonth()]    , '->' , monthInput.currentIndex ,
             curr.getDate()     , '->' , dayInput.currentIndex
         )
         */

        // test if current date changed
        if (
            yearInput.currentText  == root.currentDate.getFullYear()                      &&
            monthInput.currentText == root.currentDate.toLocaleString(gui.locale, "MMM") &&
            dayInput.currentText   == gui.prependZeros(root.currentDate.getDate(),2)
        ) {
            //console.log(" currentDate NOT changed", label, root.currentDate.toDateString())
            return
        }

        root.currentDate = root.getCurrentDate()
        // console.log(" currentDate changed", label, root.currentDate.toDateString())
    }

    // get current date from selected
    function getCurrentDate() {
        if (isNaN(root.currentDate.getTime())) { // wrong current ?
            console.log("!WARNING! Wrong current date format", root.currentDate)
            root.currentDate = new Date(0)
        }
        var currentString = ""
        var currentUnix = root.currentDate.getTime()
        if (
            yearInput.currentText  != ""                         &&
            yearInput.currentText  != yearInput.placeholderText  &&
            monthInput.currentText != ""                         &&
            monthInput.currentText != monthInput.placeholderText
        ) {
            var day = gui.daysInMonth(yearInput.currentText, monthInput.currentText)
            if (!isNaN(parseInt(dayInput.currentText))) {
                day = Math.min(day, parseInt(dayInput.currentText))
            }
            var month = gui.allMonths.indexOf(monthInput.currentText)
            var year = parseInt(yearInput.currentText)
            var pickedDate = new Date(year, month, day)
            // Compensate automatic DST in windows
            if (pickedDate.getDate() != day) {
                pickedDate.setTime(pickedDate.getTime() + 60*60*1000) // add hour
            }
            currentUnix = pickedDate.getTime()
        }
        return new Date(Math.max(
            minDate.getTime(),
            Math.min(maxDate.getTime(), currentUnix)
        ))
    }
}
