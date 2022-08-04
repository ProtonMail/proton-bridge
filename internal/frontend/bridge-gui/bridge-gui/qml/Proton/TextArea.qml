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

import QtQml
import QtQuick
import QtQuick.Controls
import QtQuick.Controls.impl
import QtQuick.Templates as T
import QtQuick.Layouts

import "." as Proton

FocusScope {
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
    // We are using our own type of validators. It should be a function
    // returning an error string in case of error and undefined if no error
    property var validator
    property alias verticalAlignment: control.verticalAlignment
    property alias wrapMode: control.wrapMode

    implicitWidth: children[0].implicitWidth
    implicitHeight: children[0].implicitHeight

    property alias label: label.text
    property alias hint: hint.text
    property string assistiveText
    property string errorString

    property bool error: false

    signal editingFinished()

    function append(text) { return control.append(text) }
    function clear() { return control.clear() }
    function copy() { return control.copy() }
    function cut() { return control.cut() }
    function deselect() { return control.deselect() }
    function getFormattedText(start, end) { return control.getFormattedText(start, end) }
    function getText(start, end) { return control.getText(start, end) }
    function insert(position, text) { return control.insert(position, text) }
    function isRightToLeft(start, end) { return control.isRightToLeft(start, end) }
    function linkAt(x, y) { return control.linkAt(x, y) }
    function moveCursorSelection(position, mode) { return control.moveCursorSelection(position, mode) }
    function paste() { return control.paste() }
    function positionAt(x, y) { return control.positionAt(x, y) }
    function positionToRectangle(position) { return control.positionToRectangle(position) }
    function redo() { return control.redo() }
    function remove(start, end) { return control.remove(start, end) }
    function select(start, end) { return control.select(start, end) }
    function selectAll() { return control.selectAll() }
    function selectWord() { return control.selectWord() }
    function undo() { return control.undo() }

    // Calculates the height of the component to make exactly lineNum visible in edit area
    function heightForLinesVisible(lineNum) {
        var totalHeight = 0
        totalHeight += headerLayout.height
        totalHeight += footerLayout.height
        totalHeight += control.topPadding + control.bottomPadding
        totalHeight += lineNum * fontMetrics.height
        return totalHeight
    }

    FontMetrics {
        id: fontMetrics
        font: control.font
    }

    ColumnLayout {
        anchors.fill: parent
        spacing: 0

        RowLayout {
            id: headerLayout
            Layout.fillWidth: true
            spacing: 0

            Proton.Label {
                colorScheme: root.colorScheme
                id: label

                Layout.fillWidth: true

                color: root.enabled ? root.colorScheme.text_norm : root.colorScheme.text_disabled

                type: Proton.Label.LabelType.Body_semibold
            }

            Proton.Label {
                colorScheme: root.colorScheme
                id: hint

                Layout.fillWidth: true

                color: root.enabled ? root.colorScheme.text_weak : root.colorScheme.text_disabled
                horizontalAlignment: Text.AlignRight
                type: Proton.Label.LabelType.Caption
            }
        }

        ScrollView {
            id: controlView

            Layout.fillHeight: true
            Layout.fillWidth: true

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

                topPadding: 8
                bottomPadding: 8
                leftPadding: 12
                rightPadding: 12

                font.family: ProtonStyle.font_family
                font.weight: ProtonStyle.fontWeight_400
                font.pixelSize: ProtonStyle.body_font_size
                font.letterSpacing: ProtonStyle.body_letter_spacing

                color: control.enabled ? root.colorScheme.text_norm : root.colorScheme.text_disabled
                placeholderTextColor: control.enabled ? root.colorScheme.text_hint : root.colorScheme.text_disabled
                selectionColor: control.palette.highlight
                selectedTextColor: control.palette.highlightedText

                onEditingFinished: root.editingFinished()

                wrapMode: TextInput.Wrap

                // enforcing default focus here within component
                focus: root.focus

                KeyNavigation.priority: root.KeyNavigation.priority
                KeyNavigation.backtab: root.KeyNavigation.backtab
                KeyNavigation.tab: root.KeyNavigation.tab
                KeyNavigation.up: root.KeyNavigation.up
                KeyNavigation.down: root.KeyNavigation.down
                KeyNavigation.left: root.KeyNavigation.left
                KeyNavigation.right: root.KeyNavigation.right

                selectByMouse: true

                cursorDelegate: Rectangle {
                    id: cursor
                    width: 1
                    color: root.colorScheme.interaction_norm
                    visible: control.activeFocus && !control.readOnly && control.selectionStart === control.selectionEnd

                    Connections {
                        target: control
                        function onCursorPositionChanged() {
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

                background: Rectangle {
                    anchors.fill: parent

                    radius: ProtonStyle.input_radius
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
            }
        }

        RowLayout {
            id: footerLayout
            Layout.fillWidth: true
            spacing: 0

            ColorImage {
                id: errorIcon

                Layout.rightMargin: 4

                visible: root.error && (assistiveText.text.length > 0)
                source: "/qml/icons/ic-exclamation-circle-filled.svg"
                color: root.colorScheme.signal_danger
                height: assistiveText.height
                sourceSize.height: assistiveText.height
            }

            Proton.Label {
                colorScheme: root.colorScheme
                id: assistiveText

                Layout.fillWidth: true

                text: root.error ? root.errorString : root.assistiveText

                color: {
                    if (!root.enabled) {
                        return root.colorScheme.text_disabled
                    }

                    if (root.error) {
                        return root.colorScheme.signal_danger
                    }

                    return root.colorScheme.text_weak
                }

                type: root.error ? Proton.Label.LabelType.Caption_semibold : Proton.Label.LabelType.Caption
            }
        }
    }

    property bool validateOnEditingFinished: true
    onEditingFinished: {
        if (!validateOnEditingFinished) {
            return
        }
        validate()
    }

    function validate() {
        if (validator === undefined) {
            return
        }

        var error = validator(text)

        if (error) {
            root.error = true
            root.errorString = error
        } else {
            root.error = false
            root.errorString = ""
        }
    }

    onTextChanged: {
        root.error = false
        root.errorString = ""
    }
}
