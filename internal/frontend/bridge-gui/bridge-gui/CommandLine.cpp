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
#include "CommandLine.h"


using namespace bridgepp;


namespace
{


QString const launcherFlag = "--launcher"; ///< launcher flag parameter used for bridge.


//****************************************************************************************************************************************************
/// \brief parse a command-line string argument as expected by go's CLI package.
/// \param[in] argc The number of arguments passed to the application.
/// \param[in] argv The list of arguments passed to the application.
/// \param[in] paramNames the list of names for the parameter
//****************************************************************************************************************************************************
QString parseGoCLIStringArgument(int argc, char *argv[], QStringList paramNames)
{
    // go cli package is pretty permissive when it comes to parsing arguments. For each name 'param', all the following seems to be accepted:
    // -param value
    // --param value
    // -param=value
    // --param=value

    for (QString const &paramName: paramNames)
        for (qsizetype i = 1; i < argc; ++i)
        {
            QString const arg(QString::fromLocal8Bit(argv[i]));
            if ((i < argc - 1) && ((arg == "-" + paramName) || (arg == "--" + paramName)))
                return QString(argv[i + 1]);

            QRegularExpressionMatch match = QRegularExpression(QString("^-{1,2}%1=(.+)$").arg(paramName)).match(arg);
            if (match.hasMatch())
                return match.captured(1);
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
Log::Level parseLogLevel(int argc, char *argv[])
{
    QString levelStr = parseGoCLIStringArgument(argc, argv, { "l", "log-level" });
    if (levelStr.isEmpty())
        return Log::defaultLevel;

    Log::Level level = Log::defaultLevel;
    Log::stringToLevel(levelStr, level);
    return level;
}


} // anonymous namespace


//****************************************************************************************************************************************************
/// \param[in]  argc number of arguments passed to the application.
/// \param[in]  argv list of arguments passed to the application.
/// \param[out] args list of arguments passed to the application as a QStringList.
/// \param[out] launcher launcher used in argument, forced to self application if not specify.
/// \param[out] outAttach The value for the 'attach' command-line parameter.
/// \param[out] outLogLevel The parsed log level. If not found, the default log level is returned.
//****************************************************************************************************************************************************
void parseCommandLineArguments(int argc, char *argv[], QStringList& args, QString& launcher, bool &outAttach, Log::Level& outLogLevel) {
    bool flagFound = false;
    launcher = QString::fromLocal8Bit(argv[0]);
    // for unknown reasons, on Windows QCoreApplication::arguments() frequently returns an empty list, which is incorrect, so we rebuild the argument
    // list from the original argc and argv values.
    for (int i = 1; i < argc; i++) {
        QString const &arg = QString::fromLocal8Bit(argv[i]);
        // we can't use QCommandLineParser here since it will fail on unknown options.
        // Arguments may contain some bridge flags.
        if (arg == launcherFlag)
        {
            args.append(arg);
            launcher = QString::fromLocal8Bit(argv[++i]);
            args.append(launcher);
            flagFound = true;
        }
#ifdef QT_DEBUG
        else if (arg == "--attach" || arg == "-a")
        {
            // we don't keep the attach mode within the args since we don't need it for Bridge.
            outAttach = true;
        }
#endif
        else
        {
            args.append(arg);
        }
    }
    if (!flagFound)
    {
        // add bridge-gui as launcher
        args.append(launcherFlag);
        args.append(launcher);
    }

    outLogLevel = parseLogLevel(argc, argv);
}
