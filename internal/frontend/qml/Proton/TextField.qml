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
import QtQuick.Layouts 1.12

Item {
    id: root
    property var colorScheme: parent.colorScheme ? parent.colorScheme : Style.currentStyle

    property alias background: control.background
    property alias bottomInset: control.bottomInset
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
    property alias acceptableInput: control.acceptableInput
    property alias activeFocusOnPress: control.activeFocusOnPress
    property alias autoScroll: control.autoScroll
    property alias bottomPadding: control.bottomPadding
    property alias canPaste: control.canPaste
    property alias canRedo: control.canRedo
    property alias canUndo: control.canUndo
    property alias color: control.color
    //property alias contentHeight: control.contentHeight
    //property alias contentWidth: control.contentWidth
    property alias cursorDelegate: control.cursorDelegate
    property alias cursorPosition: control.cursorPosition
    property alias cursorRectangle: control.cursorRectangle
    property alias cursorVisible: control.cursorVisible
    property alias displayText: control.displayText
    property alias effectiveHorizontalAlignment: control.effectiveHorizontalAlignment
    property alias font: control.font
    property alias horizontalAlignment: control.horizontalAlignment
    property alias inputMask: control.inputMask
    property alias inputMethodComposing: control.inputMethodComposing
    property alias inputMethodHints: control.inputMethodHints
    property alias leftPadding: control.leftPadding
    property alias length: control.length
    property alias maximumLength: control.maximumLength
    property alias mouseSelectionMode: control.mouseSelectionMode
    property alias overwriteMode: control.overwriteMode
    property alias padding: control.padding
    property alias passwordCharacter: control.passwordCharacter
    property alias passwordMaskDelay: control.passwordMaskDelay
    property alias persistentSelection: control.persistentSelection
    property alias preeditText: control.preeditText
    property alias readOnly: control.readOnly
    property alias renderType: control.renderType
    property alias rightPadding: control.rightPadding
    property alias selectByMouse: control.selectByMouse
    property alias selectedText: control.selectedText
    property alias selectedTextColor: control.selectedTextColor
    property alias selectionColor: control.selectionColor
    property alias selectionEnd: control.selectionEnd
    property alias selectionStart: control.selectionStart
    property alias text: control.text
    property alias validator: control.validator
    property alias verticalAlignment: control.verticalAlignment
    property alias wrapMode: control.wrapMode

    implicitWidth: children[0].implicitWidth
    implicitHeight: children[0].implicitHeight

    property alias label: label.text
    property alias hint: hint.text
    property alias assistiveText: assistiveText.text

    property var echoMode: TextInput.Normal

    property bool error: false

    signal accepted()
    signal editingFinished()
    signal textEdited()

    function clear() { control.clear() }
    function copy() { control.copy() }
    function cut() { control.cut() }
    function deselect() { control.deselect() }
    function ensureVisible(position) { control.ensureVisible(position) }
    function getText(start, end) { control.getText(start, end) }
    function insert(position, text) { control.insert(position, text) }
    function isRightToLeft(start, end) { control.isRightToLeft(start, end) }
    function moveCursorSelection(position, mode) { control.moveCursorSelection(position, mode) }
    function paste() { control.paste() }
    function positionAt(x, y, position) { control.positionAt(x, y, position) }
    function positionToRectangle(pos) { control.positionToRectangle(pos) }
    function redo() { control.redo() }
    function remove(start, end) { control.remove(start, end) }
    function select(start, end) { control.select(start, end) }
    function selectAll() { control.selectAll() }
    function selectWord() { control.selectWord() }
    function undo() { control.undo() }
    function forceActiveFocus() {control.forceActiveFocus()}

    ColumnLayout {
        anchors.fill: parent
        spacing: 0

        RowLayout {
            Layout.fillWidth: true
            spacing: 0

            ProtonLabel {
                id: label
                Layout.fillHeight: true
                Layout.fillWidth: true
                color: root.enabled ? colorScheme.text_norm : colorScheme.text_disabled
                font.weight: Style.fontWidth_600
                state: "body"
            }

            ProtonLabel {
                id: hint
                Layout.fillHeight: true
                Layout.fillWidth: true
                color: root.enabled ? colorScheme.text_weak : colorScheme.text_disabled
                horizontalAlignment: Text.AlignRight
                state: "caption"
            }
        }

        // Background is moved away from within control to cover eye button as well.
        // In case it will remain as control background property - control's width
        // will be adjusted to background's width making text field and eye button overlap
        Rectangle {
            id: background

            Layout.fillHeight: true
            Layout.fillWidth: true

            radius: 4
            visible: true
            color: colorScheme.background_norm
            border.color: {
                if (!control.enabled) {
                    return colorScheme.field_disabled
                }

                if (control.activeFocus) {
                    return colorScheme.interaction_norm
                }

                if (root.error) {
                    return colorScheme.signal_danger
                }

                if (control.hovered) {
                    return colorScheme.field_hover
                }

                return colorScheme.field_norm
            }
            border.width: 1

            implicitWidth: children[0].implicitWidth
            implicitHeight: children[0].implicitHeight

            RowLayout {
                anchors.fill: parent
                spacing: 0

                T.TextField {
                    id: control

                    Layout.fillHeight: true
                    Layout.fillWidth: true

                    implicitWidth: implicitBackgroundWidth + leftInset + rightInset
                    || Math.max(contentWidth, placeholder.implicitWidth) + leftPadding + rightPadding
                    implicitHeight: Math.max(implicitBackgroundHeight + topInset + bottomInset,
                    contentHeight + topPadding + bottomPadding,
                    placeholder.implicitHeight + topPadding + bottomPadding)

                    padding: 8
                    leftPadding: 12

                    color: control.enabled ? colorScheme.text_norm : colorScheme.text_disabled

                    selectionColor: control.palette.highlight
                    selectedTextColor: control.palette.highlightedText
                    placeholderTextColor: control.enabled ? colorScheme.text_hint : colorScheme.text_disabled
                    verticalAlignment: TextInput.AlignVCenter

                    cursorDelegate: Rectangle {
                        id: cursor
                        width: 1
                        color: colorScheme.interaction_norm
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

                    background: Item {
                        implicitWidth: 80
                        implicitHeight: 36
                        visible: false
                    }

                    onAccepted: {
                        root.accepted()
                    }
                    onEditingFinished: {
                        root.editingFinished()
                    }
                    onTextEdited: {
                        root.textEdited()
                    }
                }

                Button {
                    id: eyeButton

                    Layout.fillHeight: true

                    visible: root.echoMode === TextInput.Password
                    icon.source: control.echoMode == TextInput.Password ? "../icons/ic-eye.svg" : "../icons/ic-eye-slash.svg"
                    icon.color: control.color
                    background: Rectangle{color: "#00000000"}
                    onClicked: {
                        if (control.echoMode == TextInput.Password) {
                            control.echoMode = TextInput.Normal
                        } else {
                            control.echoMode = TextInput.Password
                        }
                    }
                    Component.onCompleted: control.echoMode = root.echoMode
                }
            }
        }

        RowLayout {
            Layout.fillWidth: true
            spacing: 0

            // FIXME: maybe somewhere in the future there will be an Icon component capable of setting color to the icon
            // but before that moment we need to use IconLabel
            IconLabel {
                id: errorIcon

                visible: root.error && (assistiveText.text.length > 0)
                icon.source: "../icons/ic-exclamation-circle-filled.svg"
                icon.color: colorScheme.signal_danger
            }

            ProtonLabel {
                id: assistiveText

                Layout.fillHeight: true
                Layout.fillWidth: true
                Layout.leftMargin: 4

                color: {
                    if (!root.enabled) {
                        return colorScheme.text_disabled
                    }

                    if (root.error) {
                        return colorScheme.signal_danger
                    }

                    return colorScheme.text_weak
                }

                font.weight: root.error ? Style.fontWidth_600 : Style.fontWidth_400
                state: "caption"
            }
        }
    }
}
