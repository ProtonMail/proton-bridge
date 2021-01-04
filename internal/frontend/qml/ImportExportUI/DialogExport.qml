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

// Export dialog
import QtQuick 2.8
import QtQuick.Controls 2.2
import ProtonUI 1.0
import ImportExportUI 1.0

// TODO
//  - make ErrorDialog module
//  - map decision to error code : ask (default), skip ()
//  - what happens when import fails ? heuristic to find mail where to start from

Dialog {
    id: root

    enum Page {
        LoadingStructure = 0, Options, Progress
    }

    title : set_title()

    property string address
    property alias finish: finish

    property string msgClearUnfished: qsTr ("Remove already exported files.")

    isDialogBusy : true // currentIndex == 0 ||  currentIndex == 3 

    signal cancel()
    signal okay()


    Rectangle { // 0
        id: dialogLoading
        width: root.width
        height: root.height
        color: Style.transparent

        Text {
            anchors.centerIn : dialogLoading
            font.pointSize: Style.dialog.titleSize * Style.pt
            color: Style.dialog.text
            horizontalAlignment: Text.AlignHCenter
            text: qsTr("Loading folders and labels for", "todo") +"\n" + address
        }
    }

    Rectangle { // 1
        id: dialogInput
        width: root.width
        height: root.height
        color: Style.transparent

        Row {
            id: inputRow
            anchors {
                topMargin        : root.titleHeight
                top              : parent.top
                horizontalCenter : parent.horizontalCenter
            }
            spacing: 3*Style.main.leftMargin
            property real columnWidth  : (root.width - Style.main.leftMargin - inputRow.spacing - Style.main.rightMargin) / 2
            property real columnHeight :  root.height  - root.titleHeight - Style.main.leftMargin


            ExportStructure {
                id: sourceFoldersInput
                width  : inputRow.columnWidth
                height : inputRow.columnHeight
                title  : qsTr("From: %1", "todo").arg(address)
            }

            Column {
                spacing: (inputRow.columnHeight - dateRangeInput.height - outputFormatInput.height - outputPathInput.height - buttonRow.height -  infotipEncrypted.height) / 4

                DateRange{
                    id: dateRangeInput
                }

                OutputFormat {
                    id: outputFormatInput
                }

                Row {
                    spacing: Style.dialog.spacing
                    CheckBoxLabel {
                        id: exportEncrypted
                        text: qsTr("Export emails that cannot be decrypted as ciphertext")
                        anchors {
                            bottom: parent.bottom
                            bottomMargin: Style.dialog.fontSize/1.8
                        }
                    }

                    InfoToolTip {
                        id: infotipEncrypted
                        anchors {
                            verticalCenter: exportEncrypted.verticalCenter
                        }
                        info: qsTr("Checking this option will export all emails that cannot be decrypted in ciphertext. If this option is not checked, these emails will not be exported", "todo")
                    }
                }

                FileAndFolderSelect {
                    id: outputPathInput
                    title: qsTr("Select location of export:", "todo")
                    width  : inputRow.columnWidth // stretch folder input
                }

                Row {
                    id: buttonRow
                    anchors.right : parent.right
                    spacing       : Style.dialog.rightMargin

                    ButtonRounded {
                        id:buttonCancel
                        fa_icon: Style.fa.times
                        text: qsTr("Cancel")
                        color_main: Style.main.textBlue
                        onClicked : root.cancel()
                    }

                    ButtonRounded {
                        id: buttonNext
                        fa_icon: Style.fa.check
                        text: qsTr("Export","todo")
                        enabled: transferRules != 0
                        color_main: Style.dialog.background
                        color_minor: enabled ? Style.dialog.textBlue : Style.main.textDisabled
                        isOpaque: true
                        onClicked : root.okay()
                    }
                }
            }
        }
    }

    Rectangle { // 2
        id: progressStatus
        width: root.width
        height: root.height
        color: "transparent"

        Row {
            anchors {
                bottom: progressbarExport.top
                bottomMargin: Style.dialog.heightSeparator
                left: progressbarExport.left
            }
            spacing: Style.main.rightMargin
            AccessibleText {
                id: statusLabel
                text : qsTr("Status:")
                font.pointSize: Style.main.iconSize * Style.pt
                color : Style.main.text
            }
            AccessibleText {
                anchors.baseline: statusLabel.baseline
                text :  {
                    if (progressbarExport.isFinished) return qsTr("finished")
                    if (go.progressDescription == "") return qsTr("exporting")
                    return go.progressDescription
                }
                elide: Text.ElideMiddle
                width: progressbarExport.width - parent.spacing - statusLabel.width
                font.pointSize: Style.dialog.textSize * Style.pt
                color : Style.main.textDisabled
            }
        }

        ProgressBar {
            id: progressbarExport
            implicitWidth  : 2*progressStatus.width/3
            implicitHeight : Style.exporting.rowHeight
            value: go.progress
            property int current:  go.total * go.progress
            property bool isFinished:  finishedPartBar.width == progressbarExport.width
            anchors {
                centerIn: parent
            }
            background: Rectangle {
                radius         : Style.exporting.boxRadius
                color          : Style.exporting.progressBackground
            }
            contentItem: Item {
                Rectangle {
                    id: finishedPartBar
                    width  : parent.width * progressbarExport.visualPosition
                    height : parent.height
                    radius : Style.exporting.boxRadius
                    gradient  : Gradient {
                        GradientStop { position: 0.00; color: Qt.lighter(Style.exporting.progressStatus,1.1) }
                        GradientStop { position: 0.66; color: Style.exporting.progressStatus }
                        GradientStop { position: 1.00; color: Qt.darker(Style.exporting.progressStatus,1.1) }
                    }

                    Behavior on width {
                        NumberAnimation { duration:800;  easing.type: Easing.InOutQuad }
                    }
                }
                Text {
                    anchors.centerIn: parent
                    text: {
                        if (progressbarExport.isFinished) {
                            if (go.progressDescription=="") return qsTr("Export finished","todo")
                            else return qsTr("Export failed: %1").arg(go.progressDescription)
                        }
                        if (
                            go.progressDescription == gui.enums.progressInit || 
                            (go.progress==0 && go.description=="")
                        ) {
                            if (go.total>1) return qsTr("Estimating the total number of messages (%1)","todo").arg(go.total)
                            else return qsTr("Estimating the total number of messages","todo")
                        } 
                        var msg = qsTr("Exporting message %1 of %2 (%3%)","todo")
                        if (pauseButton.paused) msg = qsTr("Exporting paused at message %1 of %2 (%3%)","todo")
                        return msg.arg(progressbarExport.current).arg(go.total).arg(Math.floor(go.progress*100))
                    }
                    color: Style.main.background
                    font {
                        pointSize: Style.dialog.fontSize * Style.pt
                    }
                }
            }
        }

        Row {
            anchors {
                top: progressbarExport.bottom
                topMargin : Style.dialog.heightSeparator
                horizontalCenter: parent.horizontalCenter
            }
            spacing: Style.dialog.rightMargin

            ButtonRounded {
                id: pauseButton
                property bool paused : false
                fa_icon    : paused ? Style.fa.play : Style.fa.pause
                text       : paused ? qsTr("Resume") : qsTr("Pause")
                color_main : Style.dialog.textBlue
                onClicked  : {
                    if (paused) {
                        if (winMain.updateState == gui.enums.statusNoInternet) {
                            go.notifyError(gui.enums.errNoInternet)
                            return
                        }
                        go.resumeProcess()
                    } else {
                        go.pauseProcess()
                    }
                    paused = !paused
                    pauseButton.focus=false
                }
                visible    : !progressbarExport.isFinished
            }

            ButtonRounded {
                fa_icon    : Style.fa.times
                text       : qsTr("Cancel")
                color_main : Style.dialog.textBlue
                visible    : !progressbarExport.isFinished
                onClicked  : root.ask_cancel_progress()
            }

            ButtonRounded {
                id: finish
                fa_icon     : Style.fa.check
                text        : qsTr("Okay","todo")
                color_main  : Style.dialog.background
                color_minor : Style.dialog.textBlue
                isOpaque    : true
                visible     : progressbarExport.isFinished
                onClicked   : root.okay()
            }
        }

        ClickIconText {
            id: buttonHelp
            anchors {
                right        : parent.right
                bottom       : parent.bottom
                rightMargin  : Style.main.rightMargin
                bottomMargin : Style.main.rightMargin
            }
            textColor  : Style.main.textDisabled
            iconText   : Style.fa.question_circle
            text       : qsTr("Help", "directs the user to the online user guide")
            textBold   : true
            onClicked  : Qt.openUrlExternally("https://protonmail.com/support/categories/import-export/")
        }
    }

    PopupMessage {
        id: errorPopup
        width: root.width
        height: root.height
    }

    function check_inputs() {
        if (currentIndex == 1) {
            // at least one email to export
            if (transferRules.rowCount() == 0){
                errorPopup.show(qsTr("No emails found to export. Please try another address.", "todo"))
                return false
            }
            // at least one source selected
            /*
             if (!transferRules.atLeastOneSelected) {
                 errorPopup.show(qsTr("Please select at least one item to export.", "todo"))
                 return false
             }
             */
            // check path
            var folderCheck = go.checkPathStatus(outputPathInput.path)
            switch (folderCheck) {
                case gui.enums.pathEmptyPath:
                errorPopup.show(qsTr("Missing export path. Please select an output folder."))
                break;
                case gui.enums.pathWrongPath:
                errorPopup.show(qsTr("Folder '%1' not found. Please select an output folder.").arg(outputPathInput.path))
                break;
                case gui.enums.pathOK | gui.enums.pathNotADir:
                errorPopup.show(qsTr("File '%1' is not a folder. Please select an output folder.").arg(outputPathInput.path))
                break;
                case gui.enums.pathWrongPermissions:
                errorPopup.show(qsTr("Cannot access folder '%1'. Please check folder permissions.").arg(outputPathInput.path))
                break;
            }
            if (
                (folderCheck&gui.enums.pathOK)==0 ||
                (folderCheck&gui.enums.pathNotADir)==gui.enums.pathNotADir
            ) return false
            if (winMain.updateState == gui.enums.statusNoInternet) {
                errorPopup.show(qsTr("Please check your internet connection."))
                return false
            }
        }
        return true
    }

    function set_title() {
        switch(root.currentIndex){
            case 1 : return qsTr("Select what you'd like to export:")
            default: return ""
        }
    }

    function clear_status() {
        go.progress=0.0
        go.total=0.0
        go.progressDescription=gui.enums.progressInit
    }

    function ask_cancel_progress(){
        errorPopup.buttonYes.visible = true
        errorPopup.buttonNo.visible = true
        errorPopup.buttonOkay.visible = false
        errorPopup.show ("Are you sure you want to cancel this export?")
    }


    onCancel : {
        switch (root.currentIndex) {
            case 0 :
            case 1 : root.hide(); break;
            case 2 : // progress bar 
            go.cancelProcess();
            // no break
            default:
            root.clear_status()
            root.currentIndex=1
        }
    }

    onOkay : {
        var isOK = check_inputs()
        if (!isOK) return
        timer.interval= currentIndex==1 ? 1 : 300
        switch (root.currentIndex) {
            case 2: // progress
            root.clear_status()
            root.hide()
            break
            case 0: // loading structure
            dateRangeInput.getRange()
            //no break
            default:
            incrementCurrentIndex()
            timer.start()
        }
    }

    onShow: {
        if (winMain.updateState==gui.enums.statusNoInternet) {
            go.checkInternet()
            if (winMain.updateState==gui.enums.statusNoInternet) {
                go.notifyError(gui.enums.errNoInternet)
                root.hide()
                return
            }
        }

        root.clear_status()
        root.currentIndex=0
        timer.interval = 300
        timer.start()
        dateRangeInput.allDates = true
    }

    Connections {
        target: timer
        onTriggered : {
            switch (currentIndex) {
                case 0:
                go.loadStructureForExport(root.address)
                sourceFoldersInput.hasItems = (transferRules.rowCount() > 0)
                break
                case 2:
                dateRangeInput.applyRange()
                go.startExport(
                    outputPathInput.path,
                    root.address,
                    outputFormatInput.checkedText,
                    exportEncrypted.checked
                )
                break
            }
        }
    }

    Connections {
        target: errorPopup

        onClickedOkay  : errorPopup.hide()
        onClickedYes : {
            root.cancel()
            errorPopup.hide()
        }
        onClickedNo  : {
            errorPopup.hide()
        }
    }
}
