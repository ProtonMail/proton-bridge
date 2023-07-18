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
import QtQuick.Layouts
import QtQuick.Controls
import QtQuick.Controls.impl

Item {
    id: root

    property bool _disabled: !enabled
    property bool checked
    property var colorScheme
    property bool hovered
    property bool loading

    signal clicked

    implicitHeight: children[0].implicitHeight
    implicitWidth: children[0].implicitWidth

    Rectangle {
        id: indicator
        color: {
            if (root.loading)
                return "transparent";
            if (root._disabled)
                return root.colorScheme.background_strong;
            return root.colorScheme.background_norm;
        }
        implicitHeight: 24
        implicitWidth: 40
        radius: width / 2

        border {
            color: (root._disabled || root.loading) ? "transparent" : colorScheme.field_norm
            width: 1
        }
        Rectangle {
            anchors.left: indicator.left
            anchors.leftMargin: root.checked ? 16 : 0
            anchors.verticalCenter: indicator.verticalCenter
            color: {
                if (root.loading)
                    return "transparent";
                if (root._disabled)
                    return root.colorScheme.field_disabled;
                if (root.checked) {
                    if (root.hovered)
                        return root.colorScheme.interaction_norm_hover;
                    return root.colorScheme.interaction_norm;
                } else {
                    if (root.hovered)
                        return root.colorScheme.field_hover;
                    return root.colorScheme.field_norm;
                }
            }
            height: 24
            radius: width / 2
            width: 24

            ColorImage {
                anchors.centerIn: parent
                color: root.colorScheme.background_norm
                height: root.colorScheme.body_font_size
                source: "/qml/icons/ic-check.svg"
                sourceSize.height: root.colorScheme.body_font_size
                visible: root.checked
            }
        }
        ColorImage {
            id: loader
            anchors.centerIn: parent
            color: root.colorScheme.text_norm
            height: root.colorScheme.body_font_size
            source: "/qml/icons/Loader_16.svg"
            sourceSize.height: root.colorScheme.body_font_size
            visible: root.loading

            RotationAnimation {
                direction: RotationAnimation.Clockwise
                duration: 1000
                from: 0
                loops: Animation.Infinite
                running: root.loading
                target: loader
                to: 360
            }
        }
        MouseArea {
            anchors.fill: indicator
            hoverEnabled: true

            onClicked: {
                if (root.enabled)
                    root.clicked();
            }
            onEntered: {
                root.hovered = true;
            }
            onExited: {
                root.hovered = false;
            }
            onPressed: {
                root.hovered = true;
            }
            onReleased: {
                root.hovered = containsMouse;
            }
        }
    }
}
