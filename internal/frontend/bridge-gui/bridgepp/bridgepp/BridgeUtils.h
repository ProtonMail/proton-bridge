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


#ifndef BRIDGE_PP_TESTER_BRIDGE_UTILS_H
#define BRIDGE_PP_TESTER_BRIDGE_UTILS_H


#include "User/User.h"


namespace bridgepp {


//****************************************************************************************************************************************************
/// \brief Enumeration for the operating system.
//****************************************************************************************************************************************************
enum class OS {
    Linux = 0, ///< The Linux OS.
    MacOS = 1, ///< The macOS OS.
    Windows = 2, ///< The Windows OS.
};


QString userConfigDir(); ///< Get the path of the user configuration folder.
QString userCacheDir(); ///< Get the path of the user cache folder.
QString userDataDir(); ///< Get the path of the user data folder.
QString sentryCacheDir(); ///< Get the path of the sentry cache folder.
QString goos(); ///< return the value of Go's  GOOS for the current platform ("darwin", "linux" and "windows"  are supported).
qint64 randN(qint64 n); ///< return a random integer in the half open range  [0,n)
QString randomFirstName(); ///< Get a random first name from a pre-determined list.
QString randomLastName(); ///< Get a random first name from a pre-determined list.
SPUser defaultUser(); ///< Return The default user, with the name and addresses used on screenshots on proton.me
SPUser randomUser(QString const &firstName = "", QString const &lastName = ""); ///< Get a random user.
OS os(); ///< Return the operating system.
bool onLinux(); ///< Check if the OS is Linux.
bool onMacOS(); ///< Check if the OS is macOS.
bool onWindows(); ///< Check if the OS in Windows.
QString elideLongString(QString const &str, qint32 maxLength); ///< Elide a string in the middle if its length exceed maxLength.


} // namespace


#endif // BRIDGE_PP_TESTER_BRIDGE_UTILS_H
