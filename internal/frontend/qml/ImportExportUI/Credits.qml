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

// credits

import QtQuick 2.8
import ProtonUI 1.0
import ImportExportUI 1.0

Item {
    id: root
    Rectangle {
        anchors.centerIn: parent
        width: Style.main.width
        height: root.parent.height - 6*Style.dialog.titleSize
        color: "transparent"

        ListView {
            anchors.fill: parent
            clip: true

            model: [
                "github.com/0xAX/notificator" ,
                "github.com/abiosoft/ishell" ,
                "github.com/allan-simon/go-singleinstance" ,
                "github.com/andybalholm/cascadia" ,
                "github.com/bgentry/speakeasy" ,
                "github.com/boltdb/bolt" ,
                "github.com/docker/docker-credential-helpers" ,
                "github.com/emersion/go-imap" ,
                "github.com/emersion/go-imap-appendlimit" ,
                "github.com/emersion/go-imap-idle" ,
                "github.com/emersion/go-imap-move" ,
                "github.com/emersion/go-imap-quota" ,
                "github.com/emersion/go-imap-specialuse" ,
                "github.com/emersion/go-smtp" ,
                "github.com/emersion/go-textwrapper" ,
                "github.com/fsnotify/fsnotify" ,
                "github.com/jaytaylor/html2text" ,
                "github.com/jhillyerd/go.enmime" ,
                "github.com/k0kubun/pp" ,
                "github.com/kardianos/osext" ,
                "github.com/keybase/go-keychain" ,
                "github.com/mattn/go-colorable" ,
                "github.com/pkg/browser" ,
                "github.com/shibukawa/localsocket" ,
                "github.com/shibukawa/tobubus" ,
                "github.com/shirou/gopsutil" ,
                "github.com/sirupsen/logrus" ,
                "github.com/skratchdot/open-golang/open" ,
                "github.com/therecipe/qt" ,
                "github.com/thomasf/systray" ,
                "github.com/ugorji/go/codec" ,
                "github.com/urfave/cli" ,
                "" ,
                "Font Awesome 4.7.0",
                "" ,
                "The Qt Company - Qt 5.9.1 LGPLv3" ,
                "" ,
            ]

            delegate: Text {
                anchors.horizontalCenter: parent.horizontalCenter
                text: modelData
                color: Style.main.text
            }

            footer: ButtonRounded {
                anchors.horizontalCenter: parent.horizontalCenter
                text: "Close"
                onClicked: {
                    root.parent.hide()
                }
            }
        }
    }
}

