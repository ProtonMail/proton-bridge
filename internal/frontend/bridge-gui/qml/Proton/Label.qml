// Copyright (c) 2022 Proton AG
//
// This file is part of Proton Mail Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

import QtQuick 2.13
import QtQuick.Controls 2.12
import QtQuick.Controls.impl 2.12
import QtQuick.Templates 2.12 as T

import "." as Proton

T.Label {
    id: root

    property ColorScheme colorScheme

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
    property int type: Proton.Label.LabelType.Body

    color: root.enabled ? root.colorScheme.text_norm : root.colorScheme.text_disabled
    linkColor: root.colorScheme.interaction_norm
    palette.link: linkColor

    font.family: Style.font_family
    lineHeightMode: Text.FixedHeight

    font.weight: {
        switch (root.type) {
        case Proton.Label.LabelType.Heading:
            return Style.fontWeight_700
        case Proton.Label.LabelType.Title:
            return Style.fontWeight_700
        case Proton.Label.LabelType.Lead:
            return Style.fontWeight_400
        case Proton.Label.LabelType.Body:
            return Style.fontWeight_400
        case Proton.Label.LabelType.Body_semibold:
            return Style.fontWeight_600
        case Proton.Label.LabelType.Body_bold:
            return Style.fontWeight_700
        case Proton.Label.LabelType.Caption:
            return Style.fontWeight_400
        case Proton.Label.LabelType.Caption_semibold:
            return Style.fontWeight_600
        case Proton.Label.LabelType.Caption_bold:
            return Style.fontWeight_700
        }
    }

    font.pixelSize: {
        switch (root.type) {
        case Proton.Label.LabelType.Heading:
            return Style.heading_font_size
        case Proton.Label.LabelType.Title:
            return Style.title_font_size
        case Proton.Label.LabelType.Lead:
            return Style.lead_font_size
        case Proton.Label.LabelType.Body:
        case Proton.Label.LabelType.Body_semibold:
        case Proton.Label.LabelType.Body_bold:
            return Style.body_font_size
        case Proton.Label.LabelType.Caption:
        case Proton.Label.LabelType.Caption_semibold:
        case Proton.Label.LabelType.Caption_bold:
            return Style.caption_font_size
        }
    }

    lineHeight: {
        switch (root.type) {
        case Proton.Label.LabelType.Heading:
            return Style.heading_line_height
        case Proton.Label.LabelType.Title:
            return Style.title_line_height
        case Proton.Label.LabelType.Lead:
            return Style.lead_line_height
        case Proton.Label.LabelType.Body:
        case Proton.Label.LabelType.Body_semibold:
        case Proton.Label.LabelType.Body_bold:
            return Style.body_line_height
        case Proton.Label.LabelType.Caption:
        case Proton.Label.LabelType.Caption_semibold:
        case Proton.Label.LabelType.Caption_bold:
            return Style.caption_line_height
        }
    }

    font.letterSpacing: {
        switch (root.type) {
        case Proton.Label.LabelType.Heading:
        case Proton.Label.LabelType.Title:
        case Proton.Label.LabelType.Lead:
            return 0
        case Proton.Label.LabelType.Body:
        case Proton.Label.LabelType.Body_semibold:
        case Proton.Label.LabelType.Body_bold:
            return Style.body_letter_spacing
        case Proton.Label.LabelType.Caption:
        case Proton.Label.LabelType.Caption_semibold:
        case Proton.Label.LabelType.Caption_bold:
            return Style.caption_letter_spacing
        }
    }

    verticalAlignment: Text.AlignBottom

    function link(url, text) {
        return `<a href="${url}">${text}</a>`
    }
}
