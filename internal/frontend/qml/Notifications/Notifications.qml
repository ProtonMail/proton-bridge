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
import Qt.labs.platform 1.1
import QtQuick.Controls 2.12
import ".."

QtObject {
    id: root

    property var backend

    property MainWindow frontendMain
    property StatusWindow frontendStatus
    property SystemTrayIcon frontendTray

    signal askEnableBeta()
    signal askEnableSplitMode(var user)
    signal askDisableLocalCache()
    signal askEnableLocalCache(var path)
    signal askResetBridge()
    signal askDeleteAccount(var user)

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
        root.disableLocalCache,
        root.enableLocalCache,
        root.resetBridge,
        root.deleteAccount,
        root.noKeychain,
        root.rebuildKeychain,
        root.addressChanged
    ]

    // Connection
    property Notification noInternet: Notification {
        description: qsTr("Bridge is not able to contact the server, please check your internet connection.")
        brief: qsTr("No connection")
        icon: "./icons/ic-no-connection.svg"
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Connection

        Connections {
            target: root.backend

            onInternetOff: {
                root.noInternet.active = true
            }
            onInternetOn: {
                root.noInternet.active = false
            }
        }
    }

    // Updates
    property Notification updateManualReady: Notification {
        title: qsTr("Update to Bridge %1").arg(data ? data.version : "")
        description:  {
            var descr = qsTr("A new version of Proton Mail Bridge is available.")
            var text = qsTr("See what's changed.")
            var link = root.backend.releaseNotesLink
            return `${descr} <a href="${link}">${text}</a>`
        }
        brief: qsTr("Update available.")
        icon: "./icons/ic-info-circle-filled.svg"
        type: Notification.NotificationType.Info
        group: Notifications.Group.Update | Notifications.Group.Dialogs

        Connections {
            target: root.backend
            onUpdateManualReady: {
                root.updateManualReady.data = { version: version }
                root.updateManualReady.active = true
            }
        }

        action: [
            Action {
                text: qsTr("Install update")

                onTriggered: {
                    root.backend.installUpdate()
                    root.updateManualReady.active = false
                }
            },
            Action {
                text: qsTr("Update manually")

                onTriggered: {
                    Qt.openUrlExternally(root.backend.landingPageLink)
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
            target: root.backend
            onUpdateManualRestartNeeded: {
                root.updateManualRestartNeeded.active = true
            }
        }

        action: Action {
            text: qsTr("Restart Bridge")

            onTriggered: {
                root.backend.restart()
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
            target: root.backend
            onUpdateManualError: {
                root.updateManualError.active = true
            }
        }

        action: [
            Action {
                text: qsTr("Update manually")

                onTriggered: {
                    Qt.openUrlExternally(root.backend.landingPageLink)
                    root.updateManualError.active = false
                    root.backend.quit()
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
            target: root.backend

            onUpdateForce: {
                root.updateForce.data = { version: version }
                root.updateForce.active = true
            }
        }

        action: [
            Action {
                text: qsTr("Install update")

                onTriggered: {
                    root.backend.installUpdate()
                    root.updateForce.active = false
                }
            },
            Action {
                text: qsTr("Update manually")

                onTriggered: {
                    Qt.openUrlExternally(root.backend.landingPageLink)
                    root.updateForce.active = false
                }
            },
            Action {
                text: qsTr("Quit Bridge")

                onTriggered: {
                    root.backend.quit()
                    root.updateForce.active = false
                }
            }
        ]
    }

    property Notification updateForceError: Notification {
        title: qsTr("Bridge coudn’t update")
        description: qsTr("You must update manually. Go to: https:/protonmail.com/bridge/download")
        brief: title
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Update | Notifications.Group.Dialogs

        Connections {
            target: root.backend

            onUpdateForceError: {
                root.updateForceError.active = true
            }
        }

        action: [
            Action {
                text: qsTr("Update manually")

                onTriggered: {
                    Qt.openUrlExternally(root.backend.landingPageLink)
                    root.updateForceError.active = false
                }
            },
            Action {
                text: qsTr("Quit Bridge")

                onTriggered: {
                    root.backend.quit()
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
            target: root.backend
            onUpdateSilentRestartNeeded: {
                root.updateSilentRestartNeeded.active = true
            }
        }

        action: Action {
            text: qsTr("Restart Bridge")

            onTriggered: {
                root.backend.restart()
                root.updateSilentRestartNeeded.active = false
            }
        }
    }

    property Notification updateSilentError: Notification {
        description: qsTr("Bridge couldn’t update")
        brief: description
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Warning
        group: Notifications.Group.Update

        Connections {
            target: root.backend
            onUpdateSilentError: {
                root.updateSilentError.active = true
            }
        }

        action: Action {
            text: qsTr("Update manually")

            onTriggered: {
                Qt.openUrlExternally(root.backend.landingPageLink)
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
            target: root.backend
            onUpdateIsLatestVersion: {
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
            onAskEnableBeta: {
                root.enableBeta.active = true
            }
        }

        action: [
            Action {
                text: qsTr("Enable")
                onTriggered: {
                    root.backend.toggleBeta(true)
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
        brief: description
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Configuration

        Connections {
            target: root.backend
            onLoginConnectionError: {
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
        description: qsTr("Bridge is exclusive to our paid plans. Upgrade your account to use Bridge.")
        brief: description
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Configuration

        Connections {
            target: root.backend
            onLoginFreeUserError: {
                root.onlyPaidUsers.active = true
            }
        }

        action: [
            Action {
                text: qsTr("OK")
                onTriggered: {
                    root.onlyPaidUsers.active = false
                }
            }
        ]
    }

    property Notification alreadyLoggedIn: Notification {
        description: qsTr("This account is already signed it.")
        brief: description
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Info
        group: Notifications.Group.Configuration

        Connections {
            target: root.backend
            onLoginAlreadyLoggedIn: {
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
        brief: description
        icon: "./icons/ic-info-circle-filled.svg"
        type: Notification.NotificationType.Success
        group: Notifications.Group.Configuration

        Connections {
            target: root.backend
            onBugReportSendSuccess: {
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
        brief: description
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Configuration

        Connections {
            target: root.backend
            onBugReportSendError: {
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
        description: qsTr("Check the directory or change it in your settings.")
        brief: qsTr("The current cache location is unavailable. Check the directory or change it in your settings.")
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Warning
        group: Notifications.Group.Configuration | Notifications.Group.Dialogs

        Connections {
            target: root.backend
            onCacheUnavailable: {
                root.cacheUnavailable.active = true
            }
        }

        action: [
            Action {
                text: qsTr("Quit Bridge")
                onTriggered: {
                    root.backend.quit()
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
            target: root.backend
            onCacheCantMove: {
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
            target: root.backend
            onCacheLocationChangeSuccess: {
                console.log("notify location changed succesfully")
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
        brief: qsTr("The address list for your account has changed. Reconfigure your email client.")
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
        brief: qsTr("Your disk is almost full. Free disk space or disable the local cache.")
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Warning
        group: Notifications.Group.Configuration | Notifications.Group.Dialogs

        Connections {
            target: root.backend
            onDiskFull: {
                root.diskFull.active = true
            }
        }

        action: [
            Action {
                text: qsTr("Quit Bridge")
                onTriggered: {
                    root.backend.quit()
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
        icon: "./icons/ic-question-circle.svg"
        type: Notification.NotificationType.Warning
        group: Notifications.Group.Configuration | Notifications.Group.Dialogs

        property var user

        Connections {
            target: root
            onAskEnableSplitMode: {
                root.enableSplitMode.user = user
                root.enableSplitMode.active = true
            }
        }

        Connections {
            target: (root && root.enableSplitMode && root.enableSplitMode.user ) ? root.enableSplitMode.user : null
            onToggleSplitModeFinished: {
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

    property Notification disableLocalCache: Notification {
        title: qsTr("Disable local cache?")
        brief: title
        description: qsTr("This action will clear your local cache, including locally stored messages. Bridge will restart.")
        icon: "./icons/ic-question-circle.svg"
        type: Notification.NotificationType.Warning
        group: Notifications.Group.Configuration | Notifications.Group.Dialogs

        Connections {
            target: root
            onAskDisableLocalCache: {
                root.disableLocalCache.active = true
            }
        }

        Connections {
            target: root.backend
            onChangeLocalCacheFinished: {
                root.disableLocalCache.active = false

                disableLocalCache_disable.loading = false
                disableLocalCache_cancel.enabled = true
            }
        }

        action: [
            Action {
                id: disableLocalCache_cancel
                text: qsTr("Cancel")
                onTriggered: {
                    root.disableLocalCache.active = false
                }
            },
            Action {
                id: disableLocalCache_disable
                text: qsTr("Disable and restart")
                onTriggered: {
                    disableLocalCache_disable.loading = true
                    disableLocalCache_cancel.enabled = false
                    root.backend.changeLocalCache(false, root.backend.diskCachePath)
                }
            }
        ]
    }

    property Notification enableLocalCache: Notification {
        title: qsTr("Enable local cache")
        brief: title
        description: qsTr("Bridge will restart.")
        icon: "./icons/ic-question-circle.svg"
        type: Notification.NotificationType.Warning
        group: Notifications.Group.Configuration | Notifications.Group.Dialogs

        property url path

        Connections {
            target: root
            onAskEnableLocalCache: {
                root.enableLocalCache.active = true
                root.enableLocalCache.path = path
            }
        }

        Connections {
            target: root.backend
            onChangeLocalCacheFinished: {
                root.enableLocalCache.active = false

                enableLocalCache_enable.loading = false
                enableLocalCache_cancel.enabled = true
            }
        }

        action: [
            Action {
                id: enableLocalCache_enable
                text: qsTr("Enable and restart")
                onTriggered: {
                    enableLocalCache_enable.loading = true
                    enableLocalCache_cancel.enabled = false
                    root.backend.changeLocalCache(true, root.enableLocalCache.path)
                }
            },
            Action {
                id: enableLocalCache_cancel
                text: qsTr("Cancel")
                onTriggered: {
                    root.enableLocalCache.active = false
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
            onAskResetBridge: {
                root.resetBridge.active = true
            }
        }

        Connections {
            target: root.backend
            onResetFinished: {
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
                    root.backend.triggerReset()
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
            onAskDeleteAccount: {
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
        description: qsTr("Bridge is not able to detect a supported password manager (pass or secret-service). Please install and setup supported password manager and restart the application.")
        brief: title
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Dialogs | Notifications.Group.Configuration

        Connections {
            target: root.backend

            onNotifyHasNoKeychain: {
                root.noKeychain.active = true
            }
        }

        action: [
            Action {
                text: qsTr("Quit Bridge")

                onTriggered: {
                    root.backend.quit()
                }
            },
            Action {
                text: qsTr("Restart Bridge")

                onTriggered: {
                    root.backend.restart()
                }
            }
        ]
    }

    property Notification rebuildKeychain: Notification {
        title: qsTr("Your macOS keychain might be corrupted")
        description: qsTr("Bridge is not able to access your macOS keychain. Please consult the instructions on our support page.")
        brief: title
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Dialogs | Notifications.Group.Configuration

        property var supportLink: "https://protonmail.com/support/knowledge-base/macos-keychain-corrupted"


        Connections {
            target: root.backend

            onNotifyRebuildKeychain: {
                console.log("notifications")
                root.rebuildKeychain.active = true
            }
        }

        action: [
            Action {
                text: qsTr("Open the support page")

                onTriggered: {
                    Qt.openUrlExternally(root.rebuildKeychain.supportLink)
                    root.backend.quit()
                }
            }
        ]
    }

    property Notification addressChanged: Notification {
        title: qsTr("Address list changes")
        description: qsTr("The address list for your account has changed. You might need to reconfigure your email client.")
        brief: description
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Warning
        group: Notifications.Group.Configuration

        Connections {
            target: root.backend

            onAddressChanged: {
                root.addressChanged.description = qsTr("The address list for your account %1 has changed. You might need to reconfigure your email client.").arg(address)
                root.addressChanged.active = true
            }

            onAddressChangedLogout: {
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
}
