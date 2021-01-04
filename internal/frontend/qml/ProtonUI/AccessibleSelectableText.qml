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

// default options to make text accessible and selectable

import QtQuick 2.8
import ProtonUI 1.0

TextEdit {
    function clearText(value) {
        // substitue the copyright symbol by the text and remove the font-awesome chars and HTML tags
        return value.replace(/\uf1f9/g,'Copyright').replace(/[\uf000-\uf2e0]/g,'').replace(/<[^>]+>/g,'')
    }

    readOnly: true
    selectByKeyboard: true
    selectByMouse: true

    Accessible.role: Accessible.StaticText
    Accessible.name: clearText(text)
    Accessible.description: clearText(text)
    Accessible.focusable: true
    Accessible.ignored: !enabled || !visible || text == ""
}


