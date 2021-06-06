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
import QtQuick.Templates 2.12 as T
import QtQuick.Controls 2.12
import QtQuick.Controls.impl 2.12

T.Switch {
    property var colorScheme: parent.colorScheme ? parent.colorScheme : Style.currentStyle

    property bool loading: false

    // TODO: store previous enabled state and restore it?
    // For now assuming that only enabled buttons could have loading state
    onLoadingChanged: {
        if (loading) {
            enabled = false
        } else {
            enabled = true
        }
    }

    id: control

    implicitWidth: Math.max(implicitBackgroundWidth + leftInset + rightInset,
                            implicitContentWidth + leftPadding + rightPadding)
    implicitHeight: Math.max(implicitBackgroundHeight + topInset + bottomInset,
                             implicitContentHeight + topPadding + bottomPadding,
                             implicitIndicatorHeight + topPadding + bottomPadding)

    padding: 0
    spacing: 7

    indicator: PaddedRectangle {
        implicitWidth: 40
        implicitHeight: 24

        x: text ? (control.mirrored ? control.width - width - control.rightPadding : control.leftPadding) : control.leftPadding + (control.availableWidth - width) / 2
        y: control.topPadding + (control.availableHeight - height) / 2

        radius: 12
        leftPadding: 0
        rightPadding: 0
        padding: 0
        color: control.enabled || control.loading ? colorScheme.background_norm : colorScheme.background_strong
        border.width: control.enabled && !loading ? 1 : 0
        border.color: control.hovered ? colorScheme.field_hover : colorScheme.field_norm

        Rectangle {
            x: Math.max(0, Math.min(parent.width - width, control.visualPosition * parent.width - (width / 2)))
            y: (parent.height - height) / 2
            width: 24
            height: 24
            radius: 12

            visible: !loading

            color: {
                if (!control.enabled) {
                    return colorScheme.field_disabled
                }

                if (control.checked) {
                    if (control.hovered) {
                        return colorScheme.interaction_norm_hover
                    }

                    return colorScheme.interaction_norm
                }

                if (control.hovered) {
                    return colorScheme.field_hover
                }

                return colorScheme.field_norm
            }

            ColorImage {
                x: (parent.width - width) / 2
                y: (parent.height - height) / 2

                width: 16
                height: 16
                color: "#FFFFFF"
                source: "../icons/ic-check.svg"
                visible: control.checked
            }

            Behavior on x {
                enabled: !control.down
                SmoothedAnimation { velocity: 200 }
            }
        }

        ColorImage {
            id: loadingImage
            x: parent.width - width
            y: (parent.height - height) / 2

            width: 18
            height: 18
            color: colorScheme.interaction_norm_hover
            source: "../icons/Loader_16.svg"
            visible: control.loading

            RotationAnimation {
                target: loadingImage
                loops: Animation.Infinite
                duration: 1000
                from: 0
                to: 360
                direction: RotationAnimation.Clockwise
                running: control.loading
            }
        }
    }

    contentItem: CheckLabel {
        id: label
        leftPadding: control.indicator && !control.mirrored ? control.indicator.width + control.spacing : 0
        rightPadding: control.indicator && control.mirrored ? control.indicator.width + control.spacing : 0

        text: control.text

        color: control.enabled || control.loading ? colorScheme.text_norm : colorScheme.text_disabled

        font.family: Style.font_family
        font.weight: Style.fontWidth_400
        font.pixelSize: 14
        lineHeight: 20
        lineHeightMode: Text.FixedHeight
        font.letterSpacing: 0.2
    }
}
