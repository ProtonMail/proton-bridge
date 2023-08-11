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
import Proton

SettingsView {
    id: root

    property var questions:Backend.bugQuestions
    property var categoryId:0
    property var questionSet:ListModel{}
    property bool error: questionRepeater.error
    signal questionAnswered

    function setCategoryId(catId) {
        root.categoryId = catId;
    }
    function submit() {
        root.questionAnswered();
    }

    fillHeight: true

    onCategoryIdChanged: {
        root.questionSet = Backend.getQuestionSet(root.categoryId)
    }

    Label {
        Layout.fillWidth: true
        colorScheme: root.colorScheme
        text: qsTr("Provide more details")
        type: Label.Heading
    }

    Label {
        Layout.fillWidth: true
        colorScheme: root.colorScheme
        text: qsTr(Backend.getBugCategory(root.categoryId))
        type: Label.Title
    }

    TextEdit {
        Layout.fillWidth: true
        color: root.colorScheme.text_weak
        font.family: ProtonStyle.font_family
        font.letterSpacing: ProtonStyle.caption_letter_spacing
        font.pixelSize: ProtonStyle.caption_font_size
        font.weight: ProtonStyle.fontWeight_400
        textFormat: Text.MarkdownText
        readOnly: true
        selectByMouse: true
        selectedTextColor: root.colorScheme.text_invert
        // No way to set lineHeight: ProtonStyle.caption_line_height
        selectionColor: root.colorScheme.interaction_norm
        text: qsTr("* Mandatory questions")
        wrapMode: Text.WordWrap
    }

    Repeater {
        id: questionRepeater
        model: root.questionSet
        property bool error :{
            for (var i = 0; i < questionRepeater.count; i++) {
                if (questionRepeater.itemAt(i).error)
                    return true;
            }
            return false;
        }

        function validate(){
            for (var i = 0; i < questionRepeater.count; i++) {
                questionRepeater.itemAt(i).validate()
            }
        }

        QuestionItem {
            Layout.fillWidth: true

            colorScheme: root.colorScheme
            showSeparator: index < (root.questionSet.length - 1)

            text: root.questions[modelData].text
            tips: root.questions[modelData].tips ? root.questions[modelData].tips : ""
            label: root.questions[modelData].label ? root.questions[modelData].label : ""
            type: root.questions[modelData].type
            mandatory: root.questions[modelData].mandatory ? root.questions[modelData].mandatory : false
            answerList: root.questions[modelData].answerList ? root.questions[modelData].answerList : []
            maxChar: root.questions[modelData].maxChar ? root.questions[modelData].maxChar : 150

            onAnswerChanged: {
                Backend.setQuestionAnswer(modelData, answer);
            }

            Connections {
                function onVisibleChanged() {
                    setDefaultValue(Backend.getQuestionAnswer(modelData))
                }
                target: root
            }
        }
    }
    // fill height so the footer label will always be attached to the bottom
    Item {
        Layout.fillHeight: true
    }
    Button {
        id: continueButton
        colorScheme: root.colorScheme
        enabled: !loading && !root.error
        text: qsTr("Continue")

        onClicked: {
            questionRepeater.validate()
            if (!root.error)
                submit();
        }
    }
}