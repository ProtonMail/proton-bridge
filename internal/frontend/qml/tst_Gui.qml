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
import QtTest 1.2
import BridgeUI 1.0
import ProtonUI 1.0
import QtQuick.Controls 2.1
import QtQuick.Window 2.2

Window {
    id: testroot
    width   : 150
    height  : 600
    flags   : Qt.Window | Qt.Dialog | Qt.FramelessWindowHint
    visible : true
    title   : "GUI test Window"
    color   : "transparent"

    property bool newVersion : true

    Accessible.name: testroot.title
    Accessible.description: "Window with buttons testing the GUI events"


    Rectangle {
        id:test_systray
        anchors{
            top: parent.top
            horizontalCenter: parent.horizontalCenter
        }
        height: 40
        width: testroot.width
        color: "yellow"
        Image {
            id: sysImg
            anchors {
                left : test_systray.left
                top  : test_systray.top
            }
            height: test_systray.height
            mipmap: true
            fillMode : Image.PreserveAspectFit
            source: ""
        }
        Text {
            id: systrText
            anchors {
                right : test_systray.right
                verticalCenter: test_systray.verticalCenter
            }
            text: "unset"
        }

        function normal() {
            test_systray.color = "#22ee22"
            systrText.text = "norm"
            sysImg.source= "../share/icons/black-systray.png"
        }
        function highlight() {
            test_systray.color = "#eeee22"
            systrText.text = "highl"
            sysImg.source= "../share/icons/black-syswarn.png"
        }
        function error() {
            test_systray.color = "#ee2222"
            systrText.text = "error"
            sysImg.source= "../share/icons/black-syserror.png"
        }

        MouseArea {
            property point diff: "0,0"
            anchors.fill: parent
            onPressed: {
                diff = Qt.point(testroot.x, testroot.y)
                var mousePos = mapToGlobal(mouse.x, mouse.y)
                diff.x -= mousePos.x
                diff.y -= mousePos.y
            }
            onPositionChanged: {
                var currPos = mapToGlobal(mouse.x, mouse.y)
                testroot.x = currPos.x + diff.x
                testroot.y = currPos.y + diff.y
            }
        }
    }

    ListModel {
        id: buttons

        ListElement { title: "Show window"    }
        ListElement { title: "Show help"      }
        ListElement { title: "Show quit"      }
        ListElement { title: "Logout bridge"  }
        ListElement { title: "Internet on"    }
        ListElement { title: "Internet off"   }
        ListElement { title: "UpToDate"       }
        ListElement { title: "NotifyManualUpdate(CanInstall)" }
        ListElement { title: "NotifyManualUpdate(CantInstall)" }
        ListElement { title: "NotifyManualUpdateRestart" }
        ListElement { title: "NotifyManualUpdateError" }
        ListElement { title: "ForceUpdate" }
        ListElement { title: "NotifySilentUpdateRestartNeeded" }
        ListElement { title: "NotifySilentUpdateError" }
        ListElement { title: "Linux"          }
        ListElement { title: "Windows"        }
        ListElement { title: "Macos"          }
        ListElement { title: "FirstDialog"    }
        ListElement { title: "AutostartError" }
        ListElement { title: "BusyPortIMAP"   }
        ListElement { title: "BusyPortSMTP"   }
        ListElement { title: "BusyPortBOTH"   }
        ListElement { title: "Minimize this"  }
        ListElement { title: "SendAlertPopup" }
        ListElement { title: "TLSCertError"   }
    }

    ListView {
        id: view
        anchors {
            top    : test_systray.bottom
            bottom : parent.bottom
            left   : parent.left
            right  : parent.right
        }

        orientation : ListView.Vertical
        model       : buttons
        focus       : true

        delegate : ButtonRounded {
            text        : title
            color_main  : "orange"
            color_minor : "#aa335588"
            isOpaque    : true
            width: testroot.width
            height : 20*Style.px
            anchors.horizontalCenter: parent.horizontalCenter
            onClicked : {
                console.log("Clicked on ", title)
                switch (title)  {
                    case "Show window" :
                    go.showWindow();
                    break;
                    case "Show help" :
                    go.showHelp();
                    break;
                    case "Show quit" :
                    go.showQuit();
                    break;
                    case "Logout bridge" :
                    go.checkLoggedOut("bridge");
                    break;
                    case "Internet on" :
                    go.setConnectionStatus(true);
                    break;
                    case "Internet off" :
                    go.setConnectionStatus(false);
                    break;
                    case "Linux" :
                    go.goos = "linux";
                    break;
                    case "Macos" :
                    go.goos = "darwin";
                    break;
                    case "Windows" :
                    go.goos = "windows";
                    break;
                    case "FirstDialog" :
                    testgui.winMain.dialogFirstStart.show();
                    break;
                    case "AutostartError" :
                    go.notifyBubble(1,go.failedAutostart);
                    break;
                    case "BusyPortIMAP" :
                    go.notifyPortIssue(true,false);
                    break;
                    case "BusyPortSMTP" :
                    go.notifyPortIssue(false,true);
                    break;
                    case "BusyPortBOTH" :
                    go.notifyPortIssue(true,true);
                    break;
                    case "Minimize this" :
                    testroot.visibility = Window.Minimized
                    break;
                    case "UpToDate" :
                    testroot.newVersion = false
                    break;
                    case "NotifyManualUpdate(CanInstall)" :
                    go.notifyManualUpdate()
                    go.updateCanInstall = true
                    break;
                    case "NotifyManualUpdate(CantInstall)" :
                    go.notifyManualUpdate()
                    go.updateCanInstall = false
                    break;
                    case "NotifyManualUpdateRestart":
                    go.notifyManualUpdateRestartNeeded()
                    break;
                    case "NotifyManualUpdateError":
                    go.notifyManualUpdateError()
                    break;
                    case "ForceUpdate" :
                    go.notifyForceUpdate()
                    break;
                    case "NotifySilentUpdateRestartNeeded" :
                    go.notifySilentUpdateRestartNeeded()
                    break;
                    case "NotifySilentUpdateError" :
                    go.notifySilentUpdateError()
                    break;
                    case "SendAlertPopup" :
                    go.showOutgoingNoEncPopup("Alert sending unencrypted!")
                    break;
                    case "TLSCertError" :
                    go.showCertIssue()
                    break;
                    default :
                    console.log("Not implemented " + data)
                }
            }
        }
    }


    Component.onCompleted : {
        testroot.x= 10
        testroot.y= 100
    }

    //InstanceExistsWindow { id: ie_test }

    Gui {
        id: testgui

        ListModel{
            id: accountsModel
            ListElement{ account : "bridge"                                           ; status : "connected";    isExpanded: false; isCombinedAddressMode: false; hostname : "127.0.0.1"; password : "ZI9tKp+ryaxmbpn2E12"; security : "StarTLS"; portSMTP : 1025; portIMAP : 1143; aliases : "bridge@pm.com;bridge2@pm.com;theHorriblySlowMurderWithExtremelyInefficientWeapon@youtube.com" }
            ListElement{ account : "exteremelongnamewhichmustbeeladed@protonmail.com" ; status : "connected";    isExpanded: true;  isCombinedAddressMode: true;  hostname : "127.0.0.1"; password : "ZI9tKp+ryaxmbpn2E12"; security : "StarTLS"; portSMTP : 1025; portIMAP : 1143; aliases : "bridge@pm.com;bridge2@pm.com;hu@hu.hu"                                                        }
            ListElement{ account : "bridge2@protonmail.com"                           ; status : "disconnected"; isExpanded: false; isCombinedAddressMode: false; hostname : "127.0.0.1"; password : "ZI9tKp+ryaxmbpn2E12"; security : "StarTLS"; portSMTP : 1025; portIMAP : 1143; aliases : "bridge@pm.com;bridge2@pm.com;hu@hu.hu"                                                        }
        }

        Component.onCompleted : {
            winMain.x = testroot.x + testroot.width
            winMain.y = testroot.y
        }
    }


    QtObject {
        id: go

        property bool isAutoStart : true
        property bool isAutoUpdate : false
        property bool isEarlyAccess : false
        property bool isProxyAllowed : false
        property bool isFirstStart : false
        property bool isFreshVersion : false
        property bool isOutdateVersion : true
        property string currentAddress : "none"
        //property string goos : "windows"
        property string goos : "linux"
        ////property string goos : "darwin"
        property bool isDefaultPort : false
        property bool isShownOnStart : true

        property bool hasNoKeychain : true

        property string wrongCredentials
        property string wrongMailboxPassword
        property string canNotReachAPI
        property string versionCheckFailed
        property string credentialsNotRemoved
        property string bugNotSent
        property string bugReportSent
        property string failedAutostartPerm
        property string failedAutostart
        property string genericErrSeeLogs

        property string programTitle : "ProtonMail Bridge"
        property string fullversion : "QA.1.0 (d9f8sdf9) 2020-02-19T10:57:23+01:00"
        property string downloadLink: "https://protonmail.com/download/beta/protonmail-bridge-1.1.5-1.x86_64.rpm;https://www.protonmail.com/downloads/beta/Desktop-Bridge-link1.exe;https://www.protonmail.com/downloads/beta/Desktop-Bridge-link1.exe;https://www.protonmail.com/downloads/beta/Desktop-Bridge-link1.exe;"

        property string updateVersion : "QA.1.0"
        property bool updateCanInstall: true
        property string updateLandingPage : "https://protonmail.com/bridge/download/"
        property string updateReleaseNotesLink  : "" // "https://protonmail.com/download/bridge/release_notes.html"
        signal notifyManualUpdate()
        signal notifyManualUpdateRestartNeeded()
        signal notifyManualUpdateError()
        signal notifyForceUpdate()
        signal notifySilentUpdateRestartNeeded()
        signal notifySilentUpdateError()
        function checkForUpdates() {
            console.log("checkForUpdates")
        }
        function startManualUpdate() {
            console.log("startManualUpdate")
        }
        function checkAndOpenReleaseNotes() {
            console.log("check for release notes")
            go.updateReleaseNotesLink = "https://protonmail.com/download/bridge/release_notes.html"
            go.openReleaseNotesExternally()
        }


        property string credits    : "here;goes;list;;of;;used;packages;"

        property real progress: 0.3
        property int progressDescription: 2

        function setToRestart() {
            console.log("setting to restart")
        }

        signal toggleMainWin(int systX, int systY, int systW, int systH)

        signal showWindow()
        signal showHelp()
        signal showQuit()

        signal notifyPortIssue(bool busyPortIMAP, bool busyPortSMTP)
        signal notifyVersionIsTheLatest()
        signal setUpdateState(string updateState)
        signal notifyKeychainRebuild()
        signal notifyHasNoKeychain()

        signal processFinished()
        signal toggleAutoStart()
        signal toggleEarlyAccess()
        signal toggleAutoUpdate()
        signal notifyBubble(int tabIndex, string message)
        signal silentBubble(int tabIndex, string message)
        signal setAddAccountWarning(string message)

        signal notifyFirewall()
        signal notifyLogout(string accname)
        signal notifyAddressChanged(string accname)
        signal notifyAddressChangedLogout(string accname)
        signal failedAutostartCode(string code)

        signal openReleaseNotesExternally()
        signal showCertIssue()

        signal updateFinished(bool hasError)


        signal  showOutgoingNoEncPopup(string subject)
        signal  setOutgoingNoEncPopupCoord(real x, real y)
        signal  showNoActiveKeyForRecipient(string recipient)

        function delay(duration) {
            var timeStart = new Date().getTime();

            while (new Date().getTime() - timeStart < duration) {
                // Do nothing
            }
        }

        function getLastMailClient() {
            return "Mutt is the best"
        }

        function sendBug(desc,client,address){
            console.log("bug report ", "desc '"+desc+"'", "client '"+client+"'", "address '"+address+"'")
            return !desc.includes("fail")
        }

        function deleteAccount(index,remove) {
            console.log ("Test: Delete account ",index," and remove prefences "+remove)
            workAndClose()
            accountsModel.remove(index)
        }

        function logoutAccount(index) {
            accountsModel.get(index).status="disconnected"
            workAndClose()
        }

        function login(username,password) {
            delay(700)
            if (password=="wrong") {
                setAddAccountWarning("Wrong password")
                return -1
            }
            if (username=="2fa") {
                return 1
            }
            if (username=="mbox") {
                return 2
            }
            return 0
        }

        function auth2FA(twoFACode){
            delay(700)
            if (twoFACode=="wrong") {
                setAddAccountWarning("Wrong 2FA")
                return -1
            }
            if (twoFACode=="mbox") {
                return 1
            }
            return 0
        }

        function addAccount(mailboxPass) {
            delay(700)
            if (mailboxPass=="wrong") {
                setAddAccountWarning("Wrong mailbox password")
                return -1
            }
            accountsModel.append({
                "account" : testgui.winMain.dialogAddUser.username,
                "status" : "connected",
                "isExpanded":true,
                "hostname" : "127.0.0.1",
                "password" : "ZI9tKp+ryaxmbpn2E12",
                "security" : "StarTLS",
                "portSMTP" : 1025,
                "portIMAP" : 1143,
                "aliases" : "bridge@pm.com;bridges@pm.com;theHorriblySlowMurderWithExtremelyInefficientWeapon@youtube.com",
                "isCombinedAddressMode": true
            })
            workAndClose()
        }

        function checkInternet() {
            var delay = Qt.createQmlObject("import QtQuick 2.8; Timer{}",go)
            delay.interval = 2000
            delay.repeat = false
            delay.triggered.connect(function(){ go.setConnectionStatus(false) })
            delay.start()
        }

        property SequentialAnimation animateProgressBar : SequentialAnimation {
            // version
            PropertyAnimation{ target: go; properties: "progressDescription"; to: 1; duration: 1; }
            PropertyAnimation{ duration: 2000; }

            // download
            PropertyAnimation{ target: go; properties: "progressDescription"; to: 2; duration: 1; }
            PropertyAnimation{ duration: 500; }
            PropertyAnimation{ target: go; properties: "progress"; to: 0.01; duration: 1; }
            PropertyAnimation{ duration: 500; }
            PropertyAnimation{ target: go; properties: "progress"; to: 0.1; duration: 1; }
            PropertyAnimation{ duration: 500; }
            PropertyAnimation{ target: go; properties: "progress"; to: 0.3; duration: 1; }
            PropertyAnimation{ duration: 500; }
            PropertyAnimation{ target: go; properties: "progress"; to: 0.5; duration: 1; }
            PropertyAnimation{ duration: 500; }
            PropertyAnimation{ target: go; properties: "progress"; to: 0.8; duration: 1; }
            PropertyAnimation{ duration: 500; }
            PropertyAnimation{ target: go; properties: "progress"; to: 1.0; duration: 1; }
            PropertyAnimation{ duration: 1000; }

            // verify
            PropertyAnimation{ target: go; properties: "progress"; to: 0.0; duration: 1; }
            PropertyAnimation{ target: go; properties: "progressDescription"; to: 3; duration: 1; }
            PropertyAnimation{ duration: 2000; }

            // unzip
            PropertyAnimation{ target: go; properties: "progressDescription"; to: 4; duration: 1; }
            PropertyAnimation{ duration: 500; }
            PropertyAnimation{ target: go; properties: "progress"; to: 0.01; duration: 1; }
            PropertyAnimation{ duration: 500; }
            PropertyAnimation{ target: go; properties: "progress"; to: 0.1; duration: 1; }
            PropertyAnimation{ duration: 500; }
            PropertyAnimation{ target: go; properties: "progress"; to: 0.3; duration: 1; }
            PropertyAnimation{ duration: 500; }
            PropertyAnimation{ target: go; properties: "progress"; to: 0.5; duration: 1; }
            PropertyAnimation{ duration: 500; }
            PropertyAnimation{ target: go; properties: "progress"; to: 0.8; duration: 1; }
            PropertyAnimation{ duration: 500; }
            PropertyAnimation{ target: go; properties: "progress"; to: 1.0; duration: 1; }
            PropertyAnimation{ duration: 2000; }

            // update
            PropertyAnimation{ target: go; properties: "progress"; to: 0.0; duration: 1; }
            PropertyAnimation{ target: go; properties: "progressDescription"; to: 5; duration: 1; }
            PropertyAnimation{ duration: 2000; }

            // quit
            PropertyAnimation{ target: go; properties: "progressDescription"; to: 6; duration: 1; }
            PropertyAnimation{ duration: 2000; }

        }

        property Timer timer : Timer {
            id: timer
            interval : 700
            repeat : false
            property string work
            onTriggered : {
                console.log("triggered "+timer.work)
                switch (timer.work) {
                    case "wait":
                    break
                    default:
                    go.processFinished()
                }
            }
        }
        function workAndClose() {
            timer.work="default"
            timer.start()
        }

        function loadAccounts() {
            console.log("Test: Account loaded")
        }


        function openDownloadLink(){
        }

        function switchAddressMode(username){
            for (var iAcc=0; iAcc < accountsModel.count; iAcc++) {
                if (accountsModel.get(iAcc).account == username ) {
                    accountsModel.get(iAcc).isCombinedAddressMode = !accountsModel.get(iAcc).isCombinedAddressMode
                    break
                }
            }
            workAndClose()
        }

        function getLocalVersionInfo(){
            go.updateVersion = "QA.1.0"
        }

        function getBackendVersion() {
            return "BridgeUI 1.0"
        }

        property bool isConnectionOK : true
        signal setConnectionStatus(bool isAvailable)

        function configureAppleMail(iAccount,iAddress) {
            console.log ("Test: autoconfig account ",iAccount," address ",iAddress)
        }

        function openLogs() {
            Qt.openUrlExternally("file:///home/dev/")
        }

        function highlightSystray() {
            test_systray.highlight()
        }

        function errorSystray() {
            test_systray.error()
        }

        function normalSystray() {
            test_systray.normal()
        }

        signal bubbleClosed()

        function getIMAPPort() {
            return 1143
        }
        function getSMTPPort() {
            return 1025
        }

        function isPortOpen(portstring){
            if (isNaN(portstring)) {
                return 1
            }
            var portnum = parseInt(portstring,10)
            if (portnum < 3333) {
                return 1
            }
            return 0
        }

        function setPortsAndSecurity(portIMAP, portSMTP, secSMTP) {
            console.log("Test: ports changed", portIMAP, portSMTP, secSMTP)
        }

        function isSMTPSTARTTLS() {
            return true
        }

        signal openManual()

        function clearCache() {
            workAndClose()
        }

        function clearKeychain() {
            workAndClose()
        }

        property bool isReportingOutgoingNoEnc : true

        function toggleIsReportingOutgoingNoEnc() {
            go.isReportingOutgoingNoEnc = !go.isReportingOutgoingNoEnc
            console.log("Reporting changed to ", go.isReportingOutgoingNoEnc)
        }

        function saveOutgoingNoEncPopupCoord(x,y) {
            console.log("Triggered saveOutgoingNoEncPopupCoord: ",x,y)
        }

        function shouldSendAnswer (messageID, shouldSend) {
            if (shouldSend) console.log("answered to send email")
            else console.log("answered to cancel email")
        }

        onToggleAutoStart: {
            workAndClose()
            isAutoStart = (isAutoStart!=false) ? false : true
            console.log (" Test: toggleAutoStart "+isAutoStart)
        }

        onToggleAutoUpdate: {
            workAndClose()
            isAutoUpdate = (isAutoUpdate!=false) ? false : true
            console.log (" Test: onToggleAutoUpdate "+isAutoUpdate)
        }

        onToggleEarlyAccess: {
            workAndClose()
            isEarlyAccess = (isEarlyAccess!=false) ? false : true
            console.log (" Test: onToggleEarlyAccess "+isEarlyAccess)
        }
    }
}

