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
import QtQml
import QtQuick
import QtQuick.Controls
import QtQuick.Controls.impl
import QtQuick.Templates as T
import QtQuick.Layouts
import "." as Proton

FocusScope {
    id: root

    property alias acceptableInput: control.acceptableInput
    property alias activeFocusOnPress: control.activeFocusOnPress
    property string assistiveText
    property alias autoScroll: control.autoScroll
    property alias background: control.background
    property alias bottomInset: control.bottomInset
    property alias bottomPadding: control.bottomPadding
    property alias canPaste: control.canPaste
    property alias canRedo: control.canRedo
    property alias canUndo: control.canUndo
    property alias color: control.color
    property ColorScheme colorScheme
    //property alias contentHeight: control.contentHeight
    //property alias contentWidth: control.contentWidth
    property alias cursorDelegate: control.cursorDelegate
    property alias cursorPosition: control.cursorPosition
    property alias cursorRectangle: control.cursorRectangle
    property alias cursorVisible: control.cursorVisible
    property alias displayText: control.displayText
    property int echoMode: TextInput.Normal
    property alias effectiveHorizontalAlignment: control.effectiveHorizontalAlignment
    property bool error: false
    property string errorString
    property alias focusReason: control.focusReason
    property alias font: control.font
    property alias hint: hint.text
    property alias horizontalAlignment: control.horizontalAlignment
    property alias hoverEnabled: control.hoverEnabled
    property alias hovered: control.hovered
    property alias implicitBackgroundHeight: control.implicitBackgroundHeight
    property alias implicitBackgroundWidth: control.implicitBackgroundWidth
    property alias inputMask: control.inputMask
    property alias inputMethodComposing: control.inputMethodComposing
    property alias inputMethodHints: control.inputMethodHints
    property alias label: label.text
    property alias leftInset: control.leftInset
    property alias leftPadding: control.leftPadding
    property alias length: control.length
    property alias maximumLength: control.maximumLength
    property alias mouseSelectionMode: control.mouseSelectionMode
    property alias overwriteMode: control.overwriteMode
    property alias padding: control.padding
    property alias palette: control.palette
    property alias passwordCharacter: control.passwordCharacter
    property alias passwordMaskDelay: control.passwordMaskDelay
    property alias persistentSelection: control.persistentSelection
    property alias placeholderText: control.placeholderText
    property alias placeholderTextColor: control.placeholderTextColor
    property alias preeditText: control.preeditText
    property alias readOnly: control.readOnly
    property alias renderType: control.renderType
    property alias rightInset: control.rightInset
    property alias rightPadding: control.rightPadding
    property alias selectByMouse: control.selectByMouse
    property alias selectedText: control.selectedText
    property alias selectedTextColor: control.selectedTextColor
    property alias selectionColor: control.selectionColor
    property alias selectionEnd: control.selectionEnd
    property alias selectionStart: control.selectionStart
    property alias text: control.text
    property alias topInset: control.topInset
    property bool validateOnEditingFinished: true
    // We are using our own type of validators. It should be a function
    // returning an error string in case of error and undefined if no error
    property var validator
    property alias verticalAlignment: control.verticalAlignment
    property alias wrapMode: control.wrapMode

    signal accepted
    signal editingFinished
    signal textEdited

    function clear() {
        control.clear();
    }
    function copy() {
        control.copy();
    }
    function cut() {
        control.cut();
    }
    function deselect() {
        control.deselect();
    }
    function ensureVisible(position) {
        control.ensureVisible(position);
    }
    function forceActiveFocus() {
        control.forceActiveFocus();
    }
    function getText(start, end) {
        control.getText(start, end);
    }
    function hidePassword() {
        eyeButton.checked = false;
    }
    function insert(position, text) {
        control.insert(position, text);
    }
    function isRightToLeft(start, end) {
        control.isRightToLeft(start, end);
    }
    function moveCursorSelection(position, mode) {
        control.moveCursorSelection(position, mode);
    }
    function paste() {
        control.paste();
    }
    function positionAt(x, y, position) {
        control.positionAt(x, y, position);
    }
    function positionToRectangle(pos) {
        control.positionToRectangle(pos);
    }
    function redo() {
        control.redo();
    }
    function remove(start, end) {
        control.remove(start, end);
    }
    function select(start, end) {
        control.select(start, end);
    }
    function selectAll() {
        control.selectAll();
    }
    function selectWord() {
        control.selectWord();
    }
    function showPassword() {
        eyeButton.checked = true;
    }
    function undo() {
        control.undo();
    }
    function validate() {
        if (validator === undefined) {
            return;
        }
        const error = validator(text);
        if (error) {
            root.error = true;
            root.errorString = error;
        } else {
            root.error = false;
            root.errorString = "";
        }
    }

    implicitHeight: children[0].implicitHeight
    implicitWidth: children[0].implicitWidth
    Accessible.role: Accessible.Grouping
    Accessible.name: label.text

    onEditingFinished: {
        if (!validateOnEditingFinished) {
            return;
        }
        validate();
    }
    onTextChanged: {
        root.error = false;
        root.errorString = "";
    }

    ColumnLayout {
        anchors.fill: parent
        spacing: 0

        RowLayout {
            Layout.fillWidth: true
            spacing: 0

            Proton.Label {
                id: label
                Layout.fillHeight: true
                Layout.fillWidth: true
                colorScheme: root.colorScheme
                type: Proton.Label.LabelType.Body_semibold
            }
            Proton.Label {
                id: hint
                Layout.fillHeight: true
                Layout.fillWidth: true
                color: root.enabled ? root.colorScheme.text_weak : root.colorScheme.text_disabled
                colorScheme: root.colorScheme
                horizontalAlignment: Text.AlignRight
                type: Proton.Label.LabelType.Caption
            }
        }

        // Background is moved away from within control to cover eye button as well.
        // In case it will remain as control background property - control's width
        // will be adjusted to background's width making text field and eye button overlap
        Rectangle {
            id: background
            Layout.fillHeight: true
            Layout.fillWidth: true
            border.color: {
                if (!control.enabled) {
                    return root.colorScheme.field_disabled;
                }
                if (control.activeFocus) {
                    return root.colorScheme.interaction_norm;
                }
                if (root.error) {
                    return root.colorScheme.signal_danger;
                }
                if (control.hovered) {
                    return root.colorScheme.field_hover;
                }
                return root.colorScheme.field_norm;
            }
            border.width: 1
            color: root.colorScheme.background_norm
            implicitHeight: children[0].implicitHeight
            implicitWidth: children[0].implicitWidth
            radius: ProtonStyle.input_radius
            visible: true

            RowLayout {
                anchors.fill: parent
                spacing: 0

                T.TextField {
                    id: control
                    KeyNavigation.backtab: root.KeyNavigation.backtab
                    KeyNavigation.down: root.KeyNavigation.down
                    KeyNavigation.left: root.KeyNavigation.left
                    KeyNavigation.priority: root.KeyNavigation.priority
                    KeyNavigation.right: root.KeyNavigation.right
                    KeyNavigation.tab: root.KeyNavigation.tab
                    KeyNavigation.up: root.KeyNavigation.up
                    Layout.fillHeight: true
                    Layout.fillWidth: true
                    bottomPadding: 8
                    color: control.enabled ? root.colorScheme.text_norm : root.colorScheme.text_disabled
                    echoMode: eyeButton.checked ? TextInput.Normal : root.echoMode

                    // enforcing default focus here within component
                    focus: true
                    font.family: ProtonStyle.font_family
                    font.letterSpacing: ProtonStyle.body_letter_spacing
                    font.pixelSize: ProtonStyle.body_font_size
                    font.weight: ProtonStyle.fontWeight_400
                    implicitHeight: Math.max(implicitBackgroundHeight + topInset + bottomInset, contentHeight + topPadding + bottomPadding, placeholder.implicitHeight + topPadding + bottomPadding)
                    implicitWidth: implicitBackgroundWidth + leftInset + rightInset || Math.max(contentWidth, placeholder.implicitWidth) + leftPadding + rightPadding
                    leftPadding: 12
                    placeholderTextColor: control.enabled ? root.colorScheme.text_hint : root.colorScheme.text_disabled
                    rightPadding: 12
                    selectByMouse: true
                    selectedTextColor: control.palette.highlightedText
                    selectionColor: control.palette.highlight
                    topPadding: 8
                    verticalAlignment: TextInput.AlignVCenter
                    Accessible.name: label.text + qsTr(" edit")

                    background: Item {
                        implicitHeight: 36
                        implicitWidth: 80
                        visible: false
                    }
                    cursorDelegate: Rectangle {
                        id: cursor
                        color: root.colorScheme.interaction_norm
                        visible: control.activeFocus && !control.readOnly && control.selectionStart === control.selectionEnd
                        width: 1

                        Connections {
                            function onCursorPositionChanged() {
                                // keep a moving cursor visible
                                cursor.opacity = 1;
                                timer.restart();
                            }

                            target: control
                        }
                        Timer {
                            id: timer
                            interval: Qt.styleHints.cursorFlashTime / 2
                            repeat: true
                            running: control.activeFocus && !control.readOnly

                            // force the cursor visible when gaining focus
                            onRunningChanged: cursor.opacity = 1
                            onTriggered: cursor.opacity = !cursor.opacity ? 1 : 0
                        }
                    }

                    onAccepted: {
                        root.accepted();
                    }
                    onEditingFinished: {
                        root.editingFinished();
                    }
                    onTextEdited: {
                        root.textEdited();
                    }

                    PlaceholderText {
                        id: placeholder
                        color: control.placeholderTextColor
                        elide: Text.ElideRight
                        font: control.font
                        height: control.height - (control.topPadding + control.bottomPadding)
                        renderType: control.renderType
                        text: control.placeholderText
                        verticalAlignment: control.verticalAlignment
                        visible: !control.length && !control.preeditText && (!control.activeFocus || control.horizontalAlignment !== Qt.AlignHCenter)
                        width: control.width - (control.leftPadding + control.rightPadding)
                        x: control.leftPadding
                        y: control.topPadding
                    }

                    Proton.ContextMenu {
                        parentObject: control
                        colorScheme: root.colorScheme
                        isPassword: control.echoMode === TextInput.Password
                        readOnly: control.readOnly
                    }


                }
                Proton.Button {
                    id: eyeButton
                    Layout.fillHeight: true
                    checkable: true
                    colorScheme: root.colorScheme
                    icon.color: control.color
                    icon.source: checked ? "../icons/ic-eye-slash.svg" : "../icons/ic-eye.svg"
                    visible: root.echoMode === TextInput.Password
                    Accessible.name: label.text + qsTr(" show check")
                }
            }
        }
        RowLayout {
            Layout.fillWidth: true
            spacing: 0

            ColorImage {
                id: errorIcon
                Layout.rightMargin: 4
                color: root.colorScheme.signal_danger
                height: assistiveText.lineHeight
                source: "../icons/ic-exclamation-circle-filled.svg"
                sourceSize.height: assistiveText.lineHeight
                visible: root.error && (assistiveText.text.length > 0)
            }
            Proton.Label {
                id: assistiveText
                Layout.fillHeight: true
                Layout.fillWidth: true
                color: {
                    if (!root.enabled) {
                        return root.colorScheme.text_disabled;
                    }
                    if (root.error) {
                        return root.colorScheme.signal_danger;
                    }
                    return root.colorScheme.text_weak;
                }
                colorScheme: root.colorScheme
                text: root.error ? root.errorString : root.assistiveText
                type: root.error ? Proton.Label.LabelType.Caption_semibold : Proton.Label.LabelType.Caption
                wrapMode: Text.WordWrap
            }
        }
    }
}
