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
import QtQuick.Window
import QtQuick.Controls
import QtQuick.Controls.impl
import QtQuick.Templates as T

T.ComboBox {
    id: root

    property ColorScheme colorScheme

    bottomPadding: 5
    font.family: ProtonStyle.font_family
    font.letterSpacing: ProtonStyle.body_letter_spacing
    font.pixelSize: ProtonStyle.body_font_size
    font.weight: ProtonStyle.fontWeight_400
    implicitHeight: Math.max(implicitBackgroundHeight + topInset + bottomInset, implicitContentHeight + topPadding + bottomPadding, implicitIndicatorHeight + topPadding + bottomPadding)
    implicitWidth: Math.max(implicitBackgroundWidth + leftInset + rightInset, implicitContentWidth + leftPadding + rightPadding)
    leftPadding: 12 + (!root.mirrored || !indicator || !indicator.visible ? 0 : indicator.width + spacing)
    rightPadding: 12 + (root.mirrored || !indicator || !indicator.visible ? 0 : indicator.width + spacing)
    spacing: 8
    topPadding: 5

    background: Rectangle {
        border.color: root.colorScheme.border_norm
        border.width: 1
        color: {
            if (root.down) {
                return root.colorScheme.interaction_default_active;
            }
            if (root.enabled && root.hovered || root.activeFocus) {
                return root.colorScheme.interaction_default_hover;
            }
            if (!root.enabled) {
                return root.colorScheme.interaction_default;
            }
            return root.colorScheme.background_norm;
        }
        implicitHeight: 36
        implicitWidth: 140
        radius: ProtonStyle.context_item_radius
    }
    contentItem: T.TextField {
        autoScroll: root.editable
        color: root.enabled ? root.colorScheme.text_norm : root.colorScheme.text_disabled
        enabled: root.editable
        font: root.font
        inputMethodHints: root.inputMethodHints
        padding: 5
        placeholderTextColor: root.enabled ? root.colorScheme.text_hint : root.colorScheme.text_disabled
        readOnly: root.down
        selectedTextColor: root.colorScheme.text_invert
        selectionColor: root.colorScheme.interaction_norm
        text: root.editable ? root.editText : root.displayText
        validator: root.validator
        verticalAlignment: TextInput.AlignVCenter

        background: Rectangle {
            border.color: {
                if (root.activeFocus) {
                    return root.colorScheme.interaction_norm;
                }
                if (root.hovered || root.activeFocus) {
                    return root.colorScheme.field_hover;
                }
                return root.colorScheme.field_norm;
            }
            border.width: 1
            color: root.colorScheme.background_norm
            radius: ProtonStyle.context_item_radius
            visible: root.enabled && root.editable && !root.flat
        }
    }
    delegate: ItemDelegate {
        property bool selected: root.currentIndex === index

        font: root.font
        highlighted: root.highlightedIndex === index
        hoverEnabled: root.hoverEnabled
        palette.highlightedText: selected ? root.colorScheme.text_invert : root.colorScheme.text_norm
        palette.text: {
            if (!root.enabled) {
                return root.colorScheme.text_disabled;
            }
            if (selected) {
                return root.colorScheme.text_invert;
            }
            return root.colorScheme.text_norm;
        }
        text: root.textRole ? (Array.isArray(root.model) ? modelData[root.textRole] : model[root.textRole]) : modelData
        width: parent.width

        background: PaddedRectangle {
            color: {
                if (parent.down) {
                    return root.colorScheme.interaction_default_active;
                }
                if (parent.selected) {
                    return root.colorScheme.interaction_norm;
                }
                if (parent.hovered || parent.highlighted) {
                    return root.colorScheme.interaction_default_hover;
                }
                return root.colorScheme.interaction_default;
            }
            radius: ProtonStyle.context_item_radius
        }
    }
    indicator: ColorImage {
        color: root.enabled ? root.colorScheme.text_norm : root.colorScheme.text_disabled
        source: popup.visible ? "/qml/icons/ic-chevron-up.svg" : "/qml/icons/ic-chevron-down.svg"
        sourceSize.height: 16
        sourceSize.width: 16
        x: root.mirrored ? 12 : root.width - width - 12
        y: root.topPadding + (root.availableHeight - height) / 2
    }
    popup: T.Popup {
        bottomMargin: 8
        height: Math.min(contentItem.implicitHeight, root.Window.height - topMargin - bottomMargin)
        topMargin: 8
        width: root.width
        y: root.height

        background: Rectangle {
            border.color: root.colorScheme.border_weak
            border.width: 1
            color: root.colorScheme.background_norm
            radius: ProtonStyle.dialog_radius
        }
        contentItem: Item {
            implicitHeight: children[0].implicitHeight + children[0].anchors.topMargin + children[0].anchors.bottomMargin
            implicitWidth: children[0].implicitWidth + children[0].anchors.leftMargin + children[0].anchors.rightMargin

            ListView {
                anchors.fill: parent
                anchors.margins: 8
                currentIndex: root.highlightedIndex
                implicitHeight: contentHeight
                model: root.delegateModel
                spacing: 4

                T.ScrollIndicator.vertical: ScrollIndicator {
                }
            }
        }
    }
}
