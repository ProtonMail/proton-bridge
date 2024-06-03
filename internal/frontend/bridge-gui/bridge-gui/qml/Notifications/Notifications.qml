// Copyright (c) 2024 Proton AG
// This file is part of Proton Mail Bridge.
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.
import QtQml
import Qt.labs.platform
import QtQuick.Controls
import QtQuick.Layouts
import QtQuick
import "../"

QtObject {
    id: root
    enum Group {
        Connection = 1,
        Update,
        Configuration = 4,
        ForceUpdate = 8,
        API = 32,

        // Special group for notifications that require dialog popup instead of banner
        Dialogs = 64
    }

    property Notification addressChanged: Notification {
        brief: title
        description: qsTr("The address list for your account has changed. You might need to reconfigure your email client.")
        group: Notifications.Group.Configuration
        icon: "./icons/ic-exclamation-circle-filled.svg"
        linkText: qsTr("Learn more about address list changes")
        linkUrl: "https://proton.me/support/bridge-address-list-has-changed"
        title: qsTr("Address list changes")
        type: Notification.NotificationType.Warning
        action: [
            Action {
                text: qsTr("OK")

                onTriggered: {
                    root.addressChanged.active = false;
                }
            }
        ]

        Connections {
            function onAddressChanged(address) {
                root.addressChanged.description = qsTr("The address list for your account %1 has changed. You might need to reconfigure your email client.").arg(address);
                root.addressChanged.active = true;
            }
            function onAddressChangedLogout(address) {
                root.addressChanged.description = qsTr("The address list for your account %1 has changed. You have to reconfigure your email client.").arg(address);
                root.addressChanged.active = true;
            }

            target: Backend
        }
    }
    property var all: [root.noInternet, root.imapPortStartupError, root.smtpPortStartupError, root.imapPortChangeError, root.smtpPortChangeError, root.imapConnectionModeChangeError, root.smtpConnectionModeChangeError, root.updateManualReady, root.updateManualRestartNeeded, root.updateManualError, root.updateForce, root.updateForceError, root.updateSilentRestartNeeded, root.updateSilentError, root.updateIsLatestVersion, root.loginConnectionError, root.onlyPaidUsers, root.alreadyLoggedIn, root.enableBeta, root.bugReportSendSuccess, root.bugReportSendError, root.bugReportSendFallback, root.cacheCantMove, root.cacheLocationChangeSuccess, root.enableSplitMode, root.resetBridge, root.changeAllMailVisibility, root.deleteAccount, root.noKeychain, root.rebuildKeychain, root.addressChanged, root.apiCertIssue, root.userBadEvent, root.imapLoginWhileSignedOut, root.genericError, root.genericQuestion, root.hvErrorEvent, root.repairBridge]
    property Notification alreadyLoggedIn: Notification {
        brief: qsTr("Already signed in")
        description: qsTr("This account is already signed in.")
        group: Notifications.Group.Configuration
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Info

        action: [
            Action {
                text: qsTr("OK")

                onTriggered: {
                    root.alreadyLoggedIn.active = false;
                }
            }
        ]

        Connections {
            function onLoginAlreadyLoggedIn(_) {
                root.alreadyLoggedIn.active = true;
            }

            target: Backend
        }
    }
    property Notification apiCertIssue: Notification {
        brief: qsTr("Cannot establish secure connection")
        description: qsTr("Bridge cannot verify the authenticity of Proton servers on your current network due to a TLS certificate error. Start Bridge again after ensuring your connection is secure and/or connecting to a VPN.")
        group: Notifications.Group.Dialogs | Notifications.Group.Connection
        icon: "./icons/ic-exclamation-circle-filled.svg"
        linkText: qsTr("Learn mode about TLS pinning")
        linkUrl: "https://proton.me/blog/tls-ssl-certificate#Extra-security-precautions-taken-by-ProtonMail"
        title: qsTr("Unable to establish a \nsecure connection to \nProton servers")
        type: Notification.NotificationType.Danger

        action: [
            Action {
                text: qsTr("Quit Bridge")

                onTriggered: {
                    root.apiCertIssue.active = false;
                    Backend.quit();
                }
            }
        ]

        Connections {
            function onApiCertIssue() {
                root.apiCertIssue.active = true;
            }

            target: Backend
        }
    }
    property Notification bugReportSendError: Notification {
        brief: qsTr("Error sending report")
        description: qsTr("Report could not be sent. Try again or email us directly.")
        group: Notifications.Group.Configuration
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Danger

        action: Action {
            text: qsTr("OK")

            onTriggered: {
                root.bugReportSendError.active = false;
            }
        }

        Connections {
            function onBugReportSendError() {
                root.bugReportSendError.active = true;
            }

            target: Backend
        }
    }

    property Notification bugReportSendFallback: Notification {
        brief: qsTr("Error sharing debug data")
        description: qsTr("Report was sent but debug data could not be shared. Please consider sharing it using "+ "<a href=\"https://proton.me/support/send-large-files-proton-drive\">Proton Drive</a>.")
        group: Notifications.Group.Configuration
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Info

        action: Action {
            text: qsTr("OK")

            onTriggered: {
                root.bugReportSendFallback.active = false;
            }
        }

        Connections {
            function onBugReportSendFallback() {
                root.bugReportSendFallback.active = true;
            }

            target: Backend
        }
    }

    // Bug reports
    property Notification bugReportSendSuccess: Notification {
        brief: qsTr("Report sent")
        description: qsTr("Thank you for the report. We'll get back to you as soon as we can.")
        group: Notifications.Group.Configuration
        icon: "./icons/ic-info-circle-filled.svg"
        type: Notification.NotificationType.Success

        action: [
            Action {
                text: qsTr("OK")

                onTriggered: {
                    root.bugReportSendSuccess.active = false;
                }
            }
        ]

        Connections {
            function onBugReportSendSuccess() {
                root.bugReportSendSuccess.active = true;
            }

            target: Backend
        }
    }
    property Notification cacheCantMove: Notification {
        brief: title
        description: qsTr("The location you have selected is not available. Make sure you have enough free space or choose another location.")
        group: Notifications.Group.Configuration | Notifications.Group.Dialogs
        icon: "./icons/ic-exclamation-circle-filled.svg"
        linkText: qsTr("Learn more about cache relocation issues")
        linkUrl: "https://proton.me/support/bridge-cant-move-cache"
        title: qsTr("Can’t move cache")
        type: Notification.NotificationType.Warning

        action: [
            Action {
                text: qsTr("Cancel")

                onTriggered: {
                    root.cacheCantMove.active = false;
                }
            },
            Action {
                text: qsTr("Change location")

                onTriggered: {
                    root.cacheCantMove.active = false;
                    root.frontendMain.showLocalCacheSettings();
                }
            }
        ]

        Connections {
            function onCantMoveDiskCache() {
                root.cacheCantMove.active = true;
            }

            target: Backend
        }
    }
    property Notification cacheLocationChangeSuccess: Notification {
        brief: description
        description: qsTr("Cache location successfully changed")
        group: Notifications.Group.Configuration
        icon: "./icons/ic-info-circle-filled.svg"
        type: Notification.NotificationType.Success

        action: [
            Action {
                text: qsTr("OK")

                onTriggered: {
                    root.cacheLocationChangeSuccess.active = false;
                }
            }
        ]

        Connections {
            function onDiskCachePathChanged() {
                console.log("notify location changed successfully");
                root.cacheLocationChangeSuccess.active = true;
            }

            target: Backend
        }
    }
    property Notification changeAllMailVisibility: Notification {
        property var isVisibleNow

        brief: title
        description: qsTr("Switching between showing and hiding the All Mail folder will require you to restart your client.")
        group: Notifications.Group.Configuration | Notifications.Group.Dialogs
        icon: "./icons/ic-info-circle-filled.svg"
        title: root.changeAllMailVisibility.isVisibleNow ? qsTr("Hide All Mail folder?") : qsTr("Show All Mail folder?")
        type: Notification.NotificationType.Info

        action: [
            Action {
                id: allMail_change
                text: root.changeAllMailVisibility.isVisibleNow ? qsTr("Hide All Mail folder") : qsTr("Show All Mail folder")

                onTriggered: {
                    Backend.changeIsAllMailVisible(!root.changeAllMailVisibility.isVisibleNow);
                    root.changeAllMailVisibility.active = false;
                }
            },
            Action {
                id: allMail_cancel
                text: qsTr("Cancel")

                onTriggered: {
                    root.changeAllMailVisibility.active = false;
                }
            }
        ]

        Connections {
            function onAskChangeAllMailVisibility(isVisibleNow) {
                root.changeAllMailVisibility.isVisibleNow = isVisibleNow;
                root.changeAllMailVisibility.active = true;
            }

            target: root
        }
    }
    property Notification deleteAccount: Notification {
        property var user

        brief: title
        description: qsTr("Are you sure you want to remove this account from Bridge and delete locally stored preferences and data?")
        group: Notifications.Group.Configuration | Notifications.Group.Dialogs
        icon: "./icons/ic-exclamation-circle-filled.svg"
        title: qsTr("Remove this account?")
        type: Notification.NotificationType.Danger

        action: [
            Action {
                id: deleteAccount_cancel
                text: qsTr("Cancel")

                onTriggered: {
                    root.deleteAccount.active = false;
                }
            },
            Action {
                id: deleteAccount_delete
                text: qsTr("Remove this account")

                onTriggered: {
                    root.deleteAccount.user.remove();
                    root.deleteAccount.active = false;
                }
            }
        ]

        Connections {
            function onAskDeleteAccount(user) {
                root.deleteAccount.user = user;
                root.deleteAccount.active = true;
            }

            target: root
        }
    }
    property Notification enableBeta: Notification {
        brief: title
        description: qsTr("Be the first to get new updates and use new features. Bridge will update to the latest beta version.")
        group: Notifications.Group.Update | Notifications.Group.Dialogs
        icon: "./icons/ic-info-circle-filled.svg"
        title: qsTr("Enable Beta access")
        type: Notification.NotificationType.Info

        action: [
            Action {
                text: qsTr("Enable")

                onTriggered: {
                    Backend.toggleBeta(true);
                    root.enableBeta.active = false;
                }
            },
            Action {
                text: qsTr("Cancel")

                onTriggered: {
                    root.enableBeta.active = false;
                }
            }
        ]

        Connections {
            function onAskEnableBeta() {
                root.enableBeta.active = true;
            }

            target: root
        }
    }
    property Notification enableSplitMode: Notification {
        property var user

        brief: title
        description: qsTr("Changing between split and combined address mode will require you to delete your account(s) from your email client and begin the setup process from scratch.")
        group: Notifications.Group.Configuration | Notifications.Group.Dialogs
        icon: "./icons/ic-question-circle.svg"
        linkText: qsTr("Learn more about split mode")
        linkUrl: "https://proton.me/support/difference-combined-addresses-mode-split-addresses-mode"
        title: qsTr("Enable split mode?")
        type: Notification.NotificationType.Warning

        action: [
            Action {
                id: enableSplitMode_cancel
                text: qsTr("Cancel")

                onTriggered: {
                    root.enableSplitMode.active = false;
                }
            },
            Action {
                id: enableSplitMode_enable
                text: qsTr("Enable split mode")

                onTriggered: {
                    enableSplitMode_enable.loading = true;
                    enableSplitMode_cancel.enabled = false;
                    root.enableSplitMode.user.toggleSplitMode(true);
                }
            }
        ]

        Connections {
            function onAskEnableSplitMode(user) {
                root.enableSplitMode.user = user;
                root.enableSplitMode.active = true;
            }

            target: root
        }
        Connections {
            function onToggleSplitModeFinished() {
                root.enableSplitMode.active = false;
                enableSplitMode_enable.loading = false;
                enableSplitMode_cancel.enabled = true;
            }

            target: (root && root.enableSplitMode && root.enableSplitMode.user) ? root.enableSplitMode.user : null
        }
    }
    property MainWindow frontendMain
    property Notification genericError: Notification {
        brief: title
        description: ""
        group: Notifications.Group.Dialogs
        icon: "./icons/ic-exclamation-circle-filled.svg"
        title: ""
        type: Notification.NotificationType.Danger

        action: [
            Action {
                text: qsTr("OK")

                onTriggered: {
                    root.genericError.active = false;
                }
            }
        ]

        Connections {
            function onGenericError(title, description) {
                root.genericError.title = title;
                root.genericError.description = description;
                root.genericError.active = true;
            }

            target: Backend
        }
    }
    property Notification genericQuestion: Notification {
        property variant action1: null
        property variant action2: null
        property var option1: ""
        property var option2: ""

        brief: title
        description: ""
        group: Notifications.Group.Dialogs
        title: ""
        type: Notification.NotificationType.Warning

        action: [
            Action {
                text: root.genericQuestion.option1

                onTriggered: {
                    root.genericQuestion.active = false;
                    if (root.genericQuestion.action1)
                        root.genericQuestion.action1();
                }
            },
            Action {
                text: root.genericQuestion.option2

                onTriggered: {
                    root.genericQuestion.active = false;
                    if (root.genericQuestion.action2)
                        root.genericQuestion.action2();
                }
            }
        ]

        Connections {
            function onAskQuestion(title, description, option1, option2, action1, action2) {
                root.genericQuestion.title = title;
                root.genericQuestion.description = description;
                root.genericQuestion.option1 = option1;
                root.genericQuestion.option2 = option2;
                root.genericQuestion.action1 = action1;
                root.genericQuestion.action2 = action2;
                root.genericQuestion.active = true;
            }

            target: root
        }
    }
    property Notification imapConnectionModeChangeError: Notification {
        brief: qsTr("IMAP Connection mode error")
        description: qsTr("The IMAP connection mode could not be changed.")
        group: Notifications.Group.Connection
        icon: "./icons/ic-alert.svg"
        type: Notification.NotificationType.Danger

        action: Action {
            text: qsTr("OK")

            onTriggered: {
                root.imapConnectionModeChangeError.active = false;
            }
        }

        Connections {
            function onImapConnectionModeChangeError() {
                root.imapConnectionModeChangeError.active = true;
            }

            target: Backend
        }
    }
    property Notification imapLoginWhileSignedOut: Notification {
        brief: title
        description: "#PlaceHolderText"
        group: Notifications.Group.Connection
        icon: "./icons/ic-exclamation-circle-filled.svg"
        linkText: qsTr("Learn more about IMAP login issues")
        linkUrl: "https://proton.me/support/bridge-imap-login-failed"
        title: qsTr("IMAP login failed")
        type: Notification.NotificationType.Danger

        action: [
            Action {
                text: qsTr("OK")

                onTriggered: {
                    root.imapLoginWhileSignedOut.active = false;
                }
            }
        ]

        Connections {
            function onImapLoginWhileSignedOut(username) {
                root.imapLoginWhileSignedOut.description = qsTr("An email client tried to connect to the account %1, but this account is signed " + "out. Please sign-in to continue.").arg(username);
                root.imapLoginWhileSignedOut.active = true;
            }

            target: Backend
        }
    }
    property Notification imapPortChangeError: Notification {
        brief: qsTr("IMAP port error")
        description: qsTr("The IMAP port could not be changed.")
        group: Notifications.Group.Connection
        icon: "./icons/ic-alert.svg"
        linkText: qsTr("Learn more about IMAP port issues")
        linkUrl: "https://proton.me/support/port-already-occupied-error"
        type: Notification.NotificationType.Danger

        Connections {
            function onImapPortChangeError() {
                root.imapPortChangeError.active = true;
            }

            target: Backend
        }
    }
    property Notification imapPortStartupError: Notification {
        brief: qsTr("IMAP port error")
        description: qsTr("The IMAP server could not be started. Please check or change the IMAP port.")
        group: Notifications.Group.Connection
        icon: "./icons/ic-alert.svg"
        linkText: qsTr("Learn more about IMAP port issues")
        linkUrl: "https://proton.me/support/port-already-occupied-error"
        type: Notification.NotificationType.Danger

        Connections {
            function onImapPortStartupError() {
                root.imapPortStartupError.active = true;
            }

            target: Backend
        }
    }

    // login
    property Notification loginConnectionError: Notification {
        brief: qsTr("Connection error")
        description: qsTr("Bridge is not able to contact the server, please check your internet connection.")
        group: Notifications.Group.Configuration
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Danger

        action: [
            Action {
                text: qsTr("OK")

                onTriggered: {
                    root.loginConnectionError.active = false;
                }
            }
        ]

        Connections {
            function onLoginConnectionError(_) {
                root.loginConnectionError.active = true;
            }

            target: Backend
        }
    }

    // Connection
    property Notification noInternet: Notification {
        brief: qsTr("No connection")
        description: qsTr("Bridge is not able to contact the server, please check your internet connection.")
        group: Notifications.Group.Connection
        icon: "./icons/ic-no-connection.svg"
        type: Notification.NotificationType.Danger
        Connections {
            function onInternetOff() {
                root.noInternet.active = true;
            }
            function onInternetOn() {
                root.noInternet.active = false;
            }

            target: Backend
        }
    }
    property Notification noKeychain: Notification {
        brief: title
        description: Backend.goos === "darwin" ?
            qsTr("Bridge is not able to access your keychain. Please make sure your keychain is not locked and restart the application.") :
            qsTr("Bridge is not able to detect a supported password manager (pass or secret-service). Please install and setup a supported password manager and restart the application.")
        group: Notifications.Group.Dialogs | Notifications.Group.Configuration
        icon: "./icons/ic-exclamation-circle-filled.svg"
        linkText: qsTr("Learn more about keychain issues")
        linkUrl: "https://proton.me/support/bridge-cannot-access-keychain"
        title: Backend.goos === "darwin" ? qsTr("Cannot access keychain") : qsTr("No keychain available")
        type: Notification.NotificationType.Danger

        action: [
            Action {
                text: qsTr("Quit Bridge")

                onTriggered: {
                    Backend.quit();
                }
            },
            Action {
                text: qsTr("Restart Bridge")

                onTriggered: {
                    Backend.restart();
                }
            }
        ]

        Connections {
            function onNotifyHasNoKeychain() {
                root.noKeychain.active = true;
            }

            target: Backend
        }
    }
    property Notification onlyPaidUsers: Notification {
        property var pricingLink: "https://proton.me/mail/pricing"

        brief: qsTr("Upgrade your account")
        description: qsTr("Bridge is exclusive to our mail paid plans. Upgrade your account to use Bridge.")
        group: Notifications.Group.Configuration
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Danger

        action: [
            Action {
                text: qsTr("Upgrade")

                onTriggered: {
                    Backend.openExternalLink(root.onlyPaidUsers.pricingLink);
                    root.onlyPaidUsers.active = false;
                }
            }
        ]

        Connections {
            function onLoginFreeUserError() {
                root.onlyPaidUsers.active = true;
            }

            target: Backend
        }
    }
    property Notification rebuildKeychain: Notification {
        brief: title
        description: qsTr("Bridge is not able to access your macOS keychain. Please consult the instructions on our support page.")
        group: Notifications.Group.Dialogs | Notifications.Group.Configuration
        icon: "./icons/ic-exclamation-circle-filled.svg"
        title: qsTr("Your macOS keychain might be corrupted")
        type: Notification.NotificationType.Danger

        action: [
            Action {
                text: qsTr("Open the support page")

                onTriggered: {
                    Backend.openExternalLink();
                    Backend.quit();
                }
            }
        ]

        Connections {
            function onNotifyRebuildKeychain() {
                console.log("notifications");
                root.rebuildKeychain.active = true;
            }

            target: Backend
        }
    }
    property Notification resetBridge: Notification {
        property var user

        brief: title
        description: qsTr("This will clear your accounts, preferences, and cached data. You will need to reconfigure your email client. Bridge will automatically restart.")
        group: Notifications.Group.Configuration | Notifications.Group.Dialogs
        icon: "./icons/ic-exclamation-circle-filled.svg"
        title: qsTr("Reset Bridge?")
        type: Notification.NotificationType.Danger

        action: [
            Action {
                id: resetBridge_cancel
                text: qsTr("Cancel")

                onTriggered: {
                    root.resetBridge.active = false;
                }
            },
            Action {
                id: resetBridge_reset
                text: qsTr("Reset and restart")

                onTriggered: {
                    resetBridge_reset.loading = true;
                    resetBridge_cancel.enabled = false;
                    Backend.triggerReset();
                }
            }
        ]

        Connections {
            function onAskResetBridge() {
                root.resetBridge.active = true;
            }

            target: root
        }
        Connections {
            function onResetFinished() {
                root.resetBridge.active = false;
                resetBridge_reset.loading = false;
                resetBridge_cancel.enabled = true;
            }

            target: Backend
        }
    }
    property Notification smtpConnectionModeChangeError: Notification {
        brief: qsTr("SMTP Connection mode error")
        description: qsTr("The SMTP connection mode could not be changed.")
        group: Notifications.Group.Connection
        icon: "./icons/ic-alert.svg"
        type: Notification.NotificationType.Danger

        action: Action {
            text: qsTr("OK")

            onTriggered: {
                root.smtpConnectionModeChangeError.active = false;
            }
        }

        Connections {
            function onSmtpConnectionModeChangeError() {
                root.smtpConnectionModeChangeError.active = true;
            }

            target: Backend
        }
    }
    property Notification smtpPortChangeError: Notification {
        brief: qsTr("SMTP port error")
        description: qsTr("The SMTP port could not be changed.")
        group: Notifications.Group.Connection
        icon: "./icons/ic-alert.svg"
        linkText: qsTr("Learn more about SMTP port issues")
        linkUrl: "https://proton.me/support/port-already-occupied-error"
        type: Notification.NotificationType.Danger

        Connections {
            function onSmtpPortChangeError() {
                root.smtpPortChangeError.active = true;
            }

            target: Backend
        }
    }
    property Notification smtpPortStartupError: Notification {
        brief: qsTr("SMTP port error")
        description: qsTr("The SMTP server could not be started. Please check or change the SMTP port.")
        group: Notifications.Group.Connection
        icon: "./icons/ic-alert.svg"
        linkText: qsTr("Learn more about SMTP port issues")
        linkUrl: "https://proton.me/support/port-already-occupied-error"
        type: Notification.NotificationType.Danger

        Connections {
            function onSmtpPortStartupError() {
                root.smtpPortStartupError.active = true;
            }

            target: Backend
        }
    }
    property Notification updateForce: Notification {
        brief: qsTr("Bridge is outdated")
        description: qsTr("This version of Bridge is no longer supported, please update.")
        group: Notifications.Group.Update | Notifications.Group.ForceUpdate | Notifications.Group.Dialogs
        icon: "./icons/ic-exclamation-circle-filled.svg"
        title: qsTr("Update to Bridge %1").arg(data ? data.version : "")
        type: Notification.NotificationType.Danger

        action: [
            Action {
                text: qsTr("Install update")

                onTriggered: {
                    Backend.installUpdate();
                    root.updateForce.active = false;
                }
            },
            Action {
                text: qsTr("Update manually")

                onTriggered: {
                    Backend.openExternalLink(Backend.landingPageLink);
                    root.updateForce.active = false;
                }
            },
            Action {
                text: qsTr("Quit Bridge")

                onTriggered: {
                    Backend.quit();
                    root.updateForce.active = false;
                }
            }
        ]

        Connections {
            function onUpdateForce(version) {
                root.updateForce.data = {
                    "version": version
                };
                root.updateForce.active = true;
            }

            target: Backend
        }
    }
    property Notification updateForceError: Notification {
        brief: title
        description: qsTr("You must update manually. Go to: https://proton.me/mail/bridge#download")
        group: Notifications.Group.Update | Notifications.Group.Dialogs
        icon: "./icons/ic-exclamation-circle-filled.svg"
        linkText: qsTr("Learn more about Bridge updates")
        linkUrl: "https://proton.me/support/protonmail-bridge-manual-update"
        title: qsTr("Bridge couldn't update")
        type: Notification.NotificationType.Danger

        action: [
            Action {
                text: qsTr("Update manually")

                onTriggered: {
                    Backend.openExternalLink(Backend.landingPageLink);
                    root.updateForceError.active = false;
                }
            },
            Action {
                text: qsTr("Quit Bridge")

                onTriggered: {
                    Backend.quit();
                    root.updateForceError.active = false;
                }
            }
        ]

        Connections {
            function onUpdateForceError() {
                root.updateForceError.active = true;
            }

            target: Backend
        }
    }
    property Notification updateIsLatestVersion: Notification {
        brief: description
        description: qsTr("Bridge is up to date")
        group: Notifications.Group.Update
        icon: "./icons/ic-info-circle-filled.svg"
        type: Notification.NotificationType.Info

        action: Action {
            text: qsTr("OK")

            onTriggered: {
                root.updateIsLatestVersion.active = false;
            }
        }

        Connections {
            function onUpdateIsLatestVersion() {
                root.updateIsLatestVersion.active = true;
            }

            target: Backend
        }
    }
    property Notification updateManualError: Notification {
        brief: title
        description: qsTr("Please follow manual installation in order to update Bridge.")
        group: Notifications.Group.Update
        icon: "./icons/ic-exclamation-circle-filled.svg"
        linkText: qsTr("Learn more about Bridge updates")
        linkUrl: "https://proton.me/support/protonmail-bridge-manual-update"
        title: qsTr("Bridge couldn’t update")
        type: Notification.NotificationType.Warning

        action: [
            Action {
                text: qsTr("Update manually")

                onTriggered: {
                    Backend.openExternalLink(Backend.landingPageLink);
                    root.updateManualError.active = false;
                    Backend.quit();
                }
            },
            Action {
                text: qsTr("Remind me later")

                onTriggered: {
                    root.updateManualError.active = false;
                }
            }
        ]

        Connections {
            function onUpdateManualError() {
                root.updateManualError.active = true;
            }

            target: Backend
        }
    }

    // Updates
    property Notification updateManualReady: Notification {
        brief: qsTr("Update available")
        description: {
            const descr = qsTr("A new version of Proton Mail Bridge is available.");
            const text = qsTr("See what's changed.");
            const link = Backend.releaseNotesLink;
            return `${descr} <a href="${link}">${text}</a>`;
        }
        group: Notifications.Group.Update | Notifications.Group.Dialogs
        icon: "./icons/ic-info-circle-filled.svg"
        title: qsTr("Update to Bridge %1").arg(data ? data.version : "")
        type: Notification.NotificationType.Info

        action: [
            Action {
                text: qsTr("Install update")

                onTriggered: {
                    Backend.installUpdate();
                    root.updateManualReady.active = false;
                }
            },
            Action {
                text: qsTr("Update manually")

                onTriggered: {
                    Backend.openExternalLink(Backend.landingPageLink);
                    root.updateManualReady.active = false;
                }
            },
            Action {
                text: qsTr("Remind me later")

                onTriggered: {
                    root.updateManualReady.active = false;
                }
            }
        ]

        Connections {
            function onUpdateManualReady(version) {
                root.updateManualReady.data = {
                    "version": version
                };
                root.updateManualReady.active = true;
            }

            target: Backend
        }
    }
    property Notification updateManualRestartNeeded: Notification {
        brief: description
        description: qsTr("Bridge update is ready")
        group: Notifications.Group.Update
        icon: "./icons/ic-info-circle-filled.svg"
        type: Notification.NotificationType.Info

        action: Action {
            text: qsTr("Restart Bridge")

            onTriggered: {
                Backend.restart();
                root.updateManualRestartNeeded.active = false;
            }
        }

        Connections {
            function onUpdateManualRestartNeeded() {
                root.updateManualRestartNeeded.active = true;
            }

            target: Backend
        }
    }
    property Notification updateSilentError: Notification {
        brief: description
        description: qsTr("Bridge couldn't update")
        group: Notifications.Group.Update
        icon: "./icons/ic-exclamation-circle-filled.svg"
        linkText: qsTr("Learn more about Bridge updates")
        linkUrl: "https://proton.me/support/protonmail-bridge-manual-update"
        type: Notification.NotificationType.Warning

        action: Action {
            text: qsTr("Update manually")

            onTriggered: {
                Backend.openExternalLink(Backend.landingPageLink);
                root.updateSilentError.active = false;
            }
        }

        Connections {
            function onUpdateSilentError() {
                root.updateSilentError.active = true;
            }

            target: Backend
        }
    }
    property Notification updateSilentRestartNeeded: Notification {
        brief: description
        description: qsTr("Bridge update is ready")
        group: Notifications.Group.Update
        icon: "./icons/ic-info-circle-filled.svg"
        type: Notification.NotificationType.Info

        action: Action {
            text: qsTr("Restart Bridge")

            onTriggered: {
                Backend.restart();
                root.updateSilentRestartNeeded.active = false;
            }
        }

        Connections {
            function onUpdateSilentRestartNeeded() {
                root.updateSilentRestartNeeded.active = true;
            }

            target: Backend
        }
    }
    property Notification userBadEvent: Notification {
        property var userID: ""

        brief: title
        description: "#PlaceHolderText"
        group: Notifications.Group.Connection | Notifications.Group.Dialogs
        icon: "./icons/ic-exclamation-circle-filled.svg"
        linkText: qsTr("Learn more about internal errors")
        linkUrl: "https://proton.me/support/bridge-internal-error"
        title: qsTr("Internal error")
        type: Notification.NotificationType.Danger

        action: [
            Action {
                text: qsTr("Synchronize")

                onTriggered: {
                    root.userBadEvent.active = false;
                    Backend.sendBadEventUserFeedback(root.userBadEvent.userID, true);
                }
            },
            Action {
                text: qsTr("Logout")

                onTriggered: {
                    root.userBadEvent.active = false;
                    Backend.sendBadEventUserFeedback(root.userBadEvent.userID, false);
                }
            }
        ]

        Connections {
            function onUserBadEvent(userID, errorMessage) {
                root.userBadEvent.userID = userID;
                root.userBadEvent.description = errorMessage;
                root.userBadEvent.active = true;
            }

            target: Backend
        }
    }
    property Notification hvErrorEvent: Notification {
        group: Notifications.Group.Configuration
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Danger

        action: Action {
            text: qsTr("OK")
            onTriggered: {
                root.hvErrorEvent.active = false;
            }
        }

        Connections {
            function onLoginHvError(errorMsg) {
                root.hvErrorEvent.active = true;
                root.hvErrorEvent.description = errorMsg;
            }
            target: Backend
        }

    }
    property Notification repairBridge: Notification {
        brief: title
        description: qsTr("This action will reload all accounts, cached data, and re-download emails. Messages may temporarily disappear but will reappear progressively. Email clients stay connected to Bridge.")
        group: Notifications.Group.Configuration | Notifications.Group.Dialogs
        icon: "./icons/ic-exclamation-circle-filled.svg"
        title: qsTr("Repair Bridge?")
        type: Notification.NotificationType.Danger

        action: [
            Action {
                id: repairBridge_cancel
                text: qsTr("Cancel")
                onTriggered: {
                    root.repairBridge.active = false;
                }
            },
            Action {
                id: repairBridge_repair
                text: qsTr("Repair")
                onTriggered: {
                    repairBridge_repair.loading = true;
                    repairBridge_repair.enabled = false;
                    repairBridge_cancel.enabled = false;
                    Backend.triggerRepair();
                }
            }
        ]

        Connections {
            function onAskRepairBridge() {
                root.repairBridge.active = true;
            }
            target: root
        }
        
        Connections {
            function onRepairStarted() {
                root.repairBridge.active = false;
                repairBridge_repair.loading = false;
                repairBridge_repair.enabled = true;
                repairBridge_cancel.enabled = true;
            }
            target: Backend
        }

    }

    signal askChangeAllMailVisibility(var isVisibleNow)
    signal askDeleteAccount(var user)
    signal askEnableBeta
    signal askEnableSplitMode(var user)
    signal askQuestion(var title, var description, var option1, var option2, var action1, var action2)
    signal askResetBridge
    signal askRepairBridge
}
