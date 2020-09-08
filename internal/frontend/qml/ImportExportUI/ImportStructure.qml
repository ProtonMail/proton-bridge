// Copyright (c) 2020 Proton Technologies AG
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

// List of import folder and their target
import QtQuick 2.8
import QtQuick.Controls 2.2
import ProtonUI 1.0
import ImportExportUI 1.0

Rectangle {
    id: root
    property string titleFrom
    property string titleTo
    property bool   hasItems: true

    color : Style.transparent

    Rectangle {
        anchors.fill: root
        radius : Style.dialog.radiusButton
        color : Style.transparent
        border {
            color : Style.main.line
            width : 1.5*Style.dialog.borderInput
        }


        Text { // placeholder
            visible: !root.hasItems
            anchors.centerIn: parent
            color: Style.main.textDisabled
            font {
                pointSize: Style.dialog.fontSize * Style.pt
            }
            horizontalAlignment: Text.AlignHCenter
            verticalAlignment: Text.AlignVCenter
            text: qsTr("No emails found for this source.","todo")
        }
    }

    anchors {
        left   : parent.left
        right  : parent.right
        top    : parent.top
        bottom : parent.bottom

        leftMargin   : Style.main.leftMargin
        rightMargin  : Style.main.leftMargin
        topMargin    : Style.main.topMargin
        bottomMargin : Style.main.bottomMargin
    }

    ListView {
        id: listview
        clip           : true
        orientation    : ListView.Vertical
        boundsBehavior : Flickable.StopAtBounds
        model          : transferRules
        cacheBuffer    : 10000
        delegate       : ImportDelegate {
            width: root.width
        }

        anchors {
            top: titleBox.bottom
            bottom: root.bottom
            left: root.left
            right: root.right
            margins : Style.dialog.borderInput
            bottomMargin: Style.dialog.radiusButton
        }

        ScrollBar.vertical: ScrollBar {
            anchors {
                right: parent.right
                rightMargin: Style.main.rightMargin/4
            }
            width: Style.main.rightMargin/3
            Accessible.ignored: true
        }
    }

    Rectangle {
        id: titleBox
        anchors {
            top: parent.top
            left: parent.left
            right: parent.right
        }
        height: Style.main.fontSize *2
        color : Style.transparent

        Text {
            id: textTitleFrom
            anchors {
                left: parent.left
                verticalCenter: parent.verticalCenter
                leftMargin: {
                    if (listview.currentItem === null) return 0
                    else return listview.currentItem.leftMargin1
                }
            }
            text: "<b>"+qsTr("From:")+"</b> " + root.titleFrom
            color: Style.main.text
            width: listview.currentItem === null  ? 0 : (listview.currentItem.leftMargin2 - listview.currentItem.leftMargin1 - Style.dialog.spacing)
            elide: Text.ElideMiddle
        }

        Text {
            id: textTitleTo
            anchors {
                left: parent.left
                verticalCenter: parent.verticalCenter
                leftMargin: {
                    if (listview.currentIndex<0) return root.width/3
                    else return listview.currentItem.leftMargin2
                }
            }
            text: "<b>"+qsTr("To:")+"</b> " + root.titleTo
            color: Style.main.text
        }
    }

    Rectangle {
        id: line
        anchors {
            left  : titleBox.left
            right : titleBox.right
            top   : titleBox.bottom
        }
        height: Style.dialog.borderInput
        color: Style.main.line
    }
}
