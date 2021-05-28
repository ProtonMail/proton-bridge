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

import QtQuick 2.12
import QtQuick.Controls 2.12
import QtQuick.Controls.impl 2.12
import QtQuick.Templates 2.12 as T
import "."

T.Button {
    property var colorScheme: parent.colorScheme ? parent.colorScheme : Style.currentStyle

    property alias secondary: control.flat
    readonly property bool primary: !secondary
    readonly property bool isIcon: control.text === ""

    id: control

    implicitWidth: Math.max(implicitBackgroundWidth + leftInset + rightInset,
    implicitContentWidth + leftPadding + rightPadding)
    implicitHeight: Math.max(implicitBackgroundHeight + topInset + bottomInset,
    implicitContentHeight + topPadding + bottomPadding)

    padding: 8
    horizontalPadding: 16
    spacing: 10

    icon.width: 12
    icon.height: 12
    icon.color: control.checked || control.highlighted ? control.palette.brightText :
    control.flat && !control.down ? (control.visualFocus ? control.palette.highlight : control.palette.windowText) : control.palette.buttonText

    contentItem: IconLabel {
        spacing: control.spacing
        mirrored: control.mirrored
        display: control.display

        icon: control.icon
        text: control.text
        font: control.font
        color: {
            if (!secondary) {
                // Primary colors
                return "#FFFFFF"
            } else {
                // Secondary colors
                return colorScheme.text_norm
            }
        }
    }

    background: Rectangle {
        implicitWidth: 72
        implicitHeight: 36
        radius: 4
        visible: !control.flat || control.down || control.checked || control.highlighted
        color: {
            if (!isIcon) {
                if (!secondary) {
                    // Primary colors

                    if (control.down) {
                        return colorScheme.interaction_norm_active
                    }

                    if (control.enabled && (control.highlighted || control.hovered || control.checked)) {
                        return colorScheme.interaction_norm_hover
                    }

                    return colorScheme.interaction_norm
                } else {
                    // Secondary colors

                    if (control.down) {
                        return colorScheme.interaction_default_active
                    }

                    if (control.enabled && (control.highlighted || control.hovered || control.checked)) {
                        return colorScheme.interaction_default_hover
                    }

                    return colorScheme.interaction_default
                }
            } else {
                if (!secondary) {
                    // Primary icon colors

                    if (control.down) {
                        return colorScheme.interaction_default_active
                    }

                    if (control.enabled && (control.highlighted || control.hovered || control.checked)) {
                        return colorScheme.interaction_default_hover
                    }

                    return colorScheme.interaction_default
                } else {
                    // Secondary icon colors

                    if (control.down) {
                        return colorScheme.interaction_default_active
                    }

                    if (control.enabled && (control.highlighted || control.hovered || control.checked)) {
                        return colorScheme.interaction_default_hover
                    }

                    return colorScheme.interaction_default
                }
            }
        }
        opacity: control.enabled ? 1.0 : 0.5
    }
}
