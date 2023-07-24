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

    signal resume
    signal questionAnswered

    function setDefaultValue() {
    }

    function next() {

        if (stackLayout.currentIndex >=(stackLayout.count - 1))  {
            root.questionAnswered();
        }
        else
            {
            ++stackLayout.currentIndex
            root.setDefaultValue();
        }
    }

    function previous() {
        if (stackLayout.currentIndex === 0) {
            root.resume()
        }
        else {
            --stackLayout.currentIndex
            root.setDefaultValue();
        }
    }

    fillHeight: true

    onBack: {
        root.previous();
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

    Label {
        Layout.fillWidth: true
        colorScheme: root.colorScheme
        text: qsTr("Step " + (stackLayout.currentIndex + 1) + " of " + stackLayout.count )
        type: Label.Caption
    }

    StackLayout {
        id: stackLayout
        QuestionItem {
            Layout.fillWidth: true
            colorScheme: root.colorScheme

            text: "question 1"
            type: QuestionItem.InputType.TextInput
            mandatory: true
            tips: ""
            errorString: "please answer the question"
        }

        QuestionItem {
            Layout.fillWidth: true
            colorScheme: root.colorScheme

            text: "question 2"
            type: QuestionItem.InputType.Radio
            mandatory: true
            answerList: ["answer A", "answer B", "answer C","answer D"]
            tips: ""
            errorString: "please answer the question"
        }

        QuestionItem {
            Layout.fillWidth: true
            colorScheme: root.colorScheme

            text: "question 3"
            type: QuestionItem.InputType.Checkbox
            mandatory: true
            answerList: ["answer 1", "answer 2", "answer 3","answer 4"]
            tips: ""
            errorString: "please answer the question"
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
            if (stackLayout.children[stackLayout.currentIndex].validate()) {
                next();
            }

        }
    }
}