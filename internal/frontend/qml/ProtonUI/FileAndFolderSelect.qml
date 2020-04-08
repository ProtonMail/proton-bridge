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

// one line input text field with label
import QtQuick 2.8
import QtQuick.Controls 2.2
import QtQuick.Dialogs 1.0
import ProtonUI 1.0

Row {
    id: root
    spacing: Style.dialog.spacing

    property string title : "title"

    property alias path: inputPath.text
    property alias inputPath: inputPath
    property alias dialogVisible: pathDialog.visible

    InputBox {
        id: inputPath
        anchors {
            bottom: parent.bottom
        }
        spacing: Style.dialog.spacing
        field {
            height: browseButton.height
            width: root.width - root.spacing - browseButton.width
        }

        label: title
        Component.onCompleted: sanitizePath(pathDialog.shortcuts.home)
    }

    ButtonRounded {
        id: browseButton
        anchors {
            bottom: parent.bottom
        }
        height: Style.dialog.heightInput
        color_main: Style.main.textBlue
        fa_icon: Style.fa.folder_open
        text: qsTr("Browse", "click to look through directory for a file or folder")
        onClicked: pathDialog.visible = true
    }

    FileDialog {
        id: pathDialog
        title: root.title + ":"
        folder: shortcuts.home
        onAccepted: sanitizePath(pathDialog.fileUrl.toString())
        selectFolder: true
    }

    function  sanitizePath(path) {
        var pattern = "file://"
        if (go.goos=="windows") pattern+="/"
        inputPath.text = path.replace(pattern, "")
    }

    function checkNonEmpty() {
        return inputPath.text != ""
    }
}
