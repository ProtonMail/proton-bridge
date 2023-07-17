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

#include "CLIUtils.h"


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


} // namespace bridgepp
