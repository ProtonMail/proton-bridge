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
import QtQuick.Layouts 1.12
import QtQuick.Controls.impl 2.12

Rectangle {
    id: root

    width: layout.width
    height: layout.height

    radius: 10

    signal accepted()


    property alias text: description.text
    property var actionText: ""

    property var colorText: Style.currentStyle.text_invert
    property var colorMain: "#000"
    property var colorHover: "#000"
    property var colorActive: "#000"
    property var iconSource: "../icons/ic-exclamation-circle-filled.svg"

    color: root.colorMain
    border.color: root.colorActive
    border.width: 1

    property var maxWidth: 600
    property var minWidth: 400
    property var usedWidth: button.width + icon.width

    RowLayout {
        id: layout

        IconLabel {
            id:icon
            Layout.alignment: Qt.AlignCenter
            Layout.leftMargin: 17.5
            Layout.topMargin: 15.5
            Layout.bottomMargin: 15.5
            color: root.colorText
            icon.source: root.iconSource
            icon.color: root.colorText
            icon.height: Style.title_line_height
        }

        ProtonLabel {
            id: description
            Layout.alignment: Qt.AlignCenter
            Layout.leftMargin: 9.5
            Layout.minimumWidth: root.minWidth - root.usedWidth
            Layout.maximumWidth: root.maxWidth - root.usedWidth

            color: root.colorText
            state: "body"

            wrapMode: Text.WordWrap
            verticalAlignment: Text.AlignVCenter
        }

        Button {
            id:button
            Layout.fillHeight: true

            hoverEnabled: true

            text: root.actionText.toUpperCase()

            onClicked: root.accepted()

            background: RoundedRectangle {
                width:parent.width
                height:parent.height
                strokeColor: root.colorActive
                strokeWidth: root.border.width

                radiusTopRight    : root.radius
                radiusBottomRight : root.radius
                radiusTopLeft     : 0
                radiusBottomLeft  : 0

                fillColor: button.down ? root.colorActive : (
                    button.hovered ? root.colorHover :
                    root.colorMain
                )
            }
        }
    }

    state: "info"
    states: [
        State{ name : "danger"  ; PropertyChanges{ target : root ; colorMain : Style.currentStyle.signal_danger  ; colorHover : Style.currentStyle.signal_danger_hover  ; colorActive : Style.currentStyle.signal_danger_active  ; iconSource: "../icons/ic-exclamation-circle-filled.svg"}} ,
        State{ name : "warning" ; PropertyChanges{ target : root ; colorMain : Style.currentStyle.signal_warning ; colorHover : Style.currentStyle.signal_warning_hover ; colorActive : Style.currentStyle.signal_warning_active ; iconSource: "../icons/ic-exclamation-circle-filled.svg"}} ,
        State{ name : "success" ; PropertyChanges{ target : root ; colorMain : Style.currentStyle.signal_success ; colorHover : Style.currentStyle.signal_success_hover ; colorActive : Style.currentStyle.signal_success_active ; iconSource: "../icons/ic-info-circle-filled.svg"}} ,
        State{ name : "info"    ; PropertyChanges{ target : root ; colorMain : Style.currentStyle.signal_info    ; colorHover : Style.currentStyle.signal_info_hover    ; colorActive : Style.currentStyle.signal_info_active    ; iconSource: "../icons/ic-info-circle-filled.svg"}}
    ]
}
