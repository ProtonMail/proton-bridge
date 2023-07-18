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
import QtQuick.Controls.impl
import QtQuick.Templates as T
import "." as Proton

T.Label {
    id: root
    enum LabelType {
        // weight 700, size 28, height 36
        Heading,
        // weight 700, size 20, height 24
        Title,
        // weight 400, size 18, height 26
        Lead,
        // weight 400, size 14, height 20, spacing 0.2
        Body,
        // weight 600, size 14, height 20, spacing 0.2
        Body_semibold,
        // weight 700, size 14, height 20, spacing 0.2
        Body_bold,
        // weight 400, size 12, height 16, spacing 0.4
        Caption,
        // weight 600, size 12, height 16, spacing 0.4
        Caption_semibold,
        // weight 700, size 12, height 16, spacing 0.4
        Caption_bold
    }

    property ColorScheme colorScheme
    property int type: Proton.Label.LabelType.Body

    function link(url, text) {
        return `<a href="${url}">${text}</a>`;
    }

    color: root.enabled ? root.colorScheme.text_norm : root.colorScheme.text_disabled
    font.family: ProtonStyle.font_family
    font.letterSpacing: {
        switch (root.type) {
        case Proton.Label.LabelType.Heading:
        case Proton.Label.LabelType.Title:
        case Proton.Label.LabelType.Lead:
            return 0;
        case Proton.Label.LabelType.Body:
        case Proton.Label.LabelType.Body_semibold:
        case Proton.Label.LabelType.Body_bold:
            return ProtonStyle.body_letter_spacing;
        case Proton.Label.LabelType.Caption:
        case Proton.Label.LabelType.Caption_semibold:
        case Proton.Label.LabelType.Caption_bold:
            return ProtonStyle.caption_letter_spacing;
        }
    }
    font.pixelSize: {
        switch (root.type) {
        case Proton.Label.LabelType.Heading:
            return ProtonStyle.heading_font_size;
        case Proton.Label.LabelType.Title:
            return ProtonStyle.title_font_size;
        case Proton.Label.LabelType.Lead:
            return ProtonStyle.lead_font_size;
        case Proton.Label.LabelType.Body:
        case Proton.Label.LabelType.Body_semibold:
        case Proton.Label.LabelType.Body_bold:
            return ProtonStyle.body_font_size;
        case Proton.Label.LabelType.Caption:
        case Proton.Label.LabelType.Caption_semibold:
        case Proton.Label.LabelType.Caption_bold:
            return ProtonStyle.caption_font_size;
        }
    }
    font.weight: {
        switch (root.type) {
        case Proton.Label.LabelType.Heading:
            return ProtonStyle.fontWeight_700;
        case Proton.Label.LabelType.Title:
            return ProtonStyle.fontWeight_700;
        case Proton.Label.LabelType.Lead:
            return ProtonStyle.fontWeight_400;
        case Proton.Label.LabelType.Body:
            return ProtonStyle.fontWeight_400;
        case Proton.Label.LabelType.Body_semibold:
            return ProtonStyle.fontWeight_600;
        case Proton.Label.LabelType.Body_bold:
            return ProtonStyle.fontWeight_700;
        case Proton.Label.LabelType.Caption:
            return ProtonStyle.fontWeight_400;
        case Proton.Label.LabelType.Caption_semibold:
            return ProtonStyle.fontWeight_600;
        case Proton.Label.LabelType.Caption_bold:
            return ProtonStyle.fontWeight_700;
        }
    }
    lineHeight: {
        switch (root.type) {
        case Proton.Label.LabelType.Heading:
            return ProtonStyle.heading_line_height;
        case Proton.Label.LabelType.Title:
            return ProtonStyle.title_line_height;
        case Proton.Label.LabelType.Lead:
            return ProtonStyle.lead_line_height;
        case Proton.Label.LabelType.Body:
        case Proton.Label.LabelType.Body_semibold:
        case Proton.Label.LabelType.Body_bold:
            return ProtonStyle.body_line_height;
        case Proton.Label.LabelType.Caption:
        case Proton.Label.LabelType.Caption_semibold:
        case Proton.Label.LabelType.Caption_bold:
            return ProtonStyle.caption_line_height;
        }
    }
    lineHeightMode: Text.FixedHeight
    linkColor: root.colorScheme.interaction_norm
    palette.link: linkColor
    verticalAlignment: Text.AlignBottom
}
