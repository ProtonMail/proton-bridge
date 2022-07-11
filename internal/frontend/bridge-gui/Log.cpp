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


#include "Pch.h"
#include "Log.h"


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
Log::Log()
    : QObject()
    , stdout_(stdout)
    , stderr_(stderr)
{
}


//****************************************************************************************************************************************************
/// \param[in] level The log level.
//****************************************************************************************************************************************************
void Log::setLevel(Log::Level level)
{
    QMutexLocker locker(&mutex_);
    level_ = level;
}


//****************************************************************************************************************************************************
/// \return The log level.
//****************************************************************************************************************************************************
Log::Level Log::level() const
{
    QMutexLocker locker(&mutex_);
    return Log::Level::Debug;
}


//****************************************************************************************************************************************************
/// \param[in] value Should the log entries be sent to STDOUT/STDERR.
//****************************************************************************************************************************************************
void Log::setEchoInConsole(bool value)
{
    QMutexLocker locker(&mutex_);
    echoInConsole_ = value;
}


//****************************************************************************************************************************************************
/// \return true iff the log entries be should sent to STDOUT/STDERR.
//****************************************************************************************************************************************************
bool Log::echoInConsole() const
{
    QMutexLocker locker(&mutex_);
    return echoInConsole_;
}


//****************************************************************************************************************************************************
/// \param[in] message The message
//****************************************************************************************************************************************************
void Log::debug(QString const &message)
{
    QMutexLocker locker(&mutex_);

    if (qint32(level_) > qint32(Level::Debug))
        return;

    emit debugEntryAdded(message);
    QString const withPrefix = "[DEBUG] " + message;
    emit entryAdded(withPrefix);
    if (echoInConsole_)
    {
        stdout_ << withPrefix << "\n";
        stdout_.flush();
    }
}

//****************************************************************************************************************************************************
/// \param[in] message The message
//****************************************************************************************************************************************************
void Log::info(QString const &message)
{
    QMutexLocker locker(&mutex_);

    if (qint32(level_) > qint32(Level::Info))
        return;

    emit infoEntryAdded(message);
    QString const withPrefix = "[INFO] " + message;
    emit entryAdded(withPrefix);
    if (echoInConsole_)
    {
        stdout_ << withPrefix << "\n";
        stdout_.flush();
    }
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void Log::warn(QString const &message)
{
    QMutexLocker locker(&mutex_);

    if (qint32(level_) > qint32(Level::Warn))
        return;

    emit infoEntryAdded(message);
    QString const withPrefix = "[WARNING] " + message;
    emit entryAdded(withPrefix);
    if (echoInConsole_)
    {
        stdout_ << withPrefix << "\n";
        stdout_.flush();
    }
}


//****************************************************************************************************************************************************
/// message
//****************************************************************************************************************************************************
void Log::error(QString const &message)
{
    QMutexLocker locker(&mutex_);

    if (qint32(level_) > qint32(Level::Error))
        return;

    emit infoEntryAdded(message);
    QString const withPrefix = "[ERROR] " + message;
    emit entryAdded(withPrefix);
    if (echoInConsole_)
    {
        stderr_ << withPrefix << "\n";
        stderr_.flush();
    }
}





