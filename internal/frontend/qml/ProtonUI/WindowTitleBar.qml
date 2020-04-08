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

// simulating window title bar with different color

import QtQuick 2.8
import QtQuick.Window 2.2
import ProtonUI 1.0

Rectangle {
    id: root
    height: root.isDarwin ? Style.titleMacOS.height : Style.title.height
    color: "transparent"
    property bool isDarwin : (go.goos == "darwin")
    property QtObject window
    anchors {
        left  : parent.left
        right : parent.right
        top   : parent.top
    }

    MouseArea {
        property point diff: "0,0"
        anchors {
            top: root.top
            bottom: root.bottom
            left: root.left
            right: root.isDarwin ? root.right : iconRowWin.left
        }
        onPressed: {
            diff = Qt.point(window.x, window.y)
            var mousePos = mapToGlobal(mouse.x, mouse.y)
            diff.x -= mousePos.x
            diff.y -= mousePos.y
        }
        onPositionChanged: {
            var currPos = mapToGlobal(mouse.x, mouse.y)
            window.x = currPos.x + diff.x
            window.y = currPos.y + diff.y
        }
    }

    // top background
    Rectangle {
        id: upperBackground
        anchors.fill: root
        color: (isDarwin? Style.titleMacOS.background : Style.title.background )
        radius: (isDarwin? Style.titleMacOS.radius : 0)
        border {
            width: Style.main.border
            color: Style.title.background
        }
    }
    // bottom background
    Rectangle {
        id: lowerBorder
        anchors {
            top: root.verticalCenter
            left: root.left
            right: root.right
            bottom: root.bottom
        }
        color: Style.title.background
        Rectangle {
            id: lowerBackground
            anchors{
                fill        : parent
                leftMargin  : Style.main.border
                rightMargin : Style.main.border
            }
            color: upperBackground.color

        }
    }

    // Title
    TextMetrics {
        id: titleMetrics
        text : window.title
        font : isDarwin ? titleMac.font : titleWin.font
        elide: Qt.ElideMiddle
        elideWidth : window.width/2
    }
    Text {
        id: titleWin
        visible: !isDarwin
        anchors {
            baseline   : logo.bottom
            left       : logo.right
            leftMargin : Style.title.leftMargin/1.5
        }
        color          : window.active ? Style.title.text : Style.main.textDisabled
        text           : titleMetrics.elidedText
        font.pointSize : Style.main.fontSize * Style.pt
    }
    Text {
        id: titleMac
        visible: isDarwin
        anchors {
            verticalCenter : parent.verticalCenter
            left           : parent.left
            leftMargin     : (parent.width-width)/2
        }
        color          : window.active ? Style.title.text : Style.main.textDisabled
        text           : titleMetrics.elidedText
        font.pointSize : Style.main.fontSize * Style.pt
    }


    // MACOS
    MouseArea {
        anchors.fill: iconRowMac
        property string beforeHover
        hoverEnabled: true
        onEntered: {
            beforeHover=iconRed.state
            //iconYellow.state="hover"
            iconRed.state="hover"
        }
        onExited: {
            //iconYellow.state=beforeHover
            iconRed.state=beforeHover
        }
    }
    Connections {
        target: window
        onActiveChanged : {
            if (window.active) {
                //iconYellow.state="normal"
                iconRed.state="normal"
            } else {
                //iconYellow.state="disabled"
                iconRed.state="disabled"
            }
        }
    }
    Row {
        id: iconRowMac
        visible : isDarwin
        spacing : Style.titleMacOS.leftMargin
        anchors {
            left           : parent.left
            verticalCenter : parent.verticalCenter
            leftMargin     : Style.title.leftMargin
        }
        Image {
            id: iconRed
            width    : Style.titleMacOS.imgHeight
            height   : Style.titleMacOS.imgHeight
            fillMode : Image.PreserveAspectFit
            smooth   : true
            state    : "normal"
            states: [
                State { name: "normal"   ; PropertyChanges { target: iconRed ; source: "images/macos_red.png"      } },
                State { name: "hover"    ; PropertyChanges { target: iconRed ; source: "images/macos_red_hl.png"   } },
                State { name: "pressed"  ; PropertyChanges { target: iconRed ; source: "images/macos_red_dark.png" } },
                State { name: "disabled" ; PropertyChanges { target: iconRed ; source: "images/macos_gray.png"     } }
            ]
            MouseArea {
                anchors.fill: parent
                property string beforePressed : "normal"
                onClicked : {
                    window.close()
                }
                onPressed: {
                    beforePressed = parent.state
                    parent.state="pressed"
                }
                onReleased: {
                    parent.state=beforePressed
                }
                Accessible.role: Accessible.Button
                Accessible.name: qsTr("Close", "Close the window button")
                Accessible.description: Accessible.name
                Accessible.ignored: !parent.visible
                Accessible.onPressAction: {
                    window.close()
                }
            }
        }
        Image {
            id: iconYellow
            width    : Style.titleMacOS.imgHeight
            height   : Style.titleMacOS.imgHeight
            fillMode : Image.PreserveAspectFit
            smooth   : true
            state    : "disabled"
            states: [
                State { name: "normal"   ; PropertyChanges { target: iconYellow ; source: "images/macos_yellow.png"      } },
                State { name: "hover"    ; PropertyChanges { target: iconYellow ; source: "images/macos_yellow_hl.png"   } },
                State { name: "pressed"  ; PropertyChanges { target: iconYellow ; source: "images/macos_yellow_dark.png" } },
                State { name: "disabled" ; PropertyChanges { target: iconYellow ; source: "images/macos_gray.png"        } }
            ]
            /*
             MouseArea {
                 anchors.fill: parent
                 property string beforePressed : "normal"
                 onClicked : {
                     window.visibility = Window.Minimized
                 }
                 onPressed: {
                     beforePressed = parent.state
                     parent.state="pressed"
                 }
                 onReleased: {
                     parent.state=beforePressed
                 }

                 Accessible.role: Accessible.Button
                 Accessible.name: qsTr("Minimize", "Minimize the window button")
                 Accessible.description: Accessible.name
                 Accessible.ignored: !parent.visible
                 Accessible.onPressAction: {
                     window.visibility = Window.Minimized
                 }
             }
             */
        }
        Image {
            id: iconGreen
            width    : Style.titleMacOS.imgHeight
            height   : Style.titleMacOS.imgHeight
            fillMode : Image.PreserveAspectFit
            smooth   : true
            source   : "images/macos_gray.png"
            Component.onCompleted : {
                visible = false // (window.flags&Qt.Dialog) != Qt.Dialog
            }
        }
    }


    // Windows
    Image {
        id: logo
        visible: !isDarwin
        anchors {
            left           : parent.left
            verticalCenter : parent.verticalCenter
            leftMargin     : Style.title.leftMargin
        }
        height   : Style.title.fontSize-2*Style.px
        fillMode : Image.PreserveAspectFit
        mipmap   : true
        source   : "images/pm_logo.png"
    }

    Row {
        id: iconRowWin
        visible: !isDarwin
        anchors {
            right          : parent.right
            verticalCenter : root.verticalCenter
        }
        Rectangle {
            height : root.height
            width  : 1.5*height
            color: Style.transparent
            Image {
                id: iconDash
                anchors.centerIn: parent
                height   : iconTimes.height*0.90
                fillMode : Image.PreserveAspectFit
                mipmap   : true
                source   : "images/win10_Dash.png"
            }
            MouseArea {
                anchors.fill : parent
                hoverEnabled : true
                onClicked    : {
                    window.visibility = Window.Minimized
                }
                onPressed: {
                    parent.scale=0.92
                }
                onReleased: {
                    parent.scale=1
                }
                onEntered: {
                    parent.color= Qt.lighter(Style.title.background,1.2)
                }
                onExited: {
                    parent.color=Style.transparent
                }

                Accessible.role          : Accessible.Button
                Accessible.name          : qsTr("Minimize", "Minimize the window button")
                Accessible.description   : Accessible.name
                Accessible.ignored       : !parent.visible
                Accessible.onPressAction : {
                    window.visibility = Window.Minimized
                }
            }
        }
        Rectangle {
            height : root.height
            width  : 1.5*height
            color : Style.transparent
            Image {
                id: iconTimes
                anchors.centerIn : parent
                mipmap   : true
                height   : parent.height/1.5
                fillMode : Image.PreserveAspectFit
                source   : "images/win10_Times.png"
            }
            MouseArea {
                anchors.fill : parent
                hoverEnabled : true
                onClicked    : window.close()
                onPressed    : {
                    iconTimes.scale=0.92
                }
                onReleased: {
                    parent.scale=1
                }
                onEntered: {
                    parent.color=Style.main.textRed
                }
                onExited: {
                    parent.color=Style.transparent
                }

                Accessible.role          : Accessible.Button
                Accessible.name          : qsTr("Close", "Close the window button")
                Accessible.description   : Accessible.name
                Accessible.ignored       : !parent.visible
                Accessible.onPressAction : {
                    window.close()
                }
            }
        }
    }
}
