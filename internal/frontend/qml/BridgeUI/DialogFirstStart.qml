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

// Dialog with Yes/No buttons

import QtQuick 2.8
import ProtonUI 1.0

Dialog {
    id: root

    title : ""
    isDialogBusy: false
    property string firstParagraph : qsTr("The Bridge is an application that runs on your computer in the background and seamlessly encrypts and decrypts your mail as it enters and leaves your computer.", "instructions that appear on welcome screen at first start")
    property string secondParagraph : qsTr("To add your ProtonMail account to the Bridge and <strong>generate your Bridge password</strong>, please see <a href=\"https://protonmail.com/bridge/install\">the installation guide</a> for detailed setup instructions.", "confirms and dismisses a notification (URL that leads to installation guide should stay intact)")

    Column {
        id: dialogMessage
        property int heightInputs : welcome.height + middleSep.height + instructions.height + buttSep.height + buttonOkay.height + imageSep.height + logo.height

        Rectangle { color : "transparent"; width : Style.main.dummy; height : (root.height-dialogMessage.heightInputs)/2 }

        Text {
            id:welcome
            color: Style.main.text
            font.bold: true
            font.pointSize: 1.5*Style.main.fontSize*Style.pt
            anchors.horizontalCenter: parent.horizontalCenter
            horizontalAlignment: Text.AlignHCenter
            text: qsTr("Welcome to the", "welcome screen that appears on first start")
        }

        Rectangle {id: imageSep; color : "transparent"; width : Style.main.dummy; height : Style.dialog.heightSeparator }


        Row {
            anchors.horizontalCenter: parent.horizontalCenter
            spacing: Style.dialog.spacing
            Image {
                id: logo
                anchors.bottom : pmbridge.baseline
                height   : 2*Style.main.fontSize
                fillMode : Image.PreserveAspectFit
                mipmap   : true
                source   : "../ProtonUI/images/pm_logo.png"
            }
            AccessibleText {
                id:pmbridge
                color: Style.main.text
                font.bold: true
                font.pointSize: 2.2*Style.main.fontSize*Style.pt
                horizontalAlignment: Text.AlignHCenter
                text: qsTr("ProtonMail Bridge", "app title")

                Accessible.name: this.clearText(pmbridge.text)
                Accessible.description: this.clearText(welcome.text+ " " + pmbridge.text + ". " + root.firstParagraph + ". " + root.secondParagraph)
            }
        }



        Rectangle { id:middleSep; color : "transparent"; width : Style.main.dummy; height : Style.dialog.heightSeparator }


        Text {
            id:instructions
            color: Style.main.text
            font.pointSize: Style.main.fontSize*Style.pt
            anchors.horizontalCenter: parent.horizontalCenter
            horizontalAlignment: Text.AlignHCenter
            width: root.width/1.5
            wrapMode: Text.Wrap
            textFormat: Text.RichText
            text: "<html><style>a { color: "+Style.main.textBlue+"; text-decoration: none;}</style><body>"+
            root.firstParagraph +
            "<br/><br/>"+
            root.secondParagraph +
            "</body></html>"
            onLinkActivated: {
                Qt.openUrlExternally(link)
            }
        }

        Rectangle { id:buttSep; color : "transparent"; width : Style.main.dummy; height : 2*Style.dialog.heightSeparator }


        ButtonRounded {
            id:buttonOkay
            color_main: Style.dialog.text
            color_minor: Style.main.textBlue
            isOpaque: true
            fa_icon: Style.fa.check
            text: qsTr("Okay", "confirms and dismisses a notification")
            onClicked : root.hide()
            anchors.horizontalCenter: parent.horizontalCenter
        }
    }

    timer.interval : 3000

    Connections {
        target: timer
        onTriggered: {
        }
    }

    onShow : {
        pmbridge.Accessible.selected = true
    }
}
