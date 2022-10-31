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


#ifndef BRIDGE_PP_TESTER_BRIDGE_UTILS_H
#define BRIDGE_PP_TESTER_BRIDGE_UTILS_H


#include <bridgepp/User/User.h>


namespace bridgepp
{


QString userConfigDir(); ///< Get the path of the user configuration folder.
QString userCacheDir(); ///< Get the path of the user cache folder.
QString goos(); ///< return the value of Go's  GOOS for the current platform ("darwin", "linux" and "windows"  are supported).
qint64 randN(qint64 n); ///< return a random integer in the half open range  [0,n)
QString randomFirstName(); ///< Get a random first name from a pre-determined list.
QString randomLastName(); ///< Get a random first name from a pre-determined list.
SPUser randomUser(); ///< Get a random user.


} // namespace


#endif // BRIDGE_PP_TESTER_BRIDGE_UTILS_H
