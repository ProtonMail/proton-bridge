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
import QtQml
import QtQuick
import QtQuick.Layouts
import QtQuick.Controls
import QtQuick.Controls.impl
import QtWebView

Item {
    id: root

    property ColorScheme colorScheme
    property bool overlay: true
    property string url: ""

    function showBlankPage() {
        webView.loadHtml("<!doctype html><meta charset=utf-8><title>blank</title>", "blank.html");
    }

    Rectangle {
        anchors.fill: parent
        color: "#000"
        opacity: ProtonStyle.web_view_overlay_opacity
        visible: overlay
    }
    Rectangle {
        anchors.fill: parent
        anchors.margins: overlay ? ProtonStyle.web_view_overlay_margin : 0
        color: root.colorScheme.background_norm
        radius: ProtonStyle.web_view_corner_radius

        ColumnLayout {
            anchors.bottomMargin: 0
            anchors.fill: parent
            anchors.leftMargin: overlay ? ProtonStyle.web_view_overlay_horizontal_margin : 0
            anchors.rightMargin: overlay ? ProtonStyle.web_view_overlay_horizontal_margin : 0
            anchors.topMargin: overlay ? ProtonStyle.web_view_overlay_vertical_margin : 0
            spacing: 0

            Rectangle {
                Layout.fillHeight: true
                Layout.fillWidth: true
                border.color: root.colorScheme.border_norm
                border.width: overlay ? ProtonStyle.web_view_overley_border_width : 0

                WebView {
                    id: webView
                    anchors.fill: parent
                    anchors.margins: ProtonStyle.web_view_overley_border_width
                    url: root.url
                }
            }
            Button {
                Layout.alignment: Qt.AlignCenter
                Layout.bottomMargin: ProtonStyle.web_view_overlay_button_vertical_margin
                Layout.preferredWidth: ProtonStyle.web_view_button_width
                Layout.topMargin: ProtonStyle.web_view_overlay_button_vertical_margin
                colorScheme: root.colorScheme
                text: qsTr("Close")
                visible: overlay

                onClicked: {
                    root.url = "";
                    root.visible = false;
                }
            }
        }
    }
}
