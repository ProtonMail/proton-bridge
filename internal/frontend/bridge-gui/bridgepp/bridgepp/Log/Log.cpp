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


#include "Log.h"
#include "../Exception/Exception.h"


namespace bridgepp {


namespace {

Log *qtHandlerLog { nullptr }; ///< The log instance handling qt logs.
QMutex qtHandlerMutex; ///< A mutex used to access qtHandlerLog.

// Mapping of log levels to string. Maybe used to lookup using both side a key, so a list of pair is more convenient that a map.
QList<QPair<Log::Level, QString>> const logLevelStrings {
    { Log::Level::Panic, "panic", },
    { Log::Level::Fatal, "fatal", },
    { Log::Level::Error, "error", },
    { Log::Level::Warn,  "warn", },
    { Log::Level::Info,  "info", },
    { Log::Level::Debug, "debug", },
    { Log::Level::Trace, "trace", },
};


//****************************************************************************************************************************************************
/// \param[in] log The log handling qt log entries. Can be null.
//****************************************************************************************************************************************************
void setQtMessageHandlerLog(Log *log) {
    QMutexLocker locker(&qtHandlerMutex);
    qtHandlerLog = log;
}


//****************************************************************************************************************************************************
/// \return The log handling qt log entries. Can be null.
//****************************************************************************************************************************************************
Log *qtMessageHandlerLog() {
    QMutexLocker locker(&qtHandlerMutex);
    return qtHandlerLog;
}


//****************************************************************************************************************************************************
/// \param[in] type The message type.
/// \param[in] message The message.
//****************************************************************************************************************************************************
void qtMessageHandler(QtMsgType type, QMessageLogContext const &, QString const &message) {
    Log *log = qtMessageHandlerLog();
    if (!log) {
        return;
    }
    switch (type) {
    case QtDebugMsg:
        log->debug(message);
        break;

    case QtWarningMsg:
        log->warn(message);
        break;

    case QtCriticalMsg:
    case QtFatalMsg:
        log->error(message);
        break;

    case QtInfoMsg:
    default:
        log->info(message);
        break;
    }
}
} // anonymous namespace


//****************************************************************************************************************************************************
/// \brief return a string representing the log entry, in a format similar to the one used by logrus.
///
/// \param[in] level The log entry level.
/// \param[in] message The log entry message.
/// \return The string for the log entry
//****************************************************************************************************************************************************
QString Log::logEntryToString(Log::Level level, QDateTime const &dateTime, QString const &message) {
    return QString("%1[%2] %3").arg(levelToString(level).left(4).toUpper(), dateTime.toString("MMM dd HH:mm:ss.zzz"), message);
}


//****************************************************************************************************************************************************
/// \param[in] level The level.
/// \return A string describing the level.
//****************************************************************************************************************************************************
QString Log::levelToString(Log::Level level) {
    QList<QPair<Log::Level, QString>>::const_iterator it = std::find_if(logLevelStrings.begin(), logLevelStrings.end(),
        [&level](QPair<Log::Level, QString> const &pair) -> bool {
            return pair.first == level;
        });
    return (it == logLevelStrings.end()) ? QString() : it->second;
}


//****************************************************************************************************************************************************
/// The matching is case-insensitive.
///
/// \param[in] str The string
/// \param[out] outLevel The log level parsed. if not found, the value of the variable is not modified.
/// \return true iff parsing was successful.
//****************************************************************************************************************************************************
bool Log::stringToLevel(QString const &str, Log::Level &outLevel) {
    QList<QPair<Log::Level, QString>>::const_iterator it = std::find_if(logLevelStrings.begin(), logLevelStrings.end(),
        [&str](QPair<Log::Level, QString> const &pair) -> bool {
            return 0 == QString::compare(str, pair.second, Qt::CaseInsensitive);
        });

    bool const found = (it != logLevelStrings.end());
    if (found) {
        outLevel = it->first;
    }

    return found;
}


//****************************************************************************************************************************************************
/// the message handle process the message from the Qt logging system.
//****************************************************************************************************************************************************
void Log::registerAsQtMessageHandler() {
    setQtMessageHandlerLog(this);
    qInstallMessageHandler(qtMessageHandler);
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
Log::Log()
    : QObject()
    , stdout_(stdout)
    , stderr_(stderr) {
}


//****************************************************************************************************************************************************
/// \param[in] level The log level.
//****************************************************************************************************************************************************
void Log::setLevel(Log::Level level) {
    QMutexLocker locker(&mutex_);
    level_ = level;
}


//****************************************************************************************************************************************************
/// \return The log level.
//****************************************************************************************************************************************************
Log::Level Log::level() const {
    QMutexLocker locker(&mutex_);
    return level_;
}


//****************************************************************************************************************************************************
/// \param[in] value Should the log entries be sent to STDOUT/STDERR.
//****************************************************************************************************************************************************
void Log::setEchoInConsole(bool value) {
    QMutexLocker locker(&mutex_);
    echoInConsole_ = value;
}


//****************************************************************************************************************************************************
/// \return true iff the log entries be should sent to STDOUT/STDERR.
//****************************************************************************************************************************************************
bool Log::echoInConsole() const {
    QMutexLocker locker(&mutex_);
    return echoInConsole_;
}


//****************************************************************************************************************************************************
/// \param[in] path The path of the file to write to.
/// \param[out] outError if an error occurs and this pointer in not null, on exit it contains a description of the error.
/// \return true if and only if the operation was successful.
//****************************************************************************************************************************************************
bool Log::startWritingToFile(QString const &path, QString *outError) {
    QMutexLocker locker(&mutex_);
    file_ = std::make_unique<QFile>(path);
    if (file_->open(QIODevice::WriteOnly | QIODevice::Text)) {
        return true;
    }

    if (outError) {
        *outError = QString("Could not open log file '%1' for writing.");
    }
    file_.reset();
    return false;
}


//****************************************************************************************************************************************************
///
//****************************************************************************************************************************************************
void Log::stopWritingToFile() {
    QMutexLocker locker(&mutex_);
    file_.reset();
}


//****************************************************************************************************************************************************
/// \param[in] message The message.
//****************************************************************************************************************************************************
void Log::panic(QString const &message) {
    return this->addEntry(Level::Panic, message);
}


//****************************************************************************************************************************************************
/// \param[in] message The message.
//****************************************************************************************************************************************************
void Log::fatal(QString const &message) {
    return this->addEntry(Level::Fatal, message);
}


//****************************************************************************************************************************************************
/// \param[in] message The message.
//****************************************************************************************************************************************************
void Log::error(QString const &message) {
    return this->addEntry(Level::Error, message);
}


//****************************************************************************************************************************************************
/// \param[in] message The message.
//****************************************************************************************************************************************************
void Log::warn(QString const &message) {
    return this->addEntry(Level::Warn, message);
}


//****************************************************************************************************************************************************
/// \param[in] message The message.
//****************************************************************************************************************************************************
void Log::info(QString const &message) {
    return this->addEntry(Level::Info, message);
}


//****************************************************************************************************************************************************
/// \param[in] message The message.
//****************************************************************************************************************************************************
void Log::debug(QString const &message) {
    return this->addEntry(Level::Debug, message);
}


//****************************************************************************************************************************************************
/// \param[in] message The message.
//****************************************************************************************************************************************************
void Log::trace(QString const &message) {
    return this->addEntry(Level::Trace, message);
}


//****************************************************************************************************************************************************
/// \param[in] level The level.
/// \param[in] message The message.
//****************************************************************************************************************************************************
void Log::addEntry(Log::Level level, QString const &message) {
    QDateTime const dateTime = QDateTime::currentDateTime();
    QMutexLocker locker(&mutex_);
    if (qint32(level) > qint32(level_)) {
        return;
    }

    emit entryAdded(level, message);

    if (!(echoInConsole_ || file_)) {
        return;
    }

    QString const entryStr = logEntryToString(level, dateTime, message) + "\n";
    if (echoInConsole_) {
        QTextStream &stream = (qint32(level) <= (qint32(Level::Warn))) ? stderr_ : stdout_;
        stream << entryStr;
        stream.flush();
    }

    if (file_) {
        file_->write(entryStr.toLocal8Bit());
        file_->flush();
    }
}


} // namespace bridgepp
