// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

import QtQuick
import QtQuick.Layouts
import QtQuick.Controls

import Proton

Item {
    id: root

    property ColorScheme colorScheme
    property var user

    property var _spacing: 12 * ProtonStyle.px

    property color usedSpaceColor : {
        if (!root.enabled) return root.colorScheme.text_weak
        if (root.type == AccountDelegate.SmallView) return root.colorScheme.text_weak
        if (root.usedFraction < .50) return root.colorScheme.signal_success
        if (root.usedFraction < .75) return root.colorScheme.signal_warning
        return root.colorScheme.signal_danger
    }
    property real usedFraction: root.user ? reasonableFracion(root.user.usedBytes, root.user.totalBytes) : 0
    property string totalSpace: root.spaceWithUnits(root.user ? root.reasonableBytes(root.user.totalBytes) : 0)
    property string usedSpace: root.spaceWithUnits(root.user ? root.reasonableBytes(root.user.usedBytes) : 0)

    function reasonableFracion(used, total){
        var usedSafe = root.reasonableBytes(used)
        var totalSafe = root.reasonableBytes(total)
        if (totalSafe == 0 || usedSafe == 0) return 0
        if (totalSafe <= usedSafe) return 1
        return usedSafe / totalSafe
    }

    function reasonableBytes(bytes){
        var safeBytes = bytes+0
        if (safeBytes != bytes) return 0
        if (safeBytes < 0) return 0
        return Math.ceil(safeBytes)
    }

    function spaceWithUnits(bytes){
        if (bytes*1 !== bytes || bytes == 0 ) return "0 kB"
        var units = ['B',"kB", "MB", "GB", "TB"];
        var i = parseInt(Math.floor(Math.log(bytes)/Math.log(1024)));

        return Math.round(bytes*10 / Math.pow(1024, i))/10 + " " + units[i]
    }

    // width expected to be set by parent object
    implicitHeight : children[0].implicitHeight

    enum ViewType{
        SmallView, LargeView
    }
    property var type : AccountDelegate.SmallView

    RowLayout {
        spacing: root._spacing

        anchors {
            top: root.top
            left: root.left
            right: root.rigth
        }

        Rectangle {
            id: avatar

            Layout.fillHeight: true
            Layout.preferredWidth: height

            radius: ProtonStyle.avatar_radius

            color: root.colorScheme.background_avatar

            Label {
                colorScheme: root.colorScheme
                anchors.fill: parent
                text: root.user ? root.user.avatarText.toUpperCase(): ""
                type: {
                    switch (root.type) {
                        case AccountDelegate.SmallView: return Label.Body
                        case AccountDelegate.LargeView: return Label.Title
                    }
                }
                font.weight: Font.Normal
                color: "#FFFFFF"
                horizontalAlignment: Qt.AlignHCenter
                verticalAlignment: Qt.AlignVCenter
            }
        }

        ColumnLayout {
            id: account
            Layout.fillHeight: true
            Layout.fillWidth: true

            spacing: 0

            Label {
                Layout.maximumWidth: root.width - (
                    root._spacing + avatar.width
                )

                colorScheme: root.colorScheme
                text: root.user ? user.username : ""
                type: {
                    switch (root.type) {
                        case AccountDelegate.SmallView: return Label.Body
                        case AccountDelegate.LargeView: return Label.Title
                    }
                }
                elide: Text.ElideMiddle
            }

            Item { implicitHeight: root.type == AccountDelegate.LargeView ? 6 * ProtonStyle.px : 0 }

            RowLayout {
                spacing: 0
                Label {
                    colorScheme: root.colorScheme
                    text: root.user && root.user.loggedIn ? root.usedSpace : qsTr("Signed out")
                    color: root.usedSpaceColor
                    type: {
                        switch (root.type) {
                            case AccountDelegate.SmallView: return Label.Caption
                            case AccountDelegate.LargeView: return Label.Body
                        }
                    }
                }

                Label {
                    colorScheme: root.colorScheme
                    text: root.user && root.user.loggedIn ? " / " + root.totalSpace : ""
                    color: root.colorScheme.text_weak
                    type: {
                        switch (root.type) {
                            case AccountDelegate.SmallView: return Label.Caption
                            case AccountDelegate.LargeView: return Label.Body
                        }
                    }
                }
            }


            Rectangle {
                id: storage_bar
                visible: root.user ? root.type == AccountDelegate.LargeView : false
                width: 140 * ProtonStyle.px
                height: 4 * ProtonStyle.px
                radius: ProtonStyle.storage_bar_radius
                color: root.colorScheme.border_weak

                Rectangle {
                    id: storage_bar_filled
                    radius: ProtonStyle.storage_bar_radius
                    color: root.usedSpaceColor
                    visible: root.user ? parent.visible && root.user.loggedIn : false
                    anchors {
                        top    : parent.top
                        bottom : parent.bottom
                        left   : parent.left
                    }
                    width: Math.min(1,Math.max(0.02,root.usedFraction)) * parent.width
                }
            }
        }

        Item {
            Layout.fillWidth: true
        }
    }
}
