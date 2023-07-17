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


#include "AppController.h"
#include "QMLBackend.h"
#include "SentryUtils.h"
#include "Settings.h"
#include <bridgepp/CLI/CLIUtils.h>
#include <bridgepp/GRPC/GRPCClient.h>
#include <bridgepp/Exception/Exception.h>
#include <bridgepp/ProcessMonitor.h>
#include <bridgepp/Log/Log.h>
#include <sentry.h>


using namespace bridgepp;

namespace {
QString const noWindowFlag = "--no-window"; ///< The no-window command-line flag.
}


//****************************************************************************************************************************************************
/// \return The AppController instance.
//****************************************************************************************************************************************************
AppController &app() {
    static AppController app;
    return app;
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
AppController::AppController()
    : backend_(std::make_unique<QMLBackend>())
    , grpc_(std::make_unique<GRPCClient>())
    , log_(std::make_unique<Log>())
    , settings_(new Settings) {
}



//****************************************************************************************************************************************************
// The following is in the implementation file because of unique pointers with incomplete types in headers.
// See https://stackoverflow.com/questions/6012157/is-stdunique-ptrt-required-to-know-the-full-definition-of-t
//****************************************************************************************************************************************************
AppController::~AppController() = default;


//****************************************************************************************************************************************************
/// \return The bridge worker, which can be null if the application was run in 'attach' mode (-a command-line switch).
//****************************************************************************************************************************************************
ProcessMonitor *AppController::bridgeMonitor() const {
    if (!bridgeOverseer_) {
        return nullptr;
    }

    // null bridgeOverseer is OK, it means we run in 'attached' mode (app attached to an already running instance of Bridge).
    // but if bridgeOverseer is not null, its attached worker must be a valid ProcessMonitor instance.
    auto *monitor = dynamic_cast<ProcessMonitor *>(bridgeOverseer_->worker());
    if (!monitor) {
        throw Exception("Could not retrieve bridge monitor");
    }

    return monitor;
}


//****************************************************************************************************************************************************
/// \return A reference to the application settings.
//****************************************************************************************************************************************************
Settings &AppController::settings() {
    return *settings_;
}


//****************************************************************************************************************************************************
/// \param[in] exception The exception that triggered the fatal error.
//****************************************************************************************************************************************************
void AppController::onFatalError(Exception const &exception) {
    sentry_uuid_t uuid = reportSentryException("AppController got notified of a fatal error", exception);

    QMessageBox::critical(nullptr, tr("Error"), exception.what());
    restart(true);
    log().fatal(QString("reportID: %1 Captured exception: %2").arg(QByteArray(uuid.bytes, 16).toHex(), exception.detailedWhat()));
    qApp->exit(EXIT_FAILURE);
}

//****************************************************************************************************************************************************
/// \param[in] isCrashing Is the restart triggered by a crash.
//****************************************************************************************************************************************************
void AppController::restart(bool isCrashing) {
    if (launcher_.isEmpty()) {
        return;
    }

    QProcess p;
    QStringList args = stripStringParameterFromCommandLine("--session-id", launcherArgs_);
    if (isCrashing) {
        args.append(noWindowFlag);
    }

    log_->info(QString("Restarting - App : %1 - Args : %2").arg(launcher_, args.join(" ")));
    p.startDetached(launcher_, args);
    p.waitForStarted();
}


//****************************************************************************************************************************************************
/// \param[in] launcher The launcher.
/// \param[in] args The launcher arguments.
//****************************************************************************************************************************************************
void AppController::setLauncherArgs(const QString &launcher, const QStringList &args) {
    launcher_ = launcher;
    launcherArgs_ = args;
}


//****************************************************************************************************************************************************
/// \param[in] sessionID The sessionID.
//****************************************************************************************************************************************************
void AppController::setSessionID(const QString &sessionID) {
    sessionID_ = sessionID;
}


//****************************************************************************************************************************************************
/// \return The sessionID.
//****************************************************************************************************************************************************
QString AppController::sessionID() {
    return sessionID_;
}
