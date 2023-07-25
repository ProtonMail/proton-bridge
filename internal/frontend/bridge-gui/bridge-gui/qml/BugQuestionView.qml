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

    signal questionAnswered

    function setDefaultValue() {
    }

    function setCategoryId(catId) {
        root.categoryId = catId;
    }
    function submit() {
        root.questionAnswered();
    }

    fillHeight: true

    onCategoryIdChanged: {
        root.questionSet = Backend.getQuestionSet(root.categoryId)
        root.setDefaultValue();
    }

    onVisibleChanged: {
        root.setDefaultValue();
    }

    Label {
        Layout.fillWidth: true
        colorScheme: root.colorScheme
        text: qsTr("Give us more details")
        type: Label.Heading
    }

    Repeater {
        model: root.questionSet

        QuestionItem {
            Layout.fillWidth: true

            implicitWidth: parent.implicitWidth

            colorScheme: root.colorScheme
            showSeparator: index < (root.questionSet.length - 1)

            text: root.questions[modelData].text
            tips: root.questions[modelData].tips ? root.questions[modelData].tips : ""
            label: root.questions[modelData].label ? root.questions[modelData].label : ""
            type: root.questions[modelData].type
            answerList: root.questions[modelData].answerList ? root.questions[modelData].answerList : []
        }
    }
    // fill height so the footer label will always be attached to the bottom
    Item {
        Layout.fillHeight: true
    }
    Button {
        id: continueButton
        colorScheme: root.colorScheme
        enabled: !loading
        text: qsTr("Continue")

        onClicked: {
            submit();
        }
    }
}