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

import QtQml
import QtQuick
import QtQuick.Templates as T
import QtQuick.Controls
import QtQuick.Controls.impl

T.Dialog {
    id: root
    property ColorScheme colorScheme

    Component.onCompleted: {
        if (!ApplicationWindow.window) {
            return
        }

        if (ApplicationWindow.window.popups === undefined) {
            return
        }

        var obj = this
        ApplicationWindow.window.popups.append( { obj } )
    }

    readonly property int popupType: ApplicationWindow.PopupType.Dialog

    property bool shouldShow: false
    readonly property var occurred: shouldShow ? new Date() : undefined
    function open() {
        root.shouldShow = true
    }

    function close() {
        root.shouldShow = false
    }

    anchors.centerIn: Overlay.overlay

    implicitWidth: Math.max(implicitBackgroundWidth + leftInset + rightInset,
    contentWidth + leftPadding + rightPadding,
    implicitHeaderWidth,
    implicitFooterWidth)
    implicitHeight: Math.max(implicitBackgroundHeight + topInset + bottomInset,
    contentHeight + topPadding + bottomPadding
    + (implicitHeaderHeight > 0 ? implicitHeaderHeight + spacing : 0)
    + (implicitFooterHeight > 0 ? implicitFooterHeight + spacing : 0))

    padding: 24

    background: Rectangle {
        color: root.colorScheme.background_norm
        radius: ProtonStyle.dialog_radius
    }

    // TODO: Add DropShadow here

    T.Overlay.modal: Rectangle {
        color: root.colorScheme.backdrop_norm
    }

    T.Overlay.modeless: Rectangle {
        color: "transparent"
    }
}
