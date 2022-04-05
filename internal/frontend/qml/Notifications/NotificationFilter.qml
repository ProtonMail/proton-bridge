// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

import QtQml 2.12
import QtQml.Models 2.12

// contains notifications that satisfy black- and whitelist and are sorted in time-occurred order
ListModel {
    id: root

    enum FilterConsts {
        None = 0,
        All = 255
    }

    property int whitelist: NotificationFilter.FilterConsts.All
    property int blacklist: NotificationFilter.FilterConsts.None

    property Notification topmost
    property var source

    property bool componentCompleted: false
    Component.onCompleted: {
        root.componentCompleted = true
        root.rebuildList()
    }

    // overriding get method to ignore any role and return directly object itself
    function get(row) {
        if (row < 0 || row >= count) {
            return undefined
        }
        return data(index(row, 0), Qt.DisplayRole)
    }

    function rebuildList() {
        // avoid evaluation of the list before Component.onCompleted
        if (!root.componentCompleted) {
            return
        }

        for (var i = 0; i < root.count; i++) {
            root.get(i).onActiveChanged.disconnect( root.updateList )
        }

        root.clear()

        if (!root.source) {
            return
        }

        for (i = 0; i < root.source.length; i++) {
            var obj = root.source[i]
            if (obj.group & root.blacklist) {
                continue
            }

            if (!(obj.group & root.whitelist)) {
                continue
            }

            root.append({obj})
            obj.onActiveChanged.connect( root.updateList )
        }
    }

    function updateList() {
        var topmost = null

        for (var i = 0; i < root.count; i++) {
            var obj = root.get(i)

            if (!obj.active) {
                continue
            }

            if (topmost && (topmost.type > obj.type)) {
                continue
            }

            if (topmost && (topmost.type === obj.type) && (topmost.occurred > obj.occurred)) {
                continue
            }

            topmost = obj
        }

        root.topmost = topmost
    }

    onWhitelistChanged: {
        root.rebuildList()
    }
    onBlacklistChanged: {
        root.rebuildList()
    }
    onSourceChanged: {
        root.rebuildList()
    }
}
