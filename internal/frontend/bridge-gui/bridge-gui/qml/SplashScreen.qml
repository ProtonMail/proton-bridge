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
import QtQuick
import QtQuick.Layouts
import QtQuick.Controls
import QtQuick.Controls.impl

import Proton

Dialog {
    id: root

    shouldShow: Backend.showSplashScreen
    modal: true

    topPadding   : 0
    leftPadding  : 0
    rightPadding : 0

    ColumnLayout {
        spacing: 20

        Image {
            Layout.alignment: Qt.AlignHCenter

            sourceSize.width: 384
            sourceSize.height: 144

            Layout.preferredWidth: 384
            Layout.preferredHeight: 144

            source: "./icons/img-splash.png"
        }

        Label {
            colorScheme: root.colorScheme;

            Layout.alignment: Qt.AlignHCenter;
            Layout.leftMargin: 24
            Layout.rightMargin: 24
            Layout.preferredWidth: 336

            type: Label.Title
            horizontalAlignment: Text.AlignHCenter
            text: qsTr("What's new in Bridge")
        }

        RowLayout {
            width: root.width

            Item {
                Layout.fillHeight: true
                width: 24
                Layout.leftMargin: 32
                Layout.rightMargin: 16
                Image {
                    anchors.horizontalCenter: parent.horizontalCenter
                    anchors.verticalCenter: parent.verticalCenter
                    sourceSize.width: 24
                    sourceSize.height: 24
                    source: "./icons/ic-splash-check.svg"
                }
            }

            Label {
                colorScheme: root.colorScheme

                Layout.fillWidth: true
                Layout.alignment: Qt.AlignHCenter;
                Layout.preferredWidth: 264
                Layout.leftMargin: 0
                Layout.rightMargin: 24
                wrapMode: Text.WordWrap

                type: Label.Body
                horizontalAlignment: Text.AlignLeft
                textFormat: Text.StyledText
                text: qsTr("<b>New IMAP engine</b><br/>For improved stability and performances.")
            }
        }

        RowLayout {
            width: root.width

            Item {
                Layout.fillHeight: true
                width: 24
                Layout.leftMargin: 32
                Layout.rightMargin: 16
                Image {
                    anchors.horizontalCenter: parent.horizontalCenter
                    anchors.verticalCenter: parent.verticalCenter
                    sourceSize.width: 24
                    sourceSize.height: 24
                    source: "./icons/ic-splash-check.svg"
                }
            }

            Label {
                colorScheme: root.colorScheme

                Layout.fillWidth: true
                Layout.alignment: Qt.AlignHCenter;
                Layout.preferredWidth: 264
                Layout.leftMargin: 0
                Layout.rightMargin: 24
                wrapMode: Text.WordWrap

                type: Label.Body
                horizontalAlignment: Text.AlignLeft
                textFormat: Text.StyledText
                text: qsTr("<b>Faster than ever</b><br/>Up to 10x faster syncing and receiving.")
            }
        }

        RowLayout {
            width: root.width

            Item {
                Layout.fillHeight: true
                width: 24
                Layout.leftMargin: 32
                Layout.rightMargin: 16
                Image {
                    anchors.horizontalCenter: parent.horizontalCenter
                    anchors.verticalCenter: parent.verticalCenter
                    sourceSize.width: 24
                    sourceSize.height: 24
                    source: "./icons/ic-splash-check.svg"
                }
            }


            Label {
                colorScheme: root.colorScheme

                Layout.fillWidth: true
                Layout.alignment: Qt.AlignHCenter;
                Layout.preferredWidth: 264
                Layout.leftMargin: 0
                Layout.rightMargin: 24
                wrapMode: Text.WordWrap

                type: Label.Body
                horizontalAlignment: Text.AlignLeft
                textFormat: Text.StyledText
                text: qsTr("<b>Extra security</b><br/>New, encrypted local database and keychain improvements.")
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
            colorScheme: root.colorScheme
            Layout.fillWidth: true
            Layout.alignment: Qt.AlignHCenter;
            Layout.preferredWidth: 336
            Layout.leftMargin: 24
            Layout.rightMargin: 24
            wrapMode: Text.WordWrap

            type: Label.Body
            horizontalAlignment: Text.AlignHCenter
            textFormat: Text.StyledText
            text: qsTr("Note that your client will redownload all the emails.<br/>") + link("https://proton.me/blog/new-proton-mail-bridge", qsTr("Learn more about new Bridge."))

            onLinkActivated: function(link) { Qt.openUrlExternally(link) }
        }
    }
}

