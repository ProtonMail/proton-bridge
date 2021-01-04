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

// Popup message
import QtQuick 2.8
import QtQuick.Controls 2.2
import ProtonUI 1.0

Rectangle {
    id: root
    color: Style.transparent
    property alias text         : message.text
    property alias checkbox     : checkbox
    property alias buttonQuit   : buttonQuit
    property alias buttonOkay   : buttonOkay
    property alias buttonYes    : buttonYes
    property alias buttonNo     : buttonNo
    property alias buttonRetry  : buttonRetry
    property alias buttonSkip   : buttonSkip
    property alias buttonCancel : buttonCancel
    property alias msgWidth     : backgroundInp.width
    property string msgID       : ""
    visible: false

    signal clickedOkay()
    signal clickedYes()
    signal clickedNo()
    signal clickedRetry()
    signal clickedSkip()
    signal clickedCancel()

    MouseArea { // prevent action below
        anchors.fill: parent
        hoverEnabled: true
    }

    Rectangle {
        id: backgroundInp
        anchors.centerIn : root
        color  : Style.errorDialog.background
        radius : Style.errorDialog.radius
        width  : parent.width/3.
        height : contentInp.height

        Column {
            id: contentInp
            anchors.horizontalCenter: backgroundInp.horizontalCenter
            spacing: Style.dialog.heightSeparator
            topPadding: Style.dialog.heightSeparator
            bottomPadding: Style.dialog.heightSeparator

            AccessibleText {
                id: message
                font {
                    pointSize : Style.errorDialog.fontSize * Style.pt
                    bold      : true
                }
                color: Style.errorDialog.text
                horizontalAlignment: Text.AlignHCenter
                width : backgroundInp.width - 2*Style.main.rightMargin
                wrapMode: Text.Wrap
            }

            CheckBoxLabel {
                id: checkbox
                text: ""
                checked: false
                visible: (text != "")
                textColor : Style.errorDialog.text
                checkedColor: Style.errorDialog.text
                uncheckedColor: Style.errorDialog.text
                anchors.horizontalCenter : parent.horizontalCenter
            }

            Row {
                spacing: Style.dialog.spacing
                anchors.horizontalCenter : parent.horizontalCenter

                ButtonRounded { id : buttonQuit   ; text : qsTr ( "Stop & quit", ""          )  ; onClicked : root.clickedYes    (  )  ; visible : false ; isOpaque : true  ; color_main : Style.errorDialog.text ; color_minor : Style.dialog.textBlue ; }
                ButtonRounded { id : buttonNo     ; text : qsTr ( "No"     , "Button No"     )  ; onClicked : root.clickedNo     (  )  ; visible : false ; isOpaque : false ; color_main : Style.errorDialog.text ; color_minor : Style.transparent     ; }
                ButtonRounded { id : buttonYes    ; text : qsTr ( "Yes"    , "Button Yes"    )  ; onClicked : root.clickedYes    (  )  ; visible : false ; isOpaque : true  ; color_main : Style.errorDialog.text ; color_minor : Style.dialog.textBlue ; }
                ButtonRounded { id : buttonRetry  ; text : qsTr ( "Retry"  , "Button Retry"  )  ; onClicked : root.clickedRetry  (  )  ; visible : false ; isOpaque : false ; color_main : Style.errorDialog.text ; color_minor : Style.transparent     ; }
                ButtonRounded { id : buttonSkip   ; text : qsTr ( "Skip"   , "Button Skip"   )  ; onClicked : root.clickedSkip   (  )  ; visible : false ; isOpaque : false ; color_main : Style.errorDialog.text ; color_minor : Style.transparent     ; }
                ButtonRounded { id : buttonCancel ; text : qsTr ( "Cancel" , "Button Cancel" )  ; onClicked : root.clickedCancel (  )  ; visible : false ; isOpaque : true  ; color_main : Style.errorDialog.text ; color_minor : Style.dialog.textBlue ; }
                ButtonRounded { id : buttonOkay   ; text : qsTr ( "Okay"   , "Button Okay"   )  ; onClicked : root.clickedOkay   (  )  ; visible : true  ; isOpaque : true  ; color_main : Style.errorDialog.text ; color_minor : Style.dialog.textBlue ; }
            }
        }
    }

    function show(text) {
        root.text = text
        root.visible = true
    }

    function hide() {
        root.visible=false

        root     .text = ""
        checkbox .text = ""

        buttonNo     .visible = false
        buttonYes    .visible = false
        buttonRetry  .visible = false
        buttonSkip   .visible = false
        buttonCancel .visible = false
        buttonOkay   .visible = true
    }
}
