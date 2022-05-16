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

import QtQuick 2.13
import QtQuick.Layouts 1.12
import QtQuick.Controls 2.13
import QtQuick.Controls.impl 2.13

Item {
    id: root
    property var colorScheme
    property bool checked
    property bool hovered
    property bool loading

    signal clicked

    property bool _disabled: !enabled

    implicitHeight: children[0].implicitHeight
    implicitWidth: children[0].implicitWidth

    Rectangle {
        id: indicator
        implicitWidth: 40
        implicitHeight: 24

        radius: width/2
        color: {
            if (root.loading) return "transparent"
            if (root._disabled) return root.colorScheme.background_strong
            return root.colorScheme.background_norm
        }
        border {
            width: 1
            color: (root._disabled || root.loading) ? "transparent" : colorScheme.field_norm
        }

        Rectangle {
            anchors.verticalCenter: indicator.verticalCenter
            anchors.left: indicator.left
            anchors.leftMargin: root.checked ? 16 : 0
            width: 24
            height: 24
            radius: width/2
            color: {
                if (root.loading) return "transparent"
                if (root._disabled) return root.colorScheme.field_disabled

                if (root.checked) {
                    if (root.hovered) return root.colorScheme.interaction_norm_hover
                    return root.colorScheme.interaction_norm
                } else {
                    if (root.hovered) return root.colorScheme.field_hover
                    return root.colorScheme.field_norm
                }
            }

            ColorImage {
                anchors.centerIn: parent
                source: "../icons/ic-check.svg"
                color: root.colorScheme.background_norm
                height: root.colorScheme.body_font_size
                sourceSize.height: root.colorScheme.body_font_size
                visible: root.checked
            }
        }

        ColorImage {
            id: loader
            anchors.centerIn: parent
            source: "../icons/Loader_16.svg"
            color: root.colorScheme.text_norm
            height: root.colorScheme.body_font_size
            sourceSize.height: root.colorScheme.body_font_size
            visible: root.loading

            RotationAnimation {
                target: loader
                loops: Animation.Infinite
                duration: 1000
                from: 0
                to: 360
                direction: RotationAnimation.Clockwise
                running: root.loading
            }
        }

        MouseArea {
            anchors.fill: indicator
            hoverEnabled: true
            onEntered: {root.hovered = true }
            onExited: {root.hovered = false }
            onClicked: { if (root.enabled) root.clicked();}
            onPressed: {root.hovered = true }
            onReleased: { root.hovered = containsMouse }
        }
    }
}
