// Copyright (c) 2023 Proton AG
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
import QtQuick
import QtQuick.Controls
import QtQuick.Controls.impl
import QtQuick.Templates as T
import QtQuick.Window
import "."

T.Menu {
    id: control

    property ColorScheme colorScheme

    implicitHeight: Math.max(implicitBackgroundHeight + topInset + bottomInset, contentHeight + topPadding + bottomPadding)
    implicitWidth: Math.max(implicitBackgroundWidth + leftInset + rightInset, contentWidth + leftPadding + rightPadding)
    margins: 0
    overlap: 1

    background: Rectangle {
        border.color: colorScheme.border_weak
        border.width: 1
        color: colorScheme.background_norm
        implicitHeight: 40
        implicitWidth: 200
        radius: ProtonStyle.account_row_radius
    }
    contentItem: Item {
        implicitHeight: children[0].implicitHeight + children[0].anchors.topMargin + children[0].anchors.bottomMargin
        implicitWidth: children[0].implicitWidth + children[0].anchors.leftMargin + children[0].anchors.rightMargin

        ListView {
            anchors.fill: parent
            anchors.margins: 8
            clip: true
            currentIndex: control.currentIndex
            implicitHeight: contentHeight
            interactive: Window.window ? contentHeight > Window.window.height : false
            model: control.contentModel

            ScrollIndicator.vertical: ScrollIndicator {
            }
        }
    }
    delegate: MenuItem {
        colorScheme: control.colorScheme
    }
}
