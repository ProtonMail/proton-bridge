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
import QtQuick.Controls 2.12

Label {
    id: root

    color: Style.currentStyle.text_norm
    palette.link: Style.currentStyle.interaction_norm

    font.family: ProtonStyle.font_family
    font.weight: ProtonStyle.fontWidth_400
    lineHeightMode: Text.FixedHeight

    function putLink(linkURL,linkText) {
        return `<a href="${linkURL}">${linkText}</a>`
    }

    state: "title"
    states: [
        State { name : "heading" ; PropertyChanges { target : root ; font.pixelSize : Style.heading_font_size ; lineHeight : Style.heading_line_height } },
        State { name : "title"   ; PropertyChanges { target : root ; font.pixelSize : Style.title_font_size   ; lineHeight : Style.title_line_height   } },
        State { name : "lead"    ; PropertyChanges { target : root ; font.pixelSize : Style.lead_font_size    ; lineHeight : Style.lead_line_height    } },
        State { name : "body"    ; PropertyChanges { target : root ; font.pixelSize : Style.body_font_size    ; lineHeight : Style.body_line_height    ; font.letterSpacing : Style.body_letter_spacing    } },
        State { name : "caption" ; PropertyChanges { target : root ; font.pixelSize : Style.caption_font_size ; lineHeight : Style.caption_line_height ; font.letterSpacing : Style.caption_letter_spacing } }
    ]
}
