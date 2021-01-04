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

// Important information under title bar

import QtQuick 2.8
import QtQuick.Window 2.2
import QtQuick.Controls 2.1
import ProtonUI 1.0

Rectangle {
    id: root
    property var iTry: 0
    property var secLeft: 0
    property var second: 1000 // convert millisecond to second
    property var checkInterval: [ 5, 10, 30, 60, 120, 300, 600 ] // seconds
    property bool isVisible: true
    property var fontSize : 1.2 * Style.main.fontSize
    color : "black"
    state: "upToDate"

    Timer {
        id: retryInternet
        interval: second
        triggeredOnStart: false
        repeat: true
        onTriggered : {
            secLeft--
            if (secLeft <= 0) {
                retryInternet.stop()
                go.checkInternet()
                if (iTry < checkInterval.length-1) {
                    iTry++
                }
                secLeft=checkInterval[iTry]
                retryInternet.start()
            }
        }
    }

    Row {
        anchors.centerIn: root
        visible: root.isVisible
        spacing: Style.main.leftMarginButton

        AccessibleText {
            id: message
            font.pointSize: root.fontSize * Style.pt
        }

        ClickIconText {
            anchors.verticalCenter : message.verticalCenter
            text      : "("+go.newversion+" " + qsTr("release notes", "display the release notes from the new version")+")"
            visible   : root.state=="oldVersion" && ( go.changelog!="" || go.bugfixes!="")
            iconText  : ""
            onClicked : {
                dialogVersionInfo.show()
            }
            fontSize : root.fontSize
        }

        ClickIconText {
            anchors.verticalCenter : message.verticalCenter
            text : root.state=="oldVersion" || root.state == "forceUpdate" ?
            qsTr("Update", "click to update to a new version when one is available") :
            qsTr("Retry now", "click to try to connect to the internet when the app is disconnected from the internet")
            visible   : root.state!="internetCheck"
            iconText  : ""
            onClicked : {
                if (root.state=="oldVersion" || root.state=="forceUpdate" ) {
                    winMain.dialogUpdate.show()
                } else {
                    go.checkInternet()
                }
            }
            fontSize : root.fontSize
            textUnderline: true
        }
        Text {
            anchors.baseline : message.baseline
            color: Style.main.text
            font {
                pointSize : root.fontSize * Style.pt
                bold      : true
            }
            visible: root.state=="oldVersion" || root.state=="noInternet"
            text : "|"
        }
        ClickIconText {
            anchors.verticalCenter : message.verticalCenter
            iconText  : ""
            text      : root.state == "noInternet" ?
            qsTr("Troubleshoot", "Show modal screen with additional tips for troubleshooting connection issues") :
            qsTr("Remind me later", "Do not install new version and dismiss a notification")
            visible   : root.state=="oldVersion" || root.state=="noInternet"
            onClicked : {
                if (root.state == "oldVersion") {
                    root.state = "upToDate"
                }
                if (root.state == "noInternet") {
                    dialogConnectionTroubleshoot.show()
                }
            }
            fontSize : root.fontSize
            textUnderline: true
        }
    }

    onStateChanged : {
        switch (root.state) {
            case "forceUpdate" :
            gui.warningFlags |= Style.errorInfoBar
            break;
            case "upToDate" :
            gui.warningFlags &= ~Style.warnInfoBar
            iTry = 0
            secLeft=checkInterval[iTry]
            break;
            case "noInternet" :
            gui.warningFlags |= Style.warnInfoBar
            retryInternet.start()
            secLeft=checkInterval[iTry]
            break;
            default :
            gui.warningFlags |= Style.warnInfoBar
        }

        if (root.state!="noInternet") {
            retryInternet.stop()
        }
    }

    function timeToRetry() {
        if (secLeft==1){
            return qsTr("a second", "time to wait till internet connection is retried")
        } else if (secLeft<60){
            return secLeft + " " + qsTr("seconds", "time to wait till internet connection is retried")
        } else {
            var leading = ""+secLeft%60
            if (leading.length < 2) {
                leading = "0" + leading
            }
            return Math.floor(secLeft/60) + ":" + leading
        }
    }

    states: [
        State {
            name: "internetCheck"
            PropertyChanges {
                target: root
                height: 2* Style.main.fontSize
                isVisible: true
                color: Style.main.textOrange
            }
            PropertyChanges {
                target: message
                color: Style.main.background
                text: qsTr("Checking connection. Please wait...", "displayed after user retries internet connection")
            }
        },
        State {
            name: "noInternet"
            PropertyChanges {
                target: root
                height: 2* Style.main.fontSize
                isVisible: true
                color: Style.main.textRed
            }
            PropertyChanges {
                target: message
                color: Style.main.line
                text: qsTr("Cannot contact server. Retrying in ", "displayed when the app is disconnected from the internet or server has problems")+timeToRetry()+"."
            }
        },
        State {
            name: "oldVersion"
            PropertyChanges {
                target: root
                height: 2* Style.main.fontSize
                isVisible: true
                color: Style.main.textBlue
            }
            PropertyChanges {
                target: message
                color: Style.main.background
                text:  qsTr("An update is available.", "displayed in a notification when an app update is available")
            }
        },
        State {
            name: "forceUpdate"
            PropertyChanges {
                target: root
                height: 2* Style.main.fontSize
                isVisible: true
                color: Style.main.textRed
            }
            PropertyChanges {
                target: message
                color: Style.main.line
                text:  qsTr("%1 is outdated.", "displayed in a notification when app is outdated").arg(go.programTitle)
            }
        },
        State {
            name: "upToDate"
            PropertyChanges {
                target: root
                height: 0
                isVisible: false
                color: Style.main.textBlue
            }
            PropertyChanges {
                target: message
                color: Style.main.background
                text:  ""
            }
        }
    ]
}
