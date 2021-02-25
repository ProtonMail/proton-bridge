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
        id: messageRow
        anchors.centerIn: root
        visible: root.isVisible
        spacing: Style.main.leftMarginButton

        AccessibleText {
            id: message
            font.pointSize: root.fontSize * Style.pt
        }

        ClickIconText {
            id: linkText
            anchors.verticalCenter : message.verticalCenter
            iconText  : " "
            fontSize : root.fontSize
            textUnderline: true
        }

        ClickIconText {
            id: actionText
            anchors.verticalCenter : message.verticalCenter
            iconText  : " "
            fontSize : root.fontSize
            textUnderline: true
        }
        Text {
            id: separatorText
            anchors.baseline : message.baseline
            color: Style.main.text
            font {
                pointSize : root.fontSize * Style.pt
                bold      : true
            }
        }
        ClickIconText {
            id: action2Text
            anchors.verticalCenter : message.verticalCenter
            iconText  : ""
            fontSize : root.fontSize
            textUnderline: true
        }
    }

    ClickIconText {
        id: closeSign
        anchors.verticalCenter : messageRow.verticalCenter
        anchors.right: root.right
        iconText  : Style.fa.close
        fontSize : root.fontSize
        textUnderline: true
    }

    onStateChanged : {
        switch (root.state) {
            case "internetCheck":
                break;
            case "noInternet" :
                retryInternet.start()
                secLeft=checkInterval[iTry]
                break;
            case "oldVersion":
                break;
            case "forceUpdate":
                break;
            case "upToDate":
                iTry = 0
                secLeft=checkInterval[iTry]
                break;
            case "updateRestart":
                break;
            case "updateError":
                break;
            default :
                break;
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
            PropertyChanges {
                target: linkText
                visible: false
            }
            PropertyChanges {
                target: actionText
                visible: false
            }
            PropertyChanges {
                target: separatorText
                visible: false
            }
            PropertyChanges {
                target: action2Text
                visible: false
            }
            PropertyChanges {
                target: closeSign
                visible: false
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
            PropertyChanges {
                target: linkText
                visible: false
            }
            PropertyChanges {
                target: actionText
                visible: true
                text: qsTr("Retry now", "click to try to connect to the internet when the app is disconnected from the internet")
                onClicked: {
                    go.checkInternet()
                }
            }
            PropertyChanges {
                target: separatorText
                visible: true
                text: "|"
            }
            PropertyChanges {
                target: action2Text
                visible: true
                text: qsTr("Troubleshoot", "Show modal screen with additional tips for troubleshooting connection issues")
                onClicked: {
                    dialogConnectionTroubleshoot.show()
                }
            }
            PropertyChanges {
                target: closeSign
                visible: false
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
                text:  qsTr("Update available", "displayed in a notification when an app update is available")
            }
            PropertyChanges {
                target: linkText
                visible: true
                text: qsTr("Release Notes", "display the release notes from the new version")
                onClicked: gui.openReleaseNotes()
            }
            PropertyChanges {
                target: actionText
                visible: true
                text: qsTr("Update", "click to update to a new version when one is available")
                onClicked: {
                    winMain.dialogUpdate.show()
                }
            }
            PropertyChanges {
                target: separatorText
                visible: false
            }
            PropertyChanges {
                target: action2Text
                visible: false
            }
            PropertyChanges {
                target: closeSign
                visible: true
                onClicked: {
                    go.updateState = "upToDate"
                }
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
            PropertyChanges {
                target: linkText
                visible: false
            }
            PropertyChanges {
                target: actionText
                visible: true
                text: qsTr("Update", "click to update to a new version when one is available")
                onClicked: {
                    winMain.dialogUpdate.show()
                }
            }
            PropertyChanges {
                target: separatorText
                visible: false
            }
            PropertyChanges {
                target: action2Text
                visible: false
            }
            PropertyChanges {
                target: closeSign
                visible: false
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
            PropertyChanges {
                target: linkText
                visible: false
            }
            PropertyChanges {
                target: actionText
                visible: false
            }
            PropertyChanges {
                target: separatorText
                visible: false
            }
            PropertyChanges {
                target: action2Text
                visible: false
            }
            PropertyChanges {
                target: closeSign
                visible: false
            }
        },
        State {
            name: "updateRestart"
            PropertyChanges {
                target: root
                height: 2* Style.main.fontSize
                isVisible: true
                color: Style.main.textBlue
            }
            PropertyChanges {
                target: message
                color: Style.main.background
                text:  qsTr("%1 update is ready", "displayed in a notification when an app update is installed and restart is needed").arg(go.programTitle)
            }
            PropertyChanges {
                target: linkText
                visible: false
            }
            PropertyChanges {
                target: actionText
                visible: true
                text: qsTr("Restart now", "click to restart application as new version was installed")
                onClicked: {
                    go.setToRestart()
                    Qt.quit()
                }
            }
            PropertyChanges {
                target: separatorText
                visible: false
            }
            PropertyChanges {
                target: action2Text
                visible: false
            }
            PropertyChanges {
                target: closeSign
                visible: false
            }
        },
        State {
            name: "updateError"
            PropertyChanges {
                target: root
                height: 2* Style.main.fontSize
                isVisible: true
                color: Style.main.textRed
            }
            PropertyChanges {
                target: message
                color: Style.main.line
                text:  qsTr("Sorry, %1 couldn't update.", "displayed in a notification when app failed to autoupdate").arg(go.programTitle)
            }
            PropertyChanges {
                target: linkText
                visible: false
            }
            PropertyChanges {
                target: actionText
                visible: true
                text: qsTr("Please update manually", "click to open download page to update manally")
                onClicked: {
                    Qt.openUrlExternally(go.updateLandingPage)
                }
            }
            PropertyChanges {
                target: separatorText
                visible: false
            }
            PropertyChanges {
                target: action2Text
                visible: false
            }
            PropertyChanges {
                target: closeSign
                visible: false
            }
        }
    ]
}
