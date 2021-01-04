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

import QtQuick 2.8
import QtQuick.Controls 2.1
import ProtonUI 1.0

Item {
    id: root

    signal addAccount()

    property alias numAccounts      : listAccounts.count
    property alias model            : listAccounts.model
    property alias delegate         : listAccounts.delegate
    property int separatorNoAccount : viewContent.height-Style.accounts.heightFooter
    property bool hasFooter : true

    // must have wrapper
    Rectangle {
        id: wrapper
        anchors.centerIn: parent
        width:  parent.width
        height: parent.height
        color: Style.main.background

        // content
        ListView {
            id: listAccounts
            anchors {
                top    : parent.top
                left   : parent.left
                right  : parent.right
                bottom : hasFooter ? addAccFooter.top : parent.bottom
            }
            orientation: ListView.Vertical
            clip: true
            cacheBuffer: 2500
            boundsBehavior: Flickable.StopAtBounds
            ScrollBar.vertical: ScrollBar {
                anchors {
                    right: parent.right
                    rightMargin: Style.main.rightMargin/4
                }
                width: Style.main.rightMargin/3
                Accessible.ignored: true
            }
            header: Rectangle {
                width  : wrapper.width
                height : root.numAccounts!=0 ? Style.accounts.heightHeader : root.separatorNoAccount
                color  : "transparent"
                AccessibleText { // Placeholder on empty
                    anchors {
                        centerIn: parent
                    }
                    visible: root.numAccounts==0
                    text  : qsTr("No accounts added", "displayed when there are no accounts added")
                    font.pointSize :  Style.main.fontSize * Style.pt
                    color : Style.main.textDisabled
                }
                Text { // Account
                    anchors {
                        left           : parent.left
                        leftMargin     : Style.main.leftMargin
                        verticalCenter : parent.verticalCenter
                    }
                    visible: root.numAccounts!=0
                    font.bold : true
                    font.pointSize :  Style.main.fontSize * Style.pt
                    text  : qsTr("ACCOUNT", "title of column that displays account name")
                    color : Style.main.textDisabled
                }
                Text { // Status
                    anchors {
                        left           : parent.left
                        leftMargin     : viewContent.width/2
                        verticalCenter : parent.verticalCenter
                    }
                    visible: root.numAccounts!=0
                    font.bold : true
                    font.pointSize :  Style.main.fontSize * Style.pt
                    text  : qsTr("STATUS", "title of column that displays connected or disconnected status")
                    color : Style.main.textDisabled
                }
                Text { // Actions
                    anchors {
                        left           : parent.left
                        leftMargin     : 5.5*viewContent.width/8
                        verticalCenter : parent.verticalCenter
                    }
                    visible: root.numAccounts!=0
                    font.bold : true
                    font.pointSize :  Style.main.fontSize * Style.pt
                    text  : qsTr("ACTIONS", "title of column that displays log out and log in actions for each account")
                    color : Style.main.textDisabled
                }
                // line
                Rectangle {
                    anchors {
                        left   : parent.left
                        right  : parent.right
                        bottom : parent.bottom
                    }
                    visible: root.numAccounts!=0
                    color: Style.accounts.line
                    height: Style.accounts.heightLine
                }
            }
        }


        AddAccountBar {
            id: addAccFooter
            visible: hasFooter
            anchors {
                left   : parent.left
                bottom : parent.bottom
            }
        }
    }

    Shortcut {
        sequence: StandardKey.SelectAll
        onActivated: root.addAccount()
    }
}
