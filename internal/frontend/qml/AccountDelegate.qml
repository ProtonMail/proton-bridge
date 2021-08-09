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

import QtQuick 2.13
import QtQuick.Layouts 1.12
import QtQuick.Controls 2.12

import Proton 4.0

Item {
    id: root

    property ColorScheme colorScheme
    property var user

    property var _spacing: 12
    property var _leftRightMargins: {
        switch(root.type) {
            case AccountDelegate.SmallView: return 12
            case AccountDelegate.LargeView: return 0
        }
    }
    property var _topBottomMargins: {
        switch(root.type) {
            case AccountDelegate.SmallView: return 10
            case AccountDelegate.LargeView: return 0
        }
    }

    property color usedSpaceColor : {
        if (!root.enabled) return root.colorScheme.text_weak
        if (root.type == AccountDelegate.SmallView) return root.colorScheme.text_weak
        if (root.usedFraction < .50) return root.colorScheme.signal_success
        if (root.usedFraction < .75) return root.colorScheme.signal_warning
        return root.colorScheme.signal_danger
    }
    property real usedFraction: root.user.totalBytes ? root.user.usedBytes / root.user.totalBytes : 0
    property string totalSpace: root.spaceWithUnits(root.user.totalBytes)
    property string usedSpace: root.spaceWithUnits(root.user.usedBytes)

    function spaceWithUnits(bytes){
        if (bytes*1 !== bytes ) return "0 kB"
        var units = ['B',"kB", "MB", "TB"];
        var i = parseInt(Math.floor(Math.log(bytes)/Math.log(1024)));

        return Math.round(bytes*10 / Math.pow(1024, i))/10 + " " + units[i]
    }

    signal clicked()

    // width expected to be set by parent object
    implicitHeight : children[0].implicitHeight + 2*root._topBottomMargins

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
            leftMargin   : root._leftRightMargins
            rightMargin  : root._leftRightMargins
            topMargin    : root._topBottomMargins
            bottomMargin : root._topBottomMargins
        }

        Rectangle {
            id: avatar

            Layout.fillHeight: true
            Layout.preferredWidth: height

            radius: 4

            color: root.colorScheme.background_avatar

            Label {
                colorScheme: root.colorScheme
                anchors.fill: parent
                text: root.user.avatarText.toUpperCase()
                type: {
                    switch (root.type) {
                        case AccountDelegate.SmallView: return Label.Body
                        case AccountDelegate.LargeView: return Label.Title
                    }
                }
                font.weight: Font.Normal
                color: {
                    switch(root.type) {
                        case AccountDelegate.SmallView: return root.colorScheme.text_norm
                        case AccountDelegate.LargeView: return root.colorScheme.text_invert
                    }
                }
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
                    root._spacing + avatar.width + 2*root._leftRightMargins
                )

                colorScheme: root.colorScheme
                text: user.username
                type: {
                    switch (root.type) {
                        case AccountDelegate.SmallView: return Label.Body
                        case AccountDelegate.LargeView: return Label.Title
                    }
                }
                elide: Text.ElideMiddle
            }

            Item { implicitHeight: root.type == AccountDelegate.LargeView ? 6 : 0 }

            RowLayout {
                Label {
                    colorScheme: root.colorScheme
                    text: root.usedSpace
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
                    text: " / " + root.totalSpace
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
                visible: root.type == AccountDelegate.LargeView

                width: 140
                height: 4
                radius: 3
                color: root.colorScheme.border_weak

                Rectangle {
                    radius: 3
                    color: root.usedSpaceColor
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

    MouseArea {
        anchors.fill: root
        onClicked: root.clicked()
    }
}
