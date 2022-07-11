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
import QtQuick 2.13
import QtQuick.Layouts 1.12
import QtQuick.Controls 2.12
import QtQuick.Controls.impl 2.12

import Proton 4.0

Dialog {
    id: root

    property var backend

    shouldShow: root.backend.showSplashScreen
    modal: true

    topPadding   : 0
    leftPadding  : 0
    rightPadding : 0

    ColumnLayout {
        spacing: 20

        Image {
            Layout.alignment: Qt.AlignHCenter

            sourceSize.width: 400
            sourceSize.height: 225

            Layout.preferredWidth: 400
            Layout.preferredHeight: 225

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
            text: qsTr("Updated Proton, unified protection")
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
            text: qsTr("Introducing Proton’s refreshed look.<br/>") +
            qsTr("Many services, one mission. Welcome to an Internet where privacy is the default. ") +
            link("https://proton.me/news/updated-proton",qsTr("Learn More"))

            onLinkActivated: Qt.openUrlExternally(link)
        }

        Button {
            Layout.fillWidth: true
            Layout.leftMargin: 24
            Layout.rightMargin: 24
            colorScheme: root.colorScheme
            text: "Got it"
            onClicked: root.backend.showSplashScreen = false
        }

        Image {
            Layout.alignment: Qt.AlignHCenter

            sourceSize.width: 164
            sourceSize.height: 32

            Layout.preferredWidth: 164
            Layout.preferredHeight: 32

            source: "./icons/img-proton-logos.svg"
        }
    }
}

