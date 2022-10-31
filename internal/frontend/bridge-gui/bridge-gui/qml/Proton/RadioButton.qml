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

import QtQuick
import QtQuick.Controls
import QtQuick.Controls.impl
import QtQuick.Templates as T

T.RadioButton {
    property ColorScheme colorScheme

    property bool error: false

    id: control

    implicitWidth: Math.max(implicitBackgroundWidth + leftInset + rightInset,
                            implicitContentWidth + leftPadding + rightPadding)
    implicitHeight: Math.max(implicitBackgroundHeight + topInset + bottomInset,
                             implicitContentHeight + topPadding + bottomPadding,
                             implicitIndicatorHeight + topPadding + bottomPadding)

    padding: 0
    spacing: 8

    indicator: Rectangle {
        implicitWidth: 20
        implicitHeight: 20
        radius: width / 2

        x: text ? (control.mirrored ? control.width - width - control.rightPadding : control.leftPadding) : control.leftPadding + (control.availableWidth - width) / 2
        y: control.topPadding + (control.availableHeight - height) / 2

        color: control.colorScheme.background_norm
        border.width: 1
        border.color: {
            if (!control.enabled) {
                return control.colorScheme.field_disabled
            }

            if (control.error) {
                return control.colorScheme.signal_danger
            }

            if (control.hovered || control.activeFocus) {
                return control.colorScheme.interaction_norm_hover
            }

            return control.colorScheme.field_norm
        }

        Rectangle {
            x: (parent.width - width) / 2
            y: (parent.height - height) / 2
            width: 8
            height: 8
            radius: width / 2
            color: {
                if (!control.enabled) {
                    return control.colorScheme.field_disabled
                }

                if (control.error) {
                    return control.colorScheme.signal_danger
                }

                if (control.hovered || control.activeFocus) {
                    return control.colorScheme.interaction_norm_hover
                }

                return control.colorScheme.interaction_norm
            }
            visible: control.checked
        }
    }

    contentItem: CheckLabel {
        leftPadding: control.indicator && !control.mirrored ? control.indicator.width + control.spacing : 0
        rightPadding: control.indicator && control.mirrored ? control.indicator.width + control.spacing : 0

        text: control.text

        color: {
            if (!enabled) {
                return control.colorScheme.text_disabled
            }

            if (error) {
                return control.colorScheme.signal_danger
            }

            return control.colorScheme.text_norm
        }

        font.family: ProtonStyle.font_family
        font.weight: ProtonStyle.fontWeight_400
        font.pixelSize: ProtonStyle.body_font_size
        lineHeight: ProtonStyle.body_line_height
        lineHeightMode: Text.FixedHeight
        font.letterSpacing: ProtonStyle.body_letter_spacing
    }
}
