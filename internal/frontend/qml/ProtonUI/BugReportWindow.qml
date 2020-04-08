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

// Window for sending a bug report

import QtQuick 2.8
import QtQuick.Window 2.2
import QtQuick.Controls 2.1
import ProtonUI 1.0
import QtGraphicalEffects 1.0

Window {
    id:root
    property alias userAddress   : userAddress
    property alias clientVersion : clientVersion

    width  : Style.bugreport.width
    height : Style.bugreport.height
    minimumWidth  : Style.bugreport.width
    maximumWidth  : Style.bugreport.width
    minimumHeight : Style.bugreport.height
    maximumHeight : Style.bugreport.height

    property color inputBorderColor : Style.main.text

    color   : "transparent"
    flags   : Qt.Window | Qt.Dialog | Qt.FramelessWindowHint
    title   : "ProtonMail Bridge - Bug report"
    visible : false

    WindowTitleBar {
        id: titleBar
        window: root
    }

    Rectangle {
        id:background
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

    Rectangle {
        id:content
        anchors {
            fill         : parent
            leftMargin   : Style.main.leftMargin
            rightMargin  : Style.main.rightMargin
            bottomMargin : Style.main.rightMargin
            topMargin    : Style.main.rightMargin + titleBar.height
        }
        color: "transparent"

        // Description in flickable
        Flickable {
            id: descripWrapper
            anchors {
                left: parent.left
                right: parent.right
                top: parent.top
            }
            height: content.height - (
                (clientVersion.visible ?  clientVersion.height + Style.dialog.fontSize : 0) +
                userAddress.height  + Style.dialog.fontSize +
                securityNote.contentHeight + Style.dialog.fontSize +
                cancelButton.height + Style.dialog.fontSize
            )
            clip: true
            contentWidth  : width
            contentHeight : height

            TextArea.flickable: TextArea {
                id: description
                focus: true
                wrapMode: TextEdit.Wrap
                placeholderText: qsTr ("Please briefly describe the bug(s) you have encountered...", "bug report instructions")
                background : Rectangle {
                    color : Style.dialog.background
                    radius:  Style.dialog.radiusButton
                    border {
                        color : root.inputBorderColor
                        width : Style.dialog.borderInput
                    }

                    layer.enabled: true
                    layer.effect: FastBlur {
                        anchors.fill: parent
                        radius: 8 * Style.px
                    }
                }
                color: Style.main.text
                font.pointSize: Style.dialog.fontSize * Style.pt
                selectionColor: Style.main.textBlue
                selectByKeyboard: true
                selectByMouse:  true
                KeyNavigation.tab: clientVersion
                KeyNavigation.priority: KeyNavigation.BeforeItem
            }

            ScrollBar.vertical : ScrollBar{}
        }

        // Client
        TextLabel {
            anchors {
                left: parent.left
                top: descripWrapper.bottom
                topMargin: Style.dialog.fontSize
            }
            visible: clientVersion.visible
            width: parent.width/2.618
            text: qsTr ("Email client:", "in the bug report form, which third-party email client is being used")
            font.pointSize: Style.dialog.fontSize * Style.pt
        }

        TextField  {
            id: clientVersion
            anchors {
                right: parent.right
                top: descripWrapper.bottom
                topMargin: Style.dialog.fontSize
            }
            placeholderText: qsTr("e.g. Thunderbird", "in the bug report form, placeholder text for email client")
            width: parent.width/1.618

            color            : Style.dialog.text
            selectionColor   : Style.main.textBlue
            selectByMouse    : true
            font.pointSize   : Style.dialog.fontSize * Style.pt
            padding          : Style.dialog.radiusButton

            background: Rectangle {
                color : Style.dialog.background
                radius:  Style.dialog.radiusButton
                border {
                    color : root.inputBorderColor
                    width : Style.dialog.borderInput
                }

                layer.enabled: true
                layer.effect: FastBlur {
                    anchors.fill: parent
                    radius: 8 * Style.px
                }
            }
            onAccepted: userAddress.focus = true
        }

        // Address
        TextLabel {
            anchors {
                left: parent.left
                top: clientVersion.visible ? clientVersion.bottom : descripWrapper.bottom
                topMargin: Style.dialog.fontSize
            }
            color: Style.dialog.text
            width: parent.width/2.618
            text: qsTr ("Contact email:", "in the bug report form, an email to contact the user at")
            font.pointSize: Style.dialog.fontSize * Style.pt
        }

        TextField  {
            id: userAddress
            anchors {
                right: parent.right
                top: clientVersion.visible ? clientVersion.bottom : descripWrapper.bottom
                topMargin: Style.dialog.fontSize
            }
            placeholderText: "benjerry@protonmail.com"
            width: parent.width/1.618

            color            : Style.dialog.text
            selectionColor   : Style.main.textBlue
            selectByMouse    : true
            font.pointSize   : Style.dialog.fontSize * Style.pt
            padding          : Style.dialog.radiusButton

            background: Rectangle {
                color : Style.dialog.background
                radius:  Style.dialog.radiusButton
                border {
                    color : root.inputBorderColor
                    width : Style.dialog.borderInput
                }

                layer.enabled: true
                layer.effect: FastBlur {
                    anchors.fill: parent
                    radius: 8 * Style.px
                }
            }
            onAccepted: root.submit()
        }

        // Note
        AccessibleText {
            id: securityNote
            anchors {
                left: parent.left
                right: parent.right
                top: userAddress.bottom
                topMargin: Style.dialog.fontSize
            }
            wrapMode: Text.Wrap
            color: Style.dialog.text
            font.pointSize : Style.dialog.fontSize * Style.pt
            text:
            "<span style='font-family: " + Style.fontawesome.name + "'>" + Style.fa.exclamation_triangle + "</span> " +
            qsTr("Bug reports are not end-to-end encrypted!", "The first part of warning in bug report form") + " " +
            qsTr("Please do not send any sensitive information.", "The second part of warning in bug report form") + " " +
            qsTr("Contact us at security@protonmail.com for critical security issues.", "The third part of warning in bug report form")
        }

        // buttons
        ButtonRounded {
            id: cancelButton
            anchors {
                left: parent.left
                bottom: parent.bottom
            }
            fa_icon: Style.fa.times
            text: qsTr ("Cancel", "dismisses current action")
            onClicked : root.hide()
        }
        ButtonRounded {
            anchors {
                right: parent.right
                bottom: parent.bottom
            }
            isOpaque: true
            color_main: "white"
            color_minor: Style.main.textBlue
            fa_icon: Style.fa.send
            text: qsTr ("Send", "button sends bug report")
            onClicked : root.submit()
        }
    }

    Rectangle {
        id: notification
        property bool isOK: true
        visible: false
        color: background.color
        anchors.fill: background

        Text {
            anchors.centerIn: parent
            color:  Style.dialog.text
            width: background.width*0.6180
            text: notification.isOK ?
            qsTr ( "Bug report successfully sent." , "notification message about bug sending" ) :
            qsTr ( "Unable to submit bug report."  , "notification message about bug sending" )
            horizontalAlignment: Text.AlignHCenter
            font.pointSize: Style.dialog.titleSize * Style.pt
        }

        Timer {
            id: notificationTimer
            interval: 3000
            repeat: false
            onTriggered : {
                notification.visible=false
                if (notification.isOK) root.hide()
            }
        }
    }

    function submit(){
        if(root.areInputsOK()){
            root.notify(go.sendBug(description.text, clientVersion.text, userAddress.text ))
        }
    }

    function isEmpty(input){
        if (input.text=="") {
            input.focus=true
            input.placeholderText = qsTr("Field required", "a field that must be filled in to submit form")
            return true
        }
        return false
    }

    function areInputsOK() {
        var isOK = true
        if (isEmpty(userAddress))   { isOK=false }
        if (clientVersion.visible && isEmpty(clientVersion)) { isOK=false }
        if (isEmpty(description))   { isOK=false }
        return isOK
    }

    function clear() {
        description.text = ""
        clientVersion.text = ""
        notification.visible = false
    }

    signal prefill()

    function notify(isOK){
        notification.isOK = isOK
        notification.visible = true
        notificationTimer.start()
    }


    function show() {
        prefill()
        root.visible=true
    }

    function hide() {
        clear()
        root.visible=false
    }
}
