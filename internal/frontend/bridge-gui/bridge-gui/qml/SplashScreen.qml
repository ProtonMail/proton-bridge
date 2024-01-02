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
import QtQuick
import QtQuick.Layouts
import QtQuick.Controls
import QtQuick.Controls.impl
import Proton

Dialog {
    id: root
    leftPadding: 0
    modal: true
    rightPadding: 0
    shouldShow: Backend.showSplashScreen
    topPadding: 0

    ColumnLayout {
        spacing: 20

        Image {
            Layout.alignment: Qt.AlignHCenter
            Layout.preferredHeight: 144
            Layout.preferredWidth: 384
            source: "./icons/img-splash.png"
            sourceSize.height: 144
            sourceSize.width: 384
        }
        Label {
            Layout.alignment: Qt.AlignHCenter
            Layout.leftMargin: 24
            Layout.preferredWidth: 336
            Layout.rightMargin: 24
            colorScheme: root.colorScheme
            horizontalAlignment: Text.AlignHCenter
            text: qsTr("What's new in Bridge")
            type: Label.Title
        }
        RowLayout {
            width: root.width

            Item {
                Layout.fillHeight: true
                Layout.leftMargin: 32
                Layout.rightMargin: 16
                width: 24

                Image {
                    anchors.horizontalCenter: parent.horizontalCenter
                    anchors.verticalCenter: parent.verticalCenter
                    source: "./icons/ic-splash-check.svg"
                    sourceSize.height: 24
                    sourceSize.width: 24
                }
            }
            Label {
                Layout.alignment: Qt.AlignHCenter
                Layout.fillWidth: true
                Layout.leftMargin: 0
                Layout.preferredWidth: 264
                Layout.rightMargin: 24
                colorScheme: root.colorScheme
                horizontalAlignment: Text.AlignLeft
                text: qsTr("<b>New IMAP engine</b><br/>For improved stability and performance.")
                textFormat: Text.StyledText
                type: Label.Body
                wrapMode: Text.WordWrap
            }
        }
        RowLayout {
            width: root.width

            Item {
                Layout.fillHeight: true
                Layout.leftMargin: 32
                Layout.rightMargin: 16
                width: 24

                Image {
                    anchors.horizontalCenter: parent.horizontalCenter
                    anchors.verticalCenter: parent.verticalCenter
                    source: "./icons/ic-splash-check.svg"
                    sourceSize.height: 24
                    sourceSize.width: 24
                }
            }
            Label {
                Layout.alignment: Qt.AlignHCenter
                Layout.fillWidth: true
                Layout.leftMargin: 0
                Layout.preferredWidth: 264
                Layout.rightMargin: 24
                colorScheme: root.colorScheme
                horizontalAlignment: Text.AlignLeft
                text: qsTr("<b>Faster than ever</b><br/>Up to 10x faster syncing and receiving.")
                textFormat: Text.StyledText
                type: Label.Body
                wrapMode: Text.WordWrap
            }
        }
        RowLayout {
            width: root.width

            Item {
                Layout.fillHeight: true
                Layout.leftMargin: 32
                Layout.rightMargin: 16
                width: 24

                Image {
                    anchors.horizontalCenter: parent.horizontalCenter
                    anchors.verticalCenter: parent.verticalCenter
                    source: "./icons/ic-splash-check.svg"
                    sourceSize.height: 24
                    sourceSize.width: 24
                }
            }
            Label {
                Layout.alignment: Qt.AlignHCenter
                Layout.fillWidth: true
                Layout.leftMargin: 0
                Layout.preferredWidth: 264
                Layout.rightMargin: 24
                colorScheme: root.colorScheme
                horizontalAlignment: Text.AlignLeft
                text: qsTr("<b>Extra security</b><br/>New, encrypted local database and keychain improvements.")
                textFormat: Text.StyledText
                type: Label.Body
                wrapMode: Text.WordWrap
            }
        }
        Button {
            Layout.fillWidth: true
            Layout.leftMargin: 24
            Layout.rightMargin: 24
            colorScheme: root.colorScheme
            text: "Got it"

            onClicked: Backend.showSplashScreen = false
        }
        Label {
            Layout.alignment: Qt.AlignHCenter
            Layout.fillWidth: true
            Layout.leftMargin: 24
            Layout.preferredWidth: 336
            Layout.rightMargin: 24
            colorScheme: root.colorScheme
            horizontalAlignment: Text.AlignHCenter
            text: qsTr("Note that your client will redownload all the emails.<br/>") + link("https://proton.me/blog/new-proton-mail-bridge", qsTr("Learn more about new Bridge."))
            textFormat: Text.StyledText
            type: Label.Body
            wrapMode: Text.WordWrap

            onLinkActivated: function (link) {
                Backend.openExternalLink(link);
            }
        }
    }
}
