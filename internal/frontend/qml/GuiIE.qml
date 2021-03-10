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

// This is main qml file

import QtQuick 2.8
import ImportExportUI 1.0
import ProtonUI 1.0

// All imports from dynamic must be loaded before
import QtQuick.Window 2.2
import QtQuick.Controls 2.1
import QtQuick.Layouts 1.3

Item {
    id: gui
    property alias winMain: winMain
    property bool isFirstWindow: true

    property var locale    : Qt.locale("en_US")
    property date netBday  : new Date("1989-03-13T00:00:00")
    property var allYears  : getYearList(1970,(new Date()).getFullYear())
    property var allMonths : getMonthList(1,12)
    property var allDays   : getDayList(1,31)

    property var enums : JSON.parse('{"pathOK":1,"pathEmptyPath":2,"pathWrongPath":4,"pathNotADir":8,"pathWrongPermissions":16,"pathDirEmpty":32,"errUnknownError":0,"errEventAPILogout":1,"errUpdateAPI":2,"errUpdateJSON":3,"errUserAuth":4,"errQApplication":18,"errEmailExportFailed":6,"errEmailExportMissing":7,"errNothingToImport":8,"errEmailImportFailed":12,"errDraftImportFailed":13,"errDraftLabelFailed":14,"errEncryptMessageAttachment":15,"errEncryptMessage":16,"errNoInternetWhileImport":17,"errUnlockUser":5,"errSourceMessageNotSelected":19,"errCannotParseMail":5000,"errWrongLoginOrPassword":5001,"errWrongServerPathOrPort":5002,"errWrongAuthMethod":5003,"errIMAPFetchFailed":5004,"errLocalSourceLoadFailed":1000,"errPMLoadFailed":1001,"errRemoteSourceLoadFailed":1002,"errLoadAccountList":1005,"errExit":1006,"errRetry":1007,"errAsk":1008,"errImportFailed":1009,"errCreateLabelFailed":1010,"errCreateFolderFailed":1011,"errUpdateLabelFailed":1012,"errUpdateFolderFailed":1013,"errFillFolderName":1014,"errSelectFolderColor":1015,"errNoInternet":1016,"folderTypeSystem":"system","folderTypeLabel":"label","folderTypeFolder":"folder","folderTypeExternal":"external","progressInit":"init","progressLooping":"looping","statusNoInternet":"noInternet","statusCheckingInternet":"internetCheck","statusNewVersionAvailable":"oldVersion","statusUpToDate":"upToDate","statusForceUpdate":"forceUpdate"}')

    IEStyle{}

    MainWindow {
        id: winMain

        visible : true
        Component.onCompleted: {
            winMain.showAndRise()
        }
    }

    BugReportWindow {
        id:bugreportWin
        clientVersion.visible: false
        onPrefill : {
            userAddress.text=""
            if (accountsModel.count>0) {
                var addressList = accountsModel.get(0).aliases.split(";")
                if (addressList.length>0) {
                    userAddress.text = addressList[0]
                }
            }
        }
    }

    // Signals from Go
    Connections {
        target: go

        onShowWindow : {
            winMain.showAndRise()
        }

        onProcessFinished :  {
            winMain.dialogAddUser.hide()
            winMain.dialogGlobal.hide()
        }
        onOpenManual : Qt.openUrlExternally("https://protonmail.com/support/categories/import-export/")

        onNotifyBubble : {
            //go.highlightSystray()
            winMain.bubleNote.text = message
            winMain.bubleNote.place(tabIndex)
            winMain.bubleNote.show()
            winMain.showAndRise()
        }
        onBubbleClosed : {
            if (winMain.updateState=="uptodate") {
                //go.normalSystray()
            }
        }

        onSetConnectionStatus: {
            go.isConnectionOK = isAvailable
            if (go.isConnectionOK) {
                if( winMain.updateState==gui.enums.statusNoInternet) {
                    go.updateState = gui.enums.statusUpToDate
                }
            } else {
                go.updateState = gui.enums.statusNoInternet
            }
        }

        onUpdateStateChanged : {
            // once app is outdated prevent from state change
            if (winMain.updateState != "forceUpdate") {
                winMain.updateState = go.updateState
            }
        }

        onSetAddAccountWarning : winMain.dialogAddUser.setWarning(message, 0)

        onNotifyVersionIsTheLatest : {
            winMain.popupMessage.show(
                qsTr("You have the latest version!", "todo")
            )
        }

        onNotifyError : {
            var sep = go.errorDescription.indexOf("\n") < 0 ? go.errorDescription.length : go.errorDescription.indexOf("\n")
            var name = go.errorDescription.slice(0, sep)
            var errorMessage = go.errorDescription.slice(sep)
            switch (errCode) {
                case gui.enums.errPMLoadFailed :
                winMain.popupMessage.show ( qsTr ( "Loading ProtonMail folders and labels was not successful."  , "Error message" )  )
                winMain.dialogExport.hide()
                break
                case gui.enums.errLocalSourceLoadFailed  :
                winMain.popupMessage.show(qsTr(
                    "Loading local folder structure was not successful. "+
                    "Folder does not contain valid MBOX or EML file.",
                    "Error message when can not find correct files in folder."
                ))
                winMain.dialogImport.hide()
                break
                case gui.enums.errRemoteSourceLoadFailed : 
                winMain.popupMessage.show ( qsTr ( "Loading remote source structure was not successful."        , "Error message" )  )
                winMain.dialogImport.hide()
                break
                case gui.enums.errWrongServerPathOrPort      :
                winMain.popupMessage.show ( qsTr ( "Cannot contact server - incorrect server address and port." , "Error message" )  )
                winMain.dialogImport.decrementCurrentIndex()
                break
                case gui.enums.errWrongLoginOrPassword   :
                winMain.popupMessage.show ( qsTr ( "Cannot authenticate - Incorrect email or password."     , "Error message" )  )
                winMain.dialogImport.decrementCurrentIndex()
                break ;
                case gui.enums.errWrongAuthMethod   :
                winMain.popupMessage.show ( qsTr ( "Cannot authenticate - Please use secured authentication method."     , "Error message" )  )
                winMain.dialogImport.decrementCurrentIndex()
                break ;


                case gui.enums.errFillFolderName: 
                winMain.popupMessage.show(qsTr (
                    "Please fill the name.",
                    "Error message when user did not fill the name of folder or label"
                ))
                break
                case gui.enums.errCreateLabelFailed: 
                winMain.popupMessage.show(qsTr(
                    "Cannot create label with name \"%1\"\n%2",
                    "Error message when it is not possible to create new label, arg1 folder name, arg2 error reason"
                ).arg(name).arg(errorMessage))
                break
                case gui.enums.errCreateFolderFailed: 
                winMain.popupMessage.show(qsTr(
                    "Cannot create folder with name \"%1\"\n%2",
                    "Error message when it is not possible to create new folder, arg1 folder name, arg2 error reason"
                ).arg(name).arg(errorMessage))
                break

                case gui.enums.errNothingToImport:
                winMain.popupMessage.show ( qsTr ( "No emails left to import after date range applied. Please, change the date range to continue."     , "Error message" )  )
                winMain.dialogImport.decrementCurrentIndex()
                break

                case gui.enums.errNoInternetWhileImport:
                case gui.enums.errNoInternet:
                go.setConnectionStatus(false)
                winMain.popupMessage.show ( go.canNotReachAPI )
                break

                case gui.enums.errPMAPIMessageTooLarge:
                case gui.enums.errIMAPFetchFailed:
                case gui.enums.errEmailImportFailed :
                case gui.enums.errDraftImportFailed :
                case gui.enums.errDraftLabelFailed :
                case gui.enums.errEncryptMessageAttachment:
                case gui.enums.errEncryptMessage:
                //winMain.dialogImport.ask_retry_skip_cancel(name, errorMessage)
                console.log("Import error", errCode, go.errorDescription)
                winMain.popupMessage.show(qsTr("Error during import: \n%1\n please see log files for more details.", "message of generic error").arg(go.errorDescription))
                winMain.dialogImport.hide()
                break;

                case gui.enums.errUnknownError  : default:
                console.log("Unknown Error", errCode, go.errorDescription)
                winMain.popupMessage.show(qsTr("The program encounter an unknown error \n%1\n please see log files for more details.", "message of generic error").arg(go.errorDescription))
                winMain.dialogExport.hide()
                winMain.dialogImport.hide()
                winMain.dialogAddUser.hide()
                winMain.dialogGlobal.hide()
            }
        }

        onNotifyManualUpdate: {
            go.updateState = "oldVersion"
        }

        onNotifyManualUpdateRestartNeeded: {
            if (!winMain.dialogUpdate.visible) {
                winMain.dialogUpdate.show()
            }
            go.updateState = "updateRestart"
            winMain.dialogUpdate.finished(false)
        
            // after manual update - just retart immidiatly
            go.setToRestart()
            Qt.quit()
        }

        onNotifyManualUpdateError: {
            if (!winMain.dialogUpdate.visible) {
                winMain.dialogUpdate.show()
            }
            go.updateState = "updateError"
            winMain.dialogUpdate.finished(true)
        }

        onNotifyForceUpdate : {
            go.updateState = "forceUpdate"
            if (!winMain.dialogUpdate.visible) {
                winMain.dialogUpdate.show()
            }
        }

        //onNotifySilentUpdateRestartNeeded: {
        //    go.updateState = "updateRestart"
        //}
        //
        //onNotifySilentUpdateError: {
        //    go.updateState = "updateError"
        //}

        onNotifyLogout : {
            go.notifyBubble(0, qsTr("Account %1 has been disconnected. Please log in to continue to use the Import-Export app with this account.").arg(accname) )
        }

        onNotifyAddressChanged : {
            go.notifyBubble(0, qsTr("The address list has been changed for account %1. You may need to reconfigure the settings in your email client.").arg(accname) )
        }

        onNotifyAddressChangedLogout : {
            go.notifyBubble(0, qsTr("The address list has been changed for account %1. You have to reconfigure the settings in your email client.").arg(accname) )
        }


        onNotifyKeychainRebuild : {
            go.notifyBubble(1, qsTr(
                "Your MacOS keychain is probably corrupted. Please consult the instructions in our <a href=\"https://protonmail.com/bridge/faq#c15\">FAQ</a>.",
                "notification message"
            ))
        }

        onNotifyHasNoKeychain : {
            gui.winMain.dialogGlobal.state="noKeychain"
            gui.winMain.dialogGlobal.show()
        }


        onExportStructureLoadFinished: {
            if (okay) winMain.dialogExport.okay()
            else winMain.dialogExport.cancel()
        }
        onImportStructuresLoadFinished: {
            if (okay) winMain.dialogImport.okay()
            else winMain.dialogImport.cancel()
        }

        onSimpleErrorHappen: {
            if (winMain.dialogImport.visible == true) {
                winMain.dialogImport.hide()
            }
            if (winMain.dialogExport.visible == true) {
                winMain.dialogExport.hide()
            }
        }

        onUpdateFinished : {
            winMain.dialogUpdate.finished(hasError)
        }

        onOpenReleaseNotesExternally: {
            Qt.openUrlExternally(go.updateReleaseNotesLink)
        }


    }

    function folderIcon(folderName, folderType) { // translations
        switch (folderName.toLowerCase()) {
            case "inbox"   : return Style.fa.inbox
            case "sent"    : return Style.fa.send
            case "spam"    :
            case "junk"    : return Style.fa.ban
            case "draft"   : return Style.fa.file_o
            case "starred" : return Style.fa.star_o
            case "trash"   : return Style.fa.trash_o
            case "archive" : return Style.fa.archive
            default: return folderType == gui.enums.folderTypeLabel ? Style.fa.tag : Style.fa.folder_open
        }
        return Style.fa.sticky_note_o
    }

    function folderTypeTitle(folderType) { // translations
        if (folderType==gui.enums.folderTypeSystem ) return ""
        if (folderType==gui.enums.folderTypeLabel  ) return qsTr("Labels"  , "todo")
        if (folderType==gui.enums.folderTypeFolder ) return qsTr("Folders" , "todo")
        return "Undef"
    }

    function isFolderEmpty() {
        return "true"
    }

    function getUnixTime(dateString) {
        var d = new Date(dateString)
        var n = d.getTime()
        if (n != n) return -1
        return n
    }

    function getYearList(minY,maxY) {
        var years  = new Array()
        for (var i=0; i<=maxY-minY;i++) {
            years[i] = (maxY-i).toString()
        }
        //console.log("getYearList:", years)
        return years
    }

    function getMonthList(minM,maxM) {
        var months = new Array()
        for (var i=0; i<=maxM-minM;i++) {
            var iMonth = new Date(1989,(i+minM-1),13)
            months[i] = iMonth.toLocaleString(gui.locale, "MMM")
        }
        //console.log("getMonthList:", months[0], months)
        return months
    }

    function getDayList(minD,maxD) {
        var days = new Array()
        for (var i=0; i<=maxD-minD;i++) {
            days[i] = gui.prependZeros(i+minD,2)
        }
        return days
    }

    function prependZeros(num,desiredLength) {
        var s = num+""
        while (s.length < desiredLength) s="0"+s
        return s
    }

    function daysInMonth(year,month) {
        if (typeof(year) !== 'number') {
            year = parseInt(year)
        }
        if (typeof(month) !== 'number') {
            month = Date.fromLocaleDateString( gui.locale, "1970-"+month+"-10", "yyyy-MMM-dd").getMonth()+1
        }
        var maxDays = (new Date(year,month,0)).getDate()
        if (isNaN(maxDays)) maxDays = 0
        //console.log(" daysInMonth", year, month, maxDays)
        return maxDays
    }

    function niceDateTime() {
        var stamp = new Date()
        var nice = getMonthList(stamp.getMonth()+1, stamp.getMonth()+1)[0]
        nice += "-" + getDayList(stamp.getDate(), stamp.getDate())[0]
        nice += "-" + getYearList(stamp.getFullYear(), stamp.getFullYear())[0]
        nice += " " + gui.prependZeros(stamp.getHours(),2)
        nice += ":" + gui.prependZeros(stamp.getMinutes(),2)
        return nice
    }

    /*
     // Debug
     Connections {
         target: structureExternal

         onDataChanged: {
             console.log("external data changed")
         }
     }

     // Debug
     Connections {
         target: structurePM

         onSelectedLabelsChanged: console.log("PM sel labels:", structurePM.selectedLabels)
         onSelectedFoldersChanged: console.log("PM sel folders:", structurePM.selectedFolders)
         onDataChanged: {
             console.log("PM data changed")
         }
     }
     */

    function openReleaseNotes(){
        if (go.updateReleaseNotesLink == "") {
            go.checkAndOpenReleaseNotes()
            return
        }
        go.openReleaseNotesExternally()
    }


    property string  areYouSureYouWantToQuit : qsTr("There are incomplete processes - some items are not yet transferred. Do you really want to stop and quit?")
    // On start
    Component.onCompleted : {
        // set spell messages
        go.wrongCredentials       = qsTr("Incorrect username or password."                               , "notification", -1)
        go.wrongMailboxPassword   = qsTr("Incorrect mailbox password."                                   , "notification", -1)
        go.canNotReachAPI         = qsTr("Cannot contact server, please check your internet connection." , "notification", -1)
        go.versionCheckFailed     = qsTr("Version check was unsuccessful. Please try again later."       , "notification", -1)
        go.credentialsNotRemoved  = qsTr("Credentials could not be removed."                             , "notification", -1)
        go.bugNotSent             = qsTr("Unable to submit bug report."                                  , "notification", -1)
        go.bugReportSent          = qsTr("Bug report successfully sent."                                 , "notification", -1)


        go.guiIsReady()

        gui.allMonths = getMonthList(1,12)
        gui.allMonthsChanged()
    }
}
