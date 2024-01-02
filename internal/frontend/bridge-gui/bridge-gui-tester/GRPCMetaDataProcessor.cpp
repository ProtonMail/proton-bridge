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


#include "GRPCMetaDataProcessor.h"
#include <bridgepp/GRPC/GRPCUtils.h>
#include <bridgepp/Exception/Exception.h>


using namespace bridgepp;
using namespace grpc;


//****************************************************************************************************************************************************
/// \param[in] serverToken The server token expected from gRPC calls
//****************************************************************************************************************************************************
GRPCMetadataProcessor::GRPCMetadataProcessor(QString const &serverToken)
    : serverToken_(serverToken.toStdString()) {

}


//****************************************************************************************************************************************************
/// \return false.
//****************************************************************************************************************************************************
bool GRPCMetadataProcessor::IsBlocking() const {
    return false;
}


//****************************************************************************************************************************************************
/// \param authMetadata The authentication metadata.
/// \return the result of the metadata processing.
//****************************************************************************************************************************************************
Status GRPCMetadataProcessor::Process(InputMetadata const &authMetadata, AuthContext *,
    OutputMetadata *, OutputMetadata *) {
    try {
        const InputMetadata::const_iterator pathIt = authMetadata.find(":path");
        QString const callName = (pathIt == authMetadata.end()) ? ("unkown gRPC call") : QString::fromLocal8Bit(pathIt->second);

        AuthMetadataProcessor::InputMetadata::size_type const count = authMetadata.count(grpcMetadataServerTokenKey);
        if (count == 0) {
            throw Exception(QString("Missing server token in gRPC client call '%1'.").arg(callName));
        }

        if (count > 1) {
            throw Exception(QString("Several server tokens were provided in gRPC client call '%1'.").arg(callName));
        }

        if (authMetadata.find(grpcMetadataServerTokenKey)->second != serverToken_) {
            throw Exception(QString("Invalid server token provided by gRPC client call '%1'.").arg(callName));
        }

        app().log().trace(QString("Server token for gRPC call '%1' was validated.").arg(callName));
        return Status::OK;
    }
    catch (Exception const &e) {
        app().log().error(e.qwhat());
        return Status(UNAUTHENTICATED, e.qwhat().toStdString());
    }
}




