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
import QtQuick.Controls
import QtQuick.Controls.impl
import QtQuick.Layouts

RowLayout {
    id: root

    property var callback: null
    property ColorScheme colorScheme
    property bool external: false
    property string link: "#"
    property string text: ""
    property color color: colorScheme.interaction_norm

    function clear() {
        root.callback = null;
        root.text = "";
        root.link = "";
        root.external = false;
    }
    function link(url, text) {
        return label.link(url, text);
    }
    function setCallback(callback, linkText, external) {
        root.callback = callback;
        root.text = linkText;
        root.link = "#"; // Cannot be empty, otherwise the text is not an hyperlink.
        root.external = external;
    }
    function setLink(linkURL, linkText, external) {
        root.callback = null;
        root.text = linkText;
        root.link = linkURL;
        root.external = external;
    }

    Label {
        id: label
        Layout.alignment: Qt.AlignVCenter
        colorScheme: root.colorScheme
        linkColor: root.color
        text: label.link(root.link, root.text)
        type: Label.LabelType.Body
        onLinkActivated: function (link) {
            if ((link !== "#") && (link.length > 0)) {
                Backend.openExternalLink(link);
            }
            if (callback) {
                callback();
            }
        }
    }
    ColorImage {
        Layout.alignment: Qt.AlignVCenter
        Layout.bottomMargin: -6
        color: label.linkColor
        height: sourceSize.height
        source: "/qml/icons/ic-external-link.svg"
        sourceSize.height: 14
        sourceSize.width: 14
        visible: external
        width: sourceSize.width

        MouseArea {
            anchors.fill: parent
            cursorShape: Qt.PointingHandCursor

            onClicked: {
                label.onLinkActivated(root.link);
            }
        }
    }
    HoverHandler {
        acceptedDevices: PointerDevice.Mouse
        cursorShape: Qt.PointingHandCursor
        enabled: true
    }
}