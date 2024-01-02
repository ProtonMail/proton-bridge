// Copyright (c) 2024 Proton AG
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


#ifndef BRIDGE_PP_LOG_H
#define BRIDGE_PP_LOG_H


namespace bridgepp {


//****************************************************************************************************************************************************
/// \brief Basic log class. No logging to file. Four levels. Rebroadcast received log entries via Qt signals.
//****************************************************************************************************************************************************
class Log : public QObject {
Q_OBJECT
public: // data types.
    /// \brief Log level class. The list matches [loggers log levels](https://pkg.go.dev/github.com/sirupsen/logrus).
    enum class Level {
        Panic, ///< Panic log level.
        Fatal, ///< Fatal log level.
        Error, ///< Error log level.
        Warn, ///< Warn log level.
        Info, ///< Info log level.
        Debug, ///< Debug log level.
        Trace ///< Trace log level.
    };

public: // static member functions.
    static QString logEntryToString(Log::Level level, QDateTime const &dateTime, QString const &message); ///< Return a string describing a log entry.
    static QString levelToString(Log::Level level); ///< return the string for a level.
    static bool stringToLevel(QString const &str, Log::Level &outLevel); ///< parse a level from a string.

public: // static data member.
    static const Level defaultLevel { Level::Debug }; ///< The default log level in Bridge.

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
    bool startWritingToFile(QString const &path, QString *outError = nullptr); ///< Start writing the log to file. Concerns only future entries.
    void stopWritingToFile();
    void registerAsQtMessageHandler(); ///< Install the Qt message handler.

public slots:
    void panic(QString const &message); ///< Adds an panic entry to the log.
    void fatal(QString const &message); ///< Adds an fatal entry to the log.
    void error(QString const &message); ///< Adds an error entry to the log.
    void warn(QString const &message); ///< Adds a warn entry to the log.
    void info(QString const &message); ///< Adds an info entry to the log.
    void debug(QString const &message); ///< Adds a debug entry in the log.
    void trace(QString const &message); ///< Adds a trace entry in the log.

    void addEntry(Log::Level level, QString const &message); ///< Adds a trace entry in the log.

signals:
    void entryAdded(Log::Level entry, QString const &); ///< Signal emitted when a log entry is added.

private: // data members
    mutable QMutex mutex_; ///< The mutex.
    Level level_ { defaultLevel }; ///< The log level
    bool echoInConsole_ { false }; ///< Set if the log messages should be sent to STDOUT/STDERR.
    std::unique_ptr<QFile> file_; ///< The file to write the log to.
    QTextStream stdout_; ///< The stdout stream.
    QTextStream stderr_; ///< The stderr stream.
};


} // namespace bridgepp


#endif //BRIDGE_PP_LOG_H
