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
#include "Exception.h"


namespace
{


/// \brief The file extension for the bridge executable file.
#ifdef Q_OS_WIN32
QString const exeSuffix = ".exe";
#else
QString const exeSuffix;
#endif

QString const exeName = "bridge" + exeSuffix; ///< The bridge executable file name.
QString const devDir = "cmd/Desktop-Bridge"; ///< The folder typically containg the bridge executable in a developer's environment.
int const maxExeUpwardSeekingDepth = 5; ///< The maximum number of parent folder that will searched when trying to locate the bridge executable.


}


//****************************************************************************************************************************************************
/// \return The path of the bridge executable.
/// \return A null string if the executable could not be located.
//****************************************************************************************************************************************************
QString BridgeMonitor::locateBridgeExe()
{
    QString const currentDir = QDir::current().absolutePath();
    QString const exeDir = QCoreApplication::applicationDirPath();
    QStringList dirs = {currentDir, exeDir};
    for (int i = 0; i <= maxExeUpwardSeekingDepth; ++i)
    {
        dirs.append(currentDir + QString("../").repeated(i) + devDir);
        dirs.append(exeDir + QString("../").repeated(i) + devDir);
    }

    for (QString const &dir: dirs)
    {
        QFileInfo const fileInfo = QDir(dir).absoluteFilePath(exeName);
        if (fileInfo.exists() && fileInfo.isFile() && fileInfo.isExecutable())
            return fileInfo.absoluteFilePath();
    }

    return QString();
}


//****************************************************************************************************************************************************
/// \param[in] exePath The path of the Bridge executable.
/// \param[in] parent The parent object of the worker.
//****************************************************************************************************************************************************
BridgeMonitor::BridgeMonitor(QString const &exePath, QObject *parent)
    : Worker(parent)
    , exePath_(exePath)
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
        p.start(exePath_);
        p.waitForStarted();

        while (!p.waitForFinished(100))
        {
            // we discard output from bridge, it's logged to file on bridge side.
            p.readAllStandardError();
            p.readAllStandardOutput();
        }
        emit processExited(p.exitCode());
        emit finished();
    }
    catch (Exception const &e)
    {
        emit error(e.qwhat());
    }
}
