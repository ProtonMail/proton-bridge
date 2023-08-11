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
import Notifications

Item {
    id: root

    property ColorScheme colorScheme
    property string selectedAddress
    property int categoryId: -1

    signal back
    signal bugReportWasSent

    onVisibleChanged: {
        root.showBugCategory();
    }

    function showBugCategory() {
        bugReportFlow.currentIndex = 0;
    }
    function showBugQuestion() {
        bugQuestion.setCategoryId(root.categoryId);
        bugQuestion.positionViewAtBegining();
        bugReportFlow.currentIndex = 1;
    }
    function showBugReport() {
        bugReport.setCategoryId(root.categoryId);
        bugReportFlow.currentIndex = 2;
    }

    Rectangle {
        anchors.fill: parent

        Layout.fillHeight: true // right content background
        Layout.fillWidth: true
        color: colorScheme.background_norm

        StackLayout {
            id: bugReportFlow

            anchors.fill: parent

            BugCategoryView {
                // 0
                id: bugCategory
                colorScheme: root.colorScheme

                onBack: {
                    root.back()
                }
                onCategorySelected: function(categoryId){
                    root.categoryId = categoryId
                    root.showBugQuestion();
                }
            }
            BugQuestionView {
                // 1
                id: bugQuestion
                colorScheme: root.colorScheme

                onBack: {
                    root.showBugCategory();
                }
                onQuestionAnswered: {
                    root.showBugReport();
                }
            }
            BugReportView {
                // 2
                id: bugReport
                colorScheme: root.colorScheme
                selectedAddress: root.selectedAddress

                onBack: {
                    root.showBugQuestion();
                }
                onBugReportWasSent: {
                    Backend.clearAnswers();
                    root.bugReportWasSent();
                }
            }
        }
    }
}