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

// Window for imap and smtp settings

import QtQuick 2.8
import QtQuick.Window 2.2
import BridgeUI 1.0
import ProtonUI 1.0


Window {
    id:root
    width  : Style.info.width
    height : Style.info.height
    minimumWidth  : Style.info.width
    minimumHeight : Style.info.height
    maximumWidth  : Style.info.width
    maximumHeight : Style.info.height
    color: "transparent"
    flags  : Qt.Window | Qt.Dialog | Qt.FramelessWindowHint
    title  : address

    Accessible.role: Accessible.Window
    Accessible.name: qsTr("Configuration information for %1").arg(address)
    Accessible.description: Accessible.name

    property QtObject accData : QtObject { // avoid null-pointer error
        property string account : "undef"
        property string aliases : "undef"
        property string hostname : "undef"
        property string password : "undef"
        property int portIMAP : 0
        property int portSMTP : 0
    }
    property string address : "undef"
    property int indexAccount : 0
    property int indexAddress : 0

    WindowTitleBar {
        id: titleBar
        window: root
    }

    Rectangle { // background
        color: Style.main.background
        anchors {
            left   : parent.left
            right  : parent.right
            top    : titleBar.bottom
            bottom : parent.bottom
        }
        border {
            width: Style.main.border
            color: Style.tabbar.background
        }
    }

    // info content
    Column {
        anchors {
            left: parent.left
            top: titleBar.bottom
            leftMargin: Style.main.leftMargin
            topMargin: Style.info.topMargin
        }
        width : root.width - Style.main.leftMargin - Style.main.rightMargin

        TextLabel { text:  qsTr("IMAP SETTINGS", "title of the portion of the configuration screen that contains IMAP settings"); state: "heading"    }
        Rectangle { width: parent.width; height: Style.info.topMargin; color: "#00000000"}
        Grid {
            columns: 2
            rowSpacing: Style.main.fontSize
            TextLabel { text: qsTr("Hostname", "in configuration screen, displays the server hostname (127.0.0.1)") + ":"} TextValue { text: root.accData.hostname }
            TextLabel { text: qsTr("Port", "in configuration screen, displays the server port (ex. 1025)") + ":"} TextValue { text: root.accData.portIMAP }
            TextLabel { text: qsTr("Username", "in configuration screen, displays the username to use with the desktop client") + ":"} TextValue { text: root.address          }
            TextLabel { text: qsTr("Password", "in configuration screen, displays the Bridge password to use with the desktop client") + ":"} TextValue { text: root.accData.password }
            TextLabel { text: qsTr("Security", "in configuration screen, displays the IMAP security settings") + ":"} TextValue { text: "STARTTLS" }
        }
        Rectangle { width: Style.main.dummy; height: Style.main.fontSize; color: "#00000000"}
        Rectangle { width: Style.main.dummy; height: Style.info.topMargin; color: "#00000000"}

        TextLabel { text:  qsTr("SMTP SETTINGS", "title of the portion of the configuration screen that contains SMTP settings"); state: "heading"    }
        Rectangle { width: Style.main.dummy; height: Style.info.topMargin; color: "#00000000"}
        Grid {
            columns: 2
            rowSpacing: Style.main.fontSize
            TextLabel { text: qsTr("Hostname", "in configuration screen, displays the server hostname (127.0.0.1)") + ":"} TextValue { text: root.accData.hostname }
            TextLabel { text: qsTr("Port", "in configuration screen, displays the server port (ex. 1025)") + ":"} TextValue { text: root.accData.portSMTP }
            TextLabel { text: qsTr("Username", "in configuration screen, displays the username to use with the desktop client") + ":"} TextValue { text: root.address          }
            TextLabel { text: qsTr("Password", "in configuration screen, displays the Bridge password to use with the desktop client") + ":"} TextValue { text: root.accData.password }
            TextLabel { text: qsTr("Security", "in configuration screen, displays the SMTP security settings") + ":"} TextValue { text: go.isSMTPSTARTTLS() ? "STARTTLS" : "SSL" }
        }
        Rectangle { width: Style.main.dummy; height: Style.main.fontSize; color: "#00000000"}
        Rectangle { width: Style.main.dummy; height: Style.info.topMargin; color: "#00000000"}
    }

    // apple mail button
    ButtonRounded{
        anchors {
            horizontalCenter: parent.horizontalCenter
            bottom: parent.bottom
            bottomMargin: Style.info.topMargin
        }
        color_main : Style.main.textBlue
        isOpaque: false
        text: qsTr("Configure Apple Mail", "button on configuration screen to automatically configure Apple Mail")
        height: Style.main.fontSize*2
        width: 2*parent.width/3
        onClicked: {
            go.configureAppleMail(root.indexAccount, root.indexAddress)
        }
        visible: go.goos == "darwin"
    }


    function showInfo(iAccount, iAddress) {
        root.indexAccount = iAccount
        root.indexAddress = iAddress
        root.accData = accountsModel.get(iAccount)
        root.address =  accData.aliases.split(";")[iAddress]
        root.show()
        root.raise()
        root.requestActivate()
    }

    function hide() {
        root.visible = false
    }
}
