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
    , log_(std::make_unique<Log>()) {
}


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
/// \param[in] exception The exception that triggered the fatal error.
//****************************************************************************************************************************************************
void AppController::onFatalError(Exception const &exception) {
    sentry_uuid_t uuid = reportSentryException("AppController got notified of a fatal error", exception);

    QMessageBox::critical(nullptr, tr("Error"), exception.what());
    restart(true);
    log().fatal(QString("reportID: %1 Captured exception: %2").arg(QByteArray(uuid.bytes, 16).toHex(), exception.detailedWhat()));
    qApp->exit(EXIT_FAILURE);
}


void AppController::restart(bool isCrashing) {
    if (!launcher_.isEmpty()) {
        QProcess p;
        log_->info(QString("Restarting - App : %1 - Args : %2").arg(launcher_, launcherArgs_.join(" ")));
        QStringList args = launcherArgs_;
        if (isCrashing) {
            args.append(noWindowFlag);
        }

        p.startDetached(launcher_, args);
        p.waitForStarted();
    }
}


void AppController::setLauncherArgs(const QString &launcher, const QStringList &args) {
    launcher_ = launcher;
    launcherArgs_ = args;
}