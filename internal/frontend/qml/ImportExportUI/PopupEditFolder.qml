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

// popup to edit folders or labels
import QtQuick 2.8
import QtQuick.Controls 2.1
import ImportExportUI 1.0
import ProtonUI 1.0


Rectangle {
    id: root
    visible: false
    color: "#aa223344"

    property string folderType    : gui.enums.folderTypeFolder
    property bool   isFolder      : folderType == gui.enums.folderTypeFolder
    property bool   isNew         : currentId == ""
    property bool   isCreateLater : currentId == "createLater" // NOTE: "createLater" is hack because folder id should be base64 string

    property string currentName   : ""
    property string currentId     : ""
    property string currentColor  : ""

    property string sourceID      : ""
    property string selectedColor : colorList[0]

    property color textColor : Style.main.background
    property color backColor : Style.bubble.paneBackground

    signal edited(string newName, string newColor)




    property var colorList : [ "#7272a7", "#8989ac", "#cf5858", "#cf7e7e", "#c26cc7", "#c793ca", "#7569d1", "#9b94d1", "#69a9d1", "#a8c4d5", "#5ec7b7", "#97c9c1", "#72bb75", "#9db99f", "#c3d261", "#c6cd97", "#e6c04c", "#e7d292", "#e6984c", "#dfb286" ]

    MouseArea { // prevent action below aka modal: true
        anchors.fill: parent
        hoverEnabled: true
    }

    Rectangle {
        id:background

        anchors {
            fill: root
            leftMargin: winMain.width/6
            topMargin: winMain.height/6
            rightMargin: anchors.leftMargin
            bottomMargin: anchors.topMargin
        }

        color: backColor
        radius: Style.errorDialog.radius
    }


    Column { // content
        anchors {
            top              : background.top
            horizontalCenter : background.horizontalCenter
        }

        topPadding    : Style.main.topMargin
        bottomPadding : topPadding
        spacing       : (background.height - title.height - inputField.height - view.height - buttonRow.height - topPadding - bottomPadding) / children.length

        Text {
            id: title

            font.pointSize: Style.dialog.titleSize * Style.pt
            color: textColor

            text: {
                if ( root.isFolder  && root.isNew  ) return qsTr ( "Create new folder" )
                if ( !root.isFolder && root.isNew  ) return qsTr ( "Create new label"  )
                if ( root.isFolder  && !root.isNew ) return qsTr ( "Edit folder %1"    ) .arg( root.currentName )
                if ( !root.isFolder && !root.isNew ) return qsTr ( "Edit label %1"     ) .arg( root.currentName )
            }

            width  : parent.width
            elide  : Text.ElideRight

            horizontalAlignment : Text.AlignHCenter

            Rectangle {
                anchors {
                    top: parent.bottom
                    topMargin: Style.dialog.spacing
                    horizontalCenter: parent.horizontalCenter
                }
                color: textColor
                height: Style.main.borderInput
            }
        }

        TextField {
            id: inputField

            anchors {
                horizontalCenter: parent.horizontalCenter
            }

            width          : parent.width
            height         : Style.dialog.button
            rightPadding   : Style.dialog.spacing
            leftPadding    : height + rightPadding
            bottomPadding  : rightPadding
            topPadding     : rightPadding
            selectByMouse  : true
            color          : textColor
            font.pointSize : Style.dialog.fontSize * Style.pt

            background: Rectangle {
                color: backColor
                border {
                    color: textColor
                    width: Style.dialog.borderInput
                }

                radius : Style.dialog.radiusButton

                Text {
                    anchors {
                        left: parent.left
                        verticalCenter: parent.verticalCenter
                    }

                    font {
                        family: Style.fontawesome.name
                        pointSize: Style.dialog.titleSize * Style.pt
                    }

                    text  : folderType == gui.enums.folderTypeFolder ? Style.fa.folder : Style.fa.tag
                    color : root.selectedColor
                    width : parent.height
                    horizontalAlignment: Text.AlignHCenter
                }

                Rectangle {
                    anchors {
                        left: parent.left
                        top: parent.top
                        leftMargin: parent.height
                    }
                    width: parent.border.width/2
                    height: parent.height
                }
            }
        }


        GridView {
            id: view

            anchors {
                horizontalCenter: parent.horizontalCenter
            }

            model      : colorList
            cellWidth  : 2*Style.dialog.titleSize
            cellHeight : cellWidth
            width      : 10*cellWidth
            height     : 2*cellHeight

            delegate: Rectangle {
                width: view.cellWidth*0.8
                height: width
                radius: width/2
                color: modelData

                border {
                    color: indicator.visible ? textColor : modelData
                    width: 2*Style.px
                }

                Text {
                    id: indicator
                    anchors.centerIn : parent
                    text: Style.fa.check
                    color: textColor
                    font {
                        family: Style.fontawesome.name
                        pointSize: Style.dialog.titleSize * Style.pt
                    }
                    visible: modelData == root.selectedColor
                }

                MouseArea {
                    anchors.fill: parent
                    onClicked : {
                        root.selectedColor = modelData
                    }
                }
            }
        }

        Row {
            id: buttonRow

            anchors {
                horizontalCenter: parent.horizontalCenter
            }

            spacing: Style.main.leftMargin

            ButtonRounded {
                text: "Cancel"
                color_main : textColor
                onClicked :{
                    root.hide()
                }
            }

            ButtonRounded {
                text: "Okay"
                color_main: Style.dialog.background
                color_minor: Style.dialog.textBlue
                isOpaque: true
                onClicked :{
                    root.okay()
                }
            }
        }
    }

    function hide() {
        root.visible=false
        root.currentId = ""
        root.currentName = ""
        root.currentColor = ""
        root.folderType = ""
        root.sourceID = ""
        inputField.text = ""
    }

    function show(currentName, currentId, currentColor, folderType, sourceID) {
        root.currentId = currentId
        root.currentName = currentName
        root.currentColor = currentColor=="" ? go.leastUsedColor() : currentColor
        root.selectedColor =  root.currentColor
        root.folderType = folderType
        root.sourceID = sourceID

        inputField.text = currentName
        root.visible=true
        //console.log(title.text , root.currentName, root.currentId, root.currentColor, root.folderType, root.sourceID) 
    }

    function okay() {
        // check inpupts
        if (inputField.text == "") {
            go.notifyError(gui.enums.errFillFolderName)
            return
        }
        if (colorList.indexOf(root.selectedColor)<0) {
            go.notifyError(gui.enums.errSelectFolderColor)
            return
        }
        var isLabel = root.folderType == gui.enums.folderTypeLabel
        if (!isLabel && !root.isFolder){
            console.log("Unknown folder type: ", root.folderType)
            go.notifyError(gui.enums.errUpdateLabelFailed)
            root.hide()
            return
        }

        if (winMain.dialogImport.address == "") {
            console.log("Unknown address", winMain.dialogImport.address)
            go.onNotifyError(gui.enums.errUpdateLabelFailed)
            root.hide()
        }

        if (root.isCreateLater) {
            root.edited(inputField.text, root.selectedColor)
            root.hide()
            return
        }


        // TODO send request (as timer)
        if (root.isNew) {
            var isOK = go.createLabelOrFolder(winMain.dialogImport.address, inputField.text, root.selectedColor, isLabel, root.sourceID)
            if (isOK) {
                root.hide()
            }
        } else {
            // TODO: check there was some change
            go.updateLabelOrFolder(winMain.dialogImport.address, root.currentId, inputField.text, root.selectedColor)
        }

        // waiting for finish
        // TODO: waiting wheel of doom
        // TODO: on close add source to sourceID
    }
}
