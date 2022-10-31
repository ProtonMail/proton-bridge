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

import QtQml

QtObject {
    // should be a pointer to ColorScheme object
    property var prominent

    // Primary
    property color primay_norm

    // Interaction-norm
    property color interaction_norm
    property color interaction_norm_hover
    property color interaction_norm_active

    // Text
    property color text_norm
    property color text_weak
    property color text_hint
    property color text_disabled
    property color text_invert

    // Field
    property color field_norm
    property color field_hover
    property color field_disabled

    // Border
    property color border_norm
    property color border_weak

    // Background
    property color background_norm
    property color background_weak
    property color background_strong
    property color background_avatar

    // Interaction-weak
    property color interaction_weak
    property color interaction_weak_hover
    property color interaction_weak_active

    // Interaction-default
    property color interaction_default
    property color interaction_default_hover
    property color interaction_default_active

    // Scrollbar
    property color scrollbar_norm
    property color scrollbar_hover

    // Signal
    property color signal_danger
    property color signal_danger_hover
    property color signal_danger_active
    property color signal_warning
    property color signal_warning_hover
    property color signal_warning_active
    property color signal_success
    property color signal_success_hover
    property color signal_success_active
    property color signal_info
    property color signal_info_hover
    property color signal_info_active

    // Shadows
    property color shadow_norm
    property color shadow_lifted

    // Backdrop
    property color backdrop_norm

    // Images
    property string welcome_img
    property string logo_img
}
