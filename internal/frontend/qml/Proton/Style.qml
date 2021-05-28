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

pragma Singleton
import QtQml 2.13

import "./"

// https://wiki.qt.io/Qml_Styling
// http://imaginativethinking.ca/make-qml-component-singleton/

QtObject {

    // TODO: Once we will use Qt >=5.15 this should be refactored with inline components as follows:
    // https://doc.qt.io/qt-5/qtqml-documents-definetypes.html#inline-components

    //component ColorScheme: QtObject {
        //    property color primay_norm
        //    ...
        //}

        // and instead of "var" later on "ColorScheme" should be used

        property var _lightStyle: ColorScheme {
            id: lightStyle

            // Primary
            primay_norm: "#657EE4"

            // Interaction-norm
            interaction_norm: "#657EE4"
            interaction_norm_hover: "#5064B6"
            interaction_norm_active: "#3C4B88"

            // Text
            text_norm: "#262A33"
            text_weak: "#696F7D"
            text_hint: "#A4A9B5"
            text_disabled: "#BABEC7"
            text_invert: "#FFFFFF"

            // Field
            field_norm: "#BABEC7"
            field_hover: "#A4A9B5"
            field_disabled: "#D0D3DA"

            // Border
            border_norm: "#D0D3DA"
            border_weak: "#E7E9EC"

            // Background
            background_norm: "#FFFFFF"
            background_weak: "#F3F4F6"
            background_strong: "#E7E9EC"
            background_avatar: "#A4A9B5"

            // Interaction-weak
            interaction_weak: "#D0D3DA"
            interaction_weak_hover: "#BABEC7"
            interaction_weak_active: "#A4A9B5"

            // Interaction-default
            interaction_default: "#00000000"
            interaction_default_hover: "#33BABEC7"
            interaction_default_active: "#4DBABEC7"

            // Scrollbar
            scrollbar_norm: "#D0D3DA"
            scrollbar_hover: "#BABEC7"

            // Signal
            signal_danger: "#D42F34"
            signal_danger_hover: "#C7262B"
            signal_danger_active: "#BA1E23"
            signal_warning: "#F5830A"
            signal_warning_hover: "#F5740A"
            signal_warning_active: "#F5640A"
            signal_success: "#1B8561"
            signal_success_hover: "#147857"
            signal_success_active: "#0F6B4C"
            signal_info: "#1578CF"
            signal_info_hover: "#0E6DC2"
            signal_info_active: "#0764B5"

            // Shadows
            shadow_norm: "#FFFFFF"
            shadow_lifted: "#FFFFFF"

            // Backdrop
            backdrop_norm: "#7A262A33"
        }

        property var _prominentStyle: ColorScheme {
            id: prominentStyle

            // Primary
            primay_norm: "#657EE4"

            // Interaction-norm
            interaction_norm: "#657EE4"
            interaction_norm_hover: "#7D92E8"
            interaction_norm_active: "#98A9EE"

            // Text
            text_norm: "#FFFFFF"
            text_weak: "#949BB9"
            text_hint: "#565F84"
            text_disabled: "#444E72"
            text_invert: "#1C223D"

            // Field
            field_norm: "#565F84"
            field_hover: "#949BB9"
            field_disabled: "#353E60"

            // Border
            border_norm: "#353E60"
            border_weak: "#2D3657"

            // Background
            background_norm: "#1C223D"
            background_weak: "#272F4F"
            background_strong: "#2D3657"
            background_avatar: "#444E72"

            // Interaction-weak
            interaction_weak: "#353E60"
            interaction_weak_hover: "#444E72"
            interaction_weak_active: "#565F84"

            // Interaction-default
            interaction_default: "#00000000"
            interaction_default_hover: "#4D444E72"
            interaction_default_active: "#66444E72"

            // Scrollbar
            scrollbar_norm: "#353E60"
            scrollbar_hover: "#444E72"

            // Signal
            signal_danger: "#ED4C51"
            signal_danger_hover: "#F7595E"
            signal_danger_active: "#FF666B"
            signal_warning: "#F5930A"
            signal_warning_hover: "#F5A716"
            signal_warning_active: "#F5B922"
            signal_success: "#349172"
            signal_success_hover: "#339C79"
            signal_success_active: "#31A67F"
            signal_info: "#2C89DB"
            signal_info_hover: "#3491E3"
            signal_info_active: "#3D99EB"

            // Shadows
            shadow_norm: "#1C223D"
            shadow_lifted: "#1C223D"

            // Backdrop
            backdrop_norm: "#52000000"
        }

        property var _darkStyle: ColorScheme {
            id: darkStyle

            // Primary
            primay_norm: "#657EE4"

            // Interaction-norm
            interaction_norm: "#657EE4"
            interaction_norm_hover: "#7D92E8"
            interaction_norm_active: "#98A9EE"

            // Text
            text_norm: "#FFFFFF"
            text_weak: "#A4A9B5"
            text_hint: "#696F7D"
            text_disabled: "#575D6B"
            text_invert: "#262A33"

            // Field
            field_norm: "#575D6B"
            field_hover: "#696F7D"
            field_disabled: "#464B58"

            // Border
            border_norm: "#464B58"
            border_weak: "#363A46"

            // Background
            background_norm: "#262A33"
            background_weak: "#2E323C"
            background_strong: "#363A46"
            background_avatar: "#575D6B"

            // Interaction-weak
            interaction_weak: "#464B58"
            interaction_weak_hover: "#575D6B"
            interaction_weak_active: "#696F7D"

            // Interaction-default
            interaction_default: "#00000000"
            interaction_default_hover: "#33575D6B"
            interaction_default_active: "#4D575D6B"

            // Scrollbar
            scrollbar_norm: "#464B58"
            scrollbar_hover: "#575D6B"

            // Signal
            signal_danger: "#ED4C51"
            signal_danger_hover: "#F7595E"
            signal_danger_active: "#FF666B"
            signal_warning: "#F5930A"
            signal_warning_hover: "#F5A716"
            signal_warning_active: "#F5B922"
            signal_success: "#349172"
            signal_success_hover: "#339C79"
            signal_success_active: "#31A67F"
            signal_info: "#2C89DB"
            signal_info_hover: "#3491E3"
            signal_info_active: "#3D99EB"

            // Shadows
            shadow_norm: "#262A33"
            shadow_lifted: "#262A33"

            // Backdrop
            backdrop_norm: "#52000000"
        }

        // TODO: if default style should be loaded from somewhere - it should be loaded here
        property var currentStyle: lightStyle

        property var _timer: Timer {
            interval: 1000
            repeat: true
            running: true
            onTriggered: {
                switch (currentStyle) {
                    case lightStyle:
                    console.debug("Dark Style")
                    currentStyle = darkStyle
                    return
                    case darkStyle:
                    console.debug("Prominent Style")
                    currentStyle = prominentStyle
                    return
                    case prominentStyle:
                    console.debug("Light Style")
                    currentStyle = lightStyle
                    return
                }
            }
        }



        property string font: {
            // TODO: add OS to backend
            return "Ubuntu"

            //switch (backend.OS) {
                //    case "Windows":
                //        return "Segoe UI"
                //    case "OSX":
                //        return "SF Pro Display"
                //    case "Linux":
                //        return "Ubuntu"
                //}
            }

            property int heading_font_size: 28
            property int heading_line_height: 36

            property int title_font_size: 20
            property int title_line_height: 24

            property int lead_font_size: 18
            property int lead_line_height: 26

            property int body_font_size: 14
            property int body_line_height: 20
            property real body_letter_spacing: 0.2

            property int caption_font_size: 12
            property int caption_line_height: 16
            property real caption_letter_spacing: 0.4
        }
