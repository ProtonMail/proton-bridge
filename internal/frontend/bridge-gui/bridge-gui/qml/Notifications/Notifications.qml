// Copyright (c) 2023 Proton AG
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

import QtQml
import Qt.labs.platform
import QtQuick.Controls
import ".."

QtObject {
    id: root

    property MainWindow frontendMain

    signal askEnableBeta()
    signal askEnableSplitMode(var user)
    signal askResetBridge()
    signal askChangeAllMailVisibility(var isVisibleNow)
    signal askDeleteAccount(var user)
    signal askQuestion(var title, var description, var option1, var option2, var action1, var action2)
    enum Group {
        Connection    = 1,
        Update        = 2,
        Configuration = 4,
        ForceUpdate   = 8,
        API           = 32,

        // Special group for notifications that require dialog popup instead of banner
        Dialogs = 64
    }

    property var all: [
        root.noInternet,
        root.imapPortStartupError,
        root.smtpPortStartupError,
        root.imapPortChangeError,
        root.smtpPortChangeError,
        root.imapConnectionModeChangeError,
        root.smtpConnectionModeChangeError,
        root.updateManualReady,
        root.updateManualRestartNeeded,
        root.updateManualError,
        root.updateForce,
        root.updateForceError,
        root.updateSilentRestartNeeded,
        root.updateSilentError,
        root.updateIsLatestVersion,
        root.loginConnectionError,
        root.onlyPaidUsers,
        root.alreadyLoggedIn,
        root.enableBeta,
        root.bugReportSendSuccess,
        root.bugReportSendError,
        root.cacheUnavailable,
        root.cacheCantMove,
        root.accountChanged,
        root.diskFull,
        root.cacheLocationChangeSuccess,
        root.enableSplitMode,
        root.resetBridge,
        root.changeAllMailVisibility,
        root.deleteAccount,
        root.noKeychain,
        root.rebuildKeychain,
        root.addressChanged,
        root.apiCertIssue,
        root.noActiveKeyForRecipient,
        root.userBadEvent,
        root.imapLoginWhileSignedOut,
        root.genericError,
        root.genericQuestion,
    ]

    // Connection
    property Notification noInternet: Notification {
        description: qsTr("Bridge is not able to contact the server, please check your internet connection.")
        brief: qsTr("No connection")
        icon: "./icons/ic-no-connection.svg"
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Connection

        Connections {
            target: Backend

            function onInternetOff() {
                root.noInternet.active = true
            }
            function onInternetOn() {
                root.noInternet.active = false
            }
        }
    }

    property Notification imapPortStartupError: Notification {
        description: qsTr("The IMAP server could not be started. Please check or change the IMAP port.")
        brief: qsTr("IMAP port error")
        icon: "./icons/ic-alert.svg"
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Connection

        Connections {
            target: Backend

            function onImapPortStartupError() {
                root.imapPortStartupError.active = true
            }
        }
    }

    property Notification smtpPortStartupError: Notification {
        description: qsTr("The SMTP server could not be started. Please check or change the SMTP port.")
        brief: qsTr("SMTP port error")
        icon: "./icons/ic-alert.svg"
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Connection

        Connections {
            target: Backend

            function onSmtpPortStartupError() {
                root.smtpPortStartupError.active = true
            }
        }
    }

    property Notification imapPortChangeError: Notification {
        description: qsTr("The IMAP port could not be changed.")
        brief: qsTr("IMAP port error")
        icon: "./icons/ic-alert.svg"
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Connection

        Connections {
            target: Backend

            function onImapPortChangeError() {
                root.imapPortChangeError.active = true
            }
        }
    }

    property Notification smtpPortChangeError: Notification {
        description: qsTr("The SMTP port could not be changed.")
        brief: qsTr("SMTP port error")
        icon: "./icons/ic-alert.svg"
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Connection

        Connections {
            target: Backend

            function onSmtpPortChangeError() {
                root.smtpPortChangeError.active = true
            }
        }
    }

    property Notification imapConnectionModeChangeError: Notification {
        description: qsTr("The IMAP connection mode could not be changed.")
        brief: qsTr("IMAP Connection mode error")
        icon: "./icons/ic-alert.svg"
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Connection

        Connections {
            target: Backend

            function onImapConnectionModeChangeError() {
                root.imapConnectionModeChangeError.active = true
            }
        }

        action: Action {
            text: qsTr("OK")

            onTriggered: {
                root.imapConnectionModeChangeError.active= false
            }
        }
    }

    property Notification smtpConnectionModeChangeError: Notification {
        description: qsTr("The SMTP connection mode could not be changed.")
        brief: qsTr("SMTP Connection mode error")
        icon: "./icons/ic-alert.svg"
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Connection

        Connections {
            target: Backend

            function onSmtpConnectionModeChangeError() {
                root.smtpConnectionModeChangeError.active = true
            }
        }

        action: Action {
            text: qsTr("OK")

            onTriggered: {
                root.smtpConnectionModeChangeError.active= false
            }
        }
    }

    // Updates
    property Notification updateManualReady: Notification {
        title: qsTr("Update to Bridge %1").arg(data ? data.version : "")
        description:  {
            var descr = qsTr("A new version of Proton Mail Bridge is available.")
            var text = qsTr("See what's changed.")
            var link = Backend.releaseNotesLink
            return `${descr} <a href="${link}">${text}</a>`
        }
        brief: qsTr("Update available")
        icon: "./icons/ic-info-circle-filled.svg"
        type: Notification.NotificationType.Info
        group: Notifications.Group.Update | Notifications.Group.Dialogs

        Connections {
            target: Backend
            function onUpdateManualReady(version) {
                root.updateManualReady.data = { version: version }
                root.updateManualReady.active = true
            }
        }

        action: [
            Action {
                text: qsTr("Install update")

                onTriggered: {
                    Backend.installUpdate()
                    root.updateManualReady.active = false
                }
            },
            Action {
                text: qsTr("Update manually")

                onTriggered: {
                    Qt.openUrlExternally(Backend.landingPageLink)
                    root.updateManualReady.active = false
                }
            },
            Action {
                text: qsTr("Remind me later")

                onTriggered: {
                    root.updateManualReady.active = false
                }
            }
        ]
    }

    property Notification updateManualRestartNeeded: Notification {
        description: qsTr("Bridge update is ready")
        brief: description
        icon: "./icons/ic-info-circle-filled.svg"
        type: Notification.NotificationType.Info
        group: Notifications.Group.Update

        Connections {
            target: Backend
            function onUpdateManualRestartNeeded() {
                root.updateManualRestartNeeded.active = true
            }
        }

        action: Action {
            text: qsTr("Restart Bridge")

            onTriggered: {
                Backend.restart()
                root.updateManualRestartNeeded.active = false
            }
        }
    }

    property Notification updateManualError: Notification {
        title: qsTr("Bridge couldn’t update")
        brief: title
        description: qsTr("Please follow manual installation in order to update Bridge.")
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Warning
        group: Notifications.Group.Update

        Connections {
            target: Backend
            function onUpdateManualError() {
                root.updateManualError.active = true
            }
        }

        action: [
            Action {
                text: qsTr("Update manually")

                onTriggered: {
                    Qt.openUrlExternally(Backend.landingPageLink)
                    root.updateManualError.active = false
                    Backend.quit()
                }
            },
            Action {
                text: qsTr("Remind me later")

                onTriggered: {
                    root.updateManualError.active = false
                }
            }
        ]
    }

    property Notification updateForce: Notification {
        title: qsTr("Update to Bridge %1").arg(data ? data.version : "")
        description: qsTr("This version of Bridge is no longer supported, please update.")
        brief: qsTr("Bridge is outdated")
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Update | Notifications.Group.ForceUpdate | Notifications.Group.Dialogs

        Connections {
            target: Backend

            function onUpdateForce(version) {
                root.updateForce.data = { version: version }
                root.updateForce.active = true
            }
        }

        action: [
            Action {
                text: qsTr("Install update")

                onTriggered: {
                    Backend.installUpdate()
                    root.updateForce.active = false
                }
            },
            Action {
                text: qsTr("Update manually")

                onTriggered: {
                    Qt.openUrlExternally(Backend.landingPageLink)
                    root.updateForce.active = false
                }
            },
            Action {
                text: qsTr("Quit Bridge")

                onTriggered: {
                    Backend.quit()
                    root.updateForce.active = false
                }
            }
        ]
    }

    property Notification updateForceError: Notification {
        title: qsTr("Bridge couldn't update")
        description: qsTr("You must update manually. Go to: https://proton.me/mail/bridge#download")
        brief: title
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Update | Notifications.Group.Dialogs

        Connections {
            target: Backend

            function onUpdateForceError() {
                root.updateForceError.active = true
            }
        }

        action: [
            Action {
                text: qsTr("Update manually")

                onTriggered: {
                    Qt.openUrlExternally(Backend.landingPageLink)
                    root.updateForceError.active = false
                }
            },
            Action {
                text: qsTr("Quit Bridge")

                onTriggered: {
                    Backend.quit()
                    root.updateForceError.active = false
                }
            }
        ]
    }

    property Notification updateSilentRestartNeeded: Notification {
        description: qsTr("Bridge update is ready")
        brief: description
        icon: "./icons/ic-info-circle-filled.svg"
        type: Notification.NotificationType.Info
        group: Notifications.Group.Update

        Connections {
            target: Backend
            function onUpdateSilentRestartNeeded() {
                root.updateSilentRestartNeeded.active = true
            }
        }

        action: Action {
            text: qsTr("Restart Bridge")

            onTriggered: {
                Backend.restart()
                root.updateSilentRestartNeeded.active = false
            }
        }
    }

    property Notification updateSilentError: Notification {
        description: qsTr("Bridge couldn't update")
        brief: description
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Warning
        group: Notifications.Group.Update

        Connections {
            target: Backend
            function onUpdateSilentError() {
                root.updateSilentError.active = true
            }
        }

        action: Action {
            text: qsTr("Update manually")

            onTriggered: {
                Qt.openUrlExternally(Backend.landingPageLink)
                root.updateSilentError.active = false
            }
        }
    }

    property Notification updateIsLatestVersion: Notification {
        description: qsTr("Bridge is up to date")
        brief: description
        icon: "./icons/ic-info-circle-filled.svg"
        type: Notification.NotificationType.Info
        group: Notifications.Group.Update

        Connections {
            target: Backend
            function onUpdateIsLatestVersion() {
                root.updateIsLatestVersion.active = true
            }
        }

        action: Action {
            text: qsTr("OK")

            onTriggered: {
                root.updateIsLatestVersion.active = false
            }
        }
    }

    property Notification enableBeta: Notification {
        title: qsTr("Enable Beta access")
        brief: title
        description: qsTr("Be the first to get new updates and use new features. Bridge will update to the latest beta version.")
        icon: "./icons/ic-info-circle-filled.svg"
        type: Notification.NotificationType.Info
        group: Notifications.Group.Update | Notifications.Group.Dialogs

        Connections {
            target: root
            function onAskEnableBeta() {
                root.enableBeta.active = true
            }
        }

        action: [
            Action {
                text: qsTr("Enable")
                onTriggered: {
                    Backend.toggleBeta(true)
                    root.enableBeta.active = false
                }
            },
            Action {
                text: qsTr("Cancel")

                onTriggered: {
                    root.enableBeta.active = false
                }
            }
        ]
    }

    // login
    property Notification loginConnectionError: Notification {
        description: qsTr("Bridge is not able to contact the server, please check your internet connection.")
        brief: qsTr("Connection error")
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Configuration

        Connections {
            target: Backend
            function onLoginConnectionError(errorMsg) {
                root.loginConnectionError.active = true
            }
        }

        action: [
            Action {
                text: qsTr("OK")
                onTriggered: {
                    root.loginConnectionError.active = false
                }
            }
        ]
    }

    property Notification onlyPaidUsers: Notification {
        description: qsTr("Bridge is exclusive to our mail paid plans. Upgrade your account to use Bridge.")
        brief: qsTr("Upgrade your account")
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Configuration
        property var pricingLink: "https://proton.me/mail/pricing"

        Connections {
            target: Backend
            function onLoginFreeUserError() {
                root.onlyPaidUsers.active = true
            }
        }

        action: [
            Action {
                text: qsTr("Upgrade")
                onTriggered: {
                    Qt.openUrlExternally(root.onlyPaidUsers.pricingLink)
                    root.onlyPaidUsers.active = false
                }
            }
        ]
    }

    property Notification alreadyLoggedIn: Notification {
        description: qsTr("This account is already signed in.")
        brief: qsTr("Already signed in")
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Info
        group: Notifications.Group.Configuration

        Connections {
            target: Backend
            function onLoginAlreadyLoggedIn(index) {
                root.alreadyLoggedIn.active = true
            }
        }

        action: [
            Action {
                text: qsTr("OK")
                onTriggered: {
                    root.alreadyLoggedIn.active = false
                }
            }
        ]
    }

    // Bug reports
    property Notification bugReportSendSuccess: Notification {
        description: qsTr("Thank you for the report. We'll get back to you as soon as we can.")
        brief: qsTr("Report sent")
        icon: "./icons/ic-info-circle-filled.svg"
        type: Notification.NotificationType.Success
        group: Notifications.Group.Configuration

        Connections {
            target: Backend
            function onBugReportSendSuccess() {
                root.bugReportSendSuccess.active = true
            }
        }

        action: [
            Action {
                text: qsTr("OK")
                onTriggered: {
                    root.bugReportSendSuccess.active = false
                }
            }
        ]
    }

    property Notification bugReportSendError: Notification {
        description: qsTr("Report could not be sent. Try again or email us directly.")
        brief: qsTr("Error sending report")
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Configuration

        Connections {
            target: Backend
            function onBugReportSendError() {
                root.bugReportSendError.active = true
            }
        }

        action: Action {
            text: qsTr("OK")
            onTriggered: {
                root.bugReportSendError.active = false
            }
        }
    }

    // Cache
    property Notification cacheUnavailable: Notification {
        title: qsTr("Cache location is unavailable")
        description: qsTr("The current cache location is unavailable. Check the directory or change it in your settings.")
        brief: title
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Warning
        group: Notifications.Group.Configuration | Notifications.Group.Dialogs

        Connections {
            target: Backend
            function onDiskCacheUnavailable() {
                root.cacheUnavailable.active = true
            }
        }

        action: [
            Action {
                text: qsTr("Quit Bridge")
                onTriggered: {
                    Backend.quit()
                    root.cacheUnavailable.active = false
                }
            },
            Action {
                text: qsTr("Change location")
                onTriggered: {
                    root.cacheUnavailable.active = false
                    root.frontendMain.showLocalCacheSettings()
                }
            }
        ]
    }

    property Notification cacheCantMove: Notification {
        title: qsTr("Can’t move cache")
        brief: title
        description: qsTr("The location you have selected is not available. Make sure you have enough free space or choose another location.")
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Warning
        group: Notifications.Group.Configuration | Notifications.Group.Dialogs

        Connections {
            target: Backend
            function onCantMoveDiskCache() {
                root.cacheCantMove.active = true
            }
        }

        action: [
            Action {
                text: qsTr("Cancel")
                onTriggered: {
                    root.cacheCantMove.active = false
                }
            },
            Action {
                text: qsTr("Change location")
                onTriggered: {
                    root.cacheCantMove.active = false
                    root.frontendMain.showLocalCacheSettings()
                }
            }
        ]
    }

    property Notification cacheLocationChangeSuccess: Notification {
        description: qsTr("Cache location successfully changed")
        brief: description
        icon: "./icons/ic-info-circle-filled.svg"
        type: Notification.NotificationType.Success
        group: Notifications.Group.Configuration

        Connections {
            target: Backend
            function onDiskCachePathChanged() {
                console.log("notify location changed successfully")
                root.cacheLocationChangeSuccess.active = true
            }
        }

        action: [
            Action {
                text: qsTr("OK")
                onTriggered: {
                    root.cacheLocationChangeSuccess.active = false
                }
            }
        ]
    }

    // Other
    property Notification accountChanged: Notification {
        description: qsTr("The address list for .... account has changed. You need to reconfigure your email client.")
        brief: qsTr("Address list changed")
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Configuration

        action: Action {
            text: qsTr("Reconfigure")

            onTriggered: {
                // TODO: open configuration window here
            }
        }
    }

    property Notification diskFull: Notification {
        title: qsTr("Your disk is almost full")
        description: qsTr("Quit Bridge and free disk space or disable the local cache (not recommended).")
        brief: title
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Warning
        group: Notifications.Group.Configuration | Notifications.Group.Dialogs

        Connections {
            target: Backend
            function onDiskFull() {
                root.diskFull.active = true
            }
        }

        action: [
            Action {
                text: qsTr("Quit Bridge")
                onTriggered: {
                    Backend.quit()
                    root.diskFull.active = false
                }
            },
            Action {
                text: qsTr("Settings")
                onTriggered: {
                    root.diskFull.active = false
                    root.frontendMain.showLocalCacheSettings()
                }
            }
        ]
    }

    property Notification enableSplitMode: Notification {
        title: qsTr("Enable split mode?")
        brief: title
        description: qsTr("Changing between split and combined address mode will require you to delete your account(s) from your email client and begin the setup process from scratch.")
        icon: "/qml/icons/ic-question-circle.svg"
        type: Notification.NotificationType.Warning
        group: Notifications.Group.Configuration | Notifications.Group.Dialogs

        property var user

        Connections {
            target: root
            function onAskEnableSplitMode(user) {
                root.enableSplitMode.user = user
                root.enableSplitMode.active = true
            }
        }

        Connections {
            target: (root && root.enableSplitMode && root.enableSplitMode.user ) ? root.enableSplitMode.user : null
            function onToggleSplitModeFinished() {
                root.enableSplitMode.active = false

                enableSplitMode_enable.loading = false
                enableSplitMode_cancel.enabled = true
            }
        }

        action: [
            Action {
                id: enableSplitMode_cancel
                text: qsTr("Cancel")
                onTriggered: {
                    root.enableSplitMode.active = false
                }
            },
            Action {
                id: enableSplitMode_enable
                text: qsTr("Enable split mode")
                onTriggered: {
                    enableSplitMode_enable.loading = true
                    enableSplitMode_cancel.enabled = false
                    root.enableSplitMode.user.toggleSplitMode(true)
                }
            }
        ]
    }

    property Notification resetBridge: Notification {
        title: qsTr("Reset Bridge?")
        brief: title
        icon: "./icons/ic-exclamation-circle-filled.svg"
        description: qsTr("This will clear your accounts, preferences, and cached data. You will need to reconfigure your email client. Bridge will automatically restart.")
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Configuration | Notifications.Group.Dialogs

        property var user

        Connections {
            target: root
            function onAskResetBridge() {
                root.resetBridge.active = true
            }
        }

        Connections {
            target: Backend
            function onResetFinished() {
                root.resetBridge.active = false

                resetBridge_reset.loading = false
                resetBridge_cancel.enabled = true
            }
        }

        action: [
            Action {
                id: resetBridge_cancel
                text: qsTr("Cancel")
                onTriggered: {
                    root.resetBridge.active = false
                }
            },
            Action {
                id: resetBridge_reset
                text: qsTr("Reset and restart")
                onTriggered: {
                    resetBridge_reset.loading = true
                    resetBridge_cancel.enabled = false
                    Backend.triggerReset()
                }
            }
        ]
    }

    property Notification changeAllMailVisibility: Notification {
        title: root.changeAllMailVisibility.isVisibleNow ?
        qsTr("Hide All Mail folder?") :
        qsTr("Show All Mail folder?")
        brief: title
        icon: "./icons/ic-info-circle-filled.svg"
        description: qsTr("Switching between showing and hiding the All Mail folder will require you to restart your client.")
        type: Notification.NotificationType.Info
        group: Notifications.Group.Configuration | Notifications.Group.Dialogs

        property var isVisibleNow

        Connections {
            target: root

            function onAskChangeAllMailVisibility(isVisibleNow) {
                root.changeAllMailVisibility.isVisibleNow = isVisibleNow
                root.changeAllMailVisibility.active = true
            }
        }

        action: [
            Action {
                id: allMail_change
                text: root.changeAllMailVisibility.isVisibleNow ? 
                qsTr("Hide All Mail folder") :
                qsTr("Show All Mail folder")
                onTriggered: {
                    Backend.changeIsAllMailVisible(!root.changeAllMailVisibility.isVisibleNow)
                    root.changeAllMailVisibility.active = false
                }
            },
            Action {
                id: allMail_cancel
                text: qsTr("Cancel")
                onTriggered: {
                    root.changeAllMailVisibility.active = false
                }
            }
        ]
    }

    property Notification deleteAccount: Notification {
        title: qsTr("Remove this account?")
        brief: title
        icon: "./icons/ic-exclamation-circle-filled.svg"
        description: qsTr("Are you sure you want to remove this account from Bridge and delete locally stored preferences and data?")
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Configuration | Notifications.Group.Dialogs

        property var user

        Connections {
            target: root
            function onAskDeleteAccount(user) {
                root.deleteAccount.user = user
                root.deleteAccount.active = true
            }
        }

        action: [
            Action {
                id: deleteAccount_cancel
                text: qsTr("Cancel")
                onTriggered: {
                    root.deleteAccount.active = false
                }
            },
            Action {
                id: deleteAccount_delete
                text: qsTr("Remove this account")
                onTriggered: {
                    root.deleteAccount.user.remove()
                    root.deleteAccount.active = false
                }
            }
        ]
    }

    property Notification noKeychain: Notification {
        title: qsTr("No keychain available")
        brief: title
        description: qsTr("Bridge is not able to detect a supported password manager (pass or secret-service). Please install and setup supported password manager and restart the application.")
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Dialogs | Notifications.Group.Configuration

        Connections {
            target: Backend

            function onNotifyHasNoKeychain() {
                root.noKeychain.active = true
            }
        }

        action: [
            Action {
                text: qsTr("Quit Bridge")

                onTriggered: {
                    Backend.quit()
                }
            },
            Action {
                text: qsTr("Restart Bridge")

                onTriggered: {
                    Backend.restart()
                }
            }
        ]
    }

    property Notification rebuildKeychain: Notification {
        title: qsTr("Your macOS keychain might be corrupted")
        brief: title
        description: qsTr("Bridge is not able to access your macOS keychain. Please consult the instructions on our support page.")
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Dialogs | Notifications.Group.Configuration

        property var supportLink: "https://proton.me/support/bridge"


        Connections {
            target: Backend

            function onNotifyRebuildKeychain() {
                console.log("notifications")
                root.rebuildKeychain.active = true
            }
        }

        action: [
            Action {
                text: qsTr("Open the support page")

                onTriggered: {
                    Qt.openUrlExternally(root.rebuildKeychain.supportLink)
                    Backend.quit()
                }
            }
        ]
    }

    property Notification addressChanged: Notification {
        title: qsTr("Address list changes")
        brief: title
        description: qsTr("The address list for your account has changed. You might need to reconfigure your email client.")
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Warning
        group: Notifications.Group.Configuration

        Connections {
            target: Backend

            function onAddressChanged(address) {
                root.addressChanged.description = qsTr("The address list for your account %1 has changed. You might need to reconfigure your email client.").arg(address)
                root.addressChanged.active = true
            }

            function onAddressChangedLogout(address) {
                root.addressChanged.description = qsTr("The address list for your account %1 has changed. You have to reconfigure your email client.").arg(address)
                root.addressChanged.active = true
            }
        }

        action: [
            Action {
                text: qsTr("OK")

                onTriggered: {
                    root.addressChanged.active = false
                }
            }
        ]
    }

    property Notification apiCertIssue: Notification {
        title: qsTr("Unable to establish a \nsecure connection to \nProton servers")
        brief: qsTr("Cannot establish secure connection")
        description: qsTr("Bridge cannot verify the authenticity of Proton servers on your current network due to a TLS certificate error. " +
        "Start Bridge again after ensuring your connection is secure and/or connecting to a VPN. Learn more about TLS pinning " +
        "<a href=\"https://proton.me/blog/tls-ssl-certificate#Extra-security-precautions-taken-by-ProtonMail\">here</a>.")

        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Dialogs | Notifications.Group.Connection

        Connections {
            target: Backend

            function onApiCertIssue() {
                root.apiCertIssue.active = true
            }
        }

        action: [
            Action {
                text: qsTr("Quit Bridge")

                onTriggered: {
                    root.apiCertIssue.active = false;
                    Backend.quit()
                }
            }
        ]
    }

    property Notification noActiveKeyForRecipient: Notification {
        title: qsTr("Unable to send \nencrypted message")
        brief: title
        description: "#PlaceholderText#"
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Dialogs | Notifications.Group.Connection

        Connections {
            target: Backend

            function onNoActiveKeyForRecipient(email) {
                root.noActiveKeyForRecipient.description = qsTr("There are no active keys to encrypt your message to %1. "+
                "Please update the setting for this contact.").arg(email)
                root.noActiveKeyForRecipient.active = true
            }
        }

        action: [
            Action {
                text: qsTr("OK")

                onTriggered: {
                    root.noActiveKeyForRecipient.active = false
                }
            }
        ]
    }

    property Notification userBadEvent: Notification {
        title: qsTr("Internal error")
        brief: title
        description: "#PlaceHolderText"
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Connection | Notifications.Group.Dialogs

        property var userID: ""

        Connections {
            target: Backend
            function onUserBadEvent(userID, errorMessage) {
                root.userBadEvent.userID = userID
                root.userBadEvent.description = errorMessage
                root.userBadEvent.active = true
            }
        }

        action: [
            Action {
                text: qsTr("Synchronize")

                onTriggered: {
                    root.userBadEvent.active = false
                    Backend.sendBadEventUserFeedback(root.userBadEvent.userID, true)
                }
            },

            Action {
                text: qsTr("Logout")

                onTriggered: {
                    root.userBadEvent.active = false
                    Backend.sendBadEventUserFeedback(root.userBadEvent.userID, false)
                }
            }
        ]

    }

    property Notification imapLoginWhileSignedOut: Notification {
        title: qsTr("IMAP Login failed")
        brief: title
        description: "#PlaceHolderText"
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Connection

        Connections {
            target: Backend
            function onImapLoginWhileSignedOut(username) {
                root.imapLoginWhileSignedOut.description = qsTr("An email client tried to connect to the account %1, but this account is signed " +
                "out. Please sign-in to continue.").arg(username)
                root.imapLoginWhileSignedOut.active = true
            }
        }

        action: [
            Action {
                text: qsTr("OK")

                onTriggered: {
                    root.imapLoginWhileSignedOut.active = false
                }
            }
        ]
    }

    property Notification genericError: Notification {
        title: ""
        brief: title
        description: ""
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Dialogs
        Connections {
            target: Backend
            function onGenericError(title, description) {
                root.genericError.title = title
                root.genericError.description = description
                root.genericError.active = true;
            }
        }

        action: [
            Action {
                text: qsTr("OK")

                onTriggered: {
                    root.genericError.active = false
                }
            }
        ]
    }

    property Notification genericQuestion: Notification {
        title: ""
        brief: title
        description: ""
        type: Notification.NotificationType.Warning
        group: Notifications.Group.Dialogs
        property var option1: ""
        property var option2: ""
        property variant action1: null
        property variant action2: null

        Connections {
            target: root
            function onAskQuestion(title, description, option1, option2, action1, action2) {
                root.genericQuestion.title  = title
                root.genericQuestion.description  = description
                root.genericQuestion.option1  = option1
                root.genericQuestion.option2  = option2
                root.genericQuestion.action1  = action1
                root.genericQuestion.action2  = action2
                root.genericQuestion.active = true
            }
        }

        action: [
            Action {
                text: root.genericQuestion.option1

                onTriggered: {
                    root.genericQuestion.active = false
                    if (root.genericQuestion.action1)
                    root.genericQuestion.action1()
                }
            },
            Action {
                text: root.genericQuestion.option2

                onTriggered: {
                    root.genericQuestion.active = false
                    if (root.genericQuestion.action2)
                    root.genericQuestion.action2()
                }
            }
        ]
    }
}
