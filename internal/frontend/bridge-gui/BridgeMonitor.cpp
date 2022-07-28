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


#include "BridgeMonitor.h"
#include <bridgepp/Exception/Exception.h>


using namespace bridgepp;


namespace
{


/// \brief The file extension for the bridge executable file.
#ifdef Q_OS_WIN32
QString const exeSuffix = ".exe";
#else
QString const exeSuffix;
#endif

QString const exeName = "proton-bridge" + exeSuffix; ///< The bridge executable file name.*


}


//****************************************************************************************************************************************************
/// \return The path of the bridge executable.
/// \return A null string if the executable could not be located.
//****************************************************************************************************************************************************
QString BridgeMonitor::locateBridgeExe()
{
    QFileInfo const fileInfo(QDir(QCoreApplication::applicationDirPath()).absoluteFilePath(exeName));
    return  (fileInfo.exists() && fileInfo.isFile() && fileInfo.isExecutable()) ? fileInfo.absoluteFilePath() : QString();
}


//****************************************************************************************************************************************************
/// \param[in] exePath The path of the Bridge executable.
/// \param[in] parent The parent object of the worker.
//****************************************************************************************************************************************************
BridgeMonitor::BridgeMonitor(QString const &exePath, QStringList const &args, QObject *parent)
    : Worker(parent)
    , exePath_(exePath)
    , args_(args)
{
    QFileInfo fileInfo(exePath);
    if (!fileInfo.exists())
        throw Exception("Could not locate Bridge executable.");
    if ((!fileInfo.isFile()) || (!fileInfo.isExecutable()))
        throw Exception("Invalid bridge executable");
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void BridgeMonitor::run()
{
    try
    {
        emit started();

        QProcess p;
        p.start(exePath_, args_);
        p.waitForStarted();

        status_.running = true;
        status_.pid = p.processId();

        while (!p.waitForFinished(100))
        {
            // we discard output from bridge, it's logged to file on bridge side.
            p.readAllStandardError();
            p.readAllStandardOutput();
        }

        status_.running = false;
        status_.returnCode =  p.exitCode();

        emit processExited(status_.returnCode );
        emit finished();
    }
    catch (Exception const &e)
    {
        emit error(e.qwhat());
    }
}

//****************************************************************************************************************************************************
/// \return status of the monitored process
//****************************************************************************************************************************************************
const BridgeMonitor::MonitorStatus& BridgeMonitor::getStatus()
{
    return status_;
}
