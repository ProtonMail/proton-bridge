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

import QtQml 2.12
import QtQuick 2.12
import QtQuick.Controls 2.12
import QtQuick.Controls.impl 2.12
import QtQuick.Templates 2.12 as T
import "."

Item {
    id: root
    property ColorScheme colorScheme

    property alias background: control.background
    property alias bottomInset: control.bottomInset
    //property alias flickable: control.flickable
    property alias focusReason: control.focusReason
    property alias hoverEnabled: control.hoverEnabled
    property alias hovered: control.hovered
    property alias implicitBackgroundHeight: control.implicitBackgroundHeight
    property alias implicitBackgroundWidth: control.implicitBackgroundWidth
    property alias leftInset: control.leftInset
    property alias palette: control.palette
    property alias placeholderText: control.placeholderText
    property alias placeholderTextColor: control.placeholderTextColor
    property alias rightInset: control.rightInset
    property alias topInset: control.topInset
    property alias activeFocusOnPress: control.activeFocusOnPress
    property alias baseUrl: control.baseUrl
    property alias bottomPadding: control.bottomPadding
    property alias canPaste: control.canPaste
    property alias canRedo: control.canRedo
    property alias canUndo: control.canUndo
    property alias color: control.color
    property alias contentHeight: control.contentHeight
    property alias contentWidth: control.contentWidth
    property alias cursorDelegate: control.cursorDelegate
    property alias cursorPosition: control.cursorPosition
    property alias cursorRectangle: control.cursorRectangle
    property alias cursorVisible: control.cursorVisible
    property alias effectiveHorizontalAlignment: control.effectiveHorizontalAlignment
    property alias font: control.font
    property alias horizontalAlignment: control.horizontalAlignment
    property alias hoveredLink: control.hoveredLink
    property alias inputMethodComposing: control.inputMethodComposing
    property alias inputMethodHints: control.inputMethodHints
    property alias leftPadding: control.leftPadding
    property alias length: control.length
    property alias lineCount: control.lineCount
    property alias mouseSelectionMode: control.mouseSelectionMode
    property alias overwriteMode: control.overwriteMode
    property alias padding: control.padding
    property alias persistentSelection: control.persistentSelection
    property alias preeditText: control.preeditText
    property alias readOnly: control.readOnly
    property alias renderType: control.renderType
    property alias rightPadding: control.rightPadding
    property alias selectByKeyboard: control.selectByKeyboard
    property alias selectByMouse: control.selectByMouse
    property alias selectedText: control.selectedText
    property alias selectedTextColor: control.selectedTextColor
    property alias selectionColor: control.selectionColor
    property alias selectionEnd: control.selectionEnd
    property alias selectionStart: control.selectionStart
    property alias tabStopDistance: control.tabStopDistance
    property alias text: control.text
    property alias textDocument: control.textDocument
    property alias textFormat: control.textFormat
    property alias textMargin: control.textMargin
    property alias topPadding: control.topPadding
    property alias verticalAlignment: control.verticalAlignment
    property alias wrapMode: control.wrapMode

    implicitWidth: background.width
    implicitHeight: control.implicitHeight + Math.max(
        label.implicitHeight + label.anchors.topMargin + label.anchors.bottomMargin,
        hint.implicitHeight + hint.anchors.topMargin + hint.anchors.bottomMargin
    ) + assistiveText.implicitHeight

    property alias label: label.text
    property alias hint: hint.text
    property alias assistiveText: assistiveText.text

    property bool error: false

    signal editingFinished()

    // Backgroud is moved away from within control as it will be clipped with scrollview
    Rectangle {
        id: background

        anchors.fill: controlView

        radius: 4
        visible: true
        color: root.colorScheme.background_norm
        border.color: {
            if (!control.enabled) {
                return root.colorScheme.field_disabled
            }

            if (control.activeFocus) {
                return root.colorScheme.interaction_norm
            }

            if (root.error) {
                return root.colorScheme.signal_danger
            }

            if (control.hovered) {
                return root.colorScheme.field_hover
            }

            return root.colorScheme.field_norm
        }
        border.width: 1
    }

    Label {
        colorScheme: root.colorScheme
        id: label

        anchors.top: root.top
        anchors.left: root.left
        anchors.bottomMargin: 4

        color: root.enabled ? root.colorScheme.text_norm : root.colorScheme.text_disabled

        type: Label.LabelType.Body_semibold
    }

    Label {
        colorScheme: root.colorScheme
        id: hint

        anchors.right: root.right
        anchors.bottom: controlView.top
        anchors.bottomMargin: 5

        color: root.enabled ? root.colorScheme.text_weak : root.colorScheme.text_disabled

        type: Label.LabelType.Caption
    }

    ColorImage {
        id: errorIcon
        visible: root.error
        anchors.left: parent.left
        anchors.top: assistiveText.top
        anchors.bottom: assistiveText.bottom
        source: "../icons/ic-exclamation-circle-filled.svg"
        sourceSize.height: height
        color: root.colorScheme.signal_danger
    }

    Label {
        colorScheme: root.colorScheme
        id: assistiveText

        anchors.left: root.error ? errorIcon.right : parent.left
        anchors.leftMargin: root.error ? 5 : 0
        anchors.bottom: root.bottom
        anchors.topMargin: 4

        color: {
            if (!root.enabled) {
                return root.colorScheme.text_disabled
            }

            if (root.error) {
                return root.colorScheme.signal_danger
            }

            return root.colorScheme.text_weak
        }

        type: root.error ? Label.LabelType.Caption_semibold : Label.LabelType.Caption
    }

    ScrollView {
        id: controlView

        anchors.top: label.bottom
        anchors.left: root.left
        anchors.right: root.right
        anchors.bottom: assistiveText.top

        clip: true

        T.TextArea {
            id: control

            implicitWidth: Math.max(
                contentWidth + leftPadding + rightPadding,
                implicitBackgroundWidth + leftInset + rightInset,
                placeholder.implicitWidth + leftPadding + rightPadding
            )
            implicitHeight: Math.max(
                contentHeight + topPadding + bottomPadding,
                implicitBackgroundHeight + topInset + bottomInset,
                placeholder.implicitHeight + topPadding + bottomPadding
            )

            padding: 8
            leftPadding: 12

            color: control.enabled ? root.colorScheme.text_norm : root.colorScheme.text_disabled
            placeholderTextColor: control.enabled ? root.colorScheme.text_hint : root.colorScheme.text_disabled

            selectionColor: control.palette.highlight
            selectedTextColor: control.palette.highlightedText

            onEditingFinished: root.editingFinished()

            cursorDelegate: Rectangle {
                id: cursor
                width: 1
                color: root.colorScheme.interaction_norm
                visible: control.activeFocus && !control.readOnly && control.selectionStart === control.selectionEnd

                Connections {
                    target: control
                    onCursorPositionChanged: {
                        // keep a moving cursor visible
                        cursor.opacity = 1
                        timer.restart()
                    }
                }

                Timer {
                    id: timer
                    running: control.activeFocus && !control.readOnly
                    repeat: true
                    interval: Qt.styleHints.cursorFlashTime / 2
                    onTriggered: cursor.opacity = !cursor.opacity ? 1 : 0
                    // force the cursor visible when gaining focus
                    onRunningChanged: cursor.opacity = 1
                }
            }

            PlaceholderText {
                id: placeholder
                x: control.leftPadding
                y: control.topPadding
                width: control.width - (control.leftPadding + control.rightPadding)
                height: control.height - (control.topPadding + control.bottomPadding)

                text: control.placeholderText
                font: control.font
                color: control.placeholderTextColor
                verticalAlignment: control.verticalAlignment
                visible: !control.length && !control.preeditText && (!control.activeFocus || control.horizontalAlignment !== Qt.AlignHCenter)
                elide: Text.ElideRight
                renderType: control.renderType
            }
        }
    }
}
