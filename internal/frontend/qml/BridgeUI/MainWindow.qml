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

// This is main window

import QtQuick 2.8
import QtQuick.Window 2.2
import QtQuick.Controls 2.1
import QtQuick.Layouts 1.3
import BridgeUI 1.0
import ProtonUI 1.0


// Main Window
Window {
    id: root
    property alias tabbar                       : tabbar
    property alias viewContent                  : viewContent
    property alias viewAccount                  : viewAccount
    property alias dialogAddUser                : dialogAddUser
    property alias dialogChangePort             : dialogChangePort
    property alias dialogCredits                : dialogCredits
    property alias dialogTlsCert                : dialogTlsCert
    property alias dialogUpdate                 : dialogUpdate
    property alias dialogFirstStart             : dialogFirstStart
    property alias dialogGlobal                 : dialogGlobal
    property alias dialogConnectionTroubleshoot : dialogConnectionTroubleshoot
    property alias bubbleNote                   : bubbleNote
    property alias addAccountTip                : addAccountTip
    property alias updateState                  : infoBar.state
    property alias tlsBarState                  : tlsBar.state
    property int heightContent                  : height-titleBar.height

    // main window appeareance
    width  : Style.main.width
    height : Style.main.height
    flags  : Qt.Window | Qt.FramelessWindowHint
    color: go.goos=="windows" ? "black" : "transparent"
    title: go.programTitle
    minimumWidth: Style.main.width
    minimumHeight: Style.main.height
    maximumWidth: Style.main.width

    property bool isOutdateVersion : root.updateState == "forceUpdate"

    property bool activeContent :
    !dialogAddUser     .visible &&
    !dialogChangePort  .visible &&
    !dialogCredits     .visible &&
    !dialogTlsCert     .visible &&
    !dialogUpdate      .visible &&
    !dialogFirstStart  .visible &&
    !dialogGlobal      .visible &&
    !bubbleNote        .visible

    Accessible.role: Accessible.Grouping
    Accessible.description: qsTr("Window %1").arg(title)
    Accessible.name: Accessible.description


    Component.onCompleted : {
        gui.winMain = root
        console.log("GraphicsInfo of", titleBar,
        "api"                   , titleBar.GraphicsInfo.api                   ,
        "majorVersion"          , titleBar.GraphicsInfo.majorVersion          ,
        "minorVersion"          , titleBar.GraphicsInfo.minorVersion          ,
        "profile"               , titleBar.GraphicsInfo.profile               ,
        "renderableType"        , titleBar.GraphicsInfo.renderableType        ,
        "shaderCompilationType" , titleBar.GraphicsInfo.shaderCompilationType ,
        "shaderSourceType"      , titleBar.GraphicsInfo.shaderSourceType      ,
        "shaderType"            , titleBar.GraphicsInfo.shaderType)

        tabbar.focusButton()
    }

    WindowTitleBar {
        id: titleBar
        window: root
    }

    Rectangle {
        anchors {
            top   : titleBar.bottom
            left  : parent.left
            right : parent.right
            bottom : parent.bottom
        }
        color: Style.title.background
    }

    TLSCertPinIssueBar {
        id: tlsBar
        anchors {
            left  : parent.left
            right : parent.right
            top   : titleBar.bottom
            leftMargin: Style.main.border
            rightMargin: Style.main.border
        }
        enabled : root.activeContent
    }

    InformationBar {
        id: infoBar
        anchors {
            left  : parent.left
            right : parent.right
            top   : tlsBar.bottom
            leftMargin: Style.main.border
            rightMargin: Style.main.border
        }
        enabled : root.activeContent
    }


    TabLabels  {
        id: tabbar
        currentIndex : 0
        enabled: root.activeContent
        anchors {
            top   : infoBar.bottom
            right : parent.right
            left  : parent.left
            leftMargin: Style.main.border
            rightMargin: Style.main.border
        }
        model: [
            { "title" : qsTr("Accounts" , "title of tab that shows account list"             ), "iconText": Style.fa.user_circle_o  },
            { "title" : qsTr("Settings" , "title of tab that allows user to change settings" ), "iconText": Style.fa.cog            },
            { "title" : qsTr("Help"     , "title of tab that shows the help menu"            ), "iconText": Style.fa.life_ring      }
        ]
    }

    // Content of tabs
    StackLayout  {
        id: viewContent
        enabled: root.activeContent
        // dimensions
        anchors {
            left   : parent.left
            right  : parent.right
            top    : tabbar.bottom
            bottom : parent.bottom
            leftMargin: Style.main.border
            rightMargin: Style.main.border
            bottomMargin: Style.main.border
        }
        // attributes
        currentIndex : { return root.tabbar.currentIndex}
        clip         : true
        // content
        AccountView  {
            id: viewAccount
            onAddAccount: dialogAddUser.show()
            model: accountsModel
            delegate: AccountDelegate {
                row_width: viewContent.width
            }
        }

        SettingsView { id: viewSettings; }
        HelpView     { id: viewHelp;     }
    }


    // Floating things

    // Triangle
    Rectangle {
        id: tabtriangle
        visible: false
        property int margin : Style.main.leftMargin+ Style.tabbar.widthButton/2
        anchors {
            top   : tabbar.bottom
            left  : tabbar.left
            leftMargin :  tabtriangle.margin - tabtriangle.width/2 + tabbar.currentIndex * tabbar.spacing
        }
        width: 2*Style.tabbar.heightTriangle
        height: Style.tabbar.heightTriangle
        color: "transparent"
        Canvas {
            anchors.fill: parent
            onPaint: {
                var ctx = getContext("2d")
                ctx.fillStyle = Style.tabbar.background
                ctx.moveTo(0 , 0)
                ctx.lineTo(width/2, height)
                ctx.lineTo(width , 0)
                ctx.closePath()
                ctx.fill()
            }
        }
    }

    // Bubble prevent action
    Rectangle {
        anchors {
            left: parent.left
            right: parent.right
            top: titleBar.bottom
            bottom: parent.bottom
        }
        visible: bubbleNote.visible
        color: "#aa222222"
        MouseArea {
            anchors.fill: parent
            hoverEnabled: true
        }
    }
    BubbleNote {
        id      : bubbleNote
        visible : false
        Component.onCompleted : {
            bubbleNote.place(0)
        }
    }

    BubbleNote {
        id:addAccountTip
        anchors.topMargin: viewAccount.separatorNoAccount - 2*Style.main.fontSize
        text : qsTr("Click here to start", "on first launch, this is displayed above the Add Account button to tell the user what to do first")
        state: (go.isFirstStart && viewAccount.numAccounts==0 && root.viewContent.currentIndex==0) ? "Visible" : "Invisible"
        bubbleColor: Style.main.textBlue

        Component.onCompleted : {
            addAccountTip.place(-1)
        }
        enabled: false

        states: [
            State {
                name: "Visible"
                // hack: opacity 100% makes buttons dialog windows quit wrong color
                PropertyChanges{target: addAccountTip; opacity: 0.999; visible: true}
            },
            State {
                name: "Invisible"
                PropertyChanges{target: addAccountTip; opacity: 0.0; visible: false}
            }
        ]

        transitions: [
            Transition {
                from: "Visible"
                to: "Invisible"

                SequentialAnimation{
                    NumberAnimation {
                        target: addAccountTip
                        property: "opacity"
                        duration: 0
                        easing.type: Easing.InOutQuad
                    }
                    NumberAnimation {
                        target: addAccountTip
                        property: "visible"
                        duration: 0
                    }
                }
            },
            Transition {
                from: "Invisible"
                to: "Visible"
                SequentialAnimation{
                    NumberAnimation {
                        target: addAccountTip
                        property: "visible"
                        duration: 300
                    }
                    NumberAnimation {
                        target: addAccountTip
                        property: "opacity"
                        duration: 500
                        easing.type: Easing.InOutQuad
                    }
                }
            }
        ]
    }

    // Dialogs
    DialogFirstStart {
        id: dialogFirstStart
        visible: go.isFirstStart && gui.isFirstWindow && !dialogGlobal.visible
    }

    // Dialogs
    DialogPortChange {
        id: dialogChangePort
    }

    DialogConnectionTroubleshoot {
        id: dialogConnectionTroubleshoot
    }

    DialogAddUser {
        id: dialogAddUser
        onCreateAccount: Qt.openUrlExternally("https://protonmail.com/signup")
    }

    DialogUpdate {
        id: dialogUpdate
        forceUpdate: root.isOutdateVersion
    }


    Dialog {
        id: dialogCredits
        title: qsTr("Credits", "link to click on to view list of credited libraries")
        Credits { }
    }

    DialogTLSCertInfo {
        id: dialogTlsCert
    }

    DialogYesNo {
        id: dialogGlobal
        question : ""
        answer   : ""
        z: 100
    }


    // resize
    MouseArea {
        property int diff: 0
        anchors {
            bottom : parent.bottom
            left   : parent.left
            right  : parent.right
        }
        cursorShape: Qt.SizeVerCursor
        height: Style.main.fontSize
        onPressed: {
            var globPos = mapToGlobal(mouse.x, mouse.y)
            diff = root.height
            diff -= globPos.y
        }
        onMouseYChanged : {
            var globPos = mapToGlobal(mouse.x, mouse.y)
            root.height =  Math.max(root.minimumHeight, globPos.y + diff)
        }
    }

    function showAndRise(){
        go.loadAccounts()
        root.show()
        root.raise()
        if (!root.active) {
            root.requestActivate()
        }
    }

    // Toggle window
    function toggle() {
        go.loadAccounts()
        if (root.visible) {
            if (!root.active) {
                root.raise()
                root.requestActivate()
            } else {
                root.hide()
            }
        } else {
            root.show()
            root.raise()
        }
    }

    onClosing: {
        close.accepted = false
        // NOTE: In order to make an initial accounts load
        root.hide()
        gui.closeMainWindow()
    }
}
