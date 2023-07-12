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


#include "LogUtils.h"
#include "BuildConfig.h"
#include <bridgepp/Log/LogUtils.h>


using namespace bridgepp;


//****************************************************************************************************************************************************
/// \return A reference to the log.
//****************************************************************************************************************************************************
Log &initLog() {
    Log &log = app().log();
    log.registerAsQtMessageHandler();
    log.setEchoInConsole(true);

    // create new GUI log file
    QString error;
    if (!log.startWritingToFile(QDir(userLogsDir()).absoluteFilePath(QString("%1_gui_000_v%2_%3.log").arg(app().sessionID(),
        PROJECT_VER, PROJECT_TAG)), &error)) {
        log.error(error);
    }

    log.info("bridge-gui starting");
    QString const qtCompileTimeVersion = QT_VERSION_STR;
    QString const qtRuntimeVersion = qVersion();
    QString msg = QString("Using Qt %1").arg(qtRuntimeVersion);
    if (qtRuntimeVersion != qtCompileTimeVersion) {
        msg += QString(" (compiled against %1)").arg(qtCompileTimeVersion);
    }
    log.info(msg);

    return log;
}
