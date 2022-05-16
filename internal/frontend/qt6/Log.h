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


#ifndef BRIDGE_QT6_LOG_H
#define BRIDGE_QT6_LOG_H


//****************************************************************************************************************************************************
/// \brief Basic log class. No logging to file. Four levels. Rebroadcast received log entries via Qt signals.
//****************************************************************************************************************************************************
class Log : public QObject
{
    Q_OBJECT
public: // data types.
    enum class Level
    {
        Debug = 0, ///< Debug
        Info = 1,  ///< Info
        Warn = 2,  ///< Warn
        Error = 3, ///< Error
    }; ///< Log level class

public: // member functions.
    Log(); ///< Default constructor.
    Log(Log const &) = delete; ///< Disabled copy-constructor.
    Log(Log &&) = delete; ///< Disabled assignment copy-constructor.
    ~Log() override = default; ///< Destructor.
    Log &operator=(Log const &) = delete; ///< Disabled assignment operator.
    Log &operator=(Log &&) = delete; ///< Disabled move assignment operator.

    void setLevel(Level level); ///< Set the log level.
    Level level() const; ///< Get the log level.
    void setEchoInConsole(bool value); ///< Set if the log entries should be echoed in STDOUT/STDERR.
    bool echoInConsole() const; ///< Check if the log entries should be echoed in STDOUT/STDERR.

public slots:
    void debug(QString const &message); ///< Adds a debug entry in the log.
    void info(QString const &message); ///< Adds an info entry to the log.
    void warn(QString const &message); ///< Adds a warning entry to the log.
    void error(QString const &message); ///< Adds an error entry to the log.

signals:
    void debugEntryAdded(QString const &); ///< Signal for debug entries.
    void infoEntryAdded(QString const &message); ///< Signal for info entries.
    void warnEntryAdded(QString const &message); ///< Signal for warning entries.
    void errorEntryAdded(QString const &message); ///< Signal for error entries.
    void entryAdded(QString const &message); ///< Signal emitted when any type of entry is added.

private: // data members
    mutable QMutex mutex_; ///< The mutex.
    Level level_{Level::Debug}; ///< The log level
    bool echoInConsole_{false}; ///< Set if the log messages should be sent to STDOUT/STDERR.

    QTextStream stdout_; ///< The stdout stream.
    QTextStream stderr_; ///< The stderr stream.
};


#endif //BRIDGE_QT6_LOG_H
