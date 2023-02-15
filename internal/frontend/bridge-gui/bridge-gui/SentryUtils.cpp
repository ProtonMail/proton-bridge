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
#include "BuildConfig.h"
#include <bridgepp/BridgeUtils.h>

#include <QByteArray>
#include <QCryptographicHash>
#include <QString>
#include <QSysInfo>

static constexpr const char *LoggerName = "bridge-gui";

QByteArray getProtectedHostname() {
    QByteArray hostname = QCryptographicHash::hash(QSysInfo::machineHostName().toUtf8(), QCryptographicHash::Sha256);
    return hostname.toHex();
}

QString getApiOS() {
#if defined(Q_OS_DARWIN)
    return "macos";
#elif defined(Q_OS_WINDOWS)
    return "windows";
#else
    return "linux";
#endif
}

QString appVersion(const QString& version) {
    return QString("%1-bridge@%2").arg(getApiOS()).arg(version);
}

void setSentryReportScope() {
    sentry_set_tag("OS", bridgepp::goos().toUtf8());
    sentry_set_tag("Client", PROJECT_FULL_NAME);
    sentry_set_tag("Version", PROJECT_REVISION);
    sentry_set_tag("HostArch", QSysInfo::currentCpuArchitecture().toUtf8());
    sentry_set_tag("server_name", getProtectedHostname());
    sentry_value_t user = sentry_value_new_object();
    sentry_value_set_by_key(user, "id", sentry_value_new_string(getProtectedHostname()));
    sentry_set_user(user);
}

sentry_options_t* newSentryOptions(const char *sentryDNS, const char *cacheDir) {
    sentry_options_t *sentryOptions = sentry_options_new();
    sentry_options_set_dsn(sentryOptions, sentryDNS);
    sentry_options_set_database_path(sentryOptions, cacheDir);
    sentry_options_set_release(sentryOptions, appVersion(PROJECT_VER).toUtf8());
    sentry_options_set_max_breadcrumbs(sentryOptions, 50);
    sentry_options_set_environment(sentryOptions, PROJECT_BUILD_ENV);
    // Enable this for debugging sentry.
    // sentry_options_set_debug(sentryOptions, 1);

    return sentryOptions;
}

sentry_uuid_t reportSentryEvent(sentry_level_t level, const char *message) {
    auto event = sentry_value_new_message_event(level, LoggerName, message);
    return sentry_capture_event(event);
}


sentry_uuid_t reportSentryException(sentry_level_t level, const char *message, const char *exceptionType, const char *exception) {
    auto event = sentry_value_new_message_event(level, LoggerName, message);
    sentry_event_add_exception(event, sentry_value_new_exception(exceptionType, exception));
    return sentry_capture_event(event);
}


