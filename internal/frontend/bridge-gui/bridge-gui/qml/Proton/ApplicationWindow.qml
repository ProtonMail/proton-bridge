// Copyright (c) 2024 Proton AG
// This file is part of Proton Mail Bridge.
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.
import QtQml
import QtQuick
import QtQuick.Window
import QtQuick.Controls
import QtQuick.Controls.impl
import QtQuick.Templates as T

T.ApplicationWindow {
    id: root

    // popup priority based on types
    enum PopupType {
        Banner,
        Dialog
    }

    property ColorScheme colorScheme

    // contains currently visible popup
    property var popupVisible: null

    // list of all popups within ApplicationWindow
    property ListModel popups: ListModel {
        // overriding get method to ignore any role and return directly object itself
        function get(row) {
            if (row < 0 || row >= count) {
                return undefined;
            }
            return data(index(row, 0), Qt.DisplayRole);
        }

        onRowsAboutToBeRemoved: function (parent, first, last) {
            for (let i = first; i <= last; i++) {
                const obj = popups.get(i);
                obj.onShouldShowChanged.disconnect(root.processPopups);

                // if currently visible popup was removed
                if (root.popupVisible === obj) {
                    root.popupVisible.visible = false;
                    root.popupVisible = null;
                }
            }
            processPopups();
        }
        onRowsInserted: function (parent, first, last) {
            for (let i = first; i <= last; i++) {
                const obj = popups.get(i);
                obj.onShouldShowChanged.connect(root.processPopups);
            }
            processPopups();
        }
    }

    function processPopups() {
        if ((root.popupVisible) && (!root.popupVisible.shouldShow)) {
            root.popupVisible.visible = false;
        }
        let topmost = null;
        for (let i = 0; i < popups.count; i++) {
            const obj = popups.get(i);
            if (obj.shouldShow === false) {
                continue;
            }
            // User notifications should have display priority
            if (obj.shouldShow && obj.isUserNotification) {
                topmost = obj;
                break;
            }
            if (topmost && (topmost.popupType > obj.popupType)) {
                continue;
            }
            if (topmost && (topmost.popupType === obj.popupType) && (topmost.occurred > obj.occurred)) {
                continue;
            }
            topmost = obj;
        }
        if (root.popupVisible !== topmost) {
            if (root.popupVisible) {
                root.popupVisible.visible = false;
            }
            root.popupVisible = topmost;
        }
        if (!root.popupVisible) {
            return;
        }
        root.popupVisible.visible = true;
    }

    color: root.colorScheme.background_norm

    Overlay.modal: Rectangle {
        color: root.colorScheme.backdrop_norm
    }
    Overlay.modeless: Rectangle {
        color: "transparent"
    }

    Connections {
        function onVisibleChanged() {
            if (root.popupVisible.visible) {
                return;
            }
            root.popupVisible = null;
            root.processPopups();
        }

        target: root.popupVisible
    }
}
