// Copyright (c) 2024 Proton AG
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
import QtQuick.Controls.impl
import Proton

Item {
    id: root

    property int _bottomMargin: 32
    property int _leftMargin: 64
    property int _rightMargin: 64
    property int _spacing: 20
    property int _topMargin: 32
    property var colorScheme

    // fillHeight indicates whether the SettingsView should fill all available explicit height set
    property bool fillHeight: false
    default property alias items: content.children

    function positionViewAtBeginning() {
        scrollView.ScrollBar.vertical.position = 0
    }
    signal back

    ScrollView {
        id: scrollView
        anchors.fill: parent
        clip: true

        ScrollBar.vertical.policy: ScrollBar.vertical.size < 1.0 ? ScrollBar.AlwaysOn : ScrollBar.AlwaysOff;
        Component.onCompleted: contentItem.boundsBehavior = Flickable.StopAtBounds // Disable the springy effect when scroll reaches top/bottom.

        Item {
            height: scrollView.availableHeight
            implicitHeight: children[0].implicitHeight + children[0].anchors.topMargin + children[0].anchors.bottomMargin
            // do not set implicitWidth because implicit width of ColumnLayout will be equal to maximum implicit width of
            // internal items. And if one of internal items would be a Text or Label - implicit width of those is always
            // equal to non-wrapped text (i.e. one line only). That will lead to enabling horizontal scroll when not needed
            implicitWidth: width
            // can't use parent here because parent is not ScrollView (Flickable inside contentItem inside ScrollView)
            width: scrollView.availableWidth

            ColumnLayout {
                anchors.fill: parent
                spacing: 0

                ColumnLayout {
                    id: content
                    Layout.bottomMargin: root._bottomMargin
                    Layout.fillWidth: true
                    Layout.leftMargin: root._leftMargin
                    Layout.rightMargin: root._rightMargin
                    Layout.topMargin: root._topMargin
                    spacing: root._spacing
                }
                Item {
                    id: filler
                    Layout.fillHeight: true
                    visible: !root.fillHeight
                }
            }
        }
    }
    Button {
        id: backButton
        colorScheme: root.colorScheme
        horizontalPadding: 8
        icon.source: "/qml/icons/ic-arrow-left.svg"
        secondary: true
        Accessible.name: qsTr("Back")

        onClicked: root.back()

        anchors {
            left: parent.left
            leftMargin: (root._leftMargin - backButton.width) / 2
            top: parent.top
            topMargin: root._topMargin
        }
    }
}
