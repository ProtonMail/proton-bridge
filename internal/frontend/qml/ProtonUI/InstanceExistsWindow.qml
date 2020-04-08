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

// This is main window

import QtQuick 2.8
import QtQuick.Window 2.2
import ProtonUI 1.0


// Main Window
Window {
    id:winMain

    // main window appeareance
    width  : Style.main.width
    height : Style.main.height
    flags  : Qt.Window | Qt.Dialog
    title: qsTr("ProtonMail Bridge", "app title")
    color  : Style.main.background
    visible : true

    Text {
        id: title
        anchors {
            horizontalCenter: parent.horizontalCenter
            top: parent.top
            topMargin: Style.main.topMargin
        }
        font{
            pointSize: Style.dialog.titleSize * Style.pt
        }
        color: Style.main.text
        text:
        "<span style='font-family: " + Style.fontawesome.name + "'>" + Style.fa.exclamation_triangle + "</span> " +
        qsTr ("Warning: Instance exists", "displayed when a version of the app is opened while another is already running")
    }

    Text {
        id: message
        anchors.centerIn : parent
        horizontalAlignment: Text.AlignHCenter
        font.pointSize: Style.dialog.fontSize * Style.pt
        color: Style.main.text
        width: 2*parent.width/3
        wrapMode: Text.Wrap
        text: qsTr("An instance of the ProtonMail Bridge is already running.", "displayed when a version of the app is opened while another is already running") + " " +
        qsTr("Please close the existing ProtonMail Bridge process before starting a new one.", "displayed when a version of the app is opened while another is already running")+ " " +
        qsTr("This program will close now.", "displayed when a version of the app is opened while another is already running")
    }

    ButtonRounded {
        anchors {
            horizontalCenter: parent.horizontalCenter
            bottom: parent.bottom
            bottomMargin: Style.main.bottomMargin
        }
        text: qsTr("Okay", "confirms and dismisses a notification")
        color_main: Style.dialog.text
        color_minor: Style.main.textBlue
        isOpaque: true
        onClicked: Qt.quit()
    }
}

