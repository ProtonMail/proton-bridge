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


#ifndef BRIDGE_PP_SESSION_ID_H
#define BRIDGE_PP_SESSION_ID_H


namespace bridgepp {


extern QString const sessionIDFlag; ///< The sessionID command-line flag (without hyphens)
extern QString const hyphenatedSessionIDFlag; ///< The sessionID command-line flag (with two hyphens)


QString newSessionID(); ///< Create a new sessions
QDateTime sessionIDToDateTime(QString const &sessionID); ///< Parse the date/time from a sessionID.


} // namespace

#endif //BRIDGE_PP_SESSION_ID_H
