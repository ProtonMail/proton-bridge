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


#ifndef BRIDGEPP_GRPC_ERRORS_H
#define BRIDGEPP_GRPC_ERRORS_H


#include "bridge.grpc.pb.h"


namespace bridgepp {


//****************************************************************************************************************************************************
/// \param[in] A structure holding information about an error.
//****************************************************************************************************************************************************
struct ErrorInfo {
    grpc::ErrorCode code;
    QString title;
    QString description;
};


ErrorInfo errorInfo(grpc::ErrorCode code); ///< Retrieve the potentially localized information about an error.


} // namespace bridgepp


#endif // BRIDGEPP_GRPC_ERRORS_H
