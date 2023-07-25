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
    property var titles: ["Category", "Description", "Confirmation"]
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
        bugReportFlow.currentIndex = 1;
        bugQuestion.setCategoryId(root.categoryId);
    }
    function showBugReport() {
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
                path: root.titles
                currPath: 0

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
                path: root.titles
                currPath: 1

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
                path: root.titles
                currPath: 2

                onBack: {
                    root.showBugQuestion();
                }
                onBugReportWasSent: {
                    root.bugReportWasSent();
                }
            }
        }
    }
}