// Copyright (c) 2023 Proton AG
// This file is part of Proton Mail Bridge.
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.
import QtQuick
import QtQuick.Layouts
import QtQuick.Controls
import Proton

Item {
    id: root
    enum ViewType {
        SmallView,
        LargeView
    }

    property var _spacing: 12
    property ColorScheme colorScheme
    property color progressColor: {
        if (!root.enabled)
            return root.colorScheme.text_weak;
        if (root.type === AccountDelegate.SmallView)
            return root.colorScheme.text_weak;
        if (root.user && root.user.isSyncing)
            return root.colorScheme.text_weak;
        if (root.progressRatio < .50)
            return root.colorScheme.signal_success;
        if (root.progressRatio < .75)
            return root.colorScheme.signal_warning;
        return root.colorScheme.signal_danger;
    }
    property real progressRatio: {
        if (!root.user)
            return 0;
        return root.user.isSyncing ? root.user.syncProgress : reasonableFraction(root.user.usedBytes, root.user.totalBytes);
    }
    property string totalSpace: root.spaceWithUnits(root.user ? root.reasonableBytes(root.user.totalBytes) : 0)
    property var type: AccountDelegate.SmallView
    property string usedSpace: root.spaceWithUnits(root.user ? root.reasonableBytes(root.user.usedBytes) : 0)
    property var user

    function primaryEmail() {
        return root.user ? root.user.primaryEmailOrUsername() : "";
    }
    function reasonableBytes(bytes) {
        const safeBytes = bytes + 0;
        if (safeBytes !== bytes)
            return 0;
        if (safeBytes < 0)
            return 0;
        return Math.ceil(safeBytes);
    }
    function reasonableFraction(used, total) {
        const usedSafe = root.reasonableBytes(used);
        const totalSafe = root.reasonableBytes(total);
        if (totalSafe === 0 || usedSafe === 0)
            return 0;
        if (totalSafe <= usedSafe)
            return 1;
        return usedSafe / totalSafe;
    }
    function spaceWithUnits(bytes) {
        if (bytes * 1 !== bytes || bytes === 0)
            return "0 kB";
        const units = ['B', "kB", "MB", "GB", "TB"];
        const i = Math.floor(Math.log(bytes) / Math.log(1024));
        return Math.round(bytes * 10 / Math.pow(1024, i)) / 10 + " " + units[i];
    }

    // width expected to be set by parent object
    implicitHeight: children[0].implicitHeight

    RowLayout {
        spacing: root._spacing

        anchors {
            left: root.left
            right: root.right
            top: root.top
        }
        Rectangle {
            id: avatar
            Layout.fillHeight: true
            Layout.preferredWidth: height
            color: root.colorScheme.background_avatar
            radius: ProtonStyle.avatar_radius

            Label {
                anchors.fill: parent
                color: "#FFFFFF"
                colorScheme: root.colorScheme
                font.weight: Font.Normal
                horizontalAlignment: Qt.AlignHCenter
                text: root.user ? root.user.avatarText.toUpperCase() : ""
                type: {
                    switch (root.type) {
                    case AccountDelegate.SmallView:
                        return Label.Body;
                    case AccountDelegate.LargeView:
                        return Label.Title;
                    }
                }
                verticalAlignment: Qt.AlignVCenter
            }
        }
        ColumnLayout {
            id: account
            Layout.fillHeight: true
            Layout.fillWidth: true
            spacing: 0

            Label {
                id: labelEmail
                Layout.maximumWidth: root.width - (root._spacing + avatar.width)
                colorScheme: root.colorScheme
                elide: Text.ElideMiddle
                text: primaryEmail()
                type: {
                    switch (root.type) {
                    case AccountDelegate.SmallView:
                        return Label.Body;
                    case AccountDelegate.LargeView:
                        return Label.Title;
                    }
                }

                MouseArea {
                    id: labelArea
                    anchors.fill: parent
                    hoverEnabled: true
                }
                ToolTip {
                    id: toolTipEmail
                    delay: 1000
                    text: primaryEmail()
                    visible: labelArea.containsMouse && labelEmail.truncated

                    background: Rectangle {
                        border.color: root.colorScheme.background_strong
                        color: root.colorScheme.background_norm
                    }
                    contentItem: Text {
                        color: root.colorScheme.text_norm
                        text: toolTipEmail.text
                    }
                }
            }
            Item {
                implicitHeight: root.type === AccountDelegate.LargeView ? 6 : 0
            }
            RowLayout {
                spacing: 0

                Label {
                    color: root.progressColor
                    colorScheme: root.colorScheme
                    text: {
                        if (!root.user)
                            return qsTr("Signed out");
                        switch (root.user.state) {
                        case EUserState.SignedOut:
                        default:
                            return qsTr("Signed out");
                        case EUserState.Locked:
                            return qsTr("Connecting") + dotsTimer.dots;
                        case EUserState.Connected:
                            if (root.user.isSyncing)
                                return qsTr("Synchronizing (%1%)").arg(Math.floor(root.user.syncProgress * 100)) + dotsTimer.dots;
                            else
                                return root.usedSpace;
                        }
                    }
                    type: {
                        switch (root.type) {
                        case AccountDelegate.SmallView:
                            return Label.Caption;
                        case AccountDelegate.LargeView:
                            return Label.Body;
                        }
                    }

                    Timer {
                        // dots animation while connecting & syncing.
                        id: dotsTimer

                        property string dots: ""

                        interval: 500
                        repeat: true
                        running: (root.user != null) && ((root.user.state === EUserState.Locked) || (root.user.isSyncing))

                        onRunningChanged: {
                            dots = "";
                        }
                        onTriggered: {
                            dots += ".";
                            if (dots.length > 3)
                                dots = "";
                        }
                    }
                }
                Label {
                    color: root.colorScheme.text_weak
                    colorScheme: root.colorScheme
                    text: root.user && root.user.state === EUserState.Connected && !root.user.isSyncing ? " / " + root.totalSpace : ""
                    type: {
                        switch (root.type) {
                        case AccountDelegate.SmallView:
                            return Label.Caption;
                        case AccountDelegate.LargeView:
                            return Label.Body;
                        }
                    }
                }
            }
            Item {
                implicitHeight: root.type === AccountDelegate.LargeView ? 3 : 0
            }
            Rectangle {
                id: progress_bar
                color: root.colorScheme.border_weak
                height: 4
                radius: ProtonStyle.progress_bar_radius
                visible: root.user ? root.type === AccountDelegate.LargeView : false
                width: 140

                Rectangle {
                    id: progress_bar_filled
                    color: root.progressColor
                    radius: ProtonStyle.progress_bar_radius
                    visible: root.user ? parent.visible && (root.user.state === EUserState.Connected) : false
                    width: Math.min(1, Math.max(0.02, root.progressRatio)) * parent.width

                    anchors {
                        bottom: parent.bottom
                        left: parent.left
                        top: parent.top
                    }
                }
            }
        }
        Item {
            Layout.fillWidth: true
        }
    }
}
