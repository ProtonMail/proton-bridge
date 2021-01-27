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
            anchors.horizontalCenter : parent.horizontalCenter

            ButtonIconText {
                id: logs
                anchors.left: parent.left
                text: qsTr("Logs", "title of button that takes user to logs directory")
                leftIcon.text  : Style.fa.align_justify
                rightIcon.text : Style.fa.chevron_circle_right
                rightIcon.font.pointSize : Style.settings.toggleSize * Style.pt
                onClicked: go.openLogs()
            }

            ButtonIconText {
                id: bugreport
                anchors.left: parent.left
                text: qsTr("Report Bug", "title of button that takes user to bug report form")
                leftIcon.text  : Style.fa.bug
                rightIcon.text : Style.fa.chevron_circle_right
                rightIcon.font.pointSize : Style.settings.toggleSize * Style.pt
                onClicked: bugreportWin.show()
            }

            ButtonIconText {
                id: manual
                anchors.left: parent.left
                text: qsTr("Setup Guide", "title of button that opens setup and installation guide")
                leftIcon.text  : Style.fa.book
                rightIcon.text : Style.fa.chevron_circle_right
                rightIcon.font.pointSize : Style.settings.toggleSize * Style.pt
                onClicked: go.openManual()
            }

            ButtonIconText {
                id: updates
                anchors.left: parent.left
                text: qsTr("Check for Updates", "title of button to check for any app updates")
                leftIcon.text  : Style.fa.refresh
                rightIcon.text : Style.fa.chevron_circle_right
                rightIcon.font.pointSize : Style.settings.toggleSize * Style.pt
                onClicked: {
                    go.checkForUpdates()
                }
            }

            // Bottom version notes
            Rectangle {
                anchors.horizontalCenter : parent.horizontalCenter
                height: viewAccount.separatorNoAccount - 3.2*manual.height
                width: wrapper.width
                color : "transparent"
                AccessibleText {
                    anchors {
                        bottom: parent.bottom
                        horizontalCenter: parent.horizontalCenter
                    }
                    color: Style.main.textDisabled
                    horizontalAlignment: Qt.AlignHCenter
                    font.pointSize :  Style.main.fontSize * Style.pt
                    text:
                    "ProtonMail Bridge "+go.getBackendVersion()+"\n"+
                    "© 2020 Proton Technologies AG"
                }
            }
            Row {
                anchors.left : parent.left

                Rectangle { height: Style.dialog.spacing; width: (wrapper.width - credits.width - licenseFile.width - release.width - sepaCreditsRelease.width)/2; color: "transparent"}

                ClickIconText {
                    id:credits
                    iconText      : ""
                    text          : qsTr("Credits", "link to click on to view list of credited libraries")
                    textColor     : Style.main.textDisabled
                    fontSize      : Style.main.fontSize
                    textUnderline : true
                    onClicked     : winMain.dialogCredits.show()
                }

                Rectangle {id: sepaLicenseFile ; height: Style.dialog.spacing; width: Style.main.dummy; color: "transparent"}

                ClickIconText {
                    id:licenseFile
                    iconText      : ""
                    text          : qsTr("License", "link to click on to view license file")
                    textColor     : Style.main.textDisabled
                    fontSize      : Style.main.fontSize
                    textUnderline : true
                    onClicked     : {
                        go.openLicenseFile()
                    }
                }

                Rectangle {id: sepaCreditsRelease ; height: Style.dialog.spacing; width: Style.main.dummy; color: "transparent"}

                ClickIconText {
                    id:release
                    iconText      : ""
                    text          : qsTr("Release notes", "link to click on to view release notes for this version of the app")
                    textColor     : Style.main.textDisabled
                    fontSize      : Style.main.fontSize
                    textUnderline : true
                    onClicked : gui.openReleaseNotes()
                }
            }
        }
    }
}
