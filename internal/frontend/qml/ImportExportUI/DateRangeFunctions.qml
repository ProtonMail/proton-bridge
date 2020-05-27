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

Item {
    id: root
    /*
     NOTE: need to be in obejct with
     id: dateRange

     property var   structure : structureExternal
     property string sourceID : structureExternal.getID ( -1 )

     property alias allDates      : allDatesBox.checked
     property alias inputDateFrom : inputDateFrom
     property alias inputDateTo   : inputDateTo

     function setRange() {common.setRange()}
     function applyRange() {common.applyRange()}
     */

    function resetRange() {
        inputDateFrom.setDate(gui.netBday.getTime())
        inputDateTo.setDate((new Date()).getTime())
    }

    function setRange(){ // unix time in seconds
        var folderFrom = dateRange.structure.getFrom(dateRange.sourceID)
        if (folderFrom===undefined) folderFrom = 0
        var folderTo = dateRange.structure.getTo(dateRange.sourceID)
        if (folderTo===undefined) folderTo = 0
        if ( folderFrom == 0 && folderTo ==0 ) {
            dateRange.allDates = true
        } else {
            dateRange.allDates = false
            inputDateFrom.setDate(folderFrom)
            inputDateTo.setDate(folderTo)
        }
    }

    function applyRange(){ // unix time is seconds
        if (dateRange.allDates)  structure.setFromToDate(dateRange.sourceID, 0, 0)
        else {
            var endOfDay = new Date(inputDateTo.unix*1000)
            endOfDay.setHours(23,59,59,999)
            var endOfDayUnix = parseInt(endOfDay.getTime()/1000)
            structure.setFromToDate(dateRange.sourceID, inputDateFrom.unix, endOfDayUnix)
        }
    }

    Connections {
        target: dateRange
        onStructureChanged: setRange()
    }

    Component.onCompleted: {
        inputDateFrom.updateRange(gui.netBday)
        inputDateTo.updateRange(new Date())
        setRange()
    }
}

