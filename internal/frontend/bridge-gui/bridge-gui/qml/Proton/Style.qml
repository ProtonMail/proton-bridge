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

pragma Singleton
import QtQml
import QtQuick

import "./"

// https://wiki.qt.io/Qml_Styling
// http://imaginativethinking.ca/make-qml-component-singleton/

QtObject {
    id: root
    // TODO: Once we will use Qt >=5.15 this should be refactored with inline components as follows:
    // https://doc.qt.io/qt-5/qtqml-documents-definetypes.html#inline-components

    // component ColorScheme: QtObject {
    //    property color primay_norm
    //    ...
    // }

    property ColorScheme lightStyle: ColorScheme {
        id: _lightStyle

        prominent: lightProminentStyle

        // Primary
        primay_norm: "#6D4AFF"

        // Interaction-norm
        interaction_norm: "#6D4AFF"
        interaction_norm_hover: "#4D34B3"
        interaction_norm_active: "#372580"

        // Text
        text_norm: "#0C0C14"
        text_weak: "#706D6B"
        text_hint: "#8F8D8A"
        text_disabled: "#C2BFBC"
        text_invert: "#FFFFFF"

        // Field
        field_norm: "#ADABA8"
        field_hover: "#8F8D8A"
        field_disabled: "#D1CFCD"

        // Border
        border_norm: "#D1CFCD"
        border_weak: "#EAE7E4"

        // Background
        background_norm: "#FFFFFF"
        background_weak: "#F5F4F2"
        background_strong: "#EAE7E4"
        background_avatar: "#C2BFBC"

        // Interaction-weak
        interaction_weak: "#D1CFCD"
        interaction_weak_hover: "#C2BFBC"
        interaction_weak_active: "#A8A6A3"

        // Interaction-default
        interaction_default: Qt.rgba(0,0,0,0)
        interaction_default_hover: Qt.rgba(194./255., 191./255., 188./255., 0.2)
        interaction_default_active: Qt.rgba(194./255., 191./255., 188./255., 0.4)

        // Scrollbar
        scrollbar_norm: "#D1CFCD"
        scrollbar_hover: "#C2BFBC"

        // Signal
        signal_danger: "#DC3251"
        signal_danger_hover: "#F74F6D"
        signal_danger_active: "#B72346"
        signal_warning: "#FF9900"
        signal_warning_hover: "#FFB800"
        signal_warning_active: "#FF851A"
        signal_success: "#1EA885"
        signal_success_hover: "#23C299"
        signal_success_active: "#198F71"
        signal_info: "#239ECE"
        signal_info_hover: "#27B1E8"
        signal_info_active: "#1F83B5"

        // Shadows
        shadow_norm: Qt.rgba(0,0,0, 0.1) // #000000 10% x:0 y:1 blur:4
        shadow_lifted: Qt.rgba(0,0,0, 0.16) // #000000 16% x:0 y:8 blur:24

        // Backdrop
        backdrop_norm: Qt.rgba(12./255., 12./255., 20./255., 0.32)

        // Images
        welcome_img: "/qml/icons/img-welcome.png"
        logo_img: "/qml/icons/product_logos.svg"
    }

    property ColorScheme lightProminentStyle: ColorScheme {
        id: _lightProminentStyle

        prominent: this

        // Primary
        primay_norm: "#8A6EFF"

        // Interaction-norm
        interaction_norm: "#6D4AFF"
        interaction_norm_hover: "#7C5CFF"
        interaction_norm_active: "#8A6EFF"

        // Text
        text_norm: "#FFFFFF"
        text_weak: "#9282D4"
        text_hint: "#544399"
        text_disabled: "#4A398F"
        text_invert: "#1B1340"

        // Field
        field_norm: "#9282D4"
        field_hover: "#7C5CFF"
        field_disabled: "#38277A"

        // Border
        border_norm: "#413085"
        border_weak: "#3C2B80"

        // Background
        background_norm: "#1B1340"
        background_weak: "#271C57"
        background_strong: "#38277A"
        background_avatar: "#6D4AFF"

        // Interaction-weak
        interaction_weak: "#4A398F"
        interaction_weak_hover: "#6D4AFF"
        interaction_weak_active: "#8A6EFF"

        // Interaction-default
        interaction_default: Qt.rgba(0,0,0,0)
        interaction_default_hover: Qt.rgba(68./255., 78./255., 114./255., 0.2)
        interaction_default_active: Qt.rgba(68./255., 78./255., 114./255., 0.3)

        // Scrollbar
        scrollbar_norm: "#413085"
        scrollbar_hover: "#4A398F"

        // Signal
        signal_danger: "#F5385A"
        signal_danger_hover: "#FF5473"
        signal_danger_active: "#DC3251"
        signal_warning: "#FF9900"
        signal_warning_hover: "#FFB800"
        signal_warning_active: "#FF8419"
        signal_success: "#1EA885"
        signal_success_hover: "#23C299"
        signal_success_active: "#198F71"
        signal_info: "#2C89DB"
        signal_info_hover: "#3491E3"
        signal_info_active: "#1F83B5"

        // Shadows
        shadow_norm: Qt.rgba(0,0,0, 0.32) // #000000 32% x:0 y:1 blur:4
        shadow_lifted: Qt.rgba(0,0,0, 0.40) // #000000 40% x:0 y:8 blur:24

        // Backdrop
        backdrop_norm: Qt.rgba(0,0,0, 0.32)

        // Images
        welcome_img: "/qml/icons/img-welcome-dark.png"
        logo_img:    "/qml/icons/product_logos_dark.svg"
    }

    property ColorScheme darkStyle: ColorScheme {
        id: _darkStyle

        prominent: darkProminentStyle

        // Primary
        primay_norm: "#8A6EFF"

        // Interaction-norm
        interaction_norm: "#6D4AFF"
        interaction_norm_hover: "#7C5CFF"
        interaction_norm_active: "#8A6EFF"

        // Text
        text_norm: "#FFFFFF"
        text_weak: "#A7A4B5"
        text_hint: "#6D697D"
        text_disabled: "#5B576B"
        text_invert: "#1C1B24"

        // Field
        field_norm: "#5B576B"
        field_hover: "#6D697D"
        field_disabled: "#3F3B4C"

        // Border
        border_norm: "#4A4658"
        border_weak: "#343140"

        // Background
        background_norm: "#1C1B24"
        background_weak: "#292733"
        background_strong: "#3F3B4C"
        background_avatar: "#6D4AFF"

        // Interaction-weak
        interaction_weak: "#4A4658"
        interaction_weak_hover: "#5B576B" 
        interaction_weak_active: "#6D697D"

        // Interaction-default
        interaction_default: "#00000000"
        interaction_default_hover: Qt.rgba(91./255.,87./255.,107./255.,0.2)
        interaction_default_active:  Qt.rgba(91./255.,87./255.,107./255.,0.4)

        // Scrollbar
        scrollbar_norm: "#4A4658"
        scrollbar_hover: "#5B576B"

        // Signal
        signal_danger: "#F5385A"
        signal_danger_hover: "#FF5473"
        signal_danger_active: "#DC3251"
        signal_warning: "#FF9900"
        signal_warning_hover: "#FFB800"
        signal_warning_active: "#FF8419"
        signal_success: "#1EA885"
        signal_success_hover: "#23C299"
        signal_success_active: "#198F71"
        signal_info: "#239ECE"
        signal_info_hover: "#27B1E8"
        signal_info_active: "#1F83B5"

        // Shadows
        shadow_norm: Qt.rgba(0,0,0,0.4) // #000000 40% x+0 y+1 blur:4
        shadow_lifted: Qt.rgba(0,0,0,0.48) // #000000 48% x+0 y+8 blur:24

        // Backdrop
        backdrop_norm: Qt.rgba(0,0,0,0.32)

        // Images
        welcome_img: "/qml/icons/img-welcome-dark.png"
        logo_img:    "/qml/icons/product_logos_dark.svg"
    }

    property ColorScheme darkProminentStyle: ColorScheme {
        id: _darkProminentStyle

        prominent: this

        // Primary
        primay_norm: "#8A6EFF"

        // Interaction-norm
        interaction_norm: "#6D4AFF"
        interaction_norm_hover: "#7C5CFF"
        interaction_norm_active: "#8A6EFF"

        // Text
        text_norm: "#FFFFFF"
        text_weak: "#A7A4B5"
        text_hint: "#6D697D"
        text_disabled: "#5B576B"
        text_invert: "#1C1B24"

        // Field
        field_norm: "#5B576B"
        field_hover: "#6D697D"
        field_disabled: "#3F3B4C"

        // Border
        border_norm: "#4A4658"
        border_weak: "#343140"

        // Background
        background_norm: "#16141c"
        background_weak: "#292733"
        background_strong: "#3F3B4C"
        background_avatar: "#6D4AFF"

        // Interaction-weak
        interaction_weak: "#4A4658"
        interaction_weak_hover: "#5B576B" 
        interaction_weak_active: "#6D697D"

        // Interaction-default
        interaction_default: "#00000000"
        interaction_default_hover: Qt.rgba(91./255.,87./255.,107./255.,0.2)
        interaction_default_active:  Qt.rgba(91./255.,87./255.,107./255.,0.4)

        // Scrollbar
        scrollbar_norm: "#4A4658"
        scrollbar_hover: "#5B576B"

        // Signal
        signal_danger: "#F5385A"
        signal_danger_hover: "#FF5473"
        signal_danger_active: "#DC3251"
        signal_warning: "#FF9900"
        signal_warning_hover: "#FFB800"
        signal_warning_active: "#FF8419"
        signal_success: "#1EA885"
        signal_success_hover: "#23C299"
        signal_success_active: "#198F71"
        signal_info: "#239ECE"
        signal_info_hover: "#27B1E8"
        signal_info_active: "#1F83B5"

        // Shadows
        shadow_norm: Qt.rgba(0,0,0,0.4) // #000000 40% x+0 y+1 blur:4
        shadow_lifted: Qt.rgba(0,0,0,0.48) // #000000 48% x+0 y+8 blur:24

        // Backdrop
        backdrop_norm: Qt.rgba(0,0,0,0.32)

        // Images
        welcome_img: "/qml/icons/img-welcome-dark.png"
        logo_img:    "/qml/icons/product_logos_dark.svg"
    }

    property ColorScheme currentStyle: lightStyle

    property string font_family: {
        switch (Qt.platform.os) {
            case "windows":
            return "Segoe UI"
            case "osx":
            return ".AppleSystemUIFont" // should be SF Pro for the foreseeable future. Using "SF Pro Display" direcly here is not allowed by the font's license.
            case "linux":
            return "Ubuntu"
            default:
            console.error("Unknown platform")
        }
    }

    property real px : 1.00 // px

    property real input_radius         : 8  * root.px // px
    property real button_radius        : 8  * root.px // px
    property real checkbox_radius      : 4  * root.px // px
    property real avatar_radius        : 8  * root.px // px
    property real big_avatar_radius    : 12 * root.px // px
    property real account_hover_radius : 12 * root.px // px
    property real account_row_radius   : 12 * root.px // px
    property real context_item_radius  : 8  * root.px // px
    property real banner_radius        : 12 * root.px // px
    property real dialog_radius        : 12 * root.px // px
    property real card_radius          : 12 * root.px // px
    property real storage_bar_radius   : 3  * root.px // px
    property real tooltip_radius       : 8  * root.px // px

    property int heading_font_size: 28
    property int heading_line_height: 36

    property int title_font_size: 20
    property int title_line_height: 24

    property int lead_font_size: 18
    property int lead_line_height: 26

    property int body_font_size: 14
    property int body_line_height: 20
    property real body_letter_spacing: 0.2 * root.px

    property int caption_font_size: 12
    property int caption_line_height: 16
    property real caption_letter_spacing: 0.4 * root.px

    property int fontWeight_100: Font.Thin
    property int fontWeight_200: Font.Light
    property int fontWeight_300: Font.ExtraLight
    property int fontWeight_400: Font.Normal
    property int fontWeight_500: Font.Medium
    property int fontWeight_600: Font.DemiBold
    property int fontWeight_700: Font.Bold
    property int fontWeight_800: Font.ExtraBold
    property int fontWeight_900: Font.Black
}
