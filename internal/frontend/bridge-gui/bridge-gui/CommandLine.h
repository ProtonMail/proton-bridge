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


#ifndef BRIDGE_GUI_COMMAND_LINE_H
#define BRIDGE_GUI_COMMAND_LINE_H


#include <bridgepp/Log/Log.h>


//****************************************************************************************************************************************************
/// \brief A struct containing the parsed command line options
//****************************************************************************************************************************************************
struct CommandLineOptions {
    QStringList bridgeArgs; ///< The command-line arguments we will pass to bridge when launching it.
    QStringList bridgeGuiArgs; ///< The command-line arguments we will pass to bridge when launching it.
    QString launcher; ///< The path to the launcher.
    bool attach { false }; ///< Is the application running in attached mode?
    bridgepp::Log::Level logLevel { bridgepp::Log::defaultLevel }; ///< The log level
    bool noWindow { false }; ///< Should the application start without displaying the main window?
    bool useSoftwareRenderer { false }; ///< Should QML be renderer in software (i.e. without rendering hardware interface).
};


CommandLineOptions parseCommandLine(QStringList const &argv); ///< Parse the command-line arguments


#endif //BRIDGE_GUI_COMMAND_LINE_H
