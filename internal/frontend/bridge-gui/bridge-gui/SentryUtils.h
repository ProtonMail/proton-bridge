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

#ifndef BRIDGE_GUI_SENTRYUTILS_H
#define BRIDGE_GUI_SENTRYUTILS_H

#include <sentry.h>

void reportSentryEvent(sentry_level_t level, const char* message);
void reportSentryException(sentry_level_t level, const char* message, const char* exceptionType, const char* exception);

#endif //BRIDGE_GUI_SENTRYUTILS_H
