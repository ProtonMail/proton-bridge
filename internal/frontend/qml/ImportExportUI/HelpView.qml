// Copyright (c) 2020 Proton Technologies AG
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
            anchors.horizontalCenter : parent.horizontalCenter


            ButtonIconText {
                id: manual
                anchors.left: parent.left
                text: qsTr("Setup Guide")
                leftIcon.text  : Style.fa.book
                rightIcon.text : Style.fa.chevron_circle_right
                rightIcon.font.pointSize : Style.settings.toggleSize * Style.pt
                onClicked: go.openManual()
            }

            ButtonIconText {
                id: updates
                anchors.left: parent.left
                text: qsTr("Check for Updates")
                leftIcon.text  : Style.fa.refresh
                rightIcon.text : Style.fa.chevron_circle_right
                rightIcon.font.pointSize : Style.settings.toggleSize * Style.pt
                onClicked: {
                    dialogGlobal.state="checkUpdates"
                    dialogGlobal.show()
                    dialogGlobal.confirmed()
                }
            }

            Rectangle {
                anchors.horizontalCenter : parent.horizontalCenter
                height: Math.max (
                    aboutText.height +
                    Style.main.fontSize,
                    wrapper.height - (
                        2*manual.height +
                        creditsLink.height +
                        Style.main.fontSize
                    )
                )
                width: wrapper.width
                color : Style.transparent
                AccessibleText {
                    id: aboutText
                    anchors {
                        bottom: parent.bottom
                        horizontalCenter: parent.horizontalCenter
                    }
                    color: Style.main.textDisabled
                    horizontalAlignment: Qt.AlignHCenter
                    font.pointSize :  Style.main.fontSize * Style.pt
                    text: "ProtonMail Import-Export app Version "+go.getBackendVersion()+"\n© 2020 Proton Technologies AG"
                }
            }

            Row {
                anchors.horizontalCenter : parent.horizontalCenter
                spacing : Style.main.dummy

                Text {
                    id: creditsLink
                    text  : qsTr("Credits", "link to click on to view list of credited libraries")
                    color : Style.main.textDisabled
                    font.pointSize: Style.main.fontSize * Style.pt
                    font.underline: true
                    MouseArea {
                        anchors.fill: parent
                        onClicked : {
                            winMain.dialogCredits.show()
                        }
                        cursorShape: Qt.PointingHandCursor
                    }
                }

                Text {
                    id: licenseFile
                    text  : qsTr("License", "link to click on to open license file")
                    color : Style.main.textDisabled
                    font.pointSize: Style.main.fontSize * Style.pt
                    font.underline: true
                    MouseArea {
                        anchors.fill: parent
                        onClicked : {
                        go.openLicenseFile()
                        }
                        cursorShape: Qt.PointingHandCursor
                    }
                }

                Text {
                    id: releaseNotes
                    text  : qsTr("Release notes", "link to click on to view release notes for this version of the app")
                    color : Style.main.textDisabled
                    font.pointSize: Style.main.fontSize * Style.pt
                    font.underline: true
                    MouseArea {
                        anchors.fill: parent
                        onClicked : {
                            go.getLocalVersionInfo()
                            winMain.dialogVersionInfo.show()
                        }
                        cursorShape: Qt.PointingHandCursor
                    }
                }
            }
        }
    }
}

