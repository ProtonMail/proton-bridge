// Copyright (c) 2024 Proton AG
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

#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wavailability"
#pragma clang diagnostic ignored "-Wdeprecated-declarations"
#pragma clang diagnostic ignored "-Wnullability-completeness"
#pragma clang diagnostic ignored "-Wdeprecated-anon-enum-enum-conversion"
#include <Cocoa/Cocoa.h>
#pragma clang diagnostic pop

#include "DockIcon.h"


#ifdef Q_OS_MACOS


void setDockIconVisibleState(bool visible) {
    if (visible) {
        [NSApp setActivationPolicy: NSApplicationActivationPolicyRegular];
        return;
    } else {
        [NSApp setActivationPolicy: NSApplicationActivationPolicyAccessory];
        return;
    }
}


bool getDockIconVisibleState() {
    switch ([NSApp activationPolicy]) {
    case NSApplicationActivationPolicyAccessory:
    case NSApplicationActivationPolicyProhibited:
        return false;
    case NSApplicationActivationPolicyRegular:
        return true;
    }
}


#endif // #ifdef Q_OS_MACOS