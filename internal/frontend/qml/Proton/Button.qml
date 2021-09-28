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

import "." as Proton

T.Button {
    property ColorScheme colorScheme

    property alias secondary: control.flat
    readonly property bool primary: !secondary
    readonly property bool isIcon: control.text === ""

    property bool loading: false

    property bool borderless: false

    property int labelType: Proton.Label.LabelType.Body

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

    implicitWidth: Math.max(
        implicitBackgroundWidth + leftInset + rightInset,
        implicitContentWidth + leftPadding + rightPadding
    )
    implicitHeight: Math.max(
        implicitBackgroundHeight + topInset + bottomInset,
        implicitContentHeight + topPadding + bottomPadding
    )

    padding: 8
    horizontalPadding: 16
    spacing: 10

    font: label.font

    icon.width: 16
    icon.height: 16
    icon.color: {
        if (primary && !isIcon) {
            return "#FFFFFF"
        } else {
            return control.colorScheme.text_norm
        }
    }

    contentItem: Item {
        id: _contentItem

        // Since contentItem is allways resized to maximum available size - we need to "incapsulate" label
        // and icon within one single item with calculated fixed implicit size

        implicitHeight: labelIcon.implicitHeight
        implicitWidth: labelIcon.implicitWidth

        Item {
            id: labelIcon

            anchors.horizontalCenter: _contentItem.horizontalCenter
            anchors.verticalCenter: _contentItem.verticalCenter

            width: Math.min(implicitWidth, control.availableWidth)
            height: Math.min(implicitHeight, control.availableHeight)

            implicitWidth: {
                var textImplicitWidth = control.text !== "" ? label.implicitWidth : 0
                var iconImplicitWidth = iconImage.source ? iconImage.implicitWidth : 0
                var spacing = (control.text !== ""  && iconImage.source && control.display === AbstractButton.TextBesideIcon) ? control.spacing : 0

                return control.display === AbstractButton.TextBesideIcon ? textImplicitWidth + iconImplicitWidth + spacing : Math.max(textImplicitWidth, iconImplicitWidth)
            }
            implicitHeight: {
                var textImplicitHeight = control.text !== "" ? label.implicitHeight : 0
                var iconImplicitHeight = iconImage.source ? iconImage.implicitHeight : 0
                var spacing = (control.text !== ""  && iconImage.source && control.display === AbstractButton.TextUnderIcon) ? control.spacing : 0

                return control.display === AbstractButton.TextUnderIcon ? textImplicitHeight + iconImplicitHeight + spacing : Math.max(textImplicitHeight, iconImplicitHeight)
            }

            Proton.Label {
                colorScheme: root.colorScheme
                id: label
                anchors.left: labelIcon.left
                anchors.top: labelIcon.top
                anchors.bottom: labelIcon.bottom
                anchors.right: control.loading ? iconImage.left : labelIcon.right
                anchors.rightMargin: control.loading ? control.spacing : 0

                elide: Text.ElideRight
                horizontalAlignment: Qt.AlignHCenter
                verticalAlignment: Qt.AlignVCenter

                text: control.text
                color: {
                    if (primary && !isIcon) {
                        return "#FFFFFF"
                    } else {
                        return control.colorScheme.text_norm
                    }
                }
                opacity: control.enabled || control.loading ? 1.0 : 0.5

                type: labelType
            }

            ColorImage {
                id: iconImage

                anchors.verticalCenter: labelIcon.verticalCenter
                anchors.right: labelIcon.right

                width: {
                    // special case for loading since we want icon to be square for rotation animation
                    if (control.loading) {
                        return Math.min(control.icon.width, availableWidth, control.icon.height, availableHeight)
                    }

                    return Math.min(control.icon.width, availableWidth)
                }
                height: {
                    if (control.loading) {
                        return width
                    }

                    Math.min(control.icon.height, availableHeight)
                }

                sourceSize.width: control.icon.width
                sourceSize.height: control.icon.height

                color: control.icon.color
                source: control.loading ? "../icons/Loader_16.svg" : control.icon.source
                visible: control.loading || control.icon.source

                RotationAnimation {
                    target: iconImage
                    loops: Animation.Infinite
                    duration: 1000
                    from: 0
                    to: 360
                    direction: RotationAnimation.Clockwise
                    running: control.loading
                }
            }
        }
    }

    background: Rectangle {
        implicitWidth: 36
        implicitHeight: 36
        radius: 4
        visible: true
        color: {
            if (!isIcon) {
                if (primary) {
                    // Primary colors

                    if (control.down) {
                        return control.colorScheme.interaction_norm_active
                    }

                    if (control.enabled && (control.highlighted || control.hovered || control.checked)) {
                        return control.colorScheme.interaction_norm_hover
                    }

                    if (control.loading) {
                        return control.colorScheme.interaction_norm_hover
                    }

                    return control.colorScheme.interaction_norm
                } else {
                    // Secondary colors

                    if (control.down) {
                        return control.colorScheme.interaction_default_active
                    }

                    if (control.enabled && (control.highlighted || control.hovered || control.checked)) {
                        return control.colorScheme.interaction_default_hover
                    }

                    if (control.loading) {
                        return control.colorScheme.interaction_default_hover
                    }

                    return control.colorScheme.interaction_default
                }
            } else {
                if (primary) {
                    // Primary icon colors

                    if (control.down) {
                        return control.colorScheme.interaction_default_active
                    }

                    if (control.enabled && (control.highlighted || control.hovered || control.checked)) {
                        return control.colorScheme.interaction_default_hover
                    }

                    if (control.loading) {
                        return control.colorScheme.interaction_default_hover
                    }

                    return control.colorScheme.interaction_default
                } else {
                    // Secondary icon colors

                    if (control.down) {
                        return control.colorScheme.interaction_default_active
                    }

                    if (control.enabled && (control.highlighted || control.hovered || control.checked)) {
                        return control.colorScheme.interaction_default_hover
                    }

                    if (control.loading) {
                        return control.colorScheme.interaction_default_hover
                    }

                    return control.colorScheme.interaction_default
                }
            }
        }

        border.color: {
            return control.colorScheme.border_norm
        }
        border.width: secondary && !borderless ? 1 : 0

        opacity: control.enabled || control.loading ? 1.0 : 0.5
    }


    Component.onCompleted: {
        if (!control.colorScheme) {
            console.trace()
            var next = root
            for (var i = 0; i<1000; i++) {
                console.log(i, next, "colorscheme", next.colorScheme)
                next = next.parent
                if (!next) break
            }
            console.error("ColorScheme not defined")
        }
    }
}
