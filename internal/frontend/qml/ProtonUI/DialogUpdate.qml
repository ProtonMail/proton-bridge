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

    property bool hasError : false
    property bool forceUpdate : false

    signal cancel()
    signal okay()

    title: forceUpdate ?
        qsTr("Update %1 now", "title of force update dialog").arg(go.programTitle):
        qsTr("Update to %1 %2", "title of normal update dialog").arg(go.programTitle).arg(go.updateVersion)

    isDialogBusy: currentIndex==1 || forceUpdate

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
                    pointSize: Style.dialog.fontSize * Style.pt
                }
                width: 2*root.width/3
                horizontalAlignment: Text.AlignHCenter
                wrapMode: Text.Wrap

                text: {
                    if (forceUpdate) {
                        if (go.updateCanInstall) {
                            return qsTr('You need to update this app to continue using it.<br>
                                Update now or manually download the most recent version here:<br>
                                <a href="%1">%1</a><br>
                                <a href="https://protonmail.com/support/knowledge-base/update-required/">Learn why</a> you need to update',
                                "Message for force-update").arg(go.updateLandingPage)
                        } else {
                            return qsTr('You need to update this app to continue using it.<br>
                                Download the most recent version here:<br>
                                <a href="%1">%1</a><br>
                                <a href="https://protonmail.com/support/knowledge-base/update-required/">Learn why</a> you need to update',
                                "Message for force-update").arg(go.updateLandingPage)
                        }
                    }

                    if (go.updateCanInstall) {
                        return qsTr('Update to the newest version or download it from:<br>
                            <a href="%1">%1</a><br>
                            <a href="%2">View release notes</a>',
                            "Message for manual update").arg(go.updateLandingPage).arg(go.updateReleaseNotesLink)
                    } else {
                         return qsTr('Update to the newest version from:<br>
                            <a href="%1">%1</a><br>
                            <a href="%2">View release notes</a>',
                            "Message for manual update").arg(go.updateLandingPage).arg(go.updateReleaseNotesLink)
                    }
                }
                

                onLinkActivated : {
                    console.log("clicked link:", link)
                    Qt.openUrlExternally(link)
                }

                MouseArea {
                    anchors.fill: parent
                    cursorShape: introduction.hoveredLink ? Qt.PointingHandCursor : Qt.ArrowCursor
                    acceptedButtons: Qt.NoButton
                }
            }

            CheckBoxLabel {
                id: autoUpdate
                anchors.horizontalCenter: parent.horizontalCenter
                text: qsTr("Automatically update in the future", "Checkbox label for using autoupdates later on")
                checked: go.isAutoUpdate
                onToggled: go.toggleAutoUpdate()
                visible: !root.forceUpdate && (go.isAutoUpdate != undefined)
            }

            Row {
                anchors.horizontalCenter: parent.horizontalCenter
                spacing: Style.dialog.spacing

                ButtonRounded {
                    fa_icon: Style.fa.times
                    text: root.forceUpdate ? qsTr("Quit") : qsTr("Cancel")
                    color_main: Style.dialog.text
                    onClicked: root.forceUpdate ? Qt.quit() : root.cancel()
                }

                ButtonRounded {
                    fa_icon: Style.fa.check
                    text: qsTr("Update")
                    visible: go.updateCanInstall
                    color_main: Style.dialog.text
                    color_minor: Style.main.textBlue
                    isOpaque: true
                    onClicked: root.okay()
                }
            }
        }
    }

    Rectangle { // 1: Installing update
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
                text:  qsTr("Updating...") 
            }

            ProgressBar {
                id: updateProgressBar
                width: 2*updateStatus.width/3
                height: Style.exporting.rowHeight
                //implicitWidth  : 2*updateStatus.width/3
                //implicitHeight : Style.exporting.rowHeight
                indeterminate: true
                //value: 0.5
                //property int current:  go.total * go.progress
                //property bool isFinished:  finishedPartBar.width == progressbar.width
                background: Rectangle {
                    radius         : Style.exporting.boxRadius
                    color          : Style.exporting.progressBackground
                }

                contentItem: Item {
                    clip: true
                    Rectangle {
                        id: progressIndicator
                        width  : updateProgressBar.indeterminate ? 50 : parent.width * updateProgressBar.visualPosition
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

                        SequentialAnimation {
                            running: updateProgressBar.visible && updateProgressBar.indeterminate
                            loops: Animation.Infinite

                            SmoothedAnimation { 
                                target: progressIndicator
                                property: "x"
                                from: 0
                                to: updateProgressBar.width - progressIndicator.width
                                duration: 2000 
                            }

                            SmoothedAnimation { 
                                target: progressIndicator
                                property: "x"
                                from: updateProgressBar.width - progressIndicator.width
                                to: 0
                                duration: 2000 
                            }
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

    Rectangle { // 2: Something went wrong / All ok, closing bridge
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
                text: !root.hasError ?  qsTr('%1 will restart now to finish the update.', "message after successful update").arg(go.programTitle) :
                qsTr('<b>The update procedure was not successful!</b><br>Please follow the download link and update manually. <br><br><a href="%1">%1</a>').arg(go.updateLandingPage)

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
        root.currentIndex = 2
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
                go.startManualUpdate()
                root.currentIndex = 1
                break
        }
    }

    onCancel: {
        root.hide()
    }
}
