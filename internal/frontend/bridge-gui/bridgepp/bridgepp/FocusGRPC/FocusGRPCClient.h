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


#ifndef BRIDGEPP_FOCUS_GRPC_CLIENT_H
#define BRIDGEPP_FOCUS_GRPC_CLIENT_H


#include "grpc++/grpc++.h"
#include "focus.grpc.pb.h"
#include "../Log/Log.h"


namespace bridgepp {


//****************************************************************************************************************************************************
/// \brief Focus GRPC client class
//****************************************************************************************************************************************************
class FocusGRPCClient {
public: // static member functions
    static void removeServiceConfigFile(QString const &configDir); ///< Delete the service config file.
    static QString grpcFocusServerConfigPath(QString const &configDir); ///< Return the path of the gRPC Focus server config file.

public: // member functions.
    FocusGRPCClient(Log& log); ///< Default constructor.
    FocusGRPCClient(FocusGRPCClient const &) = delete; ///< Disabled copy-constructor.
    FocusGRPCClient(FocusGRPCClient &&) = delete; ///< Disabled assignment copy-constructor.
    ~FocusGRPCClient() = default; ///< Destructor.
    FocusGRPCClient &operator=(FocusGRPCClient const &) = delete; ///< Disabled assignment operator.
    FocusGRPCClient &operator=(FocusGRPCClient &&) = delete; ///< Disabled move assignment operator.

    bool connectToServer(qint64 timeoutMs, quint16 port, QString *outError = nullptr); ///< Connect to the focus server
    grpc::Status raise(QString const &reason); ///< Performs the 'raise' call.
    grpc::Status version(QString &outVersion); ///< Performs the 'version' call.

private:
    Log &log_; ///< The log to use for logging calls
    std::shared_ptr<grpc::Channel> channel_ { nullptr }; ///< The gRPC channel.
    std::shared_ptr<focus::Focus::Stub> stub_ { nullptr }; ///< The gRPC stub (a.k.a. client).
};


}


#endif //BRIDGEPP_FOCUS_GRPC_CLIENT_H
