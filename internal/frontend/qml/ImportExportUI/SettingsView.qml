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
import ProtonUI 1.0
import ImportExportUI 1.0

Item {
    id: root

    // must have wrapper
    Rectangle {
        id: wrapper
        anchors.centerIn: parent
        width: parent.width
        height: parent.height
        color: Style.main.background

        // content
        Column {
            anchors.left : parent.left

            ButtonIconText {
                id: cacheKeychain
                text: qsTr("Clear Keychain")
                leftIcon.text  : Style.fa.chain_broken
                rightIcon {
                    text : qsTr("Clear")
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
                id: logs
                anchors.left: parent.left
                text: qsTr("Logs")
                leftIcon.text  : Style.fa.align_justify
                rightIcon.text : Style.fa.chevron_circle_right
                rightIcon.font.pointSize : Style.settings.toggleSize * Style.pt
                onClicked: go.openLogs()
            }

            ButtonIconText {
                id: bugreport
                anchors.left: parent.left
                text: qsTr("Report Bug")
                leftIcon.text  : Style.fa.bug
                rightIcon.text : Style.fa.chevron_circle_right
                rightIcon.font.pointSize : Style.settings.toggleSize * Style.pt
                onClicked: bugreportWin.show()
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

            /*

             ButtonIconText {
                 id: cacheClear
                 text: qsTr("Clear Cache")
                 leftIcon.text  : Style.fa.times
                 rightIcon {
                     text : qsTr("Clear")
                     color: Style.main.text
                     font {
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
                 id: autoStart
                 text: qsTr("Automatically Start Bridge")
                 leftIcon.text  : Style.fa.rocket
                 rightIcon {
                     font.pointSize : Style.settings.toggleSize * Style.pt
                     text  : go.isAutoStart!=0 ? Style.fa.toggle_on  : Style.fa.toggle_off
                     color : go.isAutoStart!=0 ? Style.main.textBlue : Style.main.textDisabled
                 }
                 onClicked: {
                     go.toggleAutoStart()
                 }
             }

             ButtonIconText {
                 id: advancedSettings
                 property bool isAdvanced : !go.isDefaultPort
                 text: qsTr("Advanced settings")
                 leftIcon.text  : Style.fa.cogs
                 rightIcon {
                     font.pointSize : Style.settings.toggleSize * Style.pt
                     text  : isAdvanced!=0 ? Style.fa.chevron_circle_up  : Style.fa.chevron_circle_right
                     color : isAdvanced!=0 ? Style.main.textDisabled : Style.main.textBlue
                 }
                 onClicked: {
                     isAdvanced = !isAdvanced
                 }
             }

             ButtonIconText {
                 id: changePort
                 visible: advancedSettings.isAdvanced
                 text: qsTr("Change SMTP/IMAP Ports")
                 leftIcon.text  : Style.fa.plug
                 rightIcon {
                     text : qsTr("Change")
                     color: Style.main.text
                     font {
                         pointSize : Style.settings.fontSize * Style.pt
                         underline : true
                     }
                 }
                 onClicked: {
                     dialogChangePort.show()
                 }
             }
             */
        }
    }
}

