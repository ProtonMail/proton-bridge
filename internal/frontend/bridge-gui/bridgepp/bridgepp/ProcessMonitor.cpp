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


#include "ProcessMonitor.h"
#include "Exception/Exception.h"


namespace bridgepp
{


//****************************************************************************************************************************************************
/// \param[in] exePath The path of the executable.
/// \param[in] parent The parent object of the worker.
//****************************************************************************************************************************************************
ProcessMonitor::ProcessMonitor(QString const &exePath, QStringList const &args, QObject *parent)
    : Worker(parent)
    , exePath_(exePath)
    , args_(args)
{
    QFileInfo fileInfo(exePath);
    if (!fileInfo.exists())
        throw Exception(QString("Could not locate %1 executable.").arg(fileInfo.baseName()));
    if ((!fileInfo.isFile()) || (!fileInfo.isExecutable()))
        throw Exception(QString("Invalid %1 executable").arg(fileInfo.baseName()));
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void ProcessMonitor::run()
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
        status_.returnCode = p.exitCode();

        emit processExited(status_.returnCode);
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
const ProcessMonitor::MonitorStatus &ProcessMonitor::getStatus()
{
    return status_;
}


} // namespace bridgepp