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
import "."

T.MenuItem {
    id: control

    property ColorScheme colorScheme

    font.family: ProtonStyle.font_family
    font.letterSpacing: ProtonStyle.body_letter_spacing
    font.pixelSize: ProtonStyle.body_font_size
    font.weight: ProtonStyle.fontWeight_400
    icon.color: control.enabled ? control.colorScheme.text_norm : control.colorScheme.text_disabled
    icon.height: 24
    icon.width: 24
    implicitHeight: Math.max(implicitBackgroundHeight + topInset + bottomInset, implicitContentHeight + topPadding + bottomPadding, implicitIndicatorHeight + topPadding + bottomPadding)
    padding: 12
    spacing: 6
    width: parent.width // required. Other item overflows to the right of the menu and get clipped.

    background: Rectangle {
        color: control.down ? control.colorScheme.interaction_default_active : control.highlighted ? control.colorScheme.interaction_default_hover : control.colorScheme.interaction_default
        implicitHeight: 36
        implicitWidth: 164
        radius: ProtonStyle.button_radius
    }
    contentItem: IconLabel {
        id: iconLabel

        readonly property real arrowPadding: control.subMenu && control.arrow ? control.arrow.width + control.spacing : 0
        readonly property real indicatorPadding: control.checkable && control.indicator ? control.indicator.width + control.spacing : 0

        alignment: Qt.AlignLeft
        color: control.enabled ? control.colorScheme.text_norm : control.colorScheme.text_disabled
        display: control.display
        font: control.font
        icon: control.icon
        leftPadding: !control.mirrored ? indicatorPadding : arrowPadding
        mirrored: control.mirrored
        rightPadding: control.mirrored ? indicatorPadding : arrowPadding
        spacing: control.spacing
        text: control.text
    }
}
