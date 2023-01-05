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


#include "FocusGRPCClient.h"
#include "../Exception/Exception.h"


using namespace focus;
using namespace grpc;
using namespace google::protobuf;


namespace {


Empty empty; ///< Empty protobuf message, re-used across calls.
qint64 const port = 1042; ///< The port for the focus service.
QString const hostname = "127.0.0.1"; ///< The hostname of the focus service.


}


namespace bridgepp {


//****************************************************************************************************************************************************
/// \param[in] timeoutMs The timeout for the connexion.
/// \param[out] outError if not null and the function returns false.
/// \return true iff the connexion was successfully established.
//****************************************************************************************************************************************************
bool FocusGRPCClient::connectToServer(qint64 timeoutMs, QString *outError) {
    try {
        QString const address = QString("%1:%2").arg(hostname).arg(port);
        channel_ = grpc::CreateChannel(address.toStdString(), grpc::InsecureChannelCredentials());
        if (!channel_) {
            throw Exception("Could not create focus service channel.");
        }
        stub_ = Focus::NewStub(channel_);

        if (!channel_->WaitForConnected(gpr_time_add(gpr_now(GPR_CLOCK_REALTIME), gpr_time_from_millis(timeoutMs, GPR_TIMESPAN)))) {
            throw Exception("Could not connect to focus service");
        }

        if (channel_->GetState(true) != GRPC_CHANNEL_READY) {
            throw Exception("Connexion check with focus service failed.");
        }

        return true;
    }
    catch (Exception const &e) {
        if (outError) {
            *outError = e.qwhat();
        }
        return false;
    }
}


//****************************************************************************************************************************************************
/// \return The status for the call.
//****************************************************************************************************************************************************
grpc::Status FocusGRPCClient::raise() {
    ClientContext ctx;
    return stub_->Raise(&ctx, empty, &empty);
}


//****************************************************************************************************************************************************
/// \param[out] outVersion The version string.
/// \return The status for the call.
//****************************************************************************************************************************************************
grpc::Status FocusGRPCClient::version(QString &outVersion) {
    ClientContext ctx;
    VersionResponse response;
    Status status = stub_->Version(&ctx, empty, &response);
    if (status.ok()) {
        outVersion = QString::fromStdString(response.version());
    }
    return status;
}


}// namespace bridgepp
