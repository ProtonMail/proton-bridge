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
import ImportExportUI 1.0
import ProtonUI 1.0


// Main Window
Window {
    id                               : root
    property alias tabbar            : tabbar
    property alias viewContent       : viewContent
    property alias viewAccount       : viewAccount
    property alias dialogAddUser     : dialogAddUser
    property alias dialogGlobal      : dialogGlobal
    property alias dialogCredits     : dialogCredits
    property alias dialogUpdate      : dialogUpdate
    property alias popupMessage      : popupMessage
    property alias popupFolderEdit   : popupFolderEdit
    property alias updateState       : infoBar.state
    property alias dialogExport      : dialogExport
    property alias dialogImport      : dialogImport
    property alias addAccountTip     : addAccountTip
    property int   heightContent     : height-titleBar.height

    property real   innerWindowBorder   : go.goos=="darwin" ? 0 : Style.main.border

    // main window appearance
    width  : Style.main.width
    height : Style.main.height
    flags  : go.goos=="darwin" ? Qt.Window : Qt.Window | Qt.FramelessWindowHint
    color: go.goos=="windows" ? Style.main.background : Style.transparent
    title: go.programTitle

    minimumWidth  : Style.main.width
    minimumHeight : Style.main.height

    property bool isOutdateVersion : root.updateState == "forceUpdate"

    property bool activeContent :
    !dialogAddUser     .visible &&
    !dialogCredits     .visible &&
    !dialogGlobal      .visible &&
    !dialogUpdate      .visible &&
    !dialogImport      .visible &&
    !dialogExport      .visible &&
    !popupFolderEdit   .visible &&
    !popupMessage      .visible

    Accessible.role: Accessible.Grouping
    Accessible.description: qsTr("Window %1").arg(title)
    Accessible.name: Accessible.description

    WindowTitleBar {
        id: titleBar
        window: root
        visible: go.goos!="darwin"
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

    InformationBar {
        id: infoBar
        anchors {
            left  : parent.left
            right : parent.right
            top   : titleBar.bottom
            leftMargin: innerWindowBorder
            rightMargin: innerWindowBorder
        }
    }

    TabLabels  {
        id: tabbar
        currentIndex : 0
        enabled: root.activeContent
        anchors {
            top   : infoBar.bottom
            right : parent.right
            left  : parent.left
            leftMargin: innerWindowBorder
            rightMargin: innerWindowBorder
        }
        model: [
            { "title" : qsTr("Import-Export" , "title of tab that shows account list"             ), "iconText": Style.fa.home      },
            { "title" : qsTr("Settings"      , "title of tab that allows user to change settings" ), "iconText": Style.fa.cogs      },
            { "title" : qsTr("Help"          , "title of tab that shows the help menu"            ), "iconText": Style.fa.life_ring }
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
            leftMargin: innerWindowBorder
            rightMargin: innerWindowBorder
            bottomMargin: innerWindowBorder
        }
        // attributes
        currentIndex : { return root.tabbar.currentIndex}
        clip         : true
        // content
        AccountView  {
            id           : viewAccount
            onAddAccount : dialogAddUser.show()
            model        : accountsModel
            hasFooter    : false
            delegate     : AccountDelegate {
                row_width : viewContent.width
            }
        }
        SettingsView { id: viewSettings; }
        HelpView     { id: viewHelp;     }
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

    DialogAddUser {
        id: dialogAddUser

        anchors {
            top   : infoBar.bottom
            bottomMargin: innerWindowBorder
            leftMargin: innerWindowBorder
            rightMargin: innerWindowBorder
        }

        onCreateAccount: Qt.openUrlExternally("https://protonmail.com/signup")
    }

    DialogUpdate {
        id: dialogUpdate
        forceUpdate: root.isOutdateVersion
    }


    DialogExport {
        id: dialogExport
        anchors {
            top   : infoBar.bottom
            bottomMargin: innerWindowBorder
            leftMargin: innerWindowBorder
            rightMargin: innerWindowBorder
        }

    }

    DialogImport {
        id: dialogImport
        anchors {
            top   : infoBar.bottom
            bottomMargin: innerWindowBorder
            leftMargin: innerWindowBorder
            rightMargin: innerWindowBorder
        }

    }

    Dialog {
        id: dialogCredits
        anchors {
            top   : infoBar.bottom
            bottomMargin: innerWindowBorder
            leftMargin: innerWindowBorder
            rightMargin: innerWindowBorder
        }

        title: qsTr("Credits", "title for list of credited libraries")

        Credits { }
    }

    DialogYesNo {
        id: dialogGlobal
        question : ""
        answer   : ""
        z: 100
    }

    PopupEditFolder {
        id: popupFolderEdit
        anchors {
            left: parent.left
            right: parent.right
            top: infoBar.bottom
            bottom: parent.bottom
        }
    }

    // Popup
    PopupMessage {
        id: popupMessage
        anchors {
            left   : parent.left
            right  : parent.right
            top    : infoBar.bottom
            bottom : parent.bottom
        }

        onClickedNo: popupMessage.hide()
        onClickedOkay: popupMessage.hide()
        onClickedCancel: popupMessage.hide()
        onClickedYes: {
            if (popupMessage.text == gui.areYouSureYouWantToQuit) Qt.quit()
        }
    }

    // resize
    MouseArea { // bottom
        id: resizeBottom
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
            resizeBottom.diff = root.height
            resizeBottom.diff -= globPos.y
        }
        onMouseYChanged : {
            var globPos = mapToGlobal(mouse.x, mouse.y)
            root.height =  Math.max(root.minimumHeight, globPos.y + resizeBottom.diff)
        }
    }

    MouseArea { // right
        id: resizeRight
        property int diff: 0
        anchors {
            top    : titleBar.bottom
            bottom : parent.bottom
            right  : parent.right
        }
        cursorShape: Qt.SizeHorCursor
        width: Style.main.fontSize/2
        onPressed: {
            var globPos = mapToGlobal(mouse.x, mouse.y)
            resizeRight.diff = root.width
            resizeRight.diff -= globPos.x
        }
        onMouseXChanged : {
            var globPos = mapToGlobal(mouse.x, mouse.y)
            root.width =  Math.max(root.minimumWidth, globPos.x + resizeRight.diff)
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

    onClosing : {
        close.accepted=false
        if (
            (dialogImport.visible && dialogImport.currentIndex == 4 && go.progress!=1) ||
            (dialogExport.visible && dialogExport.currentIndex == 2 && go.progress!=1)
        ) {
            popupMessage.buttonOkay   .visible = false
            popupMessage.buttonYes    .visible = false
            popupMessage.buttonQuit   .visible = true
            popupMessage.buttonCancel .visible = true
            popupMessage.show ( gui.areYouSureYouWantToQuit )
            return
        }

        close.accepted=true
        go.processFinished()
    }
}
