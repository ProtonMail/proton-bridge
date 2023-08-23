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
import QtQuick.Controls

Label {
    id: root

    property var callback: null

    function clear() {
        callback = null;
        text = "";
    }
    function setCallback(callback, linkText) {
        root.callback = callback;
        text = link("#", linkText);
    }
    function setLink(linkURL, linkText) {
        callback = null;
        text = link(linkURL, linkText);
    }

    type: Label.LabelType.Body

    onLinkActivated: function (link) {
        if (link !== "#") {
            Qt.openUrlExternally(link);
        }
        if (callback) {
            callback();
        }
    }

    HoverHandler {
        acceptedDevices: PointerDevice.Mouse
        cursorShape: Qt.PointingHandCursor
        enabled: true
    }
}
