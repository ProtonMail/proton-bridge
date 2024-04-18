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

#include "Pch.h"
#include "CommandLine.h"
#include "Settings.h"
#include <bridgepp/CLI/CLIUtils.h>
#include <bridgepp/SessionID/SessionID.h>

using namespace bridgepp;

namespace {

QString const hyphenatedLauncherFlag = "--launcher"; ///< launcher flag parameter used for bridge.
QString const hyphenatedWindowFlag = "--no-window"; ///< The no-window command-line flag.
QString const hyphenatedSoftwareRendererFlag = "--software-renderer"; ///< The 'software-renderer' command-line flag. enable software rendering for a single execution
QString const hyphenatedSetSoftwareRendererFlag = "--set-software-renderer"; ///< The 'set-software-renderer' command-line flag. Software rendering will be used for all subsequent executions of the application.
QString const hyphenatedSetHardwareRendererFlag = "--set-hardware-renderer"; ///< The 'set-hardware-renderer' command-line flag. Hardware rendering will be used for all subsequent executions of the application.

//****************************************************************************************************************************************************
/// \brief Parse the log level from the command-line arguments.
///
/// \param[in] args The command-line arguments.
/// \return The log level. if not specified on the command-line, the default log level is returned.
//****************************************************************************************************************************************************
Log::Level parseLogLevel(QStringList const &args) {
    QStringList levelStr = parseGoCLIStringArgument(args, {"l", "log-level"});
    if (levelStr.isEmpty()) {
        return Log::defaultLevel;
    }

    Log::Level level = Log::defaultLevel;
    Log::stringToLevel(levelStr.back(), level);
    return level;
}

} // anonymous namespace

//****************************************************************************************************************************************************
/// \param[in]  argv list of arguments passed to the application, including the exe name/path at index 0.
/// \return The parsed options.
//****************************************************************************************************************************************************
CommandLineOptions parseCommandLine(QStringList const &argv) {
    CommandLineOptions options;
    bool launcherFlagFound = false;
    options.launcher = argv[0];
    // for unknown reasons, on Windows QCoreApplication::arguments() frequently returns an empty list, which is incorrect, so we rebuild the argument
    // list from the original argc and argv values.
    for (int i = 1; i < argv.count(); i++) {
        QString const &arg = argv[i];
        // we can't use QCommandLineParser here since it will fail on unknown options.

        // we skip session-id for now we'll process it later, with a special treatment for duplicates
        if (arg == hyphenatedSessionIDFlag) {
            i++; // we skip the next param, which if the flag's value.
            continue;
        }
        if (arg.startsWith(hyphenatedSessionIDFlag + "=")) {
            continue;
        }
        
        // Arguments may contain some bridge flags.
        if (arg == hyphenatedSoftwareRendererFlag) {
            options.bridgeGuiArgs.append(arg);
            options.useSoftwareRenderer = true;
        }
        if (arg == hyphenatedSetSoftwareRendererFlag) {
            app().settings().setUseSoftwareRenderer(true);
            continue; // setting is permanent. no need to keep/pass it to bridge for restart.
        }
        if (arg == hyphenatedSetHardwareRendererFlag) {
            app().settings().setUseSoftwareRenderer(false);
            continue; // setting is permanent. no need to keep/pass it to bridge for restart.
        }
        if (arg == hyphenatedWindowFlag) {
            options.noWindow = true;
        }
        if (arg == hyphenatedLauncherFlag) {
            options.bridgeArgs.append(arg);
            options.launcher = argv[++i];
            options.bridgeArgs.append(options.launcher);
            launcherFlagFound = true;
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
    if (!launcherFlagFound) {
        // add bridge-gui as launcher
        options.bridgeArgs.append(hyphenatedLauncherFlag);
        options.bridgeArgs.append(options.launcher);
    }

    QStringList args;
    if (!argv.isEmpty()) {
        args = argv.last(argv.count() - 1);
    }

    options.logLevel = parseLogLevel(args);

    QString const sessionID = mostRecentSessionID(args);
    options.bridgeArgs.append(hyphenatedSessionIDFlag);
    options.bridgeArgs.append(sessionID);
    app().setSessionID(sessionID);

    return options;
}

