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
#include "../BridgeUtils.h"
#include "../Exception/Exception.h"


namespace bridgepp {


//****************************************************************************************************************************************************
/// \return user logs directory used by bridge.
//****************************************************************************************************************************************************
QString userLogsDir() {
    QString const path = QDir(bridgepp::userDataDir()).absoluteFilePath("logs");
    QDir().mkpath(path);
    return path;
}

//****************************************************************************************************************************************************
/// \brief Return the path of the latest bridge log.
///
/// \param[in] sessionID The sessionID.
/// \return The path of the latest bridge log file.
/// \return An empty string if no bridge log file was found.
//****************************************************************************************************************************************************
QString latestBridgeLogPath(QString const &sessionID) {
    QDir const logsDir(userLogsDir());
    if (logsDir.isEmpty()) {
        return QString();
    }

    QFileInfoList const files = logsDir.entryInfoList({ sessionID + "_bri_*.log" }, QDir::Files, QDir::Name);
    return files.isEmpty() ? QString() : files.back().absoluteFilePath();
}


//****************************************************************************************************************************************************
/// Return the maxSize last bytes of the latest bridge log.
//****************************************************************************************************************************************************
QByteArray tailOfLatestBridgeLog(QString const &sessionID) {
    QString path = latestBridgeLogPath(sessionID);
    if (path.isEmpty()) {
        return QString("We could not find a bridge log file for the current session.").toLocal8Bit();
    }

    QFile file(path);
    return file.open(QIODevice::Text | QIODevice::ReadOnly) ? file.readAll().right(Exception::attachmentMaxLength) : QByteArray();
}


} // namespace bridgepp
