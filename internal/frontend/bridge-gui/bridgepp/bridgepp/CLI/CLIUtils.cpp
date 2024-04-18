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

#include "CLIUtils.h"
#include "../SessionID/SessionID.h"

namespace bridgepp {


//****************************************************************************************************************************************************
/// \param[in] paramName The parameter name, including prefix dashes (e.g. '--string').
/// \param[in] commandLineParams The command-line parameters.
/// \return The command-line parameters where all occurrences of paramName and it associated value have been removed. Comparison is case-sensitive.
//****************************************************************************************************************************************************
QStringList stripStringParameterFromCommandLine(QString const &paramName, QStringList const &commandLineParams) {
    qint32 i = 0;
    QStringList result;
    while (i < commandLineParams.count()) {
        if (paramName == commandLineParams[i]) {
            i += 2;
            continue;
        }
        result.append(commandLineParams[i]);
        i++;
    }

    return result;
}


//****************************************************************************************************************************************************
/// The flags may be present more than once in the args. All values are returned in order of appearance.
///
/// \param[in] args The arguments
/// \param[in] paramNames the list of names for the parameter, without any prefix hypen.
/// \return The values found for the flag.
//****************************************************************************************************************************************************
QStringList parseGoCLIStringArgument(QStringList const &args, QStringList const& paramNames) {
    // go cli package is pretty permissive when it comes to parsing arguments. For each name 'param', all the following seems to be accepted:
    // -param value
    // --param value
    // -param=value
    // --param=value

    QStringList result;
    qsizetype const argCount = args.count();
    for (qsizetype i = 0; i < args.size(); ++i) {
        for (QString const &paramName: paramNames) {
            if ((i < argCount - 1) && ((args[i] == "-" + paramName) || (args[i] == "--" + paramName))) {
                result.append(args[i + 1]);
                i += 1;
                continue;
            }
            if (QRegularExpressionMatch match = QRegularExpression(QString("^-{1,2}%1=(.+)$").arg(paramName)).match(args[i]); match.hasMatch()) {
                result.append(match.captured(1));
                continue;
            }
        }
    }

    return result;
}

//****************************************************************************************************************************************************
/// \param[in] argc The number of command-line arguments.
/// \param[in] argv The list of command-line arguments.
/// \return A QStringList representing the arguments list.
//****************************************************************************************************************************************************
QStringList cliArgsToStringList(int argc, char **argv) {
    QStringList result;
    result.reserve(argc);
    for (qsizetype i = 0; i < argc; ++i) {
        result.append(QString::fromLocal8Bit(argv[i]));
    }
    return result;
}

//****************************************************************************************************************************************************
/// \param[in] args The command-line arguments.
/// \return The most recent sessionID in the list. If the list is empty, a new sessionID is created.
//****************************************************************************************************************************************************
    QString mostRecentSessionID(QStringList const& args) {
    QStringList const sessionIDs = parseGoCLIStringArgument(args, {sessionIDFlag});
    if (sessionIDs.isEmpty()) {
        return newSessionID();
    }

    return *std::max_element(sessionIDs.constBegin(), sessionIDs.constEnd(), [](QString const &lhs, QString const &rhs) -> bool {
        return sessionIDToDateTime(lhs) < sessionIDToDateTime(rhs);
    });
}


} // namespace bridgepp
