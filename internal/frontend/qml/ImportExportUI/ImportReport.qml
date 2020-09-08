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

// Import report modal
import QtQuick 2.11
import QtQuick.Controls 2.4
import ProtonUI 1.0
import ImportExportUI 1.0

Rectangle {
    id: root
    color: "#aa101021"
    visible: false

    MouseArea { // disable bellow
        anchors.fill: root
        hoverEnabled: true
    }

    Rectangle {
        id:background
        color: Style.main.background
        anchors {
            fill         : root
            topMargin    : Style.main.rightMargin
            leftMargin   : 2*Style.main.rightMargin
            rightMargin  : 2*Style.main.rightMargin
            bottomMargin : 2.5*Style.main.rightMargin
        }

        ClickIconText {
            anchors {
                top     : parent.top
                right   : parent.right
                margins : .5* Style.main.rightMargin
            }
            iconText  : Style.fa.times
            text      : ""
            textColor : Style.main.textBlue
            onClicked : root.hide()
            Accessible.description : qsTr("Close dialog %1", "Click to exit modal.").arg(title.text)
        }

        Text {
            id: title
            text : qsTr("List of errors")
            font {
                pointSize: Style.dialog.titleSize * Style.pt
            }
            anchors {
                top              : parent.top
                topMargin        : 0.5*Style.main.rightMargin
                horizontalCenter : parent.horizontalCenter
            }
        }

        ListView {
            id: errorView
            anchors {
                left    : parent.left
                right   : parent.right
                top     : title.bottom
                bottom  : detailBtn.top
                margins : Style.main.rightMargin
            }

            clip               : true
            flickableDirection : Flickable.HorizontalAndVerticalFlick
            contentWidth       : errorView.rWall
            boundsBehavior     : Flickable.StopAtBounds

            ScrollBar.vertical: ScrollBar {
                anchors {
                    right        : parent.right
                    top          : parent.top
                    rightMargin  : Style.main.rightMargin/4
                    topMargin    : Style.main.rightMargin
                }
                width: Style.main.rightMargin/3
                Accessible.ignored: true
            }
            ScrollBar.horizontal: ScrollBar {
                anchors {
                    bottom       : parent.bottom
                    right        : parent.right
                    bottomMargin : Style.main.rightMargin/4
                    rightMargin  : Style.main.rightMargin
                }
                height: Style.main.rightMargin/3
                Accessible.ignored: true
            }



            property real rW1   : 150 *Style.px
            property real rW2   : 150 *Style.px
            property real rW3   : 100 *Style.px
            property real rW4   : 150 *Style.px
            property real rW5   : 550 *Style.px
            property real rWall : errorView.rW1+errorView.rW2+errorView.rW3+errorView.rW4+errorView.rW5
            property real pH    : .5*Style.main.rightMargin

            model    : errorList
            delegate : Rectangle {
                width  : Math.max(errorView.width, row.width)
                height : row.height

                Row {
                    id: row

                    spacing       : errorView.pH
                    leftPadding   : errorView.pH
                    rightPadding  : errorView.pH
                    topPadding    : errorView.pH
                    bottomPadding : errorView.pH

                    ImportReportCell { width : errorView.rW1; text : mailSubject  }
                    ImportReportCell { width : errorView.rW2; text : mailDate     }
                    ImportReportCell { width : errorView.rW3; text : inputFolder  }
                    ImportReportCell { width : errorView.rW4; text : mailFrom     }
                    ImportReportCell { width : errorView.rW5; text : errorMessage }
                }

                Rectangle {
                    color          : Style.main.line
                    height         : .8*Style.px
                    width          : parent.width
                    anchors.left   : parent.left
                    anchors.bottom : parent.bottom
                }
            }

            headerPositioning: ListView.OverlayHeader
            header: Rectangle {
                height : viewHeader.height
                width  : Math.max(errorView.width, viewHeader.width)
                color  : Style.accounts.backgroundExpanded
                z      : 2

                Row {
                    id: viewHeader

                    spacing       : errorView.pH
                    leftPadding   : errorView.pH
                    rightPadding  : errorView.pH
                    topPadding    : .5*errorView.pH
                    bottomPadding : .5*errorView.pH

                    ImportReportCell { width : errorView.rW1 ; text : qsTr ( "SUBJECT"   ); isHeader: true }
                    ImportReportCell { width : errorView.rW2 ; text : qsTr ( "DATE/TIME" ); isHeader: true }
                    ImportReportCell { width : errorView.rW3 ; text : qsTr ( "FOLDER"    ); isHeader: true }
                    ImportReportCell { width : errorView.rW4 ; text : qsTr ( "FROM"      ); isHeader: true }
                    ImportReportCell { width : errorView.rW5 ; text : qsTr ( "ERROR"     ); isHeader: true }
                }

                Rectangle {
                    color          : Style.main.line
                    height         : .8*Style.px
                    width          : parent.width
                    anchors.left   : parent.left
                    anchors.bottom : parent.bottom
                }
            }
        }

        Rectangle {
            anchors{
                fill    : errorView
                margins : -radius
            }
            radius : 2* Style.px
            color  : Style.transparent
            border {
                width : Style.px
                color : Style.main.line
            }
        }

        ButtonRounded {
            id: detailBtn
            fa_icon    : Style.fa.file_text
            text       : qsTr("Detailed file")
            color_main : Style.dialog.textBlue
            onClicked  : go.importLogFileName == "" ? go.openLogs() : go.openReport()

            anchors {
                bottom           : parent.bottom
                bottomMargin     : 0.5*Style.main.rightMargin
                horizontalCenter : parent.horizontalCenter
            }
        }
    }


    function show() {
        root.visible = true
    }

    function hide() {
        root.visible = false
    }
}
