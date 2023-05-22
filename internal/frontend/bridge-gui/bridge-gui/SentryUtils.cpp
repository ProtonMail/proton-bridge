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
#include <bridgepp/Exception/Exception.h>


using namespace bridgepp;


static constexpr const char *LoggerName = "bridge-gui";


//****************************************************************************************************************************************************
/// \return The temporary file used for sentry attachment.
//****************************************************************************************************************************************************
QString sentryAttachmentFilePath() {
    static QString path;
    if (!path.isEmpty()) {
        return path;
    }
    while (true) {
        path = QDir::temp().absoluteFilePath(QUuid::createUuid().toString(QUuid::WithoutBraces) + ".txt"); // Sentry does not offer preview for .log files.
        if (!QFileInfo::exists(path)) {
            return path;
        }
    }
}


//****************************************************************************************************************************************************
/// \brief Get a hash of the computer's host name
//****************************************************************************************************************************************************
QByteArray getProtectedHostname() {
    QByteArray hostname = QCryptographicHash::hash(QSysInfo::machineHostName().toUtf8(), QCryptographicHash::Sha256);
    return hostname.toBase64();
}

//****************************************************************************************************************************************************
/// \return The OS String used by sentry
//****************************************************************************************************************************************************
QString getApiOS() {
    switch (os()) {
    case OS::MacOS:
        return "macos";
    case OS::Windows:
        return "windows";
    case OS::Linux:
    default:
        return "linux";
    }
}

//****************************************************************************************************************************************************
/// \return The application version number.
//****************************************************************************************************************************************************
QString appVersion(const QString& version) {
    return QString("%1-bridge@%2").arg(getApiOS(), version);
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void initSentry() {
    sentry_options_t *sentryOptions = newSentryOptions(PROJECT_DSN_SENTRY, sentryCacheDir().toStdString().c_str());
    if (!QString(PROJECT_CRASHPAD_HANDLER_PATH).isEmpty())
        sentry_options_set_handler_path(sentryOptions, PROJECT_CRASHPAD_HANDLER_PATH);

    if (sentry_init(sentryOptions) != 0) {
        QTextStream(stderr) << "Failed to initialize sentry\n";
    }
    setSentryReportScope();
}

//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
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


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
sentry_options_t* newSentryOptions(const char *sentryDNS, const char *cacheDir) {
    sentry_options_t *sentryOptions = sentry_options_new();
    sentry_options_set_dsn(sentryOptions, sentryDNS);
    sentry_options_set_database_path(sentryOptions, cacheDir);
    sentry_options_set_release(sentryOptions, appVersion(PROJECT_VER).toUtf8());
    sentry_options_set_max_breadcrumbs(sentryOptions, 50);
    sentry_options_set_environment(sentryOptions, PROJECT_BUILD_ENV);
    QByteArray const array = sentryAttachmentFilePath().toLocal8Bit();
    sentry_options_add_attachment(sentryOptions, array.constData());
    // Enable this for debugging sentry.
    // sentry_options_set_debug(sentryOptions, 1);

    return sentryOptions;
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
sentry_uuid_t reportSentryEvent(sentry_level_t level, const char *message) {
    auto event = sentry_value_new_message_event(level, LoggerName, message);
    return sentry_capture_event(event);
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
sentry_uuid_t reportSentryException(sentry_level_t level, const char *message, const char *exceptionType, const char *exception) {
    auto event = sentry_value_new_message_event(level, LoggerName, message);
    sentry_event_add_exception(event, sentry_value_new_exception(exceptionType, exception));

    // reject exception content from the fingerprint if there is not enough content
    if ( strlen(exception) < 5) {
        sentry_value_t fingerprint = sentry_value_new_list();
        sentry_value_append(fingerprint, sentry_value_new_string("level"));
        sentry_value_append(fingerprint, sentry_value_new_string(message));
        sentry_value_append(fingerprint, sentry_value_new_string(LoggerName));
        sentry_value_set_by_key(event, "fingerprint", fingerprint);
    }

    return sentry_capture_event(event);
}


//****************************************************************************************************************************************************
/// \param[in] message The message for the exception.
/// \param[in] function The name of the function that triggered the exception.
/// \param[in] exception The exception.
/// \return The Sentry exception UUID.
//****************************************************************************************************************************************************
sentry_uuid_t reportSentryException(QString const &message, bridgepp::Exception const exception) {
    QByteArray const attachment = exception.attachment();
    QFile file(sentryAttachmentFilePath());
    bool const hasAttachment = !attachment.isEmpty();
    if (hasAttachment) {
        if (file.open(QIODevice::Text | QIODevice::WriteOnly)) {
            file.write(attachment);
            file.close();
        }
    }

    sentry_uuid_t const uuid = reportSentryException(SENTRY_LEVEL_ERROR, message.toLocal8Bit(), "Exception",
        exception.detailedWhat().toLocal8Bit());

    if (hasAttachment) {
        file.remove();
    }

    return uuid;
}

