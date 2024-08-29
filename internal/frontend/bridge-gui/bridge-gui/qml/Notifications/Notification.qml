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
import QtQml
import QtQuick.Controls

QtObject {
    id: root
    enum NotificationType {
        Info,
        Success,
        Warning,
        Danger,
        UserNotification
    }

    property list<Action> action
    property bool active: false
    property string brief // brief is used in status view only
    default property var children
    property var data
    property string description // description is used in banners and in dialogs as description
    property bool dismissed: false
    property int group
    property string icon
    property string linkUrl: ""
    property string linkText: ""
    readonly property var occurred: active ? new Date() : undefined
    property string title // title is used in dialogs only
    property int type
    property string subtitle
    property string username


    onActiveChanged: {
        dismissed = false;
    }
}
