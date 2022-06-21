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

// +build darwin
// +build build_qt

#include "DockIcon.h"
#include <Cocoa/Cocoa.h>

void SetDockIconVisibleState(bool visible) {
    if (visible) {
        [NSApp setActivationPolicy: NSApplicationActivationPolicyRegular];
        return;
    } else {
        [NSApp setActivationPolicy: NSApplicationActivationPolicyAccessory];
        return;
    }
}

bool GetDockIconVisibleState() {
    switch ([NSApp activationPolicy]) {
    case NSApplicationActivationPolicyAccessory:
    case NSApplicationActivationPolicyProhibited:
            return false;
    case NSApplicationActivationPolicyRegular:
            return true;
    }
}
