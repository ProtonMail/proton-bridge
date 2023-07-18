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
pragma Singleton
import QtQml
import QtQuick
import "."

// https://wiki.qt.io/Qml_Styling
// http://imaginativethinking.ca/make-qml-component-singleton/
QtObject {
    id: root

    property real account_hover_radius: 12 * root.px // px
    property real account_row_radius: 12 * root.px // px
    property real avatar_radius: 8 * root.px // px
    property real banner_radius: 12 * root.px // px
    property real big_avatar_radius: 12 * root.px // px
    property int body_font_size: 14
    property real body_letter_spacing: 0.2 * root.px
    property int body_line_height: 20
    property real button_radius: 8 * root.px // px
    property int caption_font_size: 12
    property real caption_letter_spacing: 0.4 * root.px
    property int caption_line_height: 16
    property real card_radius: 12 * root.px // px
    property real checkbox_radius: 4 * root.px // px
    property real context_item_radius: 8 * root.px // px
    property ColorScheme currentStyle: lightStyle
    property ColorScheme darkProminentStyle: ColorScheme {
        id: _darkProminentStyle

        // Backdrop
        backdrop_norm: Qt.rgba(0, 0, 0, 0.32)
        background_avatar: "#6D4AFF"

        // Background
        background_norm: "#16141c"
        background_strong: "#3F3B4C"
        background_weak: "#292733"

        // Border
        border_norm: "#4A4658"
        border_weak: "#343140"
        field_disabled: "#3F3B4C"
        field_hover: "#6D697D"

        // Field
        field_norm: "#5B576B"

        // Interaction-default
        interaction_default: "#00000000"
        interaction_default_active: Qt.rgba(91. / 255., 87. / 255., 107. / 255., 0.4)
        interaction_default_hover: Qt.rgba(91. / 255., 87. / 255., 107. / 255., 0.2)

        // Interaction-norm
        interaction_norm: "#6D4AFF"
        interaction_norm_active: "#8A6EFF"
        interaction_norm_hover: "#7C5CFF"

        // Interaction-weak
        interaction_weak: "#4A4658"
        interaction_weak_active: "#6D697D"
        interaction_weak_hover: "#5B576B"
        logo_img: "/qml/icons/product_logos_dark.svg"

        // Primary
        primary_norm: "#8A6EFF"
        prominent: this
        scrollbar_hover: "#5B576B"

        // Scrollbar
        scrollbar_norm: "#4A4658"
        shadow_lifted: Qt.rgba(0, 0, 0, 0.48) // #000000 48% x+0 y+8 blur:24

        // Shadows
        shadow_norm: Qt.rgba(0, 0, 0, 0.4) // #000000 40% x+0 y+1 blur:4

        // Signal
        signal_danger: "#F5385A"
        signal_danger_active: "#DC3251"
        signal_danger_hover: "#FF5473"
        signal_info: "#239ECE"
        signal_info_active: "#1F83B5"
        signal_info_hover: "#27B1E8"
        signal_success: "#1EA885"
        signal_success_active: "#198F71"
        signal_success_hover: "#23C299"
        signal_warning: "#FF9900"
        signal_warning_active: "#FF8419"
        signal_warning_hover: "#FFB800"
        text_disabled: "#5B576B"
        text_hint: "#6D697D"
        text_invert: "#1C1B24"

        // Text
        text_norm: "#FFFFFF"
        text_weak: "#A7A4B5"

        // Images
        welcome_img: "/qml/icons/img-welcome-dark.png"
    }
    property ColorScheme darkStyle: ColorScheme {
        id: _darkStyle

        // Backdrop
        backdrop_norm: Qt.rgba(0, 0, 0, 0.32)
        background_avatar: "#6D4AFF"

        // Background
        background_norm: "#1C1B24"
        background_strong: "#3F3B4C"
        background_weak: "#292733"

        // Border
        border_norm: "#4A4658"
        border_weak: "#343140"
        field_disabled: "#3F3B4C"
        field_hover: "#6D697D"

        // Field
        field_norm: "#5B576B"

        // Interaction-default
        interaction_default: "#00000000"
        interaction_default_active: Qt.rgba(91. / 255., 87. / 255., 107. / 255., 0.4)
        interaction_default_hover: Qt.rgba(91. / 255., 87. / 255., 107. / 255., 0.2)

        // Interaction-norm
        interaction_norm: "#6D4AFF"
        interaction_norm_active: "#8A6EFF"
        interaction_norm_hover: "#7C5CFF"

        // Interaction-weak
        interaction_weak: "#4A4658"
        interaction_weak_active: "#6D697D"
        interaction_weak_hover: "#5B576B"
        logo_img: "/qml/icons/product_logos_dark.svg"

        // Primary
        primary_norm: "#8A6EFF"
        prominent: darkProminentStyle
        scrollbar_hover: "#5B576B"

        // Scrollbar
        scrollbar_norm: "#4A4658"
        shadow_lifted: Qt.rgba(0, 0, 0, 0.48) // #000000 48% x+0 y+8 blur:24

        // Shadows
        shadow_norm: Qt.rgba(0, 0, 0, 0.4) // #000000 40% x+0 y+1 blur:4

        // Signal
        signal_danger: "#F5385A"
        signal_danger_active: "#DC3251"
        signal_danger_hover: "#FF5473"
        signal_info: "#239ECE"
        signal_info_active: "#1F83B5"
        signal_info_hover: "#27B1E8"
        signal_success: "#1EA885"
        signal_success_active: "#198F71"
        signal_success_hover: "#23C299"
        signal_warning: "#FF9900"
        signal_warning_active: "#FF8419"
        signal_warning_hover: "#FFB800"
        text_disabled: "#5B576B"
        text_hint: "#6D697D"
        text_invert: "#1C1B24"

        // Text
        text_norm: "#FFFFFF"
        text_weak: "#A7A4B5"

        // Images
        welcome_img: "/qml/icons/img-welcome-dark.png"
    }
    property real dialog_radius: 12 * root.px // px
    property int fontWeight_100: Font.Thin
    property int fontWeight_200: Font.Light
    property int fontWeight_300: Font.ExtraLight
    property int fontWeight_400: Font.Normal
    property int fontWeight_500: Font.Medium
    property int fontWeight_600: Font.DemiBold
    property int fontWeight_700: Font.Bold
    property int fontWeight_800: Font.ExtraBold
    property int fontWeight_900: Font.Black
    property string font_family: {
        switch (Qt.platform.os) {
        case "windows":
            return "Segoe UI";
        case "osx":
            return ".AppleSystemUIFont"; // should be SF Pro for the foreseeable future. Using "SF Pro Display" directly here is not allowed by the font's license.
        case "linux":
            return "Ubuntu";
        default:
            console.error("Unknown platform");
        }
    }
    property int heading_font_size: 28
    property int heading_line_height: 36
    property real input_radius: 8 * root.px // px
    property int lead_font_size: 18
    property int lead_line_height: 26
    property ColorScheme lightProminentStyle: ColorScheme {
        id: _lightProminentStyle

        // Backdrop
        backdrop_norm: Qt.rgba(0, 0, 0, 0.32)
        background_avatar: "#6D4AFF"

        // Background
        background_norm: "#1B1340"
        background_strong: "#38277A"
        background_weak: "#271C57"

        // Border
        border_norm: "#413085"
        border_weak: "#3C2B80"
        field_disabled: "#38277A"
        field_hover: "#7C5CFF"

        // Field
        field_norm: "#9282D4"

        // Interaction-default
        interaction_default: Qt.rgba(0, 0, 0, 0)
        interaction_default_active: Qt.rgba(68. / 255., 78. / 255., 114. / 255., 0.3)
        interaction_default_hover: Qt.rgba(68. / 255., 78. / 255., 114. / 255., 0.2)

        // Interaction-norm
        interaction_norm: "#6D4AFF"
        interaction_norm_active: "#8A6EFF"
        interaction_norm_hover: "#7C5CFF"

        // Interaction-weak
        interaction_weak: "#4A398F"
        interaction_weak_active: "#8A6EFF"
        interaction_weak_hover: "#6D4AFF"
        logo_img: "/qml/icons/product_logos_dark.svg"

        // Primary
        primary_norm: "#8A6EFF"
        prominent: this
        scrollbar_hover: "#4A398F"

        // Scrollbar
        scrollbar_norm: "#413085"
        shadow_lifted: Qt.rgba(0, 0, 0, 0.40) // #000000 40% x:0 y:8 blur:24

        // Shadows
        shadow_norm: Qt.rgba(0, 0, 0, 0.32) // #000000 32% x:0 y:1 blur:4

        // Signal
        signal_danger: "#F5385A"
        signal_danger_active: "#DC3251"
        signal_danger_hover: "#FF5473"
        signal_info: "#2C89DB"
        signal_info_active: "#1F83B5"
        signal_info_hover: "#3491E3"
        signal_success: "#1EA885"
        signal_success_active: "#198F71"
        signal_success_hover: "#23C299"
        signal_warning: "#FF9900"
        signal_warning_active: "#FF8419"
        signal_warning_hover: "#FFB800"
        text_disabled: "#4A398F"
        text_hint: "#544399"
        text_invert: "#1B1340"

        // Text
        text_norm: "#FFFFFF"
        text_weak: "#9282D4"

        // Images
        welcome_img: "/qml/icons/img-welcome-dark.png"
    }
    // TODO: Once we will use Qt >=5.15 this should be refactored with inline components as follows:
    // https://doc.qt.io/qt-5/qtqml-documents-definetypes.html#inline-components

    // component ColorScheme: QtObject {
    //    property color primary_norm
    //    ...
    // }
    property ColorScheme lightStyle: ColorScheme {
        id: _lightStyle

        // Backdrop
        backdrop_norm: Qt.rgba(12. / 255., 12. / 255., 20. / 255., 0.32)
        background_avatar: "#C2BFBC"

        // Background
        background_norm: "#FFFFFF"
        background_strong: "#EAE7E4"
        background_weak: "#F5F4F2"

        // Border
        border_norm: "#D1CFCD"
        border_weak: "#EAE7E4"
        field_disabled: "#D1CFCD"
        field_hover: "#8F8D8A"

        // Field
        field_norm: "#ADABA8"

        // Interaction-default
        interaction_default: Qt.rgba(0, 0, 0, 0)
        interaction_default_active: Qt.rgba(194. / 255., 191. / 255., 188. / 255., 0.4)
        interaction_default_hover: Qt.rgba(194. / 255., 191. / 255., 188. / 255., 0.2)

        // Interaction-norm
        interaction_norm: "#6D4AFF"
        interaction_norm_active: "#372580"
        interaction_norm_hover: "#4D34B3"

        // Interaction-weak
        interaction_weak: "#D1CFCD"
        interaction_weak_active: "#A8A6A3"
        interaction_weak_hover: "#C2BFBC"
        logo_img: "/qml/icons/product_logos.svg"

        // Primary
        primary_norm: "#6D4AFF"
        prominent: lightProminentStyle
        scrollbar_hover: "#C2BFBC"

        // Scrollbar
        scrollbar_norm: "#D1CFCD"
        shadow_lifted: Qt.rgba(0, 0, 0, 0.16) // #000000 16% x:0 y:8 blur:24

        // Shadows
        shadow_norm: Qt.rgba(0, 0, 0, 0.1) // #000000 10% x:0 y:1 blur:4

        // Signal
        signal_danger: "#DC3251"
        signal_danger_active: "#B72346"
        signal_danger_hover: "#F74F6D"
        signal_info: "#239ECE"
        signal_info_active: "#1F83B5"
        signal_info_hover: "#27B1E8"
        signal_success: "#1EA885"
        signal_success_active: "#198F71"
        signal_success_hover: "#23C299"
        signal_warning: "#FF9900"
        signal_warning_active: "#FF851A"
        signal_warning_hover: "#FFB800"
        text_disabled: "#C2BFBC"
        text_hint: "#8F8D8A"
        text_invert: "#FFFFFF"

        // Text
        text_norm: "#0C0C14"
        text_weak: "#706D6B"

        // Images
        welcome_img: "/qml/icons/img-welcome.png"
    }
    property real progress_bar_radius: 3 * root.px // px
    property real px: 1.00 // px
    property int title_font_size: 20
    property int title_line_height: 24
    property real tooltip_radius: 8 * root.px // px
}
