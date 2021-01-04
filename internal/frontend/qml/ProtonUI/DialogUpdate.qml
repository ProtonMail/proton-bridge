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

// default options to make button accessible

import QtQuick 2.8
import QtQuick.Controls 2.2
import ProtonUI 1.0


Dialog {
    id: root

    title: "Bridge update "+go.newversion

    property alias introductionText : introduction.text
    property bool hasError : false

    signal cancel()
    signal okay()


    isDialogBusy: currentIndex==1

    Rectangle { // 0: Release notes and confirm
        width: parent.width
        height: parent.height
        color: Style.transparent

        Column {
            anchors.centerIn: parent
            spacing: 5*Style.dialog.spacing

            AccessibleText {
                id:introduction
                anchors.horizontalCenter: parent.horizontalCenter
                color: Style.dialog.text
                linkColor: Style.dialog.textBlue
                font {
                    pointSize: 0.8 * Style.dialog.fontSize * Style.pt
                }
                width: 2*root.width/3
                horizontalAlignment: Text.AlignHCenter
                wrapMode: Text.Wrap

                // customize message per application
                text: ' <a href="%1">Release notes</a><br> New version %2<br> <br><br> <a href="%3">%3</a>'

                onLinkActivated : {
                    console.log("clicked link:", link)
                    if (link == "releaseNotes"){
                        root.hide()
                        winMain.dialogVersionInfo.show()
                    } else {
                        root.hide()
                        Qt.openUrlExternally(link)
                    }
                }

                MouseArea {
                    anchors.fill: parent
                    cursorShape: introduction.hoveredLink ? Qt.PointingHandCursor : Qt.ArrowCursor
                    acceptedButtons: Qt.NoButton
                }
            }

            Row {
                anchors.horizontalCenter: parent.horizontalCenter
                spacing: Style.dialog.spacing

                ButtonRounded {
                    fa_icon: Style.fa.times
                    text: (go.goos=="linux" ? qsTr("Okay") : qsTr("Cancel"))
                    color_main: Style.dialog.text
                    onClicked: root.cancel()
                }

                ButtonRounded {
                    fa_icon: Style.fa.check
                    text: qsTr("Update")
                    visible: go.goos!="linux"
                    color_main: Style.dialog.text
                    color_minor: Style.main.textBlue
                    isOpaque: true
                    onClicked: root.okay()
                }
            }
        }
    }

    Rectangle { // 0: Check / download / unpack / prepare
        id: updateStatus
        width: parent.width
        height: parent.height
        color: Style.transparent

        Column {
            anchors.centerIn: parent
            spacing: Style.dialog.spacing

            AccessibleText {
                color: Style.dialog.text
                font {
                    pointSize: Style.dialog.fontSize * Style.pt
                    bold: false
                }
                width: 2*root.width/3
                horizontalAlignment: Text.AlignHCenter
                wrapMode: Text.Wrap
                text: {
                    switch (go.progressDescription) {
                        case "1": return qsTr("Checking the current version.")
                        case "2": return qsTr("Downloading the update files.")
                        case "3": return qsTr("Verifying the update files.")
                        case "4": return qsTr("Unpacking the update files.")
                        case "5": return qsTr("Starting the update.")
                        case "6": return qsTr("Quitting the application.")
                        default: return ""
                    }
                }
            }

            ProgressBar {
                id: progressbar
                implicitWidth  : 2*updateStatus.width/3
                implicitHeight : Style.exporting.rowHeight
                visible: go.progress!=0 // hack hide animation when clearing out progress bar
                value: go.progress
                property int current:  go.total * go.progress
                property bool isFinished:  finishedPartBar.width == progressbar.width
                background: Rectangle {
                    radius         : Style.exporting.boxRadius
                    color          : Style.exporting.progressBackground
                }
                contentItem: Item {
                    Rectangle {
                        id: finishedPartBar
                        width  : parent.width * progressbar.visualPosition
                        height : parent.height
                        radius : Style.exporting.boxRadius
                        gradient  : Gradient {
                            GradientStop { position: 0.00; color: Qt.lighter(Style.main.textBlue,1.1) }
                            GradientStop { position: 0.66; color: Style.main.textBlue }
                            GradientStop { position: 1.00; color: Qt.darker(Style.main.textBlue,1.1) }
                        }

                        Behavior on width {
                            NumberAnimation { duration:300;  easing.type: Easing.InOutQuad }
                        }
                    }
                    Text {
                        anchors.centerIn: parent
                        text: ""
                        color: Style.main.background
                        font {
                            pointSize: Style.dialog.fontSize * Style.pt
                        }
                    }
                }
            }
        }
    }

    Rectangle { // 1: Something went wrong / All ok, closing bridge
        width: parent.width
        height: parent.height
        color: Style.transparent

        Column {
            anchors.centerIn: parent
            spacing: 5*Style.dialog.spacing

            AccessibleText {
                color: Style.dialog.text
                linkColor: Style.dialog.textBlue
                font {
                    pointSize: Style.dialog.fontSize * Style.pt
                }
                width: 2*root.width/3
                horizontalAlignment: Text.AlignHCenter
                wrapMode: Text.Wrap
                text: !root.hasError ?  qsTr('Application will quit now to finish the update.', "message after successful update") :
                qsTr('<b>The update procedure was not successful!</b><br>Please follow the download link and update manually. <br><br><a href="%1">%1</a>').arg(go.downloadLink)

                onLinkActivated : {
                    console.log("clicked link:", link)
                    Qt.openUrlExternally(link)
                }

                MouseArea {
                    anchors.fill: parent
                    cursorShape: parent.hoveredLink ? Qt.PointingHandCursor : Qt.ArrowCursor
                    acceptedButtons: Qt.NoButton
                }
            }

            ButtonRounded{
                anchors.horizontalCenter: parent.horizontalCenter
                visible: root.hasError
                text: qsTr("Close")
                onClicked: root.cancel()
            }
        }
    }

    function clear() {
        root.hasError = false
        go.progress = 0.0
        go.progressDescription = "0"
    }

    function finished(hasError) {
        root.hasError = hasError
        root.incrementCurrentIndex()
    }

    onShow: {
        root.clear()
    }

    onHide: {
        root.clear()
    }

    onOkay: {
        switch (root.currentIndex) {
            case 0:
            go.startUpdate()
        }
        root.incrementCurrentIndex()
    }

    onCancel: {
        root.hide()
    }
}
