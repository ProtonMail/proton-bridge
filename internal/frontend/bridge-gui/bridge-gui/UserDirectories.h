//
// Created by romain on 01/08/22.
//

#ifndef PROTON_BRIDGE_GUI_USERDIRECTORIES_H
#define PROTON_BRIDGE_GUI_USERDIRECTORIES_H

#include <bridgepp/Exception/Exception.h>

using namespace bridgepp;

namespace UserDirectories {

    QString const configFolder = "protonmail/bridge";

//****************************************************************************************************************************************************
/// \return user configuration directory used by bridge (based on Golang OS/File's UserConfigDir).
//****************************************************************************************************************************************************
    static const QString UserConfigDir()
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
        QString folder = dir + "/" + configFolder;
        QDir().mkpath(folder);

        return folder;
    }


//****************************************************************************************************************************************************
/// \return user configuration directory used by bridge (based on Golang OS/File's UserCacheDir).
//****************************************************************************************************************************************************
    static const QString UserCacheDir()
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
        QString folder = dir + "/" + configFolder;
        QDir().mkpath(folder);

        return folder;
    }

};


#endif //PROTON_BRIDGE_GUI_USERDIRECTORIES_H
