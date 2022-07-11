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

import QtQuick 2.12
import QtQuick.Controls 2.12
import QtQuick.Controls.impl 2.12
import QtQuick.Templates 2.12 as T
import QtQuick.Window 2.12
import "."

T.Menu {
    id: control

    property ColorScheme colorScheme

    implicitWidth: Math.max(
        implicitBackgroundWidth + leftInset + rightInset,
        contentWidth + leftPadding + rightPadding
    )
    implicitHeight: Math.max(
        implicitBackgroundHeight + topInset + bottomInset,
        contentHeight + topPadding + bottomPadding
    )

    margins: 0
    overlap: 1

    delegate: MenuItem {
        colorScheme: control.colorScheme
    }

    contentItem: Item {
        implicitHeight: children[0].implicitHeight + children[0].anchors.topMargin + children[0].anchors.bottomMargin
        implicitWidth: children[0].implicitWidth + children[0].anchors.leftMargin + children[0].anchors.rightMargin

        ListView {
            anchors.fill: parent
            anchors.margins: 8

            implicitHeight: contentHeight
            model: control.contentModel
            interactive: Window.window ? contentHeight > Window.window.height : false
            clip: true
            currentIndex: control.currentIndex

            ScrollIndicator.vertical: ScrollIndicator {}
        }
    }

    background: Rectangle {
        implicitWidth: 200
        implicitHeight: 40
        color: colorScheme.background_norm
        border.width: 1
        border.color: colorScheme.border_weak
        radius: Style.account_row_radius
    }
}
