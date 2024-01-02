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


#include "SecondInstance.h"
#include "QMLBackend.h"
#include <bridgepp/Exception/Exception.h>
#import <Cocoa/Cocoa.h>
#import <objc/runtime.h>


using namespace bridgepp;


#ifdef Q_OS_MACOS


//****************************************************************************************************************************************************
/// \brief handle notification of attempt to re-open the application.
//****************************************************************************************************************************************************
void applicationShouldHandleReopen(id, SEL) {
    app().backend().showMainWindow("macOS applicationShouldHandleReopen notification");
}


//****************************************************************************************************************************************************
/// \brief Register our handler for the NSApplicationDelegate applicationShouldHandleReopen:hasVisibleWindows: method that is called when the user
/// tries to open a second instance of an application.
///
/// Objective-C(++) lets you add or replace selector within a class at runtime. we use it to implement our handler for the
/// applicationShouldHandleReopen:hasVisibleWindows: selector of the Cocoa application delegate.
//****************************************************************************************************************************************************
void registerSecondInstanceHandler() {
    try {
        Class cls = [[[NSApplication sharedApplication] delegate] class];
        if (!cls) {
            throw Exception("Could not retrieve Cocoa NSApplicationDelegate instance");
        }

        if (!class_replaceMethod(cls, @selector(applicationShouldHandleReopen:hasVisibleWindows:), (IMP) applicationShouldHandleReopen, "v@:")) {
            throw Exception("Could not register second application instance handler");
        }
    } catch (Exception const &e) {
        app().log().error(e.qwhat());
    }
}


#endif //  Q_OS_MACOS
