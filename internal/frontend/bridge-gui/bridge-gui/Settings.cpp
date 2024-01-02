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


#include "Settings.h"
#include <bridgepp/BridgeUtils.h>


using namespace bridgepp;


namespace {


QString const settingsFileName = "bridge-gui.ini"; ///< The name of the settings file.
QString const keyUseSoftwareRenderer = "UseSoftwareRenderer"; ///< The key for storing the 'Use software rendering' setting.


}

//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
Settings::Settings()
    : settings_(QDir(userConfigDir()).absoluteFilePath("bridge-gui.ini"), QSettings::Format::IniFormat) {
}


//****************************************************************************************************************************************************
/// \return The value for the 'Use software renderer' setting.
//****************************************************************************************************************************************************
bool Settings::useSoftwareRenderer() const {
    return settings_.value(keyUseSoftwareRenderer, onWindows()).toBool();
}


//****************************************************************************************************************************************************
/// \param[in] value The value for the 'Use software renderer' setting.
//****************************************************************************************************************************************************
void Settings::setUseSoftwareRenderer(bool value) {
    settings_.setValue(keyUseSoftwareRenderer, value);
}
