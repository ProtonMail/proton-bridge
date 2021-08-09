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

import QtQuick 2.13
import QtQuick.Layouts 1.12
import QtQuick.Controls 2.13
import QtQuick.Controls.impl 2.13

import Proton 4.0

ScrollView {
    id: root

    property var colorScheme
    property var backend
    default property alias items: content.children

    signal back()

    property int _leftRightMargins: 64
    property int _topBottomMargins: 68
    property int _spacing: 22

    clip: true
    contentWidth: pane.width
    contentHeight: pane.height

    RowLayout{
        id: pane
        width: root.width

        ColumnLayout {
            id: content
            spacing: root._spacing
            Layout.maximumWidth: root.width - 2*root._leftRightMargins
            Layout.fillWidth: true
            Layout.topMargin: root._topBottomMargins
            Layout.bottomMargin: root._topBottomMargins
            Layout.leftMargin: root._leftRightMargins
            Layout.rightMargin: root._leftRightMargins
        }
    }

    Button {
        anchors {
            top: parent.top
            left: parent.left
            topMargin: 10
            leftMargin: 18
        }
        colorScheme: root.colorScheme
        onClicked: root.back()
        icon.source: "icons/ic-arrow-left.svg"
        secondary: true
        horizontalPadding: 8
    }
}
