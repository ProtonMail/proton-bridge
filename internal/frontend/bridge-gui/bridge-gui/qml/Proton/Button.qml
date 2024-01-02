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
import QtQuick.Controls
import QtQuick.Controls.impl
import QtQuick.Templates as T
import QtQuick.Layouts
import "." as Proton

T.Button {
    id: control

    property bool borderless: false
    property ColorScheme colorScheme
    readonly property bool hasTextAndIcon: (control.text !== "") && (iconImage.source.toString().length > 0)
    property bool iconOnTheLeft: false
    readonly property bool isIcon: control.text === ""
    property int labelType: Proton.Label.LabelType.Body
    property bool loading: false
    readonly property bool primary: !secondary
    property alias secondary: control.flat
    property bool secondaryIsOpaque: false
    property alias textHorizontalAlignment: label.horizontalAlignment
    property alias textVerticalAlignment: label.verticalAlignment

    font: label.font
    horizontalPadding: 16
    icon.color: {
        if (primary && !isIcon) {
            return "#FFFFFF";
        } else {
            return control.colorScheme.text_norm;
        }
    }
    icon.height: 16
    icon.width: 16
    implicitHeight: Math.max(implicitBackgroundHeight + topInset + bottomInset, implicitContentHeight + topPadding + bottomPadding)
    implicitWidth: Math.max(implicitBackgroundWidth + leftInset + rightInset, implicitContentWidth + leftPadding + rightPadding)
    padding: 8
    spacing: 10

    background: Rectangle {
        border.color: {
            return control.colorScheme.border_norm;
        }
        border.width: secondary && !borderless ? 1 : 0
        color: {
            if (!isIcon) {
                if (primary) {
                    // Primary colors
                    if (control.down) {
                        return control.colorScheme.interaction_norm_active;
                    }
                    if (control.enabled && (control.highlighted || control.hovered || control.checked || control.activeFocus)) {
                        return control.colorScheme.interaction_norm_hover;
                    }
                    if (control.loading) {
                        return control.colorScheme.interaction_norm_hover;
                    }
                    return control.colorScheme.interaction_norm;
                } else {
                    // Secondary colors
                    if (control.down) {
                        return control.colorScheme.interaction_default_active;
                    }
                    if (control.enabled && (control.highlighted || control.hovered || control.checked || control.activeFocus)) {
                        return control.colorScheme.interaction_default_hover;
                    }
                    if (control.loading) {
                        return control.colorScheme.interaction_default_hover;
                    }
                    return secondaryIsOpaque ? control.colorScheme.background_norm : control.colorScheme.interaction_default;
                }
            } else {
                if (primary) {
                    // Primary icon colors
                    if (control.down) {
                        return control.colorScheme.interaction_default_active;
                    }
                    if (control.enabled && (control.highlighted || control.hovered || control.checked || control.activeFocus)) {
                        return control.colorScheme.interaction_default_hover;
                    }
                    if (control.loading) {
                        return control.colorScheme.interaction_default_hover;
                    }
                    return control.colorScheme.interaction_default;
                } else {
                    // Secondary icon colors
                    if (control.down) {
                        return control.colorScheme.interaction_default_active;
                    }
                    if (control.enabled && (control.highlighted || control.hovered || control.checked || control.activeFocus)) {
                        return control.colorScheme.interaction_default_hover;
                    }
                    if (control.loading) {
                        return control.colorScheme.interaction_default_hover;
                    }
                    return secondaryIsOpaque ? control.colorScheme.background_norm : control.colorScheme.interaction_default;
                }
            }
        }
        implicitHeight: 36
        implicitWidth: 36
        opacity: control.enabled || control.loading ? 1.0 : 0.5
        radius: ProtonStyle.button_radius
        visible: true
    }
    contentItem: RowLayout {
        id: _contentItem
        layoutDirection: iconOnTheLeft ? Qt.RightToLeft : Qt.LeftToRight
        spacing: control.hasTextAndIcon ? control.spacing : 0

        Proton.Label {
            id: label
            Layout.alignment: Qt.AlignVCenter
            Layout.fillWidth: true
            color: {
                if (primary && !isIcon) {
                    return "#FFFFFF";
                } else {
                    return control.colorScheme.text_norm;
                }
            }
            colorScheme: control.colorScheme
            elide: Text.ElideRight
            horizontalAlignment: Qt.AlignHCenter
            opacity: control.enabled || control.loading ? 1.0 : 0.5
            text: control.text
            type: labelType
            verticalAlignment: Text.AlignVCenter
            visible: !control.isIcon
        }
        ColorImage {
            id: iconImage
            Layout.alignment: Qt.AlignCenter
            color: control.icon.color
            height: {
                if (control.loading) {
                    return width;
                }
                Math.min(control.icon.height, availableHeight);
            }
            source: control.loading ? "/qml/icons/Loader_16.svg" : control.icon.source
            sourceSize.height: control.icon.height
            sourceSize.width: control.icon.width
            visible: control.loading || control.icon.source
            width: {
                // special case for loading since we want icon to be square for rotation animation
                if (control.loading) {
                    return Math.min(control.icon.width, availableWidth, control.icon.height, availableHeight);
                }
                return Math.min(control.icon.width, availableWidth);
            }

            RotationAnimation {
                direction: RotationAnimation.Clockwise
                duration: 1000
                from: 0
                loops: Animation.Infinite
                running: control.loading
                target: iconImage
                to: 360
            }
        }
    }

    Component.onCompleted: {
        if (!control.colorScheme) {
            console.trace();
            let next = root;
            for (let i = 0; i < 1000; i++) {
                console.log(i, next, "colorscheme", next.colorScheme);
                next = next.parent;
                if (!next)
                    break;
            }
            console.error("ColorScheme not defined");
        }
    }
}
