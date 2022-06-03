// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

import QtQml 2.12
import QtQuick 2.13
import QtQuick.Window 2.13
import QtQuick.Layouts 1.12
import QtQuick.Controls 2.13

import QtQml.Models 2.12

import Qt.labs.platform 1.1

import Proton 4.0

import "./BridgeTest"
import BridgePreview 1.0

import Notifications 1.0

Window {
    id: root

    x: 10
    y: 10
    width: 800
    height: 800

    property ColorScheme colorScheme: ProtonStyle.darkStyle

    flags   : Qt.Window | Qt.Dialog
    visible : true
    title   : "Bridge Test GUI"

    // This is needed because on MacOS if first window shown is not transparent -
    // all other windows of application will not have transparent background (black
    // instead of transparency). In our case that mean that if BridgeTest will be
    // shown before StatusWindow - StatusWindow will not have transparent corners.
    color: "transparent"

    function getCursorPos() {
        return BridgePreview.getCursorPos()
    }

    function restart() {
        root.quit()
        console.log("Restarting....")
        root.openBridge()
    }

    function openBridge() {
        bridge = bridgeComponent.createObject()
        var showSetupGuide = false
        if (showSetupGuide) {
            var newUserObject = root.userComponent.createObject(root)
            newUserObject.username = "LerooooyJenkins@protonmail.com"
            newUserObject.loggedIn = true
            newUserObject.setupGuideSeen = false
            root.users.append( { object: newUserObject } )
        }
    }


    function quit() {
        if (bridge !== undefined && bridge !== null) {
            bridge.destroy()
        }
    }

    function guiReady() {
        console.log("Gui Ready")
    }

    function _log(msg, color) {
        logTextArea.text += "<p style='color: " + color + ";'>" + msg + "</p>"
        logTextArea.text += "\n"
    }

    function log(msg) {
        console.log(msg)
        _log(msg, root.colorScheme.signal_info)
    }

    function error(msg) {
        console.error(msg)
        _log(msg, root.colorScheme.signal_danger)
    }

    // No user object should be put in this list until a successful login
    property var users: UserModel {
        id: _users

        onRowsInserted: {
            for (var i = first; i <= last; i++) {
                _usersTest.insert(i + 1, { object: get(i) } )
            }
        }

        onRowsRemoved: {
            _usersTest.remove(first + 1, first - last + 1)
        }

        onRowsMoved: {
            _usersTest.move(start + 1, row + 1, end - start + 1)
        }

        onDataChanged: {
            for (var i = topLeft.row; i <= bottomRight.row; i++) {
                _usersTest.set(i + 1, { object: get(i) } )
            }
        }
    }

    // this list is used on test gui: it contains same users list as users above + fake user to represent login request of new user on pos 0
    property var usersTest: UserModel {
        id: _usersTest
    }

    property var userComponent: Component {
        id: _userComponent

        QtObject {
            property string username: ""
            property bool loggedIn: false
            property bool splitMode: false

            property bool setupGuideSeen: true

            property var usedBytes: 5350*1024*1024
            property var totalBytes: 20*1024*1024*1024
            property string avatarText: "jd"

            property string password: "SMj975NnEYYsqu55GGmlpv"
            property var addresses: [
                "jaanedoe@protonmail.com",
                "jane@pm.me",
                "jdoe@pm.me"
            ]

            signal loginUsernamePasswordError()
            signal loginFreeUserError()
            signal loginConnectionError()
            signal login2FARequested()
            signal login2FAError()
            signal login2FAErrorAbort()
            signal login2PasswordRequested()
            signal login2PasswordError()
            signal login2PasswordErrorAbort()

            // Test purpose only:
            property bool isFakeUser: this === root.loginUser

            function userSignal(msg) {
                if (isFakeUser) {
                    return
                }

                root.log("<- User (" + username + "): " + msg)
            }

            function toggleSplitMode(makeActive) {
                userSignal("toggle split mode "+makeActive)
            }
            signal toggleSplitModeFinished()

            function configureAppleMail(address){
                userSignal("confugure apple mail "+address)
            }

            function logout(){
                userSignal("logout")
                loggedIn = false
            }
            function remove(){
                console.log("remove this", users.count)
                for (var i=0; i<users.count; i++) {
                    if (users.get(i) === this) {
                        users.remove(i,1)
                        return
                    }
                }
            }


            onLoginUsernamePasswordError: {
                userSignal("loginUsernamePasswordError")
            }
            onLoginFreeUserError: {
                userSignal("loginFreeUserError")
            }
            onLoginConnectionError: {
                userSignal("loginConnectionError")
            }
            onLogin2FARequested: {
                userSignal("login2FARequested")
            }
            onLogin2FAError: {
                userSignal("login2FAError")
            }
            onLogin2FAErrorAbort: {
                userSignal("login2FAErrorAbort")
            }
            onLogin2PasswordRequested: {
                userSignal("login2PasswordRequested")
            }
            onLogin2PasswordError: {
                userSignal("login2PasswordError")
            }
            onLogin2PasswordErrorAbort: {
                userSignal("login2PasswordErrorAbort")
            }

            function resetLoginRequests() {
                isLoginRequested = false
                isLogin2FARequested = false
                isLogin2FAProvided = false
                isLogin2PasswordRequested = false
                isLogin2PasswordProvided = false
            }

            property bool isLoginRequested: false

            property bool isLogin2FARequested: false
            property bool isLogin2FAProvided: false

            property bool isLogin2PasswordRequested: false
            property bool isLogin2PasswordProvided: false
        }
    }

    // this it fake user used only for representing first login request
    property var loginUser
    Component.onCompleted: {
        var newLoginUser = _userComponent.createObject()
        root.loginUser = newLoginUser
        root.loginUser.setupGuideSeen = false
        _usersTest.append({object: newLoginUser})

        newLoginUser.loginUsernamePasswordError.connect(root.loginUsernamePasswordError)
        newLoginUser.loginFreeUserError.connect(root.loginFreeUserError)
        newLoginUser.loginConnectionError.connect(root.loginConnectionError)
        newLoginUser.login2FARequested.connect(root.login2FARequested)
        newLoginUser.login2FAError.connect(root.login2FAError)
        newLoginUser.login2FAErrorAbort.connect(root.login2FAErrorAbort)
        newLoginUser.login2PasswordRequested.connect(root.login2PasswordRequested)
        newLoginUser.login2PasswordError.connect(root.login2PasswordError)
        newLoginUser.login2PasswordErrorAbort.connect(root.login2PasswordErrorAbort)


        // add one user on start
        var hasUserOnStart = false
        if (hasUserOnStart) {
            var newUserObject = root.userComponent.createObject(root)
            newUserObject.username = "LerooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooyJenkins@protonmail.com"
            newUserObject.loggedIn = true
            newUserObject.setupGuideSeen = true
            root.users.append( { object: newUserObject } )
        }
    }


    TabBar {
        id: tabBar
        anchors.left: parent.left
        anchors.right: parent.right

        TabButton {
            text: "Global settings"
        }

        TabButton {
            text: "User control"
        }

        TabButton {
            text: "Notifications"
        }

        TabButton {
            text: "Log"
        }

        TabButton {
            text: "Settings signals"
        }
    }

    Rectangle {
        color: root.colorScheme.background_norm

        anchors.top: tabBar.bottom
        anchors.left: parent.left
        anchors.right: parent.right
        anchors.bottom: parent.bottom

        implicitHeight: children[0].contentHeight + children[0].anchors.topMargin + children[0].anchors.bottomMargin
        implicitWidth: children[0].contentWidth + children[0].anchors.leftMargin + children[0].anchors.rightMargin

        StackLayout {
            anchors.fill: parent
            currentIndex: tabBar.currentIndex
            anchors.margins: 10

            RowLayout {
                id: globalTab
                spacing : 5

                ColumnLayout {
                    spacing : 5

                    Label {
                        colorScheme: root.colorScheme
                        text: "Global settings"
                    }

                    ButtonGroup {
                        id: styleRadioGroup
                    }

                    RadioButton {
                        colorScheme: root.colorScheme
                        Layout.fillWidth: true

                        text: "Light UI"
                        checked: ProtonStyle.currentStyle === ProtonStyle.lightStyle
                        ButtonGroup.group: styleRadioGroup

                        onCheckedChanged: {
                            if (checked && ProtonStyle.currentStyle !== ProtonStyle.lightStyle) {
                                root.colorSchemeName = "light"
                            }
                        }
                    }

                    RadioButton {
                        colorScheme: root.colorScheme
                        Layout.fillWidth: true

                        text: "Dark UI"
                        checked: ProtonStyle.currentStyle === ProtonStyle.darkStyle
                        ButtonGroup.group: styleRadioGroup

                        onCheckedChanged: {
                            if (checked && ProtonStyle.currentStyle !== ProtonStyle.darkStyle) {
                                root.colorSchemeName = "dark"
                            }
                        }
                    }

                    CheckBox {
                        id: showOnStartupCheckbox
                        colorScheme: root.colorScheme
                        text: "Show on startup"
                        checked: root.showOnStartup
                        onCheckedChanged: {
                            root.showOnStartup = checked
                        }
                    }

                    CheckBox {
                        id: showSplashScreen
                        colorScheme: root.colorScheme
                        text: "Show splash screen"
                        checked: root.showSplashScreen
                        onCheckedChanged: {
                            root.showSplashScreen = checked
                        }
                    }

                    Button {
                        colorScheme: root.colorScheme
                        //Layout.fillWidth: true

                        text: "Open Bridge"
                        enabled: bridge === undefined || bridge === null
                        onClicked: root.openBridge()
                    }

                    Button {
                        colorScheme: root.colorScheme
                        //Layout.fillWidth: true

                        text: "Close Bridge"
                        enabled: bridge !== undefined && bridge !== null
                        onClicked: {
                            bridge.destroy()
                        }
                    }

                    Item {
                        Layout.fillHeight: true
                    }
                }

                ColumnLayout {
                    spacing : 5

                    Label {
                        colorScheme: root.colorScheme
                        text: "Notifications"
                    }

                    Button {
                        colorScheme: root.colorScheme
                        text: "Notify: danger"
                        enabled: bridge !== undefined && bridge !== null
                        onClicked: {
                            bridge.mainWindow.notifyOnlyPaidUsers()
                        }
                    }

                    Button {
                        colorScheme: root.colorScheme
                        text: "Notify: warning"
                        enabled: bridge !== undefined && bridge !== null
                        onClicked: {
                            bridge.mainWindow.notifyUpdateManually()
                        }
                    }

                    Button {
                        colorScheme: root.colorScheme
                        text: "Notify: success"
                        enabled: bridge !== undefined && bridge !== null
                        onClicked: {
                            bridge.mainWindow.notifyUserAdded()
                        }
                    }

                    Item {
                        Layout.fillHeight: true
                    }
                }
            }

            RowLayout {
                id: usersTab
                UserList {
                    id: usersListView
                    Layout.fillHeight: true
                    colorScheme: root.colorScheme
                    backend: root
                }

                UserControl {
                    colorScheme: root.colorScheme
                    backend: root
                    user: ((root.usersTest.count > usersListView.currentIndex) && usersListView.currentIndex != -1) ? root.usersTest.get(usersListView.currentIndex) : undefined
                    userIndex: usersListView.currentIndex - 1 // -1 because 0 index is fake user
                }
            }

            RowLayout {
                id: notificationsTab
                spacing: 5

                ColumnLayout {
                    spacing: 5

                    Switch {
                        text: "Internet connection"
                        colorScheme: root.colorScheme
                        checked: true
                        onCheckedChanged: {
                            checked ? root.internetOn() : root.internetOff()
                        }
                    }

                    Button {
                        text: "Update manual ready"
                        colorScheme: root.colorScheme
                        onClicked: {
                            root.updateManualReady("3.14.1592")
                        }
                    }

                    Button {
                        text: "Update manual done"
                        colorScheme: root.colorScheme
                        onClicked: {
                            root.updateManualRestartNeeded()
                        }
                    }

                    Button {
                        text: "Update manual error"
                        colorScheme: root.colorScheme
                        onClicked: {
                            root.updateManualError()
                        }
                    }

                    Button {
                        text: "Update force"
                        colorScheme: root.colorScheme
                        onClicked: {
                            root.updateForce("3.14.1592")
                        }
                    }

                    Button {
                        text: "Update force error"
                        colorScheme: root.colorScheme
                        onClicked: {
                            root.updateForceError()
                        }
                    }

                    Button {
                        text: "Update silent done"
                        colorScheme: root.colorScheme
                        onClicked: {
                            root.updateSilentRestartNeeded()
                        }
                    }

                    Button {
                        text: "Update silent error"
                        colorScheme: root.colorScheme
                        onClicked: {
                            root.updateSilentError()
                        }
                    }

                    Button {
                        text: "Update is latest version"
                        colorScheme: root.colorScheme
                        onClicked: {
                            root.updateIsLatestVersion()
                        }
                    }

                    Button {
                        text: "Bug report send OK"
                        colorScheme: root.colorScheme
                        onClicked: {
                            root.reportBugFinished()
                            root.bugReportSendSuccess()
                        }

                    }
                }

                ColumnLayout {
                    spacing: 5

                    Button {
                        text: "Bug report send error"
                        colorScheme: root.colorScheme
                        onClicked: {
                            root.reportBugFinished()
                            root.bugReportSendError()
                        }
                    }

                    Button {
                        text: "Cache anavailable"
                        colorScheme: root.colorScheme
                        onClicked: {
                            root.cacheUnavailable()
                        }
                    }

                    Button {
                        text: "Cache can't move"
                        colorScheme: root.colorScheme
                        onClicked: {
                            root.cacheCantMove()
                        }
                    }

                    Button {
                        text: "Cache location change success"
                        onClicked: {
                            root.cacheLocationChangeSuccess()
                        }
                        colorScheme: root.colorScheme
                    }

                    Button {
                        text: "Disk full"
                        colorScheme: root.colorScheme
                        onClicked: {
                            root.diskFull()
                        }
                    }

                    Button {
                        text: "No keychain"
                        colorScheme: root.colorScheme
                        onClicked: {
                            root.notifyHasNoKeychain()
                        }
                    }

                    Button {
                        text: "Rebuild keychain"
                        colorScheme: root.colorScheme
                        onClicked: {
                            root.notifyRebuildKeychain()
                        }
                    }

                    Button {
                        text: "Address changed"
                        colorScheme: root.colorScheme
                        onClicked: {
                            root.addressChanged("p@v.el")
                        }
                    }

                    Button {
                        text: "Address changed + Logout"
                        colorScheme: root.colorScheme
                        onClicked: {
                            root.addressChangedLogout("p@v.el")
                        }
                    }
                }
            }

            TextArea {
                id: logTextArea
                colorScheme: root.colorScheme
                Layout.fillHeight: true
                Layout.fillWidth: true

                Layout.preferredWidth: 400
                Layout.preferredHeight: 200

                textFormat: TextEdit.RichText
                //readOnly: true
            }

            ScrollView {
                id: settingsTab
                ColumnLayout {
                    RowLayout {
                        Label {colorScheme  : root.colorScheme ; text : "GOOS     : "}
                        Button {colorScheme : root.colorScheme ; text : "Linux"   ; onClicked : root.goos = "linux"   ; enabled: root.goos != "linux"}
                        Button {colorScheme : root.colorScheme ; text : "Windows" ; onClicked : root.goos = "windows" ; enabled: root.goos != "windows"}
                        Button {colorScheme : root.colorScheme ; text : "macOS"   ; onClicked : root.goos = "darwin"  ; enabled: root.goos != "darwin"}
                    }
                    RowLayout {
                        Label {colorScheme: root.colorScheme; text: "Automatic updates:"}
                        Toggle {colorScheme: root.colorScheme; checked: root.isAutomaticUpdateOn; onClicked: root.isAutomaticUpdateOn = !root.isAutomaticUpdateOn}
                    }
                    RowLayout {
                        Label {colorScheme: root.colorScheme; text: "Autostart:"}
                        Toggle {colorScheme: root.colorScheme; checked: root.isAutostartOn; onClicked: root.isAutostartOn = !root.isAutostartOn}
                        Button {colorScheme: root.colorScheme; text: "Toggle finished"; onClicked: root.toggleAutostartFinished()}
                    }
                    RowLayout {
                        Label {colorScheme: root.colorScheme; text: "Beta:"}
                        Toggle {colorScheme: root.colorScheme; checked: root.isBetaEnabled; onClicked: root.isBetaEnabled = !root.isBetaEnabled}
                    }
                    RowLayout {
                        Label {colorScheme: root.colorScheme; text: "DoH:"}
                        Toggle {colorScheme: root.colorScheme; checked: root.isDoHEnabled; onClicked: root.isDoHEnabled = !root.isDoHEnabled}
                    }
                    RowLayout {
                        Label {colorScheme: root.colorScheme; text: "Ports:"}
                        TextField {
                            colorScheme:root.colorScheme
                            label: "IMAP"
                            text: root.portIMAP
                            onEditingFinished: root.portIMAP = this.text*1
                            validator: IntValidator {bottom: 1; top: 65536}
                        }
                        TextField {
                            colorScheme:root.colorScheme
                            label: "SMTP"
                            text: root.portSMTP
                            onEditingFinished: root.portSMTP = this.text*1
                            validator: IntValidator {bottom: 1; top: 65536}
                        }
                        Button {colorScheme: root.colorScheme; text: "Change finished"; onClicked: root.changePortFinished()}
                    }
                    RowLayout {
                        Label {colorScheme: root.colorScheme; text: "SMTP using SSL:"}
                        Toggle {colorScheme: root.colorScheme; checked: root.useSSLforSMTP; onClicked: root.useSSLforSMTP = !root.useSSLforSMTP}
                    }
                    RowLayout {
                        Label {colorScheme: root.colorScheme; text: "Local cache:"}
                        Toggle {colorScheme: root.colorScheme; checked: root.isDiskCacheEnabled; onClicked: root.isDiskCacheEnabled = !root.isDiskCacheEnabled}
                        TextField {
                            colorScheme:root.colorScheme
                            label: "Path"
                            text: root.diskCachePath.toString().replace("file://", "")
                            implicitWidth: 160
                            onEditingFinished: {
                                root.diskCachePath = Qt.resolvedUrl("file://"+text)
                            }
                        }
                        Button {colorScheme: root.colorScheme; text: "Change finished"; onClicked: root.changeLocalCacheFinished()}
                    }
                    RowLayout {
                        Label {colorScheme: root.colorScheme; text: "Reset:"}
                        Button {colorScheme: root.colorScheme; text: "Finished"; onClicked: root.resetFinished()}
                    }
                    RowLayout {
                        Label {colorScheme: root.colorScheme; text: "Check update:"}
                        Button {colorScheme: root.colorScheme; text: "Finished"; onClicked: root.checkUpdatesFinished()}
                    }
                }
            }
        }
    }

    property Bridge bridge

    property string goos: "darwin"

    property bool showOnStartup: true // this actually needs to be false, but since we use Bridge_test for testing purpose - lets default this to true just for convenience
    property bool dockIconVisible: false

    // this signals are used only when trying to login with new user (i.e. not in users model)
    signal loginUsernamePasswordError(string errorMsg)
    signal loginFreeUserError()
    signal loginConnectionError(string errorMsg)
    signal login2FARequested(string username)
    signal login2FAError(string errorMsg)
    signal login2FAErrorAbort(string errorMsg)
    signal login2PasswordRequested()
    signal login2PasswordError(string errorMsg)
    signal login2PasswordErrorAbort(string errorMsg)
    signal loginFinished(int index)
    signal loginAlreadyLoggedIn(int index)

    signal internetOff()
    signal internetOn()

    signal updateManualReady(var version)
    signal updateManualRestartNeeded()
    signal updateManualError()
    signal updateForce(var version)
    signal updateForceError()
    signal updateSilentRestartNeeded()
    signal updateSilentError()
    signal updateIsLatestVersion()
    function checkUpdates(){
        console.log("check updates")
    }
    signal checkUpdatesFinished()
    function installUpdate() {
        console.log("manuall install update triggered")
    }


    property bool   isDiskCacheEnabled: true
    // Qt.resolvedUrl("file:///C:/Users/user/AppData/Roaming/protonmail/bridge/cache/c11/messages")
    property url diskCachePath: StandardPaths.standardLocations(StandardPaths.HomeLocation)[0]
    signal cacheUnavailable()
    signal cacheCantMove()
    signal cacheLocationChangeSuccess()
    signal diskFull()
    function changeLocalCache(enableDiskCache, diskCachePath) {
        console.debug("-> disk cache", enableDiskCache, diskCachePath)
    }
    signal changeLocalCacheFinished()


    // Settings
    property bool isAutomaticUpdateOn : true
    function toggleAutomaticUpdate(makeItActive) {
        console.debug("-> silent updates", makeItActive, root.isAutomaticUpdateOn)
        var callback = function () {
            root.isAutomaticUpdateOn = makeItActive;
            console.debug("-> CHANGED silent updates", makeItActive, root.isAutomaticUpdateOn)
        }
        atimer.onTriggered.connect(callback)
        atimer.restart()
    }

    Timer {
        id: atimer
        interval: 2000
        running: false
        repeat: false
    }

    property bool isAutostartOn : true // Example of settings with loading state
    function toggleAutostart(makeItActive) {
        console.debug("-> autostart", makeItActive, root.isAutostartOn)
    }
    signal toggleAutostartFinished()

    property bool isBetaEnabled : false
    function toggleBeta(makeItActive){
        console.debug("-> beta", makeItActive, root.isBetaEnabled)
        root.isBetaEnabled = makeItActive
    }

    property bool isDoHEnabled : true
    function toggleDoH(makeItActive){
        console.debug("-> DoH", makeItActive, root.isDoHEnabled)
        root.isDoHEnabled = makeItActive
    }

    property bool useSSLforSMTP: false
    function toggleUseSSLforSMTP(makeItActive){
        console.debug("-> SMTP SSL", makeItActive, root.useSSLforSMTP)
    }
    signal toggleUseSSLFinished()

    property string hostname: "127.0.0.1"
    property int portIMAP: 1143
    property int portSMTP: 1025
    function changePorts(imapPort, smtpPort){
        console.debug("-> ports", imapPort, smtpPort)
    }
    function isPortFree(port){
        if (port == portIMAP) return false
        if (port == portSMTP) return false
        if (port == 12345) return false
        return true
    }
    signal changePortFinished()
    signal portIssueIMAP()
    signal portIssueSMTP()

    function triggerReset() {
        console.debug("-> trigger reset")
    }
    signal resetFinished()

    property string version: "2.0.X-BridePreview"
    property url logsPath: StandardPaths.standardLocations(StandardPaths.HomeLocation)[0]
    property url licensePath: StandardPaths.standardLocations(StandardPaths.HomeLocation)[0]
    property url releaseNotesLink: Qt.resolvedUrl("https://protonmail.com/download/bridge/early_releases.html")
    property url dependencyLicensesLink: Qt.resolvedUrl("https://github.com/ProtonMail/proton-bridge/v2/blob/master/COPYING_NOTES.md#dependencies")
    property url landingPageLink: Qt.resolvedUrl("https://protonmail.com/bridge")

    property string colorSchemeName: "light"
    function changeColorScheme(newScheme){
        root.colorSchemeName = newScheme
    }


    property string currentEmailClient: "" // "Apple Mail 14.0"
    function updateCurrentMailClient(){
        currentEmailClient = "Apple Mail 14.0"
    }

    function reportBug(description,address,emailClient,includeLogs){
        console.log("report bug")
        console.log("  description",description)
        console.log("  address",address)
        console.log("  emailClient",emailClient)
        console.log("  includeLogs",includeLogs)
    }
    signal reportBugFinished()
    signal bugReportSendSuccess()
    signal bugReportSendError()

    property var availableKeychain: ["gnome-keyring", "pass", "macos-keychain", "windows-credentials"]
    property string currentKeychain: availableKeychain[0]
    function changeKeychain(wantedKeychain){
        console.log("Changing keychain from", root.currentKeychain, "to", wantedKeychain)
        root.currentKeychain = wantedKeychain
        root.changeKeychainFinished()
    }
    signal changeKeychainFinished()
    signal notifyHasNoKeychain()
    signal notifyRebuildKeychain()

    signal noActiveKeyForRecipient(string email)
    signal showMainWindow()

    signal addressChanged(string address)
    signal addressChangedLogout(string address)
    signal userDisconnected(string username)
    signal apiCertIssue()

    property bool showSplashScreen: false


    function login(username, password) {
        root.log("-> login(" + username + ", " + password + ")")

        loginUser.username = username
        loginUser.isLoginRequested = true
    }

    function login2FA(username, code) {
        root.log("-> login2FA(" + username + ", " + code + ")")

        loginUser.isLogin2FAProvided = true
    }

    function login2Password(username, password) {
        root.log("-> login2FA(" + username + ", " + password + ")")

        loginUser.isLogin2PasswordProvided = true
    }

    function loginAbort(username) {
        root.log("-> loginAbort(" + username + ")")

        loginUser.resetLoginRequests()
    }


    onLoginUsernamePasswordError: {
        console.debug("<- loginUsernamePasswordError")
    }
    onLoginFreeUserError: {
        console.debug("<- loginFreeUserError")
    }
    onLoginConnectionError: {
        console.debug("<- loginConnectionError")
    }
    onLogin2FARequested: {
        console.debug("<- login2FARequested", username)
    }
    onLogin2FAError: {
        console.debug("<- login2FAError")
    }
    onLogin2FAErrorAbort: {
        console.debug("<- login2FAErrorAbort")
    }
    onLogin2PasswordRequested: {
        console.debug("<- login2PasswordRequested")
    }
    onLogin2PasswordError: {
        console.debug("<- login2PasswordError")
    }
    onLogin2PasswordErrorAbort: {
        console.debug("<- login2PasswordErrorAbort")
    }
    onLoginFinished: {
        console.debug("<- loginFinished", index)
    }
    onLoginAlreadyLoggedIn: {
        console.debug("<- loginAlreadyLoggedIn", index)
    }

    onInternetOff: {
        console.debug("<- internetOff")
    }
    onInternetOn: {
        console.debug("<- internetOn")
    }

    Component {
        id: bridgeComponent

        Bridge {
            backend: root

        }
    }

    onClosing: {
        Qt.quit()
    }
}
