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

    enum Group {
        Connection  = 1,
        Update      = 2,
        Configuration = 4,
        API         = 32,

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
        root.bugReportSendSuccess,
        root.bugReportSendError,
        root.cacheAnavailable,
        root.cacheCantMove,
        root.accountChanged,
        root.diskFull
    ]

    // Connection
    property Notification noInternet: Notification {
        text: qsTr("No connection")
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
        text: qsTr("Update to Bridge") + " " + (data ? data.version : "")
        description: qsTr("A new version of ProtonMail Bridge is available. See what's changed.")
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
                text: qsTr("Update")

                onTriggered: {
                    // TODO: call update from backend
                    root.updateManualReady.active = false
                }
            },
            Action {
                text: qsTr("Remind me later")

                onTriggered: {
                    // TODO: start timer here
                    root.updateManualReady.active = false
                }
            }
        ]
    }

    property Notification updateManualRestartNeeded: Notification {
        text: qsTr("Bridge update is ready")
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
                // TODO
                root.updateManualRestartNeeded.active = false
            }
        }
    }

    property Notification updateManualError: Notification {
        text: qsTr("Bridge couldn’t update")
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Warning
        group: Notifications.Group.Update

        Connections {
            target: root.backend
            onUpdateManualError: {
                root.updateManualError.active = true
            }
        }

        action: Action {
            text: qsTr("Update manually")

            onTriggered: {
                // TODO
                root.updateManualError.active = false
            }
        }
    }

    property Notification updateForce: Notification {
        text: qsTr("Update to ProtonMail Bridge") + " " + (data ? data.version : "")
        description: qsTr("This version of Bridge is no longer supported, please update. Learn why. To update manually, go to: https:/protonmail.com/bridge/download")
        icon: "./icons/ic-exclamation-circle-filled.svg"
        type: Notification.NotificationType.Danger
        group: Notifications.Group.Update | Notifications.Group.Dialogs

        Connections {
            target: root.backend

            onUpdateForce: {
                root.updateForce.data = { version: version }
                root.updateForce.active = true
            }
        }

        action: [
            Action {
                text: qsTr("Update")

                onTriggered: {
                    // TODO: trigger update here
                    root.updateForce.active = false
                }
            },
            Action {
                text: qsTr("Quite Bridge")

                onTriggered: {
                    // TODO: quit Bridge here
                    root.updateForce.active = false
                }
            }
        ]
    }

    property Notification updateForceError: Notification {
        text: qsTr("Bridge coudn’t update")
        description: qsTr("You must update manually. Go to: https:/protonmail.com/bridge/download")
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
                    // TODO: trigger update here
                    root.updateForceError.active = false
                }
            },
            Action {
                text: qsTr("Quite Bridge")

                onTriggered: {
                    // TODO: quit Bridge here
                    root.updateForce.active = false
                }
            }
        ]
    }

    property Notification updateSilentRestartNeeded: Notification {
        text: qsTr("Bridge update is ready")
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
                // TODO
                root.updateSilentRestartNeeded.active = false
            }
        }
    }

    property Notification updateSilentError: Notification {
        text: qsTr("Bridge couldn’t update")
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
                // TODO
                root.updateSilentError.active = false
            }
        }
    }

    // Bug reports
    property Notification bugReportSendSuccess: Notification {
        text: qsTr("Bug report sent")
        description: qsTr("We’ve received your report, thank you! Our team will get back to you as soon as we can.")
        type: Notification.NotificationType.Success
        group: Notifications.Group.Configuration | Notifications.Group.Dialogs

        Connections {
            target: root.backend
            onBugReportSendSuccess: {
                root.bugReportSendSuccess.active = true
            }
        }

        action: [
            Action {
                text: qsTr("Ok")
                onTriggered: {
                    root.bugReportSendSuccess.active = false
                }
            },
            Action {
                text: "test"
            }
        ]
    }

    property Notification bugReportSendError: Notification {
        text: qsTr("There was a problem")
        description: qsTr("There was a problem with sending your report. Please try again later or contact us directly at security@protonmail.com")
        type: Notification.NotificationType.Warning
        group: Notifications.Group.Configuration | Notifications.Group.Dialogs

        Connections {
            target: root.backend
            onBugReportSendError: {
                root.bugReportSendError.active = true
            }
        }

        action: Action {
            text: qsTr("Ok")
            onTriggered: {
                root.bugReportSendError.active = false
            }
        }
    }

    // Cache
    property Notification cacheAnavailable: Notification {
        text: qsTr("Cache location is unavailable")
        description: qsTr("Check the directory or change it in your settings.")
        type: Notification.NotificationType.Warning
        group: Notifications.Group.Configuration | Notifications.Group.Dialogs

        Connections {
            target: root.backend
            onCacheAnavailable: {
                root.cacheAnavailable.active = true
            }
        }

        action: [
            Action {
                text: qsTr("Quit Bridge")
                onTriggered: {
                    root.cacheAnavailable.active = false
                }
            },
            Action {
                text: qsTr("Change location")
                onTriggered: {
                    root.cacheAnavailable.active = false
                }
            }
        ]
    }

    property Notification cacheCantMove: Notification {
        text: qsTr("Can’t move cache")
        description: qsTr("The location you have selected is not available. Make sure you have enough free space or choose another location.")
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
                }
            }
        ]
    }

    // Other
    property Notification accountChanged: Notification {
        text: qsTr("The address list for your account has changed")
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
        text: qsTr("Your disk is almost full")
        description: qsTr("Quit Bridge and free disk space or disable the local cache (not recommended).")
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
                    root.diskFull.active = false
                }
            },
            Action {
                text: qsTr("Settings")
                onTriggered: {
                    root.diskFull.active = false
                }
            }
        ]
    }
}
