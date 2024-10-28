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
import QtQuick.Controls
import QtQuick.Controls.impl
import QtQuick.Templates as T

T.Popup {
    id: root

    property ColorScheme colorScheme
    readonly property var occurred: shouldShow ? new Date() : undefined
    property int popupPriority: ApplicationWindow.PopupPriority.Banner
    property bool shouldShow: false

    function close() {
        root.shouldShow = false;
    }
    function open() {
        root.shouldShow = true;
    }

    implicitHeight: Math.max(implicitBackgroundHeight + topInset + bottomInset, contentHeight + topPadding + bottomPadding)
    implicitWidth: Math.max(implicitBackgroundWidth + leftInset + rightInset, contentWidth + leftPadding + rightPadding)

    // TODO: Add DropShadow here
    T.Overlay.modal: Rectangle {
        color: root.colorScheme.backdrop_norm
    }
    T.Overlay.modeless: Rectangle {
        color: "transparent"
    }

    Component.onCompleted: {
        if (!ApplicationWindow.window) {
            return;
        }
        if (ApplicationWindow.window.popups === undefined) {
            return;
        }
        const obj = this;
        ApplicationWindow.window.popups.append({
                "obj": obj
            });
    }
}
