// Copyright (c) 2023 Proton AG
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

#include "SentryUtils.h"


static constexpr const char *LoggerName = "bridge-gui";


void reportSentryEvent(sentry_level_t level, const char *message) {
    auto event = sentry_value_new_message_event(level, LoggerName, message);
    sentry_capture_event(event);
}


void reportSentryException(sentry_level_t level, const char *message, const char *exceptionType, const char *exception) {
    auto event = sentry_value_new_message_event(level, LoggerName, message);
    sentry_event_add_exception(event, sentry_value_new_exception(exceptionType, exception));
    sentry_capture_event(event);
}
