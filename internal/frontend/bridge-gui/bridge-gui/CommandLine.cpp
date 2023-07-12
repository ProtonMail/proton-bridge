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


#include "Pch.h"
#include "CommandLine.h"
#include "Settings.h"
#include <bridgepp/SessionID/SessionID.h>


using namespace bridgepp;


namespace {


QString const launcherFlag = "--launcher"; ///< launcher flag parameter used for bridge.
QString const noWindowFlag = "--no-window"; ///< The no-window command-line flag.
QString const softwareRendererFlag = "--software-renderer"; ///< The 'software-renderer' command-line flag. enable software rendering for a single execution
QString const setSoftwareRendererFlag = "--set-software-renderer"; ///< The 'set-software-renderer' command-line flag. Software rendering will be used for all subsequent executions of the application.
QString const setHardwareRendererFlag = "--set-hardware-renderer"; ///< The 'set-hardware-renderer' command-line flag. Hardware rendering will be used for all subsequent executions of the application.


//****************************************************************************************************************************************************
/// \brief parse a command-line string argument as expected by go's CLI package.
/// \param[in] argc The number of arguments passed to the application.
/// \param[in] argv The list of arguments passed to the application.
/// \param[in] paramNames the list of names for the parameter
//****************************************************************************************************************************************************
QString parseGoCLIStringArgument(int argc, char *argv[], QStringList paramNames) {
    // go cli package is pretty permissive when it comes to parsing arguments. For each name 'param', all the following seems to be accepted:
    // -param value
    // --param value
    // -param=value
    // --param=value
    for (QString const &paramName: paramNames) {
        for (qsizetype i = 1; i < argc; ++i) {
            QString const arg(QString::fromLocal8Bit(argv[i]));
            if ((i < argc - 1) && ((arg == "-" + paramName) || (arg == "--" + paramName))) {
                return QString(argv[i + 1]);
            }

            QRegularExpressionMatch match = QRegularExpression(QString("^-{1,2}%1=(.+)$").arg(paramName)).match(arg);
            if (match.hasMatch()) {
                return match.captured(1);
            }
        }
    }

    return QString();
}


//****************************************************************************************************************************************************
/// \brief Parse the log level from the command-line arguments.
///
/// \param[in] argc The number of arguments passed to the application.
/// \param[in] argv The list of arguments passed to the application.
/// \return The log level. if not specified on the command-line, the default log level is returned.
//****************************************************************************************************************************************************
Log::Level parseLogLevel(int argc, char *argv[]) {
    QString levelStr = parseGoCLIStringArgument(argc, argv, { "l", "log-level" });
    if (levelStr.isEmpty()) {
        return Log::defaultLevel;
    }

    Log::Level level = Log::defaultLevel;
    Log::stringToLevel(levelStr, level);
    return level;
}


} // anonymous namespace


//****************************************************************************************************************************************************
/// \param[in]  argc number of arguments passed to the application.
/// \param[in]  argv list of arguments passed to the application.
/// \return The parsed options.
//****************************************************************************************************************************************************
CommandLineOptions parseCommandLine(int argc, char *argv[]) {
    CommandLineOptions options;
    bool flagFound = false;
    options.launcher = QString::fromLocal8Bit(argv[0]);
    // for unknown reasons, on Windows QCoreApplication::arguments() frequently returns an empty list, which is incorrect, so we rebuild the argument
    // list from the original argc and argv values.
    for (int i = 1; i < argc; i++) {
        QString const &arg = QString::fromLocal8Bit(argv[i]);
        // we can't use QCommandLineParser here since it will fail on unknown options.
        // Arguments may contain some bridge flags.
        if (arg == softwareRendererFlag) {
            options.bridgeGuiArgs.append(arg);
            options.useSoftwareRenderer = true;
        }
        if (arg == setSoftwareRendererFlag) {
            app().settings().setUseSoftwareRenderer(true);
            continue; // setting is permanent. no need to keep/pass it to bridge for restart.
        }
        if (arg == setHardwareRendererFlag) {
            app().settings().setUseSoftwareRenderer(false);
            continue; // setting is permanent. no need to keep/pass it to bridge for restart.
        }
        if (arg == noWindowFlag) {
            options.noWindow = true;
        }
        if (arg == launcherFlag) {
            options.bridgeArgs.append(arg);
            options.launcher = QString::fromLocal8Bit(argv[++i]);
            options.bridgeArgs.append(options.launcher);
            flagFound = true;
        }
#ifdef QT_DEBUG
        else if (arg == "--attach" || arg == "-a") {
            // we don't keep the attach mode within the args since we don't need it for Bridge.
            options.attach = true;
            options.bridgeGuiArgs.append(arg);
        }
#endif
        else {
            options.bridgeArgs.append(arg);
            options.bridgeGuiArgs.append(arg);
        }
    }
    if (!flagFound) {
        // add bridge-gui as launcher
        options.bridgeArgs.append(launcherFlag);
        options.bridgeArgs.append(options.launcher);
    }

    options.logLevel = parseLogLevel(argc, argv);

    QString sessionID = parseGoCLIStringArgument(argc, argv, { "session-id" });
    if (sessionID.isEmpty()) {
        // The session ID was not passed to us on the command-line -> create one and add to the command-line for bridge
        sessionID = newSessionID();
        options.bridgeArgs.append("--session-id");
        options.bridgeArgs.append(sessionID);
    }
    app().setSessionID(sessionID);

    return options;
}
