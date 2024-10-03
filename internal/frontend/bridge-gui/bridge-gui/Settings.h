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


#ifndef BRIDGE_GUI_SETTINGS_H
#define BRIDGE_GUI_SETTINGS_H


//****************************************************************************************************************************************************
/// \brief Application settings class
//****************************************************************************************************************************************************
class Settings {
public: // member functions.
    Settings(Settings const&) = delete; ///< Disabled copy-constructor.
    Settings(Settings&&) = delete; ///< Disabled assignment copy-constructor.
    ~Settings() = default; ///< Destructor.
    Settings& operator=(Settings const&) = delete; ///< Disabled assignment operator.
    Settings& operator=(Settings&&) = delete; ///< Disabled move assignment operator.

    bool useSoftwareRenderer() const; ///< Get the 'Use software renderer' settings value.
    void setUseSoftwareRenderer(bool value); ///< Set the 'Use software renderer' settings value.
    void setTrayIconVisible(bool value);  ///< Get the 'Tray icon visible' setting value.
    bool trayIconVisible() const; ///< Set the 'Tray icon visible' setting value.

private: // member functions.
    Settings(); ///< Default constructor.

private: // data members.
    QSettings settings_; ///< The settings.

    friend class AppController;
};


#endif //BRIDGE_GUI_SETTINGS_H
