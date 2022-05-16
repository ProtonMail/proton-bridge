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


#ifndef BRIDGE_QT6_GRPCUTILS_H
#define BRIDGE_QT6_GRPCUTILS_H

#include "GRPC/bridge.grpc.pb.h"
#include "grpc++/grpc++.h"
#include "User/User.h"


void logGRPCCallStatus(grpc::Status const& status, QString const &callName); ///< Log the status of a gRPC code.
SPUser parsegrpcUser(grpc::User const& grpcUser); ///< Parse a gRPC user struct and return a User.

#endif // BRIDGE_QT6_GRPCUTILS_H
