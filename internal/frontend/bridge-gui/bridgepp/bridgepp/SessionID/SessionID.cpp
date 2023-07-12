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


#include "SessionID.h"
#include "QtCore/qdatetime.h"


namespace {


QString const dateTimeFormat = "yyyyMMdd_hhmmsszzz"; ///< The format string for date/time used by the sessionID.


}


namespace bridgepp {


//****************************************************************************************************************************************************
/// \return a new session ID based on the current local date/time
//****************************************************************************************************************************************************
QString newSessionID() {
    return QDateTime::currentDateTime().toString(dateTimeFormat);
}


//****************************************************************************************************************************************************
/// \param[in] sessionID The sessionID.
/// \return The date/time corresponding to the sessionID.
/// \return An invalid date/time if an error occurs.
//****************************************************************************************************************************************************
QDateTime sessionIDToDateTime(QString const &sessionID) {
    return QDateTime::fromString(sessionID, dateTimeFormat);
}


} // namespace
