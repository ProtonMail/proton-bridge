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
#include <bridgepp/BridgeUtils.h>


using namespace bridgepp;


namespace {
qsizetype const logFileTailMaxLength = 25 * 1024; ///< The maximum length of the portion of log returned by tailOfLatestBridgeLog()
}


//****************************************************************************************************************************************************
/// \return user logs directory used by bridge.
//****************************************************************************************************************************************************
QString userLogsDir() {
    QString const path = QDir(bridgepp::userDataDir()).absoluteFilePath("logs");
    QDir().mkpath(path);
    return path;
}

//****************************************************************************************************************************************************
/// \return A reference to the log.
//****************************************************************************************************************************************************
Log &initLog() {
    Log &log = app().log();
    log.registerAsQtMessageHandler();
    log.setEchoInConsole(true);

    // remove old gui log files
    QDir const logsDir(userLogsDir());
    for (QFileInfo const fileInfo: logsDir.entryInfoList({ "gui_v*.log" }, QDir::Filter::Files)) { // entryInfolist apparently only support wildcards, not regex.
        QFile(fileInfo.absoluteFilePath()).remove();
    }

    // create new GUI log file
    QString error;
    if (!log.startWritingToFile(logsDir.absoluteFilePath(QString("gui_v%1_%2.log").arg(PROJECT_VER).arg(QDateTime::currentSecsSinceEpoch())), &error)) {
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


//****************************************************************************************************************************************************
/// \brief Return the path of the latest bridge log.
/// \return The path of the latest bridge log file.
/// \return An empty string if no bridge log file was found.
//****************************************************************************************************************************************************
QString latestBridgeLogPath() {
    QDir const logsDir(userLogsDir());
    if (logsDir.isEmpty()) {
        return QString();
    }
    QFileInfoList files = logsDir.entryInfoList({ "v*.log" }, QDir::Files); // could do sorting, but only by last modification time. we want to sort by creation time.
    std::sort(files.begin(), files.end(), [](QFileInfo const &lhs, QFileInfo const &rhs) -> bool {
        return lhs.birthTime() < rhs.birthTime();
    });
    return files.back().absoluteFilePath();
}


//****************************************************************************************************************************************************
/// Return the maxSize last bytes of the latest bridge log.
//****************************************************************************************************************************************************
QByteArray tailOfLatestBridgeLog() {
    QString path = latestBridgeLogPath();
    if (path.isEmpty()) {
        return QByteArray();
    }

    QFile file(path);
    return file.open(QIODevice::Text | QIODevice::ReadOnly) ? file.readAll().right(logFileTailMaxLength) : QByteArray();
}

