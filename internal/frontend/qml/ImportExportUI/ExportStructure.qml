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

// List of export folders / labels
import QtQuick 2.8
import QtQuick.Controls 2.2
import ProtonUI 1.0
import ImportExportUI 1.0

Rectangle {
    id: root
    color     : Style.exporting.background
    radius    : Style.exporting.boxRadius
    border {
        color : Style.exporting.line
        width : Style.dialog.borderInput
    }
    property bool hasItems: true


    Text { // placeholder
        visible: !root.hasItems
        anchors.centerIn: parent
        color: Style.main.textDisabled
        font {
            pointSize: Style.dialog.fontSize * Style.pt
        }
        horizontalAlignment: Text.AlignHCenter
        verticalAlignment: Text.AlignVCenter
        text: qsTr("No emails found for this address.","todo")
    }


    property string title  : ""

    TextMetrics {
        id: titleMetrics
        text: root.title
        elide: Qt.ElideMiddle
        elideWidth: root.width - 4*Style.exporting.leftMargin
        font {
            pointSize: Style.dialog.fontSize * Style.pt
            bold: true
        }
    }

    Rectangle {
        id: header
        anchors {
            top: root.top
            left: root.left
        }
        width  : root.width
        height : Style.dialog.fontSize*3
        color  : Style.transparent
        Rectangle {
            anchors.bottom: parent.bottom
            color  : Style.exporting.line
            height : Style.dialog.borderInput
            width  : parent.width
        }

        Text {
            anchors {
                left           : parent.left
                leftMargin     : 2*Style.exporting.leftMargin
                verticalCenter : parent.verticalCenter
            }
            color: Style.dialog.text
            font: titleMetrics.font
            text: titleMetrics.elidedText
        }
    }


    ListView {
        id: listview
        clip           : true
        orientation    : ListView.Vertical
        boundsBehavior : Flickable.StopAtBounds
        model          : transferRules
        cacheBuffer    : 10000

        anchors {
            left    : root.left
            right   : root.right
            bottom  : root.bottom
            top     : header.bottom
            margins : Style.dialog.borderInput
        }

        ScrollBar.vertical: ScrollBar {
            /*
             policy: ScrollBar.AsNeeded
             background : Rectangle {
                 color  : Style.exporting.sliderBackground
                 radius : Style.exporting.boxRadius
             }
             contentItem : Rectangle {
                 color         : Style.exporting.sliderForeground
                 radius        : Style.exporting.boxRadius
                 implicitWidth : Style.main.rightMargin / 3
             }
             */
            anchors {
                right: parent.right
                rightMargin: Style.main.rightMargin/4
            }
            width: Style.main.rightMargin/3
            Accessible.ignored: true
        }

        delegate: FolderRowButton {
            property variant modelData: model
            width      : root.width - 5*root.border.width
            type       : modelData.type
            folderIconColor  : modelData.iconColor
            title      : modelData.name
            isSelected : modelData.isActive
            onClicked  : {
                //console.log("Clicked", folderId, isSelected)
                transferRules.setIsRuleActive(modelData.mboxID,!model.isActive)
            }
        }

        section.property: "type"
        section.delegate: FolderRowButton {
            isSection  : true
            width      : root.width - 5*root.border.width
            title      : gui.folderTypeTitle(section)
            isSelected : section == gui.enums.folderTypeLabel ? transferRules.isLabelGroupSelected : transferRules.isFolderGroupSelected
            onClicked  : transferRules.setIsGroupActive(section,!isSelected)
        }
    }
}
