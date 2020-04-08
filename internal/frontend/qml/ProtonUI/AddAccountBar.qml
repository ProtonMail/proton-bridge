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

// Bar with add account button and help

import QtQuick 2.8
import ProtonUI 1.0


Rectangle {
    width  : parent.width
    height : Style.accounts.heightFooter
    color: "transparent"
    Rectangle {
        anchors {
            top   : parent.top
            left  : parent.left
            right : parent.right
        }
        height: Style.accounts.heightLine
        color: Style.accounts.line
    }
    ClickIconText {
        id: buttonAddAccount
        anchors {
            left : parent.left
            leftMargin : Style.main.leftMargin
            verticalCenter : parent.verticalCenter
        }
        textColor  : Style.main.textBlue
        iconText   : Style.fa.plus_circle
        text       : qsTr("Add Account", "begins the flow to log in to an account that is not yet listed")
        textBold   : true
        onClicked  : root.addAccount()
        Accessible.description: {
            if (gui.winMain!=null) {
                return text + (gui.winMain.addAccountTip.visible? ", "+gui.winMain.addAccountTip.text : "")
            }
            return buttonAddAccount.text
        }
    }
    ClickIconText {
        id: buttonHelp
        anchors {
            right          : parent.right
            rightMargin    : Style.main.rightMargin
            verticalCenter : parent.verticalCenter
        }
        textColor  : Style.main.textDisabled
        iconText   : Style.fa.question_circle
        text       : qsTr("Help", "directs the user to the online user guide")
        textBold   : true
        onClicked  : go.openManual()
    }
}
