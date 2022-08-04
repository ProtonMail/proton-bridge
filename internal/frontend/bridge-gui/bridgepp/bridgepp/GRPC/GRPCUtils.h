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


#ifndef BRIDGE_GUI_GRPC_UTILS_H
#define BRIDGE_GUI_GRPC_UTILS_H


#include "../User/User.h"
#include "../Log/Log.h"
#include "bridge.grpc.pb.h"


namespace bridgepp
{


typedef std::shared_ptr<grpc::StreamEvent> SPStreamEvent; ///< Type definition for shared pointer to grpc::StreamEvent.


QString serverCertificatePath(); ///< Return the path of the server certificate.
QString serverKeyPath(); ///< Return the path of the server key.
grpc::LogLevel logLevelToGRPC(Log::Level level); ///< Convert a Log::Level to gRPC enum value.
Log::Level logLevelFromGRPC(grpc::LogLevel level); ///< Convert a grpc::LogLevel to a Log::Level.
void userToGRPC(User const &user, grpc::User &outGRPCUser); ///< Convert a bridgepp::User to a grpc::User.
SPUser userFromGRPC(grpc::User const &grpcUser); ///< Create a bridgepp::User from a grpc::User.

}


#endif // BRIDGE_GUI_GRPC_UTILS_H
