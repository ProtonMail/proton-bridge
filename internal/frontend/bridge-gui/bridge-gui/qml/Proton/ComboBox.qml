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
import QtQuick.Window
import QtQuick.Controls
import QtQuick.Controls.impl
import QtQuick.Templates as T

T.ComboBox {
    id: root

    property ColorScheme colorScheme

    implicitWidth: Math.max(implicitBackgroundWidth + leftInset + rightInset,
    implicitContentWidth + leftPadding + rightPadding)
    implicitHeight: Math.max(implicitBackgroundHeight + topInset + bottomInset,
    implicitContentHeight + topPadding + bottomPadding,
    implicitIndicatorHeight + topPadding + bottomPadding)

    leftPadding: 12 + (!root.mirrored || !indicator || !indicator.visible ? 0 : indicator.width + spacing)
    rightPadding: 12 + (root.mirrored || !indicator || !indicator.visible ? 0 : indicator.width + spacing)

    topPadding: 5
    bottomPadding: 5

    spacing: 8

    font.family: ProtonStyle.font_family
    font.weight: ProtonStyle.fontWeight_400
    font.pixelSize: ProtonStyle.body_font_size
    font.letterSpacing: ProtonStyle.body_letter_spacing

    contentItem: T.TextField {
        padding: 5

        text: root.editable ? root.editText : root.displayText
        font: root.font

        enabled: root.editable
        autoScroll: root.editable
        readOnly: root.down
        inputMethodHints: root.inputMethodHints
        validator: root.validator
        verticalAlignment: TextInput.AlignVCenter

        color: root.enabled ? root.colorScheme.text_norm : root.colorScheme.text_disabled
        selectionColor: root.colorScheme.interaction_norm
        selectedTextColor: root.colorScheme.text_invert
        placeholderTextColor: root.enabled ? root.colorScheme.text_hint : root.colorScheme.text_disabled

        background: Rectangle {
            radius: ProtonStyle.context_item_radius
            visible: root.enabled && root.editable && !root.flat
            border.color: {
                if (root.activeFocus) {
                    return root.colorScheme.interaction_norm
                }

                if (root.hovered || root.activeFocus) {
                    return root.colorScheme.field_hover
                }

                return root.colorScheme.field_norm
            }
            border.width: 1
            color: root.colorScheme.background_norm
        }
    }

    background: Rectangle {
        implicitWidth: 140
        implicitHeight: 36
        radius: ProtonStyle.context_item_radius
        color: {
            if (root.down) {
                return root.colorScheme.interaction_default_active
            }

            if (root.enabled && root.hovered || root.activeFocus) {
                return root.colorScheme.interaction_default_hover
            }

            if (!root.enabled) {
                return root.colorScheme.interaction_default
            }

            return root.colorScheme.background_norm
        }

        border.color: root.colorScheme.border_norm
        border.width: 1
    }

    indicator: ColorImage {
        x: root.mirrored ? 12 : root.width - width - 12
        y: root.topPadding + (root.availableHeight - height) / 2
        color: root.enabled ? root.colorScheme.text_norm : root.colorScheme.text_disabled
        source: popup.visible ? "/qml/icons/ic-chevron-up.svg" : "/qml/icons/ic-chevron-down.svg"

        sourceSize.width: 16
        sourceSize.height: 16
    }


    delegate: ItemDelegate {
        width: parent.width
        text: root.textRole ? (Array.isArray(root.model) ? modelData[root.textRole] : model[root.textRole]) : modelData

        palette.text: {
            if (!root.enabled) {
                return root.colorScheme.text_disabled
            }

            if (selected) {
                return root.colorScheme.text_invert
            }

            return root.colorScheme.text_norm
        }
        font: root.font

        hoverEnabled: root.hoverEnabled

        property bool selected: root.currentIndex === index

        highlighted: root.highlightedIndex === index
        palette.highlightedText: selected ? root.colorScheme.text_invert : root.colorScheme.text_norm

        background: PaddedRectangle {
            radius: ProtonStyle.context_item_radius
            color: {
                if (parent.down) {
                    return root.colorScheme.interaction_default_active
                }

                if (parent.selected) {
                    return root.colorScheme.interaction_norm
                }

                if (parent.hovered || parent.highlighted) {
                    return root.colorScheme.interaction_default_hover
                }

                return root.colorScheme.interaction_default
            }
        }
    }

    popup: T.Popup {
        y: root.height
        width: root.width
        height: Math.min(contentItem.implicitHeight, root.Window.height - topMargin - bottomMargin)
        topMargin: 8
        bottomMargin: 8

        contentItem: Item {
            implicitHeight: children[0].implicitHeight + children[0].anchors.topMargin + children[0].anchors.bottomMargin
            implicitWidth: children[0].implicitWidth + children[0].anchors.leftMargin + children[0].anchors.rightMargin

            ListView {
                anchors.fill: parent
                anchors.margins: 8

                implicitHeight: contentHeight
                model: root.delegateModel
                currentIndex: root.highlightedIndex
                spacing: 4

                T.ScrollIndicator.vertical: ScrollIndicator { }
            }
        }

        background: Rectangle {
            color: root.colorScheme.background_norm
            radius: ProtonStyle.dialog_radius
            border.color: root.colorScheme.border_weak
            border.width: 1
        }
    }
}
