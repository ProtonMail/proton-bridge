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
import BridgeUI 1.0
import ProtonUI 1.0

// All imports from dynamic must be loaded before
import QtQuick.Window 2.2
import QtQuick.Controls 2.1
import QtQuick.Layouts 1.3

Item {
    id: gui
    property MainWindow winMain
    property bool isFirstWindow: true
    property int warningFlags: 0

    InfoWindow      { id: infoWin      }
    OutgoingNoEncPopup { id: outgoingNoEncPopup }
    BugReportWindow {
        id: bugreportWin
        clientVersion.visible : true

        // pre-fill the form
        onPrefill : {
            userAddress.text=""
            if (accountsModel.count>0) {
                var addressList = accountsModel.get(0).aliases.split(";")
                if (addressList.length>0) {
                    userAddress.text = addressList[0]
                }
            }
            clientVersion.text=go.getLastMailClient()
        }
    }

    onWarningFlagsChanged : {
        if (gui.warningFlags==Style.okInfoBar) {
            go.normalSystray()
            return
        }

        if ((gui.warningFlags & Style.errorInfoBar) == Style.errorInfoBar) {
            go.errorSystray()
            return
        }

        go.highlightSystray()
    }

    // Signals from Go
    Connections {
        target: go

        onShowWindow : {
            gui.openMainWindow()
        }
        onShowHelp : {
            gui.openMainWindow(false)
            winMain.tabbar.currentIndex = 2
            winMain.showAndRise()
        }
        onShowQuit : {
            gui.openMainWindow(false)
            winMain.dialogGlobal.state="quit"
            winMain.dialogGlobal.show()
            winMain.showAndRise()
        }

        onProcessFinished :  {
            winMain.dialogGlobal.hide()
            winMain.dialogAddUser.hide()
            winMain.dialogChangePort.hide()
            infoWin.hide()
        }
        onOpenManual : Qt.openUrlExternally("http://protonmail.com/bridge")

        onNotifyBubble : {
            gui.showBubble(tabIndex, message, true)
        }
        onSilentBubble : {
            gui.showBubble(tabIndex, message, false)
        }
        onBubbleClosed : {
            gui.warningFlags &= ~Style.warnBubbleMessage
        }

        onSetConnectionStatus: {
            go.isConnectionOK = isAvailable
            gui.openMainWindow(false)
            if (go.isConnectionOK) {
                if( winMain.updateState=="noInternet") {
                    go.setUpdateState("upToDate")
                }
            } else {
                go.setUpdateState("noInternet")
            }
        }

        onSetUpdateState : {
            // once app is outdated prevent from state change
            if (winMain.updateState != "forceUpdate") {
                winMain.updateState = updateState
            }
        }

        onSetAddAccountWarning : winMain.dialogAddUser.setWarning(message, 0)


        onNotifyVersionIsTheLatest : {
            go.silentBubble(2,qsTr("You have the latest version!", "notification", -1))
        }

        onNotifyManualUpdate: {
            go.setUpdateState("oldVersion")
        }

        onNotifyManualUpdateRestartNeeded: {
            if (!winMain.dialogUpdate.visible) {
                gui.openMainWindow(true)
                winMain.dialogUpdate.show()
            }
            go.setUpdateState("updateRestart")
            winMain.dialogUpdate.finished(false)

            // after manual update - just retart immidiatly
            go.setToRestart()
            Qt.quit()
        }

        onNotifyManualUpdateError: {
            if (!winMain.dialogUpdate.visible) {
                gui.openMainWindow(true)
                winMain.dialogUpdate.show()
            }
            go.setUpdateState("updateError")
            winMain.dialogUpdate.finished(true)
        }

        onNotifyForceUpdate : {
            go.setUpdateState("forceUpdate")
            if (!winMain.dialogUpdate.visible) {
                gui.openMainWindow(true)
                winMain.dialogUpdate.show()
            }
        }

        onNotifySilentUpdateRestartNeeded: {
            go.setUpdateState("updateRestart")
        }

        onNotifySilentUpdateError: {
            go.setUpdateState("updateError")
            gui.openMainWindow(true)
        }

        onNotifyLogout : {
            go.notifyBubble(0, qsTr("Account %1 has been disconnected. Please log in to continue to use the Bridge with this account.").arg(accname) )
        }

        onNotifyAddressChanged : {
            go.notifyBubble(0, qsTr("The address list has been changed for account %1. You may need to reconfigure the settings in your email client.").arg(accname) )
        }

        onNotifyAddressChangedLogout : {
            go.notifyBubble(0, qsTr("The address list has been changed for account %1. You have to reconfigure the settings in your email client.").arg(accname) )
        }

        onNotifyPortIssue : { // busyPortIMAP , busyPortSMTP
            if (!busyPortIMAP && !busyPortSMTP) { // at least one must have issues to show warning
                return
            }
            gui.openMainWindow(false)
            winMain.tabbar.currentIndex=1
            go.isDefaultPort = false
            var text
            if (busyPortIMAP && busyPortSMTP) { // both have problems
                text = qsTr("The default ports used by Bridge for IMAP (%1) and SMTP (%2) are occupied by one or more other applications." , "the first part of notification text (two ports)").arg(go.getIMAPPort()).arg(go.getSMTPPort())
                text += " "
                text += qsTr("To change the ports for these servers, go to Settings -> Advanced Settings.", "the second part of notification text (two ports)")
            } else { // only one is occupied
                var server, port
                if (busyPortSMTP) {
                    server = "SMTP"
                    port = go.getSMTPPort()
                } else {
                    server = "IMAP"
                    port = go.getIMAPPort()
                }
                text = qsTr("The default port used by Bridge for %1 (%2) is occupied by another application.", "the first part of notification text (one port)").arg(server).arg(port)
                text += " "
                text += qsTr("To change the port for this server, go to Settings -> Advanced Settings.", "the second part of notification text (one port)")
            }
            go.notifyBubble(1, text )
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

        onShowNoActiveKeyForRecipient : {
            go.notifyBubble(0, qsTr(
                "Key pinning is enabled for %1 but no active key is pinned. " +
                "You must pin the key in order to send a message to this address. " +
                "You can find instructions " +
                "<a href=\"https://protonmail.com/support/knowledge-base/key-pinning/\">here</a>."
            ).arg(recipient))
        }

        onFailedAutostartCode : {
            gui.openMainWindow(true)
            switch (code) {
                case "permission" : // linux+darwin
                case "85070005" : // windows
                go.notifyBubble(1, go.failedAutostartPerm)
                break
                case "81004003" : // windows
                go.notifyBubble(1, go.failedAutostart+" "+qsTr("Can not create instance.", "for autostart"))
                break
                case "" :
                default :
                go.notifyBubble(1, go.failedAutostart)
            }
        }

        onShowOutgoingNoEncPopup : {
            outgoingNoEncPopup.show(messageID, subject)
        }

        onSetOutgoingNoEncPopupCoord : {
            outgoingNoEncPopup.x = x
            outgoingNoEncPopup.y = y
        }

        onShowCertIssue : {
            winMain.tlsBarState="notOK"
        }

        onOpenReleaseNotesExternally: {
            Qt.openUrlExternally(go.updateReleaseNotesLink)
        }

    }

    function openMainWindow(showAndRise) {
        // wait and check until font is loaded
        while(true){
            if (Style.fontawesome.status == FontLoader.Loading) continue
            if (Style.fontawesome.status != FontLoader.Ready) console.log("Error while loading font")
            break
        }

        if (typeof(showAndRise)==='undefined') {
            showAndRise = true
        }
        if (gui.winMain == null) {
            gui.winMain = Qt.createQmlObject(
                'import BridgeUI 1.0; MainWindow {visible : false}',
                gui, "winMain"
            )
        }
        if (showAndRise) {
            gui.winMain.showAndRise()
        }
    }

    function closeMainWindow () {
        gui.winMain.hide()
        gui.winMain.destroy(5000)
        gui.winMain = null
        gui.isFirstWindow = false
    }

    function showBubble(tabIndex, message, isWarning) {
        gui.openMainWindow(true)
        if (isWarning) {
            gui.warningFlags |= Style.warnBubbleMessage
        }
        winMain.bubbleNote.text = message
        winMain.bubbleNote.place(tabIndex)
        winMain.bubbleNote.show()
    }

    function openReleaseNotes(){
        if (go.updateReleaseNotesLink == "") {
            go.checkAndOpenReleaseNotes()
            return
        }
        go.openReleaseNotesExternally()
    }


    // On start
    Component.onCompleted : {
        // set  messages for translations
        go.wrongCredentials       = qsTr("Incorrect username or password."                               , "notification", -1)
        go.wrongMailboxPassword   = qsTr("Incorrect mailbox password."                                   , "notification", -1)
        go.canNotReachAPI         = qsTr("Cannot contact server, please check your internet connection." , "notification", -1)
        go.versionCheckFailed     = qsTr("Version check was unsuccessful. Please try again later."       , "notification", -1)
        go.credentialsNotRemoved  = qsTr("Credentials could not be removed."                             , "notification", -1)
        go.failedAutostartPerm    = qsTr("Unable to configure automatic start due to permissions settings - see <a href=\"https://protonmail.com/bridge/faq#c11\">FAQ</a> for details.", "notification", -1)
        go.failedAutostart        = qsTr("Unable to configure automatic start."                           , "notification", -1)
        go.genericErrSeeLogs      = qsTr("An error happened during procedure. See logs for more details." , "notification", -1)

        // start window
        gui.openMainWindow(false)
        if (go.isShownOnStart) {
            gui.winMain.showAndRise()
        }
    }
}
