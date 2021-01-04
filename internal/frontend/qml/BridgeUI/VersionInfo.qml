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

// credits

import QtQuick 2.8
import BridgeUI 1.0
import ProtonUI 1.0

Item {
    Rectangle {
        id: wrapper
        anchors.centerIn: parent
        width: 2*Style.main.width/3
        height: Style.main.height - 6*Style.dialog.titleSize
        color: "transparent"

        Flickable {
            anchors.fill       : wrapper
            contentWidth       : wrapper.width
            contentHeight      : content.height
            flickableDirection : Flickable.VerticalFlick
            clip               : true


            Column {
                id: content
                anchors.top: parent.top
                anchors.horizontalCenter: parent.horizontalCenter
                width: wrapper.width
                spacing: Style.dialog.spacing

                AccessibleText {
                    visible: go.changelog != ""
                    anchors {
                        left: parent.left
                    }
                    font.bold: true
                    font.pointSize: Style.main.fontSize * Style.pt
                    color: Style.main.text
                    text: qsTr("Release notes", "list of release notes for this version of the app") + ":"
                }

                AccessibleSelectableText {
                    anchors {
                        left: parent.left
                        leftMargin: Style.main.leftMargin
                    }
                    font {
                        pointSize : Style.main.fontSize * Style.pt
                    }
                    width: wrapper.width - anchors.leftMargin
                    onLinkActivated: {
                        Qt.openUrlExternally(link)
                    }
                    wrapMode: Text.Wrap
                    color: Style.main.text
                    text: go.changelog
                }

                AccessibleText {
                    visible: go.bugfixes != ""
                    anchors {
                        left: parent.left
                    }
                    font.bold: true
                    font.pointSize: Style.main.fontSize * Style.pt
                    color: Style.main.text
                    text: qsTr("Fixed bugs", "list of bugs fixed for this version of the app") + ":"
                }

                AccessibleSelectableText {
                    visible: go.bugfixes!=""
                    anchors {
                        left: parent.left
                        leftMargin: Style.main.leftMargin
                    }
                    font {
                        pointSize : Style.main.fontSize * Style.pt
                    }
                    width: wrapper.width - anchors.leftMargin
                    onLinkActivated: {
                        Qt.openUrlExternally(link)
                    }
                    wrapMode: Text.Wrap
                    color: Style.main.text
                    text: go.bugfixes
                }

                Rectangle{id:spacer; color:Style.transparent; width: Style.main.dummy; height: buttonClose.height}

                ButtonRounded {
                    id: buttonClose
                    anchors.horizontalCenter: content.horizontalCenter
                    text: qsTr("Close")
                    onClicked: {
                        dialogVersionInfo.hide()
                    }
                }


                AccessibleSelectableText {
                    anchors.horizontalCenter: content.horizontalCenter
                    font {
                        pointSize : Style.main.fontSize * Style.pt
                    }
                    color: Style.main.textDisabled
                    text: "\n Current: "+go.fullversion
                }
            }
        }
    }
}
