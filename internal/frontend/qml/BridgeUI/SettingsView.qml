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

// List the settings

import QtQuick 2.8
import BridgeUI 1.0
import ProtonUI 1.0
import QtQuick.Controls 2.4

Item {
    id: root

    // must have wrapper
    ScrollView {
        id: wrapper
        anchors.centerIn: parent
        width: parent.width
        height: parent.height
        clip: true
        background: Rectangle {
            color: Style.main.background
        }

        // horizontall scrollbar sometimes showes up when vertical scrollbar coveres content
        ScrollBar.horizontal.policy: ScrollBar.AlwaysOff
        ScrollBar.vertical.policy: ScrollBar.AsNeeded

        // keeping vertical scrollbar allways visible when needed
        Connections {
            target: wrapper.ScrollBar.vertical
            onSizeChanged: {
                // ScrollBar.size == 0 at creating so no need to make it active
                if (wrapper.ScrollBar.vertical.size < 1.0 && wrapper.ScrollBar.vertical.size > 0 && !wrapper.ScrollBar.vertical.active) {
                    wrapper.ScrollBar.vertical.active = true
                }
            }
            onActiveChanged: {
                wrapper.ScrollBar.vertical.active = true
            }
        }

        // content
        Column {
            anchors.left : parent.left

            ButtonIconText {
                id: cacheClear
                text: qsTr("Clear Cache", "button to clear cache in settings")
                leftIcon.text  : Style.fa.times
                rightIcon {
                    text : qsTr("Clear", "clickable link next to clear cache button in settings")
                    color: Style.main.text
                    font {
                        family : cacheClear.font.family // use default font, not font-awesome
                        pointSize : Style.settings.fontSize * Style.pt
                        underline : true
                    }
                }
                onClicked: {
                    dialogGlobal.state="clearCache"
                    dialogGlobal.show()
                }
            }

            ButtonIconText {
                id: cacheKeychain
                text: qsTr("Clear Keychain", "button to clear keychain in settings")
                leftIcon.text  : Style.fa.chain_broken
                rightIcon {
                    text : qsTr("Clear", "clickable link next to clear keychain button in settings")
                    color: Style.main.text
                    font {
                        family : cacheKeychain.font.family // use default font, not font-awesome
                        pointSize : Style.settings.fontSize * Style.pt
                        underline : true
                    }
                }
                onClicked: {
                    dialogGlobal.state="clearChain"
                    dialogGlobal.show()
                }
            }

            ButtonIconText {
                id: autoStart
                text: qsTr("Automatically start Bridge", "label for toggle that activates and disables the automatic start")
                leftIcon.text  : Style.fa.rocket
                rightIcon {
                    font.pointSize : Style.settings.toggleSize * Style.pt
                    text  : go.isAutoStart!=false ? Style.fa.toggle_on  : Style.fa.toggle_off
                    color : go.isAutoStart!=false ? Style.main.textBlue : Style.main.textDisabled
                }
                Accessible.description: (
                    go.isAutoStart == false ?
                    qsTr("Enable"  , "Click to enable the automatic start of Bridge") :
                    qsTr("Disable" , "Click to disable the automatic start of Bridge")
                ) + " " + text
                onClicked: {
                    go.toggleAutoStart()
                }
            }

            ButtonIconText {
                id: autoUpdates
                text: qsTr("Keep the application up to date", "label for toggle that activates and disables the automatic updates")
                leftIcon.text  : Style.fa.download
                rightIcon {
                    font.pointSize : Style.settings.toggleSize * Style.pt
                    text  : go.isAutoUpdate!=false ? Style.fa.toggle_on  : Style.fa.toggle_off
                    color : go.isAutoUpdate!=false ? Style.main.textBlue : Style.main.textDisabled
                }
                Accessible.description: (
                    go.isAutoUpdate == false ?
                    qsTr("Enable"  , "Click to enable the automatic update of Bridge") :
                    qsTr("Disable" , "Click to disable the automatic update of Bridge")
                ) + " " + text
                onClicked: {
                    go.toggleAutoUpdate()
                }
            }

            ButtonIconText {
                id: earlyAccess
                text: qsTr("Early access", "label for toggle that enables and disables early access")
                leftIcon.text  : Style.fa.star
                rightIcon {
                    font.pointSize : Style.settings.toggleSize * Style.pt
                    text  : go.isEarlyAccess!=false ? Style.fa.toggle_on  : Style.fa.toggle_off
                    color : go.isEarlyAccess!=false ? Style.main.textBlue : Style.main.textDisabled
                }
                Accessible.description: (
                    go.isEarlyAccess == false ?
                    qsTr("Enable"  , "Click to enable early access") :
                    qsTr("Disable" , "Click to disable early access")
                ) + " " + text
                onClicked: {
                    if (go.isEarlyAccess == true) {
                      go.toggleEarlyAccess()
                    } else {
                      dialogGlobal.state="toggleEarlyAccess"
                      dialogGlobal.show()
                    }
                }
            }

            ButtonIconText {
                id: advancedSettings
                property bool isAdvanced : !go.isDefaultPort
                text: qsTr("Advanced settings", "button to open the advanced settings list in the settings page")
                leftIcon.text  : Style.fa.cogs
                rightIcon {
                    font.pointSize : Style.settings.toggleSize * Style.pt
                    text  : isAdvanced!=0 ? Style.fa.chevron_circle_up  : Style.fa.chevron_circle_right
                    color : isAdvanced!=0 ? Style.main.textDisabled : Style.main.textBlue
                }

                Accessible.description: (
                    isAdvanced ?
                    qsTr("Hide", "Click to hide the advance settings") :
                    qsTr("Show", "Click to show the advance settings")
                ) + " " + text
                onClicked: {
                    isAdvanced = !isAdvanced
                }
            }

            ButtonIconText {
                id: changePort
                visible: advancedSettings.isAdvanced
                text: qsTr("Change IMAP & SMTP settings", "button to change IMAP and SMTP ports in settings")
                leftIcon.text  : Style.fa.plug
                rightIcon {
                    text : qsTr("Change", "clickable link next to change ports button in settings")
                    color: Style.main.text
                    font {
                        family : changePort.font.family // use default font, not font-awesome
                        pointSize : Style.settings.fontSize * Style.pt
                        underline : true
                    }
                }
                onClicked: {
                    dialogChangePort.show()
                }
            }

            ButtonIconText {
                id: reportNoEnc
                text: qsTr("Notification of outgoing email without encryption", "Button to set whether to report or send an email without encryption")
                visible: advancedSettings.isAdvanced
                leftIcon.text  : Style.fa.ban
                rightIcon {
                    font.pointSize : Style.settings.toggleSize * Style.pt
                    text  : go.isReportingOutgoingNoEnc ? Style.fa.toggle_on  : Style.fa.toggle_off
                    color : go.isReportingOutgoingNoEnc ? Style.main.textBlue : Style.main.textDisabled
                }
                Accessible.description: (
                    go.isReportingOutgoingNoEnc == 0 ?
                    qsTr("Enable"  , "Click to report an email without encryption") :
                    qsTr("Disable" , "Click to send without asking an email without encryption")
                ) + " " + text
                onClicked: {
                    go.toggleIsReportingOutgoingNoEnc()
                }
            }

            ButtonIconText {
                id: allowProxy
                visible: advancedSettings.isAdvanced
                text: qsTr("Allow alternative routing", "label for toggle that allows and disallows using a proxy")
                leftIcon.text  : Style.fa.rocket
                rightIcon {
                    font.pointSize : Style.settings.toggleSize * Style.pt
                    text  : go.isProxyAllowed!=false ? Style.fa.toggle_on  : Style.fa.toggle_off
                    color : go.isProxyAllowed!=false ? Style.main.textBlue : Style.main.textDisabled
                }
                Accessible.description: (
                    go.isProxyAllowed == false ?
                    qsTr("Enable"  , "Click to allow alternative routing") :
                    qsTr("Disable" , "Click to disallow alternative routing")
                ) + " " + text
                onClicked: {
                    dialogGlobal.state="toggleAllowProxy"
                    dialogGlobal.show()
                }
            }
        }
    }
}
