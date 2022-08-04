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


#include "BridgeUtils.h"
#include "Exception/Exception.h"


namespace bridgepp
{

namespace {

QString const configFolder = "protonmail/bridge";


}


//****************************************************************************************************************************************************
/// \return user configuration directory used by bridge (based on Golang OS/File's UserConfigDir).
//****************************************************************************************************************************************************
QString userConfigDir()
{
    QString dir;
#ifdef Q_OS_WIN
    dir = qgetenv ("AppData");
    if (dir.isEmpty())
        throw Exception("%AppData% is not defined.");
#elif defined(Q_OS_IOS) || defined(Q_OS_DARWIN)
    dir = qgetenv ("HOME");
    if (dir.isEmpty())
        throw Exception("$HOME is not defined.");
    dir += "/Library/Application Support";
#else
    dir = qgetenv ("XDG_CONFIG_HOME");
        if (dir.isEmpty())
            dir = qgetenv ("HOME");
        if (dir.isEmpty())
            throw Exception("neither $XDG_CONFIG_HOME nor $HOME are defined");
        dir += "/.config";
#endif
    QString const folder = QDir(dir).absoluteFilePath(configFolder);
    QDir().mkpath(folder);

    return folder;
}


//****************************************************************************************************************************************************
/// \return user configuration directory used by bridge (based on Golang OS/File's UserCacheDir).
//****************************************************************************************************************************************************
QString userCacheDir()
{
    QString dir;

#ifdef Q_OS_WIN
    dir = qgetenv ("LocalAppData");
    if (dir.isEmpty())
        throw Exception("%LocalAppData% is not defined.");
#elif defined(Q_OS_IOS) || defined(Q_OS_DARWIN)
    dir = qgetenv ("HOME");
    if (dir.isEmpty())
        throw Exception("$HOME is not defined.");
    dir += "/Library/Caches";
#else
    dir = qgetenv ("XDG_CACHE_HOME");
        if (dir.isEmpty())
            dir = qgetenv ("HOME");
        if (dir.isEmpty())
            throw Exception("neither XDG_CACHE_HOME nor $HOME are defined");
        dir += "/.cache";
#endif

    QString const folder = QDir(dir).absoluteFilePath(configFolder);
    QDir().mkpath(folder);

    return folder;
}


} // namespace bridgepp
