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

import QtQml 2.12
import QtQuick 2.13
import Proton 4.0
import QtQuick.Controls 2.13

Rectangle {
    id: root

    property var window

    property bool onTop: false
    property bool blocking: root.nDangers != 0
    property int  nDangers: 0

    color: root.getTransparentVersion(window.colorScheme.text_norm,root.blocking ? 0.5 : 0)

    MouseArea {
        anchors.fill: root
        acceptedButtons: root.blocking ? Qt.AllButtons : Qt.NoButton
        enabled: root.blocking
    }

    ListModel {
        id: notifications
    }

    ListView {
        id: view
        anchors.top              : root.top
        anchors.bottom           : root.bottom
        anchors.horizontalCenter : root.horizontalCenter
        anchors.topMargin        : root.height/20
        anchors.bottomMargin     : root.height/20

        layoutDirection: ListView.Vertical
        verticalLayoutDirection: root.onTop ? ListView.TopToBottom : ListView.BottomToTop

        spacing: 5

        model: notifications
        delegate: Banner {
            id: bannerDelegate
            anchors.horizontalCenter: parent.horizontalCenter
            text: model.text
            actionText: model.buttonText
            state: model.state

            onAccepted: {
                switch (model.submitAction) {
                    case "update":
                    console.log("I am updating now")
                    break;
                    default:
                    console.log("NOOP")
                }
                if (model.state == "danger") root.nDangers-=1
                anchors.horizontalCenter = undefined
                notifications.remove(index)
            }
        }
    }

    function notify(descriptionText, buttonText, type = "info", submitAction = "noop") {
        if (type === "danger") root.nDangers+=1
        notifications.append({
            "text": descriptionText,
            "buttonText": buttonText,
            "state": type,
            "submitAction": submitAction
        })
    }

    function notifyOnlyPaidUsers(){
        root.notify(
            qsTr("Bridge is exclusive to our paid plans. Upgrade your account to use Bridge."),
            qsTr("ok"), "danger"
        )
    }

    function notifyConnectionLostWhileLogin(){
        root.notify(
            qsTr("Can't connect to the server. Check your internet connection and try again."),
            qsTr("ok"), "danger"
        )
    }

    function notifyUpdateManually(){
        root.notify(
            qsTr("Bridge could not update automatically."),
            qsTr("update"), "warning", "update"
        )
    }

    function notifyUserAdded(){
        root.notify(
            qsTr("Your account has been added to Bridge and you are now signed in."),
            qsTr("ok"), "success"
        )
    }

    function getTransparentVersion(original, transparency){
        return Qt.rgba(original.r, original.g, original.b, transparency)
    }
}
