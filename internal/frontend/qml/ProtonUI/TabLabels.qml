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

// Tab labels

import QtQuick 2.8
import ProtonUI 1.0


Rectangle {
    id: root
    // attributes
    property alias model        : tablist.model
    property alias currentIndex : tablist.currentIndex
    property int spacing : Style.tabbar.widthButton + Style.tabbar.spacingButton
    currentIndex: 0

    // appereance
    height : Style.tabbar.height
    color  : Style.tabbar.background

    // content
    ListView {
        id: tablist
        // dimensions
        anchors {
            fill: root
            leftMargin   : Style.tabbar.leftMargin
            rightMargin  : Style.main.rightMargin
            bottomMargin : Style.tabbar.bottomMargin
        }
        spacing: Style.tabbar.spacingButton
        interactive  : false
        // style
        orientation: Qt.Horizontal
        delegate: TabButton {
            anchors.bottom : parent.bottom
            title     : modelData.title
            iconText  : modelData.iconText
            state     : index == tablist.currentIndex ? "activated" : "deactivated"
            onClicked : {
                tablist.currentIndex = index
            }
        }
    }

    // Quit button
    TabButton {
        id: buttonQuit
        title    : qsTr("Close Bridge", "quits the application")
        iconText : Style.fa.power_off
        state    : "deactivated"
        visible  : Style.tabbar.rightButton=="quit"
        anchors {
            right        : root.right
            bottom       : root.bottom
            rightMargin  : Style.main.rightMargin
            bottomMargin : Style.tabbar.bottomMargin
        }

        Accessible.description: buttonQuit.title

        onClicked : {
            dialogGlobal.state = "quit"
            dialogGlobal.show()
        }
    }

    // Add account
    TabButton {
        id: buttonAddAccount
        title    : qsTr("Add account", "start the authentication to add account")
        iconText : Style.fa.plus_circle
        state    : "deactivated"
        visible  : Style.tabbar.rightButton=="add account"
        anchors {
            right        : root.right
            bottom       : root.bottom
            rightMargin  : Style.main.rightMargin
            bottomMargin : Style.tabbar.bottomMargin
        }

        Accessible.description: buttonAddAccount.title

        onClicked : dialogAddUser.show()
    }

    function focusButton() {
        tablist.currentItem.forceActiveFocus()
        tablist.currentItem.Accessible.focusedChanged(true)
    }
}

