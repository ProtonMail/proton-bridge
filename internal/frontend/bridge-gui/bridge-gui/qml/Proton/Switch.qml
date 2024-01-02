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
import QtQuick
import QtQuick.Templates as T
import QtQuick.Controls
import QtQuick.Controls.impl

T.Switch {
    id: control

    property ColorScheme colorScheme
    property bool loading: false

    implicitHeight: Math.max(implicitBackgroundHeight + topInset + bottomInset, implicitContentHeight + topPadding + bottomPadding, implicitIndicatorHeight + topPadding + bottomPadding)
    implicitWidth: Math.max(implicitBackgroundWidth + leftInset + rightInset, implicitContentWidth + leftPadding + rightPadding)
    padding: 0
    spacing: 7

    contentItem: CheckLabel {
        id: label
        color: control.enabled || control.loading ? control.colorScheme.text_norm : control.colorScheme.text_disabled
        font.family: ProtonStyle.font_family
        font.letterSpacing: ProtonStyle.body_letter_spacing
        font.pixelSize: ProtonStyle.body_font_size
        font.weight: ProtonStyle.fontWeight_400
        leftPadding: control.indicator && !control.mirrored ? control.indicator.width + control.spacing : 0
        lineHeight: ProtonStyle.body_line_height
        lineHeightMode: Text.FixedHeight
        rightPadding: control.indicator && control.mirrored ? control.indicator.width + control.spacing : 0
        text: control.text
    }
    indicator: Rectangle {
        border.color: control.hovered ? control.colorScheme.field_hover : control.colorScheme.field_norm
        border.width: control.enabled && !loading ? 1 : 0
        color: control.enabled || control.loading ? control.colorScheme.background_norm : control.colorScheme.background_strong
        implicitHeight: 24
        implicitWidth: 40
        radius: height / 2.
        x: text ? (control.mirrored ? control.width - width - control.rightPadding : control.leftPadding) : control.leftPadding + (control.availableWidth - width) / 2
        y: control.topPadding + (control.availableHeight - height) / 2

        Rectangle {
            color: {
                if (!control.enabled) {
                    return control.colorScheme.field_disabled;
                }
                if (control.checked) {
                    if (control.hovered || control.activeFocus) {
                        return control.colorScheme.interaction_norm_hover;
                    }
                    return control.colorScheme.interaction_norm;
                }
                if (control.hovered || control.activeFocus) {
                    return control.colorScheme.field_hover;
                }
                return control.colorScheme.field_norm;
            }
            height: 24
            radius: parent.radius
            visible: !loading
            width: 24
            x: Math.max(0, Math.min(parent.width - width, control.visualPosition * parent.width - (width / 2)))
            y: (parent.height - height) / 2

            Behavior on x  {
                enabled: !control.down

                SmoothedAnimation {
                    velocity: 200
                }
            }

            ColorImage {
                color: "#FFFFFF"
                height: 16
                source: "/qml/icons/ic-check.svg"
                sourceSize.height: 16
                sourceSize.width: 16
                visible: control.checked
                width: 16
                x: (parent.width - width) / 2
                y: (parent.height - height) / 2
            }
        }
        ColorImage {
            id: loadingImage
            color: control.colorScheme.interaction_norm_hover
            height: 18
            source: "/qml/icons/Loader_16.svg"
            sourceSize.height: 18
            sourceSize.width: 18
            visible: control.loading
            width: 18
            x: parent.width - width
            y: (parent.height - height) / 2

            RotationAnimation {
                direction: RotationAnimation.Clockwise
                duration: 1000
                from: 0
                loops: Animation.Infinite
                running: control.loading
                target: loadingImage
                to: 360
            }
        }
    }

    // TODO: store previous enabled state and restore it?
    // For now assuming that only enabled buttons could have loading state
    onLoadingChanged: {
        enabled = !loading;
    }
}
