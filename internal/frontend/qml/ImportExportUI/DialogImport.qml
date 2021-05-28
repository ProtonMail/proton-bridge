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
import QtQuick.Layouts 1.3
import QtQuick.Dialogs 1.0
import ProtonUI 1.0
import ImportExportUI 1.0

Dialog {
    id: root

    enum Page {
        SelectSourceType=0, ImapSource, LoadingStructure, SourceToTarget, Progress, Report
    }

    title: "" // qsTr("Importing from: %1", "todo").arg(address)

    isDialogBusy: currentIndex==3 || currentIndex==4

    property string account
    property string address
    property string inputPath  : ""
    property bool   isFromFile : inputEmail.text == "" && root.inputPath != ""
    property bool   isFromIMAP : inputEmail.text != ""
    property bool   paused : false

    property string msgDontShowAgain : qsTr("Do not show this message again")


    signal cancel()
    signal okay()


    Rectangle { // SelectSourceType
        id: sourceType
        width: parent.width
        height: parent.height
        color: "transparent"

        Text {
            anchors {
                horizontalCenter : parent.horizontalCenter
                top              : parent.top
                topMargin        : Style.dialog.titleSize
            }

            font.pointSize: Style.dialog.titleSize * Style.pt
            color: Style.dialog.text
            text: qsTr("Please select the source of the emails that you would like to import:")
        }

        Row {
            anchors {
                centerIn: parent
            }

            ImportSourceButton {
                id: imapButton
                width: winMain.width/2
                iconText: "envelope_open"
                text: qsTr("Import from account")
                onClicked:  root.incrementCurrentIndex()
            }

            ImportSourceButton {
                id: fileButton
                width: winMain.width/2
                iconText: "folder_open"
                text: qsTr("Import local files")
                onClicked: { pathDialog.visible = true }
                anchors.bottom: imapButton.bottom
            }
        }

        FileDialog {
            id: pathDialog
            title: "Select local folder to import"
            folder: shortcuts.home
            onAccepted: {
                sanitizePath(pathDialog.fileUrl.toString())
                root.okay()
            }
            selectFolder: true
        }
    }

    Rectangle { // ImapSource
        id: imapSource

        Text {
            id: imapSourceTitle
            anchors {
                top : parent.top
                topMargin: imapSourceTitle.height / 2
                horizontalCenter : parent.horizontalCenter
            }

            font.pointSize: Style.dialog.titleSize * Style.pt
            color: Style.dialog.text
            text: qsTr("Sign in to your Account")
        }

        Rectangle { // line
            id: titleLine
            anchors {
                top: imapSourceTitle.bottom
                topMargin: imapSourceTitle.height / 2
                horizontalCenter : parent.horizontalCenter
            }
            width: imapSourceContent.width
            height: Style.main.heightLine
            color: Style.main.line
        }

        Rectangle {
            id: imapSourceContent
            anchors {
                top: titleLine.bottom
                topMargin: imapSourceTitle.height / 2
                bottom: buttonRow.top
            }
            width: winMain.width
            color: Style.dialog.background

            Text {
                id: note
                anchors {
                    bottom: wrapper.top
                    bottomMargin: imapSourceTitle.height / 2
                    horizontalCenter : parent.horizontalCenter
                }
                text: qsTr(
                    "Many email providers (Gmail, Yahoo, etc.) will require you to allow remote sign-on in order to perform import through IMAP. See <a href=\"%1\">this article</a> for details about how to do this with your email account.",
                    "Note added at IMAP credential page."
                ).arg("https://protonmail.com/support/knowledge-base/allowing-imap-access-and-entering-imap-details/")
                font {
                    pointSize: Style.dialog.fontSize * Style.pt
                }
                color: Style.dialog.text
                linkColor: Style.dialog.textBlue
                wrapMode: Text.WordWrap
                textFormat: Text.StyledText
                horizontalAlignment: Text.AlignHCenter

                width: parent.width * 0.618
                onLinkActivated: Qt.openUrlExternally(link)
            }

            Rectangle {
                id: wrapper
                anchors.centerIn: parent
                width: firstRow.width
                height: firstRow.height + secondRow.height + secondRow.anchors.topMargin
                color: Style.transparent
                Row {
                    id: firstRow
                    spacing: imapSourceTitle.height

                    InputField {
                        id: inputEmail
                        iconText: Style.fa.user_circle
                        label: qsTr("Email", "todo") + ":"
                        onEditingFinished: {
                            root.guessEmailProvider()
                        }
                        onAccepted: if (root.check_inputs()) root.okay()
                        anchors.horizontalCenter: undefined
                    }

                    InputField {
                        id: inputPassword
                        label    : qsTr("Password:")
                        iconText     : Style.fa.lock
                        isPassword: true
                        onAccepted: if (root.check_inputs()) root.okay()
                        anchors.horizontalCenter: undefined
                    }
                }

                Row {
                    id: secondRow
                    spacing: imapSourceTitle.height
                    anchors {
                        top: firstRow.bottom
                        topMargin: 2*imapSourceTitle.height
                    }

                    InputField {
                        id: inputServer
                        iconText: Style.fa.server
                        label: qsTr("Server address", "todo") + ":"
                        onAccepted: if (root.check_inputs()) root.okay()
                        anchors.horizontalCenter: undefined
                    }

                    InputField {
                        id: inputPort
                        iconText: Style.fa.hashtag
                        label: qsTr("Port:")
                        onAccepted: if (root.check_inputs()) root.okay()
                        anchors.horizontalCenter: undefined
                    }
                }
            }
        }

        Row { // Buttons
            id:buttonRow
            anchors {
                right        : parent.right
                bottom       : parent.bottom
                rightMargin  : Style.dialog.rightMargin
                bottomMargin : Style.dialog.bottomMargin
            }
            spacing: Style.main.leftMargin

            ButtonRounded {
                fa_icon    : Style.fa.times
                text       : qsTr("Cancel", "todo")
                color_main : Style.dialog.textBlue
                onClicked  : root.cancel()
            }

            ButtonRounded {
                fa_icon     : Style.fa.check
                text        : qsTr("Next", "todo")
                color_main  : Style.dialog.background
                color_minor : Style.dialog.textBlue
                isOpaque    : true
                onClicked   : root.okay()
            }
        }
    }

    Rectangle { // LoadingStructure
        id: loadingStructures
        color  : Style.dialog.background
        width  : parent.width
        height : parent.height

        Text {
            anchors {
                verticalCenter   : parent.verticalCenter
                horizontalCenter : parent.horizontalCenter
                topMargin        : Style.dialog.titleSize
            }
            font.pointSize: Style.dialog.titleSize * Style.pt
            color: Style.dialog.text
            text: root.isFromFile ? qsTr("Loading folder structures, please wait...") : qsTr("Loading structure of IMAP account, please wait...")
        }
    }

    Rectangle { // SourceToTarget
        id: dialogStructure
        width  : parent.width
        height : parent.height

        // Import instructions
        ImportStructure {
            id: importInstructions
            anchors.bottom : masterImportSettings.top
            titleFrom      : root.isFromFile  ? root.inputPath : inputEmail.text
            titleTo        : root.address
        }

        Column {
            id: masterImportSettings
            anchors {
                right  : parent.right
                left   : parent.left
                bottom : parent.bottom

                leftMargin   : Style.main.leftMargin
                rightMargin  : Style.main.leftMargin
                bottomMargin : Style.main.bottomMargin
            }

            spacing: Style.main.bottomMargin

            Row {
                spacing: masterImportSettings.width - labelMasterImportSettings.width - resetSourceButton.width

                Text {
                    id: labelMasterImportSettings
                    text: qsTr("Master import settings:")

                    font {
                        bold: true
                        pointSize: Style.main.fontSize * Style.pt
                    }
                    color: Style.main.text

                    InfoToolTip {
                        anchors {
                            left: parent.right
                            bottom: parent.bottom
                            leftMargin : Style.dialog.leftMargin
                        }
                        info: qsTr(
                            "If master import date range is selected only emails within this range will be imported, unless it is specified differently in folder date range.",
                            "Text in master import settings tooltip."
                        )
                    }
                }

                // Reset all to default
                ClickIconText {
                    id: resetSourceButton
                    text:qsTr("Reset all settings to default")
                    iconText: Style.fa.refresh
                    textColor: Style.main.textBlue
                    onClicked: {
                        go.resetSource()
                        root.decrementCurrentIndex()
                        timer.start()
                    }
                }
            }

            Rectangle{
                id: line
                anchors {
                    left  : parent.left
                    right : parent.right
                    top   : labelMasterImportSettings.bottom

                    topMargin : Style.dialog.spacing
                }
                height : Style.main.border * 2
                color  : Style.main.line
            }

            InlineDateRange {
                id: globalDateRange
            }

            // Add global label (inline)
            InlineLabelSelect {
                id: globalLabels
            }

            Row {
                spacing: Style.dialog.spacing
                CheckBoxLabel {
                    id: importEncrypted
                    text: qsTr("Import encrypted emails as they are")
                    anchors {
                        bottom: parent.bottom
                        bottomMargin: Style.dialog.fontSize/1.8
                    }
                }

                InfoToolTip {
                    anchors {
                        verticalCenter: importEncrypted.verticalCenter
                    }
                    info: qsTr("When this option is enabled, encrypted emails will be imported as ciphertext. Otherwise, such messages will be skipped.", "todo")
                }
            }
        }

        // Buttons
        Row {
            spacing: Style.dialog.spacing
            anchors {
                right: parent.right
                bottom: parent.bottom
                rightMargin: Style.main.leftMargin
                bottomMargin: Style.main.bottomMargin
            }

            ButtonRounded {
                id: buttonCancelThree
                fa_icon    : Style.fa.times
                text       : qsTr("Cancel", "todo")
                color_main : Style.dialog.textBlue
                onClicked  : root.cancel()
            }

            ButtonRounded {
                id: buttonNextThree
                fa_icon     : Style.fa.check
                text        : qsTr("Import", "todo")
                color_main  : Style.dialog.background
                color_minor : Style.dialog.textBlue
                isOpaque    : true
                onClicked   : root.okay()
            }
        }
    }

    Rectangle { // Progress
        id: progressStatus
        width  : parent.width
        height : parent.height
        color: Style.transparent

        Column {
            anchors.centerIn: progressStatus
            spacing: Style.dialog.heightSeparator

            Row { // description
                spacing: Style.main.rightMargin
                AccessibleText {
                    id: statusLabel
                    text : qsTr("Status:")
                    font.pointSize: Style.main.iconSize * Style.pt
                    color : Style.main.text
                }
                AccessibleText {
                    anchors.baseline: statusLabel.baseline
                    text : go.progressDescription == "" ? qsTr("importing") : go.progressDescription
                    elide: Text.ElideMiddle
                    width: progressbarImport.width - parent.spacing - statusLabel.width
                    font.pointSize: Style.dialog.textSize * Style.pt
                    color : Style.main.textDisabled
                }
            }

            ProgressBar {
                id: progressbarImport
                implicitWidth  : 2*progressStatus.width/3
                implicitHeight : Style.exporting.rowHeight
                value: go.progress
                property int current:  go.total * go.progress
                property bool isFinished:  finishedPartBar.width == progressbarImport.width

                background: Rectangle {
                    radius         : Style.exporting.boxRadius
                    color          : Style.exporting.progressBackground
                }

                contentItem: Item {
                    Rectangle {
                        id: finishedPartBar
                        width  : parent.width * progressbarImport.visualPosition
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
                            if (progressbarImport.isFinished) return qsTr("Import finished","todo")
                            if (
                                go.progressDescription == gui.enums.progressInit ||
                                (go.progress == 0 && go.progressDescription=="")
                            ) return qsTr("Estimating the total number of messages","todo")
                            if (
                                go.progressDescription == gui.enums.progressLooping
                            ) return qsTr("Loading first message","todo")
                            //var msg = qsTr("Importing message %1 of %2 (%3%)","todo")
                            var msg = qsTr("Importing messages %1 of %2 (%3%)","todo")
                            if (root.paused) msg = qsTr("Importing paused at %1 of %2 (%3%)","todo")
                            return msg.arg(progressbarImport.current).arg(go.total).arg(Math.floor(go.progress*100))
                        }
                        color: Style.main.background
                        font {
                            pointSize: Style.dialog.fontSize * Style.pt
                        }
                    }
                }

                onIsFinishedChanged: { // show report
                    console.log("Is finished ", progressbarImport.isFinished)
                    if (progressbarImport.isFinished && root.currentIndex == DialogImport.Page.Progress) {
                        root.incrementCurrentIndex()
                    }
                }
            }

            Row {
                property int fails: go.progressFails
                visible: fails > 0
                anchors.horizontalCenter: parent.horizontalCenter

                Text {
                    color: Style.main.textRed
                    font {
                        pointSize : Style.dialog.fontSize * Style.pt
                        family    : Style.fontawesome.name
                    }
                    text: Style.fa.exclamation_circle
                }

                Text {
                    property int fails: go.progressFails
                    color: Style.main.textRed
                    font.pointSize: Style.main.fontSize * Style.pt
                    text: " " + (
                        fails == 1 ?
                        qsTr("%1 message failed to be imported").arg(fails) :
                        qsTr("%1 messages failed to be imported").arg(fails)
                    )
                }
            }

            Row { // buttons
                spacing: Style.dialog.rightMargin
                anchors.horizontalCenter: parent.horizontalCenter

                ButtonRounded {
                    id: pauseButton
                    fa_icon    : root.paused ? Style.fa.play :  Style.fa.pause
                    text       : root.paused ? qsTr("Resume") : qsTr("Pause")
                    color_main : Style.dialog.textBlue
                    onClicked  : {
                        if (root.paused) {
                            if (winMain.updateState == gui.enums.statusNoInternet) {
                                go.notifyError(gui.enums.errNoInternet)
                                return
                            }
                            go.resumeProcess()
                        } else {
                            go.pauseProcess()
                        }
                        root.paused = !root.paused
                        pauseButton.focus=false
                    }
                    visible    : !progressbarImport.isFinished
                }

                ButtonRounded {
                    fa_icon    : Style.fa.times
                    text       : qsTr("Cancel")
                    color_main : Style.dialog.textBlue
                    visible    : !progressbarImport.isFinished
                    onClicked  : root.ask_cancel_progress()
                }

                ButtonRounded {
                    id: finish
                    fa_icon     : Style.fa.check
                    text        : qsTr("Okay","todo")
                    color_main  : Style.dialog.background
                    color_minor : Style.dialog.textBlue
                    isOpaque    : true
                    visible     : progressbarImport.isFinished
                    onClicked   : root.okay()
                }
            }
        }

        ImportReport {
        }

        ClickIconText {
            id: buttonHelp
            anchors {
                bottom: progressStatus.bottom
                right: progressStatus.right
                margins: Style.main.rightMargin
            }

            textColor  : Style.main.textDisabled
            iconText   : Style.fa.question_circle
            text       : qsTr("Help", "directs the user to the online user guide")
            textBold   : true
            onClicked  : Qt.openUrlExternally("https://protonmail.com/support/categories/import-export/")
        }
    }

    Rectangle { // Report
        id: finalReport
        width  : parent.width
        height : parent.height
        color: Style.transparent

        property int imported: go.total - go.progressFails


        Column {
            anchors.centerIn : finalReport
            spacing          : Style.dialog.heightSeparator

            Row {
                anchors.horizontalCenter: parent.horizontalCenter

                Text {
                    font {
                        pointSize: Style.dialog.fontSize * Style.pt
                        family: Style.fontawesome.name
                    }
                    color: Style.main.textGreen
                    text: go.progressDescription!="" ? "" : Style.fa.check_circle
                }

                Text {
                    text: go.progressDescription!="" ? qsTr("Import failed: %1").arg(go.progressDescription) : " " + qsTr("Import completed successfully")
                    color: go.progressDescription!="" ? Style.main.textRed : Style.main.textGreen
                    font.bold : true
                }
            }

            Text {
                text: qsTr("<b>Import summary:</b><br>Total number of emails: %1<br>Imported emails: %2<br>Filtered out emails: %3<br>Errors: %4").arg(go.total).arg(go.progressImported).arg(go.progressSkipped).arg(go.progressFails)
                anchors.horizontalCenter: parent.horizontalCenter
                textFormat: Text.RichText
                horizontalAlignment: Text.AlignHCenter
            }

            Row {
                spacing: Style.dialog.rightMargin
                anchors.horizontalCenter: parent.horizontalCenter

                ButtonRounded {
                    fa_icon    : Style.fa.info_circle
                    text       : qsTr("View errors")
                    color_main : Style.dialog.textBlue
                    onClicked  : {
                        go.loadImportReports()
                        reportList.show()
                    }
                }

                ButtonRounded {
                    fa_icon    : Style.fa.send
                    text       : qsTr("Report files")
                    color_main : Style.dialog.textBlue
                    onClicked  : {
                        root.ask_send_report()
                    }
                }
            }

            ButtonRounded {
                text        : qsTr("Close")
                color_main  : Style.dialog.background
                color_minor : Style.dialog.textBlue
                isOpaque    : true
                anchors.horizontalCenter: parent.horizontalCenter
                onClicked: root.okay()
            }
        }

        ImportReport {
            id: reportList
            anchors.fill: finalReport
        }

        ClickIconText {
            anchors {
                bottom: finalReport.bottom
                right: finalReport.right
                margins: Style.main.rightMargin
            }

            textColor  : Style.main.textDisabled
            iconText   : Style.fa.question_circle
            text       : qsTr("Help", "directs the user to the online user guide")
            textBold   : true
            onClicked  : Qt.openUrlExternally("https://protonmail.com/support/categories/import-export/")
        }
    }

    function guessEmailProvider() {
        var splitMail = inputEmail.text.split("@")
        //console.log("finished ", splitMail)
        if (splitMail.length != 2) return
        switch  (splitMail[1]){
            case "yandex.ru":
            case "yandex.com":
            case "ya.ru":
            inputServer.text = "imap.yandex.ru"
            inputPort.text = "993"
            break
            case "outlook.com":
            case "hotmail.com":
            case "live.com":
            case "live.ru":
            inputServer.text = "imap-mail.outlook.com"
            inputPort.text = "993"
            break
            case "seznam.cz":
            case "email.cz":
            case "post.cz":
            inputServer.text = "imap.seznam.cz"
            inputPort.text = "993"
            break
            case "gmx.de":
            inputServer.text = "imap.gmx.net"
            inputPort.text = "993"
            break
            case "rundbox.com":
            inputServer.text = "mail."+splitMail[1]
            inputPort.text = "993"
            break
            case "fastmail.com":
            case "aol.com":
            case "orange.fr":
            case "hushmail.com":
            case "ntlworld.com":
            case "aol.com":
            case "gmx.com":
            case "mail.com":
            case "mail.ru":
            case "gmail.com":
            inputServer.text = "imap."+splitMail[1]
            inputPort.text = "993"
            break
            case (splitMail[1].match(/^yahoo\./) || {}).input:
            inputServer.text = "imap.mail.yahoo.com"
            inputPort.text = "993"
            break
            default:
        }

        return
    }

    function setServerParams() {
        switch (emailProvider.currentIndex) {
            case 1:
            inputServer.text = "imap.gmail.com"
            inputPort.text = "993"
            break
            case 2:
            inputServer.text = "imap.yandex.com"
            inputPort.text = "993"
            break
            case 3:
            inputServer.text = "imap.outlook.com"
            inputPort.text = "993"
            break
            case 4:
            inputServer.text = "imap.yahoo.com"
            inputPort.text = "993"
            break
        }

        return
    }

    function update_label_time() {
        var d = new Date();
        var outstring = " "
        outstring+=qsTr("Import")
        outstring+="-"
        outstring+=d.getDate()
        outstring+="-"
        outstring+=d.getMonth()+1
        outstring+="-"
        outstring+=d.getFullYear()
        outstring+=" "
        outstring+=d.getHours()
        outstring+=":"
        outstring+=d.getMinutes()
        outstring+=" "
        importLabel.text = outstring
    }

    function clear() {
        root.inputPath = ""
        clear_status()
        inputEmail.clear()
        inputPassword.clear()
        inputServer.clear()
        inputPort.clear()
        reportList.hide()
        globalLabels.reset()
    }

    PopupMessage {
        id: errorPopup
        width    : parent.width
        height   : parent.height
        msgWidth : root.width * 0.6108
    }

    Connections {
        target: errorPopup

        onClickedOkay  : errorPopup.hide()

        onClickedYes   : {
            if (errorPopup.msgID == "ask_send_report") {
                errorPopup.hide()
                root.report_sent(go.sendImportReport(root.address))
                return
            }
            root.cancel()
            errorPopup.hide()
        }
        onClickedNo  : {
            errorPopup.hide()
        }

        onClickedRetry  : {
            go.answerRetry()
            errorPopup.hide()
        }
        onClickedSkip   : {
            go.answerSkip(
                errorPopup.checkbox.text == root.msgDontShowAgain  &&
                errorPopup.checkbox.checked
            )
            errorPopup.hide()
        }
        onClickedCancel : {
            root.cancel()
            errorPopup.hide()
        }

        /*
         onClickedCancel  : {
             errorPopup.hide()
             root.ask_cancel_progress()
         }
         */
    }


    function check_inputs() {
        var isOK = true
        switch (currentIndex) {
            case 0: // select source
            var res = go.checkPathStatus(root.inputPath)
            isOK = (
                (res&gui.enums.pathOK)==gui.enums.pathOK &&
                (res&gui.enums.pathNotADir)==0 && // is a dir
                (res&gui.enums.pathDirEmpty)==0 // is nonempty
            )
            if (!isOK) errorPopup.show(qsTr("Please select non-empty folder."))
            break
            // check directory
            case 1: // imap settings
            if (!(
                inputEmail    . checkNonEmpty() &&
                inputPassword . checkNonEmpty() &&
                inputServer   . checkNonEmpty() &&
                inputPort     . checkNonEmpty() &&
                inputPort     . checkIsANumber()
                //emailProvider . currentIndex!=0
            )) isOK = false
            if (winMain.updateState == gui.enums.statusNoInternet) { // todo: use main error dialog for this
                errorPopup.show(qsTr("Please check your internet connection."))
                return false
            }
            break
            case 2: // loading structure
            if (winMain.updateState == gui.enums.statusNoInternet) {
                errorPopup.show(qsTr("Please check your internet connection."))
                return false
            }
            break
            case 3: // import insturctions
            /*
             console.log(" ====== TODO ======== ")
             if (!structureExternal.hasTarget()) {
                 errorPopup.show(qsTr("Nothing selected for import."))
                 return false
             }
             */
            break
            case 4: // import status
        }
        return isOK
    }

    onCancel: {
        switch (currentIndex) {
            case DialogImport.Page.ImapSource:
            case DialogImport.Page.LoadingStructure:
            root.clear()
            root.currentIndex=0
            break
            case DialogImport.Page.SelectSourceType:
            case DialogImport.Page.SourceToTarget:
            case DialogImport.Page.Report:
            root.hide()
            break
            case DialogImport.Page.Progress:
            go.cancelProcess()
            root.currentIndex=3
            root.clear_status()
            globalLabels.reset()
            break
        }
    }

    onOkay: {
        var isOK = check_inputs()
        if (isOK) {
            timer.interval= currentIndex==0 || currentIndex==4 ? 10 : 300
            switch (currentIndex) {
                case DialogImport.Page.SelectSourceType: // select source
                currentIndex=2
                break

                case DialogImport.Page.SourceToTarget:
                globalDateRange.applyRange()
                if (globalLabels.labelSelected) {
                    var isOK = go.createLabelOrFolder(
                        winMain.dialogImport.address,
                        globalLabels.labelName,
                        globalLabels.labelColor,
                        true,
                        "-1"
                    )
                    if (!isOK) return
                }
                incrementCurrentIndex()
                break

                case DialogImport.Page.Report:
                root.clear_status()
                root.hide()
                break

                case DialogImport.Page.LoadingStructure:
                globalLabels.reset()
                // TODO_: importInstructions.hasItems = (structureExternal.rowCount() > 0)
                importInstructions.hasItems = true
                case DialogImport.Page.ImapSource:
                default:
                incrementCurrentIndex()
            }
            timer.start()
        }
    }

    onShow : {
        root.clear()
        if (winMain.updateState==gui.enums.statusNoInternet) {
            if (winMain.updateState==gui.enums.statusNoInternet) {
                winMain.popupMessage.show(go.canNotReachAPI)
                root.hide()
            }
        }
    }

    onHide : {
        root.clear()
    }

    function clear_status() { // TODO: move this to Gui.qml
        go.progress=0.0
        go.progressFails=0.0
        go.total=0.0
        go.progressDescription=gui.enums.progressInit
    }

    function ask_send_report(){
        errorPopup.msgID="ask_send_report"
        errorPopup.buttonYes.visible = true
        errorPopup.buttonNo.visible = true
        errorPopup.buttonOkay.visible = false
        errorPopup.show (qsTr("Program will send the report of finished import process to our customer support. The report was filtered to remove all personal information.\n\nDo you want to send report?"))
    }

    function report_sent(isOK){
        errorPopup.msgID="report_sent"
        if (isOK) {
            errorPopup.show (qsTr("Report sent successfully."))
        } else {
            errorPopup.show (qsTr("Not able to send report. Please contact customer support at importexport@protonmail.com"))
        }
    }

    function ask_cancel_progress(){
        errorPopup.msgID="ask_cancel_progress"
        errorPopup.buttonYes.visible = true
        errorPopup.buttonNo.visible = true
        errorPopup.buttonOkay.visible = false
        errorPopup.show (qsTr("Are you sure you want to cancel this import?"))
    }

    function ask_retry_skip_cancel(subject,errorMessage){
        errorPopup.msgID="ask_retry_skip_cancel"
        errorPopup.buttonYes.visible  = false
        errorPopup.buttonNo.visible   = false
        errorPopup.buttonOkay.visible = false

        errorPopup.buttonRetry.visible  = true
        errorPopup.buttonSkip.visible   = true
        errorPopup.buttonCancel.visible = true

        errorPopup.checkbox.text = root.msgDontShowAgain

        errorPopup.show(
            qsTr(
                "Cannot import message \"%1\"\n\n%2\nCancel will stop the entire import.",
                "error message while importing: arg1 is message subject, arg2 is error message"
            ).arg(subject).arg(errorMessage)
        )
    }

    function sanitizePath(path) {
        var pattern = "file://"
        if (go.goos=="windows") pattern+="/"
        root.inputPath = path.replace(pattern, "")
    }

    Connections {
        target: timer
        onTriggered: {
            switch (currentIndex) {
                case DialogImport.Page.SelectSourceType:
                case DialogImport.Page.ImapSource:
                case DialogImport.Page.SourceToTarget:
                globalDateRange.getRange()
                break
                case DialogImport.Page.LoadingStructure:
                go.setupAndLoadForImport(
                    root.isFromIMAP,
                    root.inputPath,
                    inputEmail.text, inputPassword.text, inputServer.text, inputPort.text,
                    root.account,
                    root.address
                )
                break
                case DialogImport.Page.Progress:
                go.startImport(root.address, importEncrypted.checked)
                break
            }
        }
    }
}
