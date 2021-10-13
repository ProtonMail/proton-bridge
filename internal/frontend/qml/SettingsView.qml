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

Item {
    id: root

    property var colorScheme
    property var backend
    default property alias items: content.children

    signal back()

    property int _leftMargin: 64
    property int _rightMargin: 64
    property int _topMargin: 32
    property int _bottomMargin: 32
    property int _spacing: 20


    ScrollView {
        clip: true

        width:root.width
        height:root.height

        contentWidth: content.width + content.anchors.leftMargin + content.anchors.rightMargin
        contentHeight: content.height + content.anchors.topMargin + content.anchors.bottomMargin

        ColumnLayout {
            id: content
            spacing: root._spacing
            width: root.width - (root._leftMargin + root._rightMargin)

            anchors{
                top: parent.top
                left: parent.left
                topMargin: root._topMargin
                bottomMargin: root._bottomMargin
                leftMargin: root._leftMargin
                rightMargin: root._rightMargin
            }
        }
    }

    Button {
        id: backButton
        anchors {
            top: parent.top
            left: parent.left
            topMargin: root._topMargin
            leftMargin: (root._leftMargin-backButton.width) / 2
        }
        colorScheme: root.colorScheme
        onClicked: root.back()
        icon.source: "icons/ic-arrow-left.svg"
        secondary: true
        horizontalPadding: 8
    }
}
