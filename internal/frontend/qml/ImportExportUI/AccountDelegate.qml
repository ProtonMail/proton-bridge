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
import ProtonUI 1.0
import ImportExportUI 1.0

// NOTE: Keep the Column so the height and width is inherited from content
Column {
    id: root
    state: status
    anchors.left: parent.left

    property real row_width: 50 * Style.px
    property int row_height: Style.accounts.heightAccount
    property var listalias : aliases.split(";")
    property int iAccount: index

    property real spacingLastButtons: (row_width - exportAccount.anchors.leftMargin -Style.main.rightMargin - exportAccount.width - logoutAccount.width - deleteAccount.width)/2

    Accessible.role: go.goos=="windows" ? Accessible.Grouping : Accessible.Row
    Accessible.name: qsTr("Account %1, status %2", "Accessible text describing account row with arguments: account name and status (connected/disconnected), resp.").arg(account).arg(statusMark.text)
    Accessible.description: Accessible.name
    Accessible.ignored: !enabled || !visible

    // Main row
    Rectangle {
        id: mainaccRow
        anchors.left: parent.left
        width  : row_width
        height : row_height
        state: { return isExpanded ? "expanded" : "collapsed" }
        color: Style.main.background

        property string actionName : (
            isExpanded ?
            qsTr("Collapse row for account %2", "Accessible text of button showing additional configuration of account") : 
            qsTr("Expand row for account %2", "Accessible text of button hiding additional configuration of account")
        ).  arg(account)


        // override by other buttons
        MouseArea {
            id: mouseArea
            anchors.fill: parent
            onClicked : {
                if (root.state=="connected") {
                    mainaccRow.toggle_accountSettings()
                }
            }
            cursorShape : root.state == "connected" ? Qt.PointingHandCursor : Qt.ArrowCursor
            hoverEnabled: true
            onEntered: {
                if (mainaccRow.state=="collapsed") {
                    mainaccRow.color = Qt.lighter(Style.main.background,1.1)
                }
            }
            onExited: {
                if (mainaccRow.state=="collapsed") {
                    mainaccRow.color = Style.main.background
                }
            }
        }

        // toggle down/up icon
        Text {
            id: toggleIcon
            anchors {
                left           : parent.left
                verticalCenter : parent.verticalCenter
                leftMargin     : Style.main.leftMargin
            }
            color: Style.main.text
            font {
                pointSize : Style.accounts.sizeChevron * Style.pt
                family    : Style.fontawesome.name
            }
            text: Style.fa.chevron_down

            MouseArea {
                anchors.fill: parent
                onClicked : {
                    if (root.state=="connected") {
                        mainaccRow.toggle_accountSettings()
                    }
                }
                cursorShape : root.state == "connected" ? Qt.PointingHandCursor : Qt.ArrowCursor
                Accessible.role: Accessible.Button
                Accessible.name: mainaccRow.actionName
                Accessible.description: mainaccRow.actionName
                Accessible.onPressAction : {
                    if (root.state=="connected") {
                        mainaccRow.toggle_accountSettings()
                    }
                }
                Accessible.ignored: root.state!="connected" || !root.enabled
            }
        }

        // account name
        TextMetrics {
            id: accountMetrics
            font : accountName.font
            elide: Qt.ElideMiddle
            elideWidth: (
                statusMark.anchors.leftMargin
                - toggleIcon.anchors.leftMargin
            )
            text: account
        }
        Text {
            id: accountName
            anchors {
                verticalCenter : parent.verticalCenter
                left           : toggleIcon.left
                leftMargin     : Style.main.leftMargin
            }
            color: Style.main.text
            font {
                pointSize : (Style.main.fontSize+2*Style.px) * Style.pt
            }
            text: accountMetrics.elidedText
        }

        // status
        ClickIconText {
            id: statusMark
            anchors {
                verticalCenter : parent.verticalCenter
                left           : parent.left
                leftMargin     : row_width/2
            }
            text      : qsTr("connected", "status of a listed logged-in account")
            iconText  : Style.fa.circle_o
            textColor : Style.main.textGreen
            enabled   : false
            Accessible.ignored: true
        }

        // export
        ClickIconText {
            id: exportAccount
            anchors {
                verticalCenter : parent.verticalCenter
                left           : parent.left
                leftMargin     : 5.5*row_width/8
            }
            text      : qsTr("Export All", "todo")
            iconText  : Style.fa.floppy_o
            textBold  : true
            textColor : Style.main.textBlue
            onClicked: {
                dialogExport.currentIndex = 0
                dialogExport.account = account
                dialogExport.address = account
                dialogExport.show()
            }
        }

        // logout
        ClickIconText {
            id: logoutAccount
            anchors {
                verticalCenter : parent.verticalCenter
                left           : exportAccount.right
                leftMargin     : root.spacingLastButtons
            }
            text      : qsTr("Log out", "action to log out a connected account")
            iconText  : Style.fa.power_off
            textBold  : true
            textColor : Style.main.textBlue
        }

        // remove
        ClickIconText {
            id: deleteAccount
            anchors {
                verticalCenter : parent.verticalCenter
                left           : logoutAccount.right
                leftMargin     : root.spacingLastButtons
            }
            text      : qsTr("Remove", "deletes an account from the account settings page")
            iconText  : Style.fa.trash_o
            textColor : Style.main.text
            onClicked : {
                dialogGlobal.input=iAccount
                dialogGlobal.state="deleteUser"
                dialogGlobal.show()
            }
        }


        // functions
        function toggle_accountSettings() {
            if (root.state=="connected") {
                if (mainaccRow.state=="collapsed" ) {
                    mainaccRow.state="expanded"
                } else {
                    mainaccRow.state="collapsed"
                }
            }
        }

        states: [
            State {
                name: "collapsed"
                PropertyChanges { target : toggleIcon  ; text      : root.state=="connected" ? Style.fa.chevron_down : " " }
                PropertyChanges { target : accountName ; font.bold : false                 }
                PropertyChanges { target : mainaccRow  ; color     : Style.main.background }
                PropertyChanges { target : addressList ; visible   : false                 }
            },
            State {
                name: "expanded"
                PropertyChanges { target : toggleIcon  ; text      : Style.fa.chevron_up               }
                PropertyChanges { target : accountName ; font.bold : true                              }
                PropertyChanges { target : mainaccRow  ; color     : Style.accounts.backgroundExpanded }
                PropertyChanges { target : addressList ; visible   : true                              }
            }
        ]
    }

    // List of adresses
    Column {
        id: addressList
        anchors.left : parent.left
        width: row_width
        visible: false
        property alias model : repeaterAddresses.model

        Repeater {
            id: repeaterAddresses
            model: ["one", "two"]

            Rectangle {
                id: addressRow
                anchors {
                    left  : parent.left
                    right : parent.right
                }
                height: Style.accounts.heightAddrRow
                color: Style.accounts.backgroundExpanded

                // iconText level down
                Text {
                    id: levelDown
                    anchors {
                        left           : parent.left
                        leftMargin     : Style.accounts.leftMarginAddr
                        verticalCenter : wrapAddr.verticalCenter
                    }
                    text        : Style.fa.level_up
                    font.family : Style.fontawesome.name
                    color       : Style.main.textDisabled
                    rotation    : 90
                }

                Rectangle {
                    id: wrapAddr
                    anchors {
                        top         : parent.top
                        left        : levelDown.right
                        right       : parent.right
                        leftMargin  : Style.main.leftMargin
                        rightMargin : Style.main.rightMargin
                    }
                    height: Style.accounts.heightAddr
                    border {
                        width : Style.main.border
                        color : Style.main.line
                    }
                    color: Style.accounts.backgroundAddrRow

                    TextMetrics {
                        id: addressMetrics
                        font: address.font
                        elideWidth: (
                            wrapAddr.width
                            - address.anchors.leftMargin
                            - 2*exportAlias.width
                            - 3*exportAlias.anchors.rightMargin
                        )
                        elide: Qt.ElideMiddle
                        text: modelData
                    }

                    Text {
                        id: address
                        anchors {
                            verticalCenter : parent.verticalCenter
                            left: parent.left
                            leftMargin: Style.main.leftMargin
                        }
                        font.pointSize :  Style.main.fontSize * Style.pt
                        color: Style.main.text
                        text: addressMetrics.elidedText
                    }

                    // export
                    ClickIconText {
                        id: exportAlias
                        anchors {
                            verticalCenter: parent.verticalCenter
                            right: importAlias.left
                            rightMargin: Style.main.rightMargin
                        }
                        text: qsTr("Export", "todo")
                        iconText: Style.fa.floppy_o
                        textBold: true
                        textColor: Style.main.textBlue
                        onClicked: {
                            dialogExport.account = account
                            dialogExport.address = listalias[index]
                            dialogExport.show()
                        }
                    }

                    // import
                    ClickIconText {
                        id: importAlias
                        anchors {
                            verticalCenter: parent.verticalCenter
                            right: parent.right
                            rightMargin: Style.main.rightMargin
                        }
                        text: qsTr("Import", "todo")
                        iconText: Style.fa.upload
                        textBold: true
                        textColor: enabled ? Style.main.textBlue : Style.main.textDisabled
                        onClicked: {
                            dialogImport.account = account
                            dialogImport.address = listalias[index]
                            dialogImport.show()
                        }
                    }
                }
            }
        }
    }

    // line
    Rectangle {
        id: line
        color: Style.accounts.line
        height: Style.accounts.heightLine
        width: root.row_width
    }


    states: [
        State {
            name: "connected"
            PropertyChanges {
                target : addressList
                model  : listalias
            }
            PropertyChanges {
                target : toggleIcon
                color  : Style.main.text
            }
            PropertyChanges {
                target : accountName
                color  : Style.main.text
            }
            PropertyChanges {
                target    : statusMark
                textColor : Style.main.textGreen
                text      : qsTr("connected", "status of a listed logged-in account")
                iconText  : Style.fa.circle
            }
            PropertyChanges {
                target: exportAccount
                visible: true
            }
            PropertyChanges {
                target : logoutAccount
                text   : qsTr("Log out", "action to log out a connected account")
                onClicked : {
                    mainaccRow.state="collapsed"
                    dialogGlobal.state = "logout"
                    dialogGlobal.input = root.iAccount
                    dialogGlobal.show()
                    dialogGlobal.confirmed()
                }
            }
        },
        State {
            name: "disconnected"
            PropertyChanges {
                target : addressList
                model  : 0
            }
            PropertyChanges {
                target : toggleIcon
                color  : Style.main.textDisabled
            }
            PropertyChanges {
                target : accountName
                color  : Style.main.textDisabled
            }
            PropertyChanges {
                target    : statusMark
                textColor : Style.main.textDisabled
                text      : qsTr("disconnected", "status of a listed logged-out account")
                iconText  : Style.fa.circle_o
            }
            PropertyChanges {
                target : logoutAccount
                text   : qsTr("Log in", "action to log in a disconnected account")
                onClicked : {
                    dialogAddUser.username = root.listalias[0]
                    dialogAddUser.show()
                    dialogAddUser.inputPassword.focusInput = true
                }
            }
            PropertyChanges {
                target: exportAccount
                visible: false
            }
        }
    ]
}
