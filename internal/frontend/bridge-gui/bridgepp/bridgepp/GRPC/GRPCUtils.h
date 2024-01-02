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


#ifndef BRIDGE_PP_GRPC_UTILS_H
#define BRIDGE_PP_GRPC_UTILS_H


#include "../User/User.h"
#include "../Log/Log.h"
#include "bridge.grpc.pb.h"


namespace bridgepp {


extern std::string const grpcMetadataServerTokenKey; ///< The key for the server token stored in the gRPC calls context metadata.


typedef std::shared_ptr<grpc::StreamEvent> SPStreamEvent; ///< Type definition for shared pointer to grpc::StreamEvent.


QString grpcServerConfigPath(QString const &configDir); ///< Return the path of the gRPC server config file.
QString grpcClientConfigBasePath(QString const &configDir); ///< Return the path of the gRPC client config file.
QString createClientConfigFile(QString const &configDir, QString const &token, QString *outError); ///< Create the client config file the server will retrieve and return its path.
grpc::LogLevel logLevelToGRPC(Log::Level level); ///< Convert a Log::Level to gRPC enum value.
Log::Level logLevelFromGRPC(grpc::LogLevel level); ///< Convert a grpc::LogLevel to a Log::Level.
grpc::UserState userStateToGRPC(UserState state); ///< Convert a bridgepp::UserState to a grpc::UserState.
UserState userStateFromGRPC(grpc::UserState state);  ///< Convert a grpc::UserState to a bridgepp::UserState.
void userToGRPC(User const &user, grpc::User &outGRPCUser); ///< Convert a bridgepp::User to a grpc::User.
SPUser userFromGRPC(grpc::User const &grpcUser); ///< Create a bridgepp::User from a grpc::User.
bool useFileSocketForGRPC(); ///< Check whether the Bridge gRPC service should use file sockets instead of TCP sockets.
QString getAvailableFileSocketPath(); ///< Return the path of a new available socket, or a null string if none could be found.


}


#endif // BRIDGE_PP_GRPC_UTILS_H
